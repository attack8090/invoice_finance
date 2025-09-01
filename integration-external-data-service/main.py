"""
Integration External Data Service
Handles ERP systems, third-party data for credit scoring and KYC/AML checks
Supports EDI standards with automated syncing and multi-platform integration
Port: 8088
"""

from fastapi import FastAPI, HTTPException, BackgroundTasks, Depends, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.middleware.trustedhost import TrustedHostMiddleware
from pydantic import BaseModel, Field
from typing import Optional, List, Dict, Any, Union
import asyncio
import httpx
import json
import logging
from datetime import datetime, timedelta
import os
import redis
from pymongo import MongoClient
import xml.etree.ElementTree as ET
from xml.dom import minidom
import csv
import io

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


# MongoDB setup
MONGO_URL = os.getenv("MONGO_URL", "mongodb://localhost:27017/")
mongo_client = MongoClient(MONGO_URL)
mongo_db = mongo_client[os.getenv("MONGO_DB", "integration_external_data")]

# Redis setup
redis_client = redis.Redis(host=os.getenv("REDIS_HOST", "localhost"), port=6379, decode_responses=True)

# FastAPI app initialization
app = FastAPI(
    title="Integration External Data Service",
    description="ERP and third-party data integration service for credit scoring, KYC/AML, and EDI standards",
    version="1.0.0",
    docs_url="/docs",
    redoc_url="/redoc"
)

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Security middleware
app.add_middleware(TrustedHostMiddleware, allowed_hosts=["*"])

# Database Models
class ERPConnection(Base):
    __tablename__ = "erp_connections"
    
    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    name = Column(String(255), nullable=False)
    erp_type = Column(String(100), nullable=False)  # sap, oracle, quickbooks, xero, etc
    connection_config = Column(JSON, nullable=False)
    status = Column(String(50), default="active")
    last_sync = Column(DateTime, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

class DataSync(Base):
    __tablename__ = "data_syncs"
    
    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    connection_id = Column(UUID(as_uuid=True), nullable=False)
    sync_type = Column(String(100), nullable=False)  # financial_data, customer_data, transaction_data
    status = Column(String(50), default="pending")  # pending, running, completed, failed
    records_processed = Column(Integer, default=0)
    records_total = Column(Integer, default=0)
    error_message = Column(Text, nullable=True)
    sync_data = Column(JSON, nullable=True)
    started_at = Column(DateTime, nullable=True)
    completed_at = Column(DateTime, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)

## Removed SQLAlchemy models for CreditScoreData and KYCAMLData. These will be stored in MongoDB collections.

class EDIDocument(Base):
    __tablename__ = "edi_documents"
    
    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    document_type = Column(String(100), nullable=False)  # 810_invoice, 850_po, 856_asn, 820_payment
    sender_id = Column(String(100), nullable=False)
    receiver_id = Column(String(100), nullable=False)
    transaction_set = Column(String(10), nullable=False)
    control_number = Column(String(50), nullable=False)
    edi_content = Column(Text, nullable=False)
    parsed_data = Column(JSON, nullable=True)
    status = Column(String(50), default="received")  # received, parsed, processed, error
    error_message = Column(Text, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    processed_at = Column(DateTime, nullable=True)

# Create tables
Base.metadata.create_all(bind=engine)

# Dependency to get DB session
def get_db():
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()

# Pydantic Models
class ERPConnectionCreate(BaseModel):
    name: str
    erp_type: str
    connection_config: Dict[str, Any]

class ERPConnectionResponse(BaseModel):
    id: str
    name: str
    erp_type: str
    status: str
    last_sync: Optional[datetime]
    created_at: datetime

class DataSyncRequest(BaseModel):
    connection_id: str
    sync_type: str
    filters: Optional[Dict[str, Any]] = {}

class DataSyncResponse(BaseModel):
    id: str
    connection_id: str
    sync_type: str
    status: str
    records_processed: int
    records_total: int
    started_at: Optional[datetime]
    completed_at: Optional[datetime]

class CreditScoreRequest(BaseModel):
    customer_id: str
    provider: str = "experian"
    include_factors: bool = True

class CreditScoreResponse(BaseModel):
    id: str
    customer_id: str
    provider: str
    score: Optional[int]
    rating: Optional[str]
    risk_factors: Optional[List[str]]
    expires_at: datetime

class KYCAMLRequest(BaseModel):
    customer_id: str
    check_type: str  # kyc, aml, sanctions, pep
    customer_data: Dict[str, Any]
    provider: str = "jumio"

class KYCAMLResponse(BaseModel):
    id: str
    customer_id: str
    check_type: str
    provider: str
    status: str
    confidence_score: Optional[float]
    flagged_items: Optional[List[str]]
    expires_at: datetime

class EDIDocumentRequest(BaseModel):
    document_type: str
    sender_id: str
    receiver_id: str
    edi_content: str

class EDIDocumentResponse(BaseModel):
    id: str
    document_type: str
    transaction_set: str
    control_number: str
    status: str
    parsed_data: Optional[Dict[str, Any]]
    created_at: datetime

# Health Check
@app.get("/health")
async def health_check():
    return {
        "status": "healthy",
        "service": "integration-external-data-service",
        "version": "1.0.0",
        "timestamp": datetime.utcnow(),
        "features": {
            "erp_integration": True,
            "credit_scoring": True,
            "kyc_aml_checks": True,
            "edi_processing": True,
            "multi_platform_sync": True,
            "automated_syncing": True
        },
        "supported_systems": {
            "erp": ["SAP", "Oracle", "QuickBooks", "Xero", "NetSuite", "Dynamics"],
            "credit_bureaus": ["Experian", "Equifax", "TransUnion", "Dun & Bradstreet"],
            "kyc_aml": ["Jumio", "Onfido", "Refinitiv", "LexisNexis"],
            "edi_standards": ["X12", "EDIFACT", "TRADACOMS"]
        }
    }

# ERP Integration Endpoints
@app.post("/api/v1/erp/connections", response_model=ERPConnectionResponse)
async def create_erp_connection(connection: ERPConnectionCreate, db: Session = Depends(get_db)):
    """Create new ERP connection"""
    db_connection = ERPConnection(
        name=connection.name,
        erp_type=connection.erp_type,
        connection_config=connection.connection_config
    )
    db.add(db_connection)
    db.commit()
    db.refresh(db_connection)
    
    return ERPConnectionResponse(
        id=str(db_connection.id),
        name=db_connection.name,
        erp_type=db_connection.erp_type,
        status=db_connection.status,
        last_sync=db_connection.last_sync,
        created_at=db_connection.created_at
    )

@app.get("/api/v1/erp/connections", response_model=List[ERPConnectionResponse])
async def get_erp_connections(db: Session = Depends(get_db)):
    """Get all ERP connections"""
    connections = db.query(ERPConnection).all()
    return [
        ERPConnectionResponse(
            id=str(conn.id),
            name=conn.name,
            erp_type=conn.erp_type,
            status=conn.status,
            last_sync=conn.last_sync,
            created_at=conn.created_at
        ) for conn in connections
    ]

@app.post("/api/v1/erp/connections/{connection_id}/test")
async def test_erp_connection(connection_id: str, db: Session = Depends(get_db)):
    """Test ERP connection"""
    connection = db.query(ERPConnection).filter(ERPConnection.id == connection_id).first()
    if not connection:
        raise HTTPException(status_code=404, detail="Connection not found")
    
    # Simulate connection test
    test_result = {
        "connection_id": connection_id,
        "status": "success",
        "response_time": "150ms",
        "last_tested": datetime.utcnow(),
        "capabilities": ["read_financial_data", "read_customer_data", "read_transactions"]
    }
    
    return test_result

@app.post("/api/v1/erp/sync", response_model=DataSyncResponse)
async def start_data_sync(sync_request: DataSyncRequest, background_tasks: BackgroundTasks, db: Session = Depends(get_db)):
    """Start data synchronization from ERP"""
    # Create sync record
    sync_record = DataSync(
        connection_id=sync_request.connection_id,
        sync_type=sync_request.sync_type,
        status="pending"
    )
    db.add(sync_record)
    db.commit()
    db.refresh(sync_record)
    
    # Start background sync
    background_tasks.add_task(perform_erp_sync, str(sync_record.id), sync_request)
    
    return DataSyncResponse(
        id=str(sync_record.id),
        connection_id=sync_request.connection_id,
        sync_type=sync_request.sync_type,
        status=sync_record.status,
        records_processed=0,
        records_total=0,
        started_at=sync_record.started_at,
        completed_at=sync_record.completed_at
    )

@app.get("/api/v1/erp/sync/{sync_id}/status", response_model=DataSyncResponse)
async def get_sync_status(sync_id: str, db: Session = Depends(get_db)):
    """Get synchronization status"""
    sync_record = db.query(DataSync).filter(DataSync.id == sync_id).first()
    if not sync_record:
        raise HTTPException(status_code=404, detail="Sync record not found")
    
    return DataSyncResponse(
        id=str(sync_record.id),
        connection_id=str(sync_record.connection_id),
        sync_type=sync_record.sync_type,
        status=sync_record.status,
        records_processed=sync_record.records_processed,
        records_total=sync_record.records_total,
        started_at=sync_record.started_at,
        completed_at=sync_record.completed_at
    )

# Credit Scoring Endpoints
@app.post("/api/v1/credit-scoring/check", response_model=CreditScoreResponse)
async def get_credit_score(request: CreditScoreRequest, background_tasks: BackgroundTasks, db: Session = Depends(get_db)):
    """Get credit score from third-party provider"""
    # Check cache first
    cache_key = f"credit_score:{request.customer_id}:{request.provider}"
    cached_score = redis_client.get(cache_key)
    
    if cached_score:
        return json.loads(cached_score)
    
    # Create credit score record
    credit_record = CreditScoreData(
        customer_id=request.customer_id,
        provider=request.provider,
        report_data={"status": "pending"},
        expires_at=datetime.utcnow() + timedelta(days=30)
    )
    db.add(credit_record)
    db.commit()
    db.refresh(credit_record)
    
    # Start background credit check
    background_tasks.add_task(perform_credit_check, str(credit_record.id), request)
    
    # Return pending response
    response = CreditScoreResponse(
        id=str(credit_record.id),
        customer_id=request.customer_id,
        provider=request.provider,
        score=None,
        rating=None,
        risk_factors=None,
        expires_at=credit_record.expires_at
    )
    
    # Cache for 5 minutes
    redis_client.setex(cache_key, 300, json.dumps(response.dict(), default=str))
    
    return response

@app.get("/api/v1/credit-scoring/{customer_id}/history")
async def get_credit_history(customer_id: str):
    """Get credit score history for customer"""
    records = list(mongo_db.credit_score_data.find({"customer_id": customer_id}).sort("created_at", -1))
    return {
        "customer_id": customer_id,
        "records": [
            {
                "id": str(record.get("_id")),
                "provider": record.get("provider"),
                "score": record.get("score"),
                "rating": record.get("rating"),
                "created_at": record.get("created_at"),
                "expires_at": record.get("expires_at")
            } for record in records
        ]
    }

# KYC/AML Endpoints
@app.post("/api/v1/kyc-aml/check", response_model=KYCAMLResponse)
async def perform_kyc_aml_check(request: KYCAMLRequest, background_tasks: BackgroundTasks):
    """Perform KYC/AML check"""
    kyc_record = {
        "customer_id": request.customer_id,
        "check_type": request.check_type,
        "provider": request.provider,
        "status": "pending",
        "check_data": request.customer_data,
        "expires_at": datetime.utcnow() + timedelta(days=90),
        "created_at": datetime.utcnow(),
        "confidence_score": None,
        "flagged_items": None
    }
    result = mongo_db.kyc_aml_data.insert_one(kyc_record)
    kyc_record["_id"] = result.inserted_id
    background_tasks.add_task(perform_kyc_aml_verification, str(result.inserted_id), request)
    return KYCAMLResponse(
        id=str(result.inserted_id),
        customer_id=request.customer_id,
        check_type=request.check_type,
        provider=request.provider,
        status=kyc_record["status"],
        confidence_score=kyc_record["confidence_score"],
        flagged_items=kyc_record["flagged_items"],
        expires_at=kyc_record["expires_at"]
    )

@app.get("/api/v1/kyc-aml/{customer_id}/status")
async def get_kyc_aml_status(customer_id: str):
    """Get KYC/AML status for customer"""
    records = list(mongo_db.kyc_aml_data.find({"customer_id": customer_id}).sort("created_at", -1))
    return {
        "customer_id": customer_id,
        "overall_status": "compliant",
        "checks": [
            {
                "id": str(record.get("_id")),
                "check_type": record.get("check_type"),
                "provider": record.get("provider"),
                "status": record.get("status"),
                "confidence_score": record.get("confidence_score"),
                "created_at": record.get("created_at"),
                "expires_at": record.get("expires_at")
            } for record in records
        ]
    }

# EDI Processing Endpoints
@app.post("/api/v1/edi/process", response_model=EDIDocumentResponse)
async def process_edi_document(request: EDIDocumentRequest, background_tasks: BackgroundTasks, db: Session = Depends(get_db)):
    """Process EDI document"""
    # Parse basic EDI structure
    transaction_set, control_number = parse_edi_header(request.edi_content)
    
    # Create EDI document record
    edi_doc = EDIDocument(
        document_type=request.document_type,
        sender_id=request.sender_id,
        receiver_id=request.receiver_id,
        transaction_set=transaction_set,
        control_number=control_number,
        edi_content=request.edi_content,
        status="received"
    )
    db.add(edi_doc)
    db.commit()
    db.refresh(edi_doc)
    
    # Start background processing
    background_tasks.add_task(process_edi_background, str(edi_doc.id))
    
    return EDIDocumentResponse(
        id=str(edi_doc.id),
        document_type=edi_doc.document_type,
        transaction_set=edi_doc.transaction_set,
        control_number=edi_doc.control_number,
        status=edi_doc.status,
        parsed_data=edi_doc.parsed_data,
        created_at=edi_doc.created_at
    )

@app.get("/api/v1/edi/documents")
async def get_edi_documents(
    document_type: Optional[str] = None,
    status: Optional[str] = None,
    db: Session = Depends(get_db)
):
    """Get EDI documents with optional filters"""
    query = db.query(EDIDocument)
    
    if document_type:
        query = query.filter(EDIDocument.document_type == document_type)
    if status:
        query = query.filter(EDIDocument.status == status)
    
    documents = query.order_by(EDIDocument.created_at.desc()).limit(100).all()
    
    return {
        "documents": [
            {
                "id": str(doc.id),
                "document_type": doc.document_type,
                "sender_id": doc.sender_id,
                "receiver_id": doc.receiver_id,
                "transaction_set": doc.transaction_set,
                "control_number": doc.control_number,
                "status": doc.status,
                "created_at": doc.created_at
            } for doc in documents
        ]
    }

@app.get("/api/v1/edi/documents/{doc_id}/parsed")
async def get_parsed_edi_document(doc_id: str, db: Session = Depends(get_db)):
    """Get parsed EDI document data"""
    document = db.query(EDIDocument).filter(EDIDocument.id == doc_id).first()
    if not document:
        raise HTTPException(status_code=404, detail="Document not found")
    
    return {
        "id": str(document.id),
        "document_type": document.document_type,
        "status": document.status,
        "parsed_data": document.parsed_data,
        "raw_content": document.edi_content[:500] + "..." if len(document.edi_content) > 500 else document.edi_content
    }

# Multi-platform Integration Endpoints
@app.get("/api/v1/integrations/platforms")
async def get_supported_platforms():
    """Get list of supported integration platforms"""
    return {
        "erp_systems": [
            {"name": "SAP", "version": "S/4HANA", "api_type": "REST/OData"},
            {"name": "Oracle EBS", "version": "12.2+", "api_type": "REST/SOAP"},
            {"name": "QuickBooks Online", "version": "v3", "api_type": "REST"},
            {"name": "Xero", "version": "v2", "api_type": "REST"},
            {"name": "NetSuite", "version": "2021.2+", "api_type": "REST/SOAP"},
            {"name": "Dynamics 365", "version": "v9+", "api_type": "Web API"}
        ],
        "credit_bureaus": [
            {"name": "Experian", "api_version": "v1", "regions": ["US", "UK", "AU"]},
            {"name": "Equifax", "api_version": "v2", "regions": ["US", "CA", "UK"]},
            {"name": "TransUnion", "api_version": "v1", "regions": ["US", "CA", "IN"]},
            {"name": "Dun & Bradstreet", "api_version": "v5", "regions": ["Global"]}
        ],
        "kyc_aml_providers": [
            {"name": "Jumio", "api_version": "v1", "services": ["ID verification", "AML screening"]},
            {"name": "Onfido", "api_version": "v3.6", "services": ["ID verification", "Facial recognition"]},
            {"name": "Refinitiv", "api_version": "v1", "services": ["AML screening", "PEP checks"]},
            {"name": "LexisNexis", "api_version": "v3", "services": ["Identity verification", "Risk assessment"]}
        ]
    }

@app.get("/api/v1/integrations/sync-status")
async def get_integration_sync_status(db: Session = Depends(get_db)):
    """Get overall sync status across all integrations"""
    recent_syncs = db.query(DataSync).filter(
        DataSync.created_at >= datetime.utcnow() - timedelta(hours=24)
    ).all()
    
    status_summary = {
        "total_syncs_24h": len(recent_syncs),
        "successful_syncs": len([s for s in recent_syncs if s.status == "completed"]),
        "failed_syncs": len([s for s in recent_syncs if s.status == "failed"]),
        "running_syncs": len([s for s in recent_syncs if s.status == "running"]),
        "by_type": {}
    }
    
    for sync in recent_syncs:
        if sync.sync_type not in status_summary["by_type"]:
            status_summary["by_type"][sync.sync_type] = {"total": 0, "successful": 0, "failed": 0}
        status_summary["by_type"][sync.sync_type]["total"] += 1
        if sync.status == "completed":
            status_summary["by_type"][sync.sync_type]["successful"] += 1
        elif sync.status == "failed":
            status_summary["by_type"][sync.sync_type]["failed"] += 1
    
    return status_summary

# Background Tasks
async def perform_erp_sync(sync_id: str, sync_request: DataSyncRequest):
    """Background task to perform ERP synchronization"""
    db = SessionLocal()
    try:
        sync_record = db.query(DataSync).filter(DataSync.id == sync_id).first()
        if not sync_record:
            return
        
        sync_record.status = "running"
        sync_record.started_at = datetime.utcnow()
        db.commit()
        
        # Simulate ERP sync process
        await asyncio.sleep(2)  # Simulate API calls
        
        # Mock sync results
        sync_record.records_total = 1000
        sync_record.records_processed = 1000
        sync_record.status = "completed"
        sync_record.completed_at = datetime.utcnow()
        sync_record.sync_data = {
            "customers": 150,
            "invoices": 500,
            "transactions": 350,
            "last_sync": datetime.utcnow().isoformat()
        }
        
        db.commit()
        logger.info(f"ERP sync {sync_id} completed successfully")
        
    except Exception as e:
        sync_record.status = "failed"
        sync_record.error_message = str(e)
        sync_record.completed_at = datetime.utcnow()
        db.commit()
        logger.error(f"ERP sync {sync_id} failed: {e}")
    finally:
        db.close()

async def perform_credit_check(record_id: str, request: CreditScoreRequest):
    """Background task to perform credit check"""
    db = SessionLocal()
    try:
        record = db.query(CreditScoreData).filter(CreditScoreData.id == record_id).first()
        if not record:
            return
        
        # Simulate credit bureau API call
        await asyncio.sleep(3)
        
        # Mock credit score result
        mock_score = 720
        mock_rating = "Good"
        mock_factors = ["Credit utilization: 25%", "Payment history: Excellent", "Credit age: 8 years"]
        
        record.score = mock_score
        record.rating = mock_rating
        record.risk_factors = mock_factors
        record.report_data = {
            "score": mock_score,
            "rating": mock_rating,
            "factors": mock_factors,
            "provider": request.provider,
            "generated_at": datetime.utcnow().isoformat()
        }
        
        db.commit()
        
        # Update cache
        cache_key = f"credit_score:{request.customer_id}:{request.provider}"
        response = CreditScoreResponse(
            id=str(record.id),
            customer_id=request.customer_id,
            provider=request.provider,
            score=mock_score,
            rating=mock_rating,
            risk_factors=mock_factors,
            expires_at=record.expires_at
        )
        redis_client.setex(cache_key, 86400, json.dumps(response.dict(), default=str))  # 24 hours
        
        logger.info(f"Credit check {record_id} completed: Score {mock_score}")
        
    except Exception as e:
        record.report_data = {"error": str(e), "status": "failed"}
        db.commit()
        logger.error(f"Credit check {record_id} failed: {e}")
    finally:
        db.close()

async def perform_kyc_aml_verification(record_id: str, request: KYCAMLRequest):
    """Background task to perform KYC/AML verification"""
    db = SessionLocal()
    try:
        record = db.query(KYCAMLData).filter(KYCAMLData.id == record_id).first()
        if not record:
            return
        
        # Simulate KYC/AML provider API call
        await asyncio.sleep(2)
        
        # Mock verification result
        record.status = "passed"
        record.confidence_score = 0.95
        record.flagged_items = []
        record.check_data.update({
            "verification_result": "passed",
            "confidence_score": 0.95,
            "verified_at": datetime.utcnow().isoformat()
        })
        
        db.commit()
        logger.info(f"KYC/AML check {record_id} completed: {record.status}")
        
    except Exception as e:
        record.status = "failed"
        record.check_data.update({"error": str(e)})
        db.commit()
        logger.error(f"KYC/AML check {record_id} failed: {e}")
    finally:
        db.close()

async def process_edi_background(doc_id: str):
    """Background task to process EDI document"""
    db = SessionLocal()
    try:
        document = db.query(EDIDocument).filter(EDIDocument.id == doc_id).first()
        if not document:
            return
        
        document.status = "parsing"
        db.commit()
        
        # Simulate EDI parsing
        await asyncio.sleep(1)
        
        # Mock parsed data based on document type
        if document.document_type == "810_invoice":
            parsed_data = {
                "invoice_number": "INV-2024-001",
                "invoice_date": "2024-01-15",
                "due_date": "2024-02-15",
                "total_amount": 15000.00,
                "currency": "USD",
                "line_items": [
                    {"description": "Product A", "quantity": 100, "unit_price": 50.00, "amount": 5000.00},
                    {"description": "Product B", "quantity": 200, "unit_price": 50.00, "amount": 10000.00}
                ]
            }
        else:
            parsed_data = {"document_type": document.document_type, "parsed": True}
        
        document.parsed_data = parsed_data
        document.status = "processed"
        document.processed_at = datetime.utcnow()
        
        db.commit()
        logger.info(f"EDI document {doc_id} processed successfully")
        
    except Exception as e:
        document.status = "error"
        document.error_message = str(e)
        db.commit()
        logger.error(f"EDI processing {doc_id} failed: {e}")
    finally:
        db.close()

# Helper Functions
def parse_edi_header(edi_content: str):
    """Parse EDI header to extract transaction set and control number"""
    lines = edi_content.split('\n')
    transaction_set = "999"  # Default
    control_number = f"CTL{datetime.utcnow().strftime('%Y%m%d%H%M%S')}"
    
    for line in lines:
        if line.startswith("ST*"):
            parts = line.split('*')
            if len(parts) > 1:
                transaction_set = parts[1]
        elif line.startswith("GS*"):
            parts = line.split('*')
            if len(parts) > 6:
                control_number = parts[6]
    
    return transaction_set, control_number

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8088)
