"""
Document Management Service
Handles secure document storage, OCR processing, and metadata management
Supports Epic 2 compliance with AWS S3 integration and encryption
"""

import os
import asyncio
import logging
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Any
from pathlib import Path

import uvicorn
from fastapi import FastAPI, HTTPException, UploadFile, File, Depends, status, Form, BackgroundTasks
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse, StreamingResponse
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from pydantic import BaseModel, validator
from sqlalchemy.ext.asyncio import AsyncSession

# Internal imports
from config import settings
from database import get_db, init_db
from models.document import Document, DocumentType, DocumentStatus
from services.storage_service import StorageService
from services.encryption_service import EncryptionService
from services.ocr_service import OCRService
from services.virus_scanner import VirusScanner
from services.auth_service import AuthService
from middleware.security import SecurityMiddleware
from middleware.rate_limit import RateLimitMiddleware
from utils.validators import validate_file_type, validate_file_size
from utils.logger import setup_logger

# Setup logging
logger = setup_logger(__name__)

# Initialize FastAPI app
app = FastAPI(
    title="Document Management Service",
    version="1.0.0",
    description="Secure document management with OCR processing and AWS S3 storage",
    docs_url="/docs" if settings.debug else None,
    redoc_url="/redoc" if settings.debug else None
)

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.allowed_origins,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Security middleware
app.add_middleware(SecurityMiddleware)
app.add_middleware(RateLimitMiddleware)

# Security
security = HTTPBearer()

# Services
storage_service = StorageService()
encryption_service = EncryptionService()
ocr_service = OCRService()
virus_scanner = VirusScanner()
auth_service = AuthService()

# Request/Response models
class DocumentUploadRequest(BaseModel):
    document_type: DocumentType
    description: Optional[str] = None
    tags: Optional[List[str]] = None
    encrypt: bool = True
    ocr_enabled: bool = True
    
    @validator('document_type')
    def validate_document_type(cls, v):
        if v not in DocumentType:
            raise ValueError(f'Invalid document type: {v}')
        return v

class DocumentMetadata(BaseModel):
    id: str
    filename: str
    original_filename: str
    document_type: DocumentType
    file_size: int
    mime_type: str
    upload_date: datetime
    status: DocumentStatus
    encrypted: bool
    ocr_completed: bool
    description: Optional[str] = None
    tags: List[str] = []
    checksum: str
    
class DocumentSearchRequest(BaseModel):
    document_type: Optional[DocumentType] = None
    status: Optional[DocumentStatus] = None
    tags: Optional[List[str]] = None
    date_from: Optional[datetime] = None
    date_to: Optional[datetime] = None
    search_text: Optional[str] = None
    
class OCRResult(BaseModel):
    document_id: str
    extracted_text: str
    structured_data: Dict[str, Any]
    confidence_score: float
    fields_detected: List[str]
    processing_time: float

class BulkUploadRequest(BaseModel):
    document_type: DocumentType
    encrypt: bool = True
    ocr_enabled: bool = True
    tags: Optional[List[str]] = None

@app.on_event("startup")
async def startup_event():
    """Initialize services on startup"""
    await init_db()
    await storage_service.initialize()
    await ocr_service.initialize()
    logger.info("Document Management Service started successfully")

@app.on_event("shutdown")
async def shutdown_event():
    """Cleanup on shutdown"""
    await storage_service.cleanup()
    await ocr_service.cleanup()
    logger.info("Document Management Service shut down")

async def get_current_user(credentials: HTTPAuthorizationCredentials = Depends(security)):
    """Validate JWT token and get current user"""
    try:
        user = await auth_service.verify_token(credentials.credentials)
        return user
    except Exception as e:
        logger.error(f"Authentication failed: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid authentication credentials",
            headers={"WWW-Authenticate": "Bearer"},
        )

@app.get("/")
async def root():
    """Health check endpoint"""
    return {
        "service": "Document Management Service",
        "version": "1.0.0",
        "status": "healthy",
        "features": {
            "storage": "AWS S3",
            "encryption": "AES-256",
            "ocr": "Gemini AI",
            "virus_scan": settings.virus_scan_enabled,
            "compliance": "Epic-2"
        }
    }

@app.post("/api/v1/documents/upload", response_model=DocumentMetadata)
async def upload_document(
    background_tasks: BackgroundTasks,
    file: UploadFile = File(...),
    document_type: DocumentType = Form(...),
    description: Optional[str] = Form(None),
    tags: Optional[str] = Form(None),  # JSON string of tags
    encrypt: bool = Form(True),
    ocr_enabled: bool = Form(True),
    current_user = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """Upload a single document with optional OCR processing"""
    try:
        # Validate file
        validate_file_type(file.filename, file.content_type)
        validate_file_size(file.size)
        
        # Parse tags if provided
        tag_list = []
        if tags:
            import json
            tag_list = json.loads(tags)
        
        # Read file content
        content = await file.read()
        
        # Virus scan if enabled
        if settings.virus_scan_enabled:
            is_safe = await virus_scanner.scan_content(content)
            if not is_safe:
                raise HTTPException(
                    status_code=status.HTTP_400_BAD_REQUEST,
                    detail="File failed virus scan"
                )
        
        # Generate unique filename
        file_extension = Path(file.filename).suffix
        unique_filename = f"{datetime.utcnow().strftime('%Y%m%d_%H%M%S')}_{current_user['id']}{file_extension}"
        
        # Encrypt content if requested
        encrypted_content = content
        encryption_key = None
        if encrypt:
            encrypted_content, encryption_key = await encryption_service.encrypt(content)
        
        # Upload to S3
        s3_url = await storage_service.upload_file(
            file_content=encrypted_content,
            filename=unique_filename,
            content_type=file.content_type,
            user_id=current_user['id']
        )
        
        # Calculate checksum
        import hashlib
        checksum = hashlib.sha256(content).hexdigest()
        
        # Create document record
        document = Document(
            filename=unique_filename,
            original_filename=file.filename,
            document_type=document_type,
            file_size=len(content),
            mime_type=file.content_type,
            s3_url=s3_url,
            status=DocumentStatus.UPLOADED,
            encrypted=encrypt,
            encryption_key=encryption_key,
            checksum=checksum,
            user_id=current_user['id'],
            description=description,
            tags=tag_list
        )
        
        db.add(document)
        await db.commit()
        await db.refresh(document)
        
        # Schedule background tasks
        if ocr_enabled:
            background_tasks.add_task(process_ocr, document.id, content)
        
        # Log successful upload
        logger.info(f"Document uploaded: {document.id} by user {current_user['id']}")
        
        return DocumentMetadata(
            id=str(document.id),
            filename=document.filename,
            original_filename=document.original_filename,
            document_type=document.document_type,
            file_size=document.file_size,
            mime_type=document.mime_type,
            upload_date=document.created_at,
            status=document.status,
            encrypted=document.encrypted,
            ocr_completed=document.ocr_completed,
            description=document.description,
            tags=document.tags or [],
            checksum=document.checksum
        )
        
    except Exception as e:
        logger.error(f"Document upload failed: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Upload failed: {str(e)}"
        )

@app.post("/api/v1/documents/bulk-upload")
async def bulk_upload_documents(
    background_tasks: BackgroundTasks,
    files: List[UploadFile] = File(...),
    request: BulkUploadRequest = Depends(),
    current_user = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """Upload multiple documents in bulk"""
    try:
        if len(files) > settings.max_bulk_upload_count:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail=f"Too many files. Maximum {settings.max_bulk_upload_count} allowed"
            )
        
        uploaded_documents = []
        failed_uploads = []
        
        for file in files:
            try:
                # Validate each file
                validate_file_type(file.filename, file.content_type)
                validate_file_size(file.size)
                
                # Process upload (similar to single upload)
                content = await file.read()
                
                # Virus scan
                if settings.virus_scan_enabled:
                    is_safe = await virus_scanner.scan_content(content)
                    if not is_safe:
                        failed_uploads.append({
                            "filename": file.filename,
                            "error": "Failed virus scan"
                        })
                        continue
                
                # Generate filename and upload
                file_extension = Path(file.filename).suffix
                unique_filename = f"{datetime.utcnow().strftime('%Y%m%d_%H%M%S')}_{current_user['id']}_{len(uploaded_documents)}{file_extension}"
                
                encrypted_content = content
                encryption_key = None
                if request.encrypt:
                    encrypted_content, encryption_key = await encryption_service.encrypt(content)
                
                s3_url = await storage_service.upload_file(
                    file_content=encrypted_content,
                    filename=unique_filename,
                    content_type=file.content_type,
                    user_id=current_user['id']
                )
                
                checksum = hashlib.sha256(content).hexdigest()
                
                document = Document(
                    filename=unique_filename,
                    original_filename=file.filename,
                    document_type=request.document_type,
                    file_size=len(content),
                    mime_type=file.content_type,
                    s3_url=s3_url,
                    status=DocumentStatus.UPLOADED,
                    encrypted=request.encrypt,
                    encryption_key=encryption_key,
                    checksum=checksum,
                    user_id=current_user['id'],
                    tags=request.tags or []
                )
                
                db.add(document)
                await db.commit()
                await db.refresh(document)
                
                uploaded_documents.append(str(document.id))
                
                # Schedule OCR processing
                if request.ocr_enabled:
                    background_tasks.add_task(process_ocr, document.id, content)
                    
            except Exception as e:
                failed_uploads.append({
                    "filename": file.filename,
                    "error": str(e)
                })
        
        logger.info(f"Bulk upload completed: {len(uploaded_documents)} successful, {len(failed_uploads)} failed")
        
        return {
            "uploaded_documents": uploaded_documents,
            "failed_uploads": failed_uploads,
            "total_processed": len(files),
            "success_count": len(uploaded_documents),
            "failure_count": len(failed_uploads)
        }
        
    except Exception as e:
        logger.error(f"Bulk upload failed: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Bulk upload failed: {str(e)}"
        )

@app.get("/api/v1/documents/{document_id}", response_model=DocumentMetadata)
async def get_document_metadata(
    document_id: str,
    current_user = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """Get document metadata by ID"""
    try:
        document = await db.get(Document, document_id)
        if not document:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail="Document not found"
            )
        
        # Check access permissions
        if document.user_id != current_user['id'] and current_user['role'] not in ['admin', 'bank']:
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail="Access denied"
            )
        
        return DocumentMetadata(
            id=str(document.id),
            filename=document.filename,
            original_filename=document.original_filename,
            document_type=document.document_type,
            file_size=document.file_size,
            mime_type=document.mime_type,
            upload_date=document.created_at,
            status=document.status,
            encrypted=document.encrypted,
            ocr_completed=document.ocr_completed,
            description=document.description,
            tags=document.tags or [],
            checksum=document.checksum
        )
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Failed to get document metadata: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to retrieve document metadata"
        )

@app.get("/api/v1/documents/{document_id}/download")
async def download_document(
    document_id: str,
    current_user = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """Download document content"""
    try:
        document = await db.get(Document, document_id)
        if not document:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail="Document not found"
            )
        
        # Check access permissions
        if document.user_id != current_user['id'] and current_user['role'] not in ['admin', 'bank']:
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail="Access denied"
            )
        
        # Get file content from S3
        file_content = await storage_service.download_file(document.s3_url)
        
        # Decrypt if encrypted
        if document.encrypted and document.encryption_key:
            file_content = await encryption_service.decrypt(file_content, document.encryption_key)
        
        # Log download
        logger.info(f"Document downloaded: {document_id} by user {current_user['id']}")
        
        return StreamingResponse(
            io.BytesIO(file_content),
            media_type=document.mime_type,
            headers={"Content-Disposition": f"attachment; filename={document.original_filename}"}
        )
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Document download failed: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Download failed"
        )

@app.post("/api/v1/documents/search")
async def search_documents(
    request: DocumentSearchRequest,
    current_user = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """Search documents with filters"""
    try:
        from sqlalchemy import and_, or_
        from sqlalchemy.orm import selectinload
        
        query = db.query(Document)
        
        # Base filter - user can only see own documents unless admin/bank
        if current_user['role'] not in ['admin', 'bank']:
            query = query.filter(Document.user_id == current_user['id'])
        
        # Apply filters
        filters = []
        
        if request.document_type:
            filters.append(Document.document_type == request.document_type)
        
        if request.status:
            filters.append(Document.status == request.status)
            
        if request.date_from:
            filters.append(Document.created_at >= request.date_from)
            
        if request.date_to:
            filters.append(Document.created_at <= request.date_to)
            
        if request.tags:
            for tag in request.tags:
                filters.append(Document.tags.contains([tag]))
        
        if request.search_text:
            filters.append(
                or_(
                    Document.original_filename.ilike(f"%{request.search_text}%"),
                    Document.description.ilike(f"%{request.search_text}%"),
                    Document.ocr_text.ilike(f"%{request.search_text}%")
                )
            )
        
        if filters:
            query = query.filter(and_(*filters))
        
        documents = await query.all()
        
        # Convert to response format
        results = []
        for doc in documents:
            results.append(DocumentMetadata(
                id=str(doc.id),
                filename=doc.filename,
                original_filename=doc.original_filename,
                document_type=doc.document_type,
                file_size=doc.file_size,
                mime_type=doc.mime_type,
                upload_date=doc.created_at,
                status=doc.status,
                encrypted=doc.encrypted,
                ocr_completed=doc.ocr_completed,
                description=doc.description,
                tags=doc.tags or [],
                checksum=doc.checksum
            ))
        
        return {
            "documents": results,
            "total_count": len(results),
            "search_criteria": request.dict()
        }
        
    except Exception as e:
        logger.error(f"Document search failed: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Search failed"
        )

@app.get("/api/v1/documents/{document_id}/ocr", response_model=OCRResult)
async def get_ocr_result(
    document_id: str,
    current_user = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """Get OCR results for a document"""
    try:
        document = await db.get(Document, document_id)
        if not document:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail="Document not found"
            )
        
        # Check access permissions
        if document.user_id != current_user['id'] and current_user['role'] not in ['admin', 'bank']:
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail="Access denied"
            )
        
        if not document.ocr_completed:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="OCR processing not completed"
            )
        
        return OCRResult(
            document_id=str(document.id),
            extracted_text=document.ocr_text or "",
            structured_data=document.ocr_data or {},
            confidence_score=document.ocr_confidence or 0.0,
            fields_detected=list(document.ocr_data.keys()) if document.ocr_data else [],
            processing_time=document.ocr_processing_time or 0.0
        )
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Failed to get OCR result: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to retrieve OCR result"
        )

@app.delete("/api/v1/documents/{document_id}")
async def delete_document(
    document_id: str,
    current_user = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """Delete a document (soft delete)"""
    try:
        document = await db.get(Document, document_id)
        if not document:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail="Document not found"
            )
        
        # Check permissions
        if document.user_id != current_user['id'] and current_user['role'] != 'admin':
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail="Access denied"
            )
        
        # Soft delete
        document.status = DocumentStatus.DELETED
        document.deleted_at = datetime.utcnow()
        
        await db.commit()
        
        logger.info(f"Document deleted: {document_id} by user {current_user['id']}")
        
        return {"message": "Document deleted successfully"}
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Document deletion failed: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Deletion failed"
        )

async def process_ocr(document_id: str, file_content: bytes):
    """Background task to process OCR"""
    try:
        async with get_db() as db:
            document = await db.get(Document, document_id)
            if not document:
                return
            
            # Update status
            document.status = DocumentStatus.PROCESSING
            await db.commit()
            
            # Perform OCR
            start_time = datetime.utcnow()
            ocr_result = await ocr_service.process_document(file_content, document.document_type)
            processing_time = (datetime.utcnow() - start_time).total_seconds()
            
            # Update document with OCR results
            document.ocr_text = ocr_result.get('text', '')
            document.ocr_data = ocr_result.get('structured_data', {})
            document.ocr_confidence = ocr_result.get('confidence', 0.0)
            document.ocr_processing_time = processing_time
            document.ocr_completed = True
            document.status = DocumentStatus.PROCESSED
            document.processed_at = datetime.utcnow()
            
            await db.commit()
            
            logger.info(f"OCR processing completed for document: {document_id}")
            
    except Exception as e:
        logger.error(f"OCR processing failed for document {document_id}: {str(e)}")
        # Update document status to failed
        try:
            async with get_db() as db:
                document = await db.get(Document, document_id)
                if document:
                    document.status = DocumentStatus.FAILED
                    await db.commit()
        except:
            pass

if __name__ == "__main__":
    uvicorn.run(
        "main:app",
        host=settings.host,
        port=settings.port,
        reload=settings.debug,
        workers=1 if settings.debug else settings.workers
    )
