"""
Payment Service
Handles fund disbursement to SMEs, invoice settlement matching, partial/early payments,
and fee calculations with Epic 5 compliance and bank API integration
"""

import os
import asyncio
import logging
from datetime import datetime, timedelta
from decimal import Decimal, ROUND_HALF_UP
from typing import Dict, List, Optional, Any, Union
from enum import Enum
import json

import uvicorn
from fastapi import FastAPI, HTTPException, Depends, status, BackgroundTasks
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from pydantic import BaseModel, validator, Field
from sqlalchemy.ext.asyncio import AsyncSession

# Internal imports
from config import settings
from database import get_db, init_db
from models.payment import Payment, PaymentStatus, PaymentType, Settlement
from models.disbursement import Disbursement, DisbursementStatus
from models.fee import Fee, FeeType, FeeCalculation
from services.auth_service import AuthService
from services.bank_integration_service import BankIntegrationService
from services.settlement_service import SettlementService
from services.fee_calculation_service import FeeCalculationService
from services.compliance_service import ComplianceService
from services.fraud_detection_service import FraudDetectionService
from services.notification_service import NotificationService
from middleware.security import SecurityMiddleware
from middleware.rate_limit import RateLimitMiddleware
from utils.currency_converter import CurrencyConverter
from utils.payment_validator import PaymentValidator
from utils.logger import setup_logger

# Setup logging
logger = setup_logger(__name__)

# Initialize FastAPI app
app = FastAPI(
    title="Payment Service",
    version="1.0.0",
    description="Epic 5 compliant payment service with bank integration and real-time settlement",
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
auth_service = AuthService()
bank_integration_service = BankIntegrationService()
settlement_service = SettlementService()
fee_calculation_service = FeeCalculationService()
compliance_service = ComplianceService()
fraud_detection_service = FraudDetectionService()
notification_service = NotificationService()

# Utilities
currency_converter = CurrencyConverter()
payment_validator = PaymentValidator()

class PaymentMethod(str, Enum):
    WIRE_TRANSFER = "wire_transfer"
    ACH = "ach"
    SWIFT = "swift"
    DOMESTIC_TRANSFER = "domestic_transfer"
    REAL_TIME_PAYMENT = "real_time_payment"
    CRYPTOCURRENCY = "cryptocurrency"

class SettlementType(str, Enum):
    FULL = "full"
    PARTIAL = "partial"
    EARLY = "early"
    OVERDUE = "overdue"

# Request/Response models
class DisbursementRequest(BaseModel):
    sme_id: str
    financing_request_id: str
    amount: Decimal = Field(..., gt=0, decimal_places=2)
    currency: str = Field(default="USD", min_length=3, max_length=3)
    payment_method: PaymentMethod
    bank_account_id: str
    reference: Optional[str] = None
    urgency: str = Field(default="normal", regex="^(low|normal|high|urgent)$")
    compliance_check: bool = True
    
    @validator('amount')
    def validate_amount(cls, v):
        if v <= 0:
            raise ValueError('Amount must be greater than zero')
        return v.quantize(Decimal('0.01'), rounding=ROUND_HALF_UP)

class SettlementRequest(BaseModel):
    invoice_id: str
    buyer_id: str
    amount: Decimal = Field(..., gt=0, decimal_places=2)
    currency: str = Field(default="USD", min_length=3, max_length=3)
    settlement_type: SettlementType
    payment_method: PaymentMethod
    payment_reference: Optional[str] = None
    expected_date: Optional[datetime] = None
    
class FeeCalculationRequest(BaseModel):
    transaction_type: str  # disbursement, settlement, transfer
    amount: Decimal = Field(..., gt=0, decimal_places=2)
    currency: str = Field(default="USD", min_length=3, max_length=3)
    payment_method: PaymentMethod
    user_tier: Optional[str] = "standard"
    urgency: Optional[str] = "normal"

class PaymentStatusUpdate(BaseModel):
    payment_id: str
    status: PaymentStatus
    bank_reference: Optional[str] = None
    failure_reason: Optional[str] = None
    processing_fee: Optional[Decimal] = None
    completed_at: Optional[datetime] = None

class ReconciliationRequest(BaseModel):
    date_from: datetime
    date_to: datetime
    currency: Optional[str] = None
    payment_method: Optional[PaymentMethod] = None
    include_fees: bool = True

@app.on_event("startup")
async def startup_event():
    """Initialize services on startup"""
    await init_db()
    await bank_integration_service.initialize()
    await settlement_service.initialize()
    await fraud_detection_service.initialize()
    
    logger.info("Payment Service started successfully")

@app.on_event("shutdown")
async def shutdown_event():
    """Cleanup on shutdown"""
    await bank_integration_service.cleanup()
    await settlement_service.cleanup()
    
    logger.info("Payment Service shut down")

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
    bank_status = await bank_integration_service.check_connectivity()
    
    return {
        "service": "Payment Service",
        "version": "1.0.0",
        "status": "healthy",
        "features": {
            "epic_5_compliance": True,
            "bank_integration": True,
            "real_time_payments": True,
            "multi_currency": True,
            "settlement_matching": True,
            "fee_calculation": True,
            "fraud_detection": True,
            "reconciliation": True
        },
        "bank_connectivity": bank_status,
        "supported_currencies": await get_supported_currencies(),
        "payment_methods": [method.value for method in PaymentMethod]
    }

@app.post("/api/v1/disbursements/create")
async def create_disbursement(
    background_tasks: BackgroundTasks,
    request: DisbursementRequest,
    current_user = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """Create and process fund disbursement to SMEs"""
    try:
        # Validate user permissions
        if not await validate_disbursement_permission(current_user, request.sme_id):
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail="Access denied to create disbursement"
            )
        
        # Validate payment details
        validation_result = await payment_validator.validate_disbursement(request)
        if not validation_result.valid:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail=f"Payment validation failed: {validation_result.errors}"
            )
        
        # Compliance check if required
        if request.compliance_check:
            compliance_result = await compliance_service.check_epic5_compliance(
                request.sme_id, request.amount, request.currency
            )
            if not compliance_result.approved:
                raise HTTPException(
                    status_code=status.HTTP_400_BAD_REQUEST,
                    detail=f"Compliance check failed: {compliance_result.reason}"
                )
        
        # Fraud detection screening
        fraud_check = await fraud_detection_service.screen_payment(
            user_id=request.sme_id,
            amount=float(request.amount),
            currency=request.currency,
            payment_method=request.payment_method.value,
            bank_account_id=request.bank_account_id
        )
        
        if fraud_check.risk_level == "high":
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="Payment flagged for fraud review"
            )
        
        # Calculate fees
        fee_calculation = await fee_calculation_service.calculate_disbursement_fee(
            amount=request.amount,
            currency=request.currency,
            payment_method=request.payment_method,
            urgency=request.urgency,
            user_tier=current_user.get('tier', 'standard')
        )
        
        # Create disbursement record
        disbursement = Disbursement(
            sme_id=request.sme_id,
            financing_request_id=request.financing_request_id,
            amount=request.amount,
            currency=request.currency,
            payment_method=request.payment_method.value,
            bank_account_id=request.bank_account_id,
            reference=request.reference,
            urgency=request.urgency,
            processing_fee=fee_calculation.total_fee,
            net_amount=request.amount - fee_calculation.total_fee,
            status=DisbursementStatus.PENDING,
            fraud_score=fraud_check.fraud_score,
            compliance_approved=request.compliance_check,
            created_by=current_user['id']
        )
        
        db.add(disbursement)
        await db.commit()
        await db.refresh(disbursement)
        
        # Schedule background processing
        background_tasks.add_task(
            process_disbursement, disbursement.id, db
        )
        
        # Send notification
        await notification_service.send_disbursement_notification(
            disbursement, "created"
        )
        
        logger.info(f"Disbursement created: {disbursement.id} for SME {request.sme_id}")
        
        return {
            "disbursement_id": str(disbursement.id),
            "status": disbursement.status.value,
            "amount": str(disbursement.amount),
            "net_amount": str(disbursement.net_amount),
            "processing_fee": str(disbursement.processing_fee),
            "fee_breakdown": fee_calculation.breakdown,
            "estimated_completion": calculate_estimated_completion(
                request.payment_method, request.urgency
            ),
            "reference": disbursement.reference,
            "fraud_score": fraud_check.fraud_score
        }
        
    except Exception as e:
        logger.error(f"Disbursement creation failed: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Disbursement creation failed: {str(e)}"
        )

@app.post("/api/v1/settlements/process")
async def process_settlement(
    background_tasks: BackgroundTasks,
    request: SettlementRequest,
    current_user = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """Process invoice settlement from buyers to lenders"""
    try:
        # Validate settlement permission
        if not await validate_settlement_permission(current_user, request.buyer_id, request.invoice_id):
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail="Access denied to process settlement"
            )
        
        # Get invoice and financing details
        invoice_details = await get_invoice_financing_details(request.invoice_id, db)
        if not invoice_details:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail="Invoice or financing details not found"
            )
        
        # Validate settlement amount
        settlement_validation = await settlement_service.validate_settlement(
            request, invoice_details
        )
        if not settlement_validation.valid:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail=f"Settlement validation failed: {settlement_validation.errors}"
            )
        
        # Calculate settlement fees and allocations
        settlement_calculation = await fee_calculation_service.calculate_settlement_fees(
            request.amount,
            request.currency,
            request.settlement_type,
            invoice_details
        )
        
        # Create settlement record
        settlement = Settlement(
            invoice_id=request.invoice_id,
            buyer_id=request.buyer_id,
            amount=request.amount,
            currency=request.currency,
            settlement_type=request.settlement_type.value,
            payment_method=request.payment_method.value,
            payment_reference=request.payment_reference,
            processing_fee=settlement_calculation.processing_fee,
            platform_fee=settlement_calculation.platform_fee,
            investor_amount=settlement_calculation.investor_amount,
            sme_amount=settlement_calculation.sme_amount,
            status=PaymentStatus.PENDING,
            expected_date=request.expected_date,
            created_by=current_user['id']
        )
        
        db.add(settlement)
        await db.commit()
        await db.refresh(settlement)
        
        # Schedule background processing
        background_tasks.add_task(
            process_settlement_payments, settlement.id, db
        )
        
        logger.info(f"Settlement processed: {settlement.id} for invoice {request.invoice_id}")
        
        return {
            "settlement_id": str(settlement.id),
            "status": settlement.status.value,
            "amount": str(settlement.amount),
            "fee_breakdown": settlement_calculation.breakdown,
            "allocations": {
                "investor_amount": str(settlement_calculation.investor_amount),
                "sme_amount": str(settlement_calculation.sme_amount),
                "platform_fee": str(settlement_calculation.platform_fee),
                "processing_fee": str(settlement_calculation.processing_fee)
            },
            "estimated_completion": calculate_settlement_completion_time(request.payment_method)
        }
        
    except Exception as e:
        logger.error(f"Settlement processing failed: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Settlement processing failed: {str(e)}"
        )

@app.get("/api/v1/payments/{payment_id}/status")
async def get_payment_status(
    payment_id: str,
    current_user = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """Get real-time payment status"""
    try:
        payment = await get_payment_by_id(payment_id, db)
        if not payment:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail="Payment not found"
            )
        
        # Validate access permissions
        if not await validate_payment_access(current_user, payment):
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail="Access denied to payment information"
            )
        
        # Get latest status from bank if in progress
        if payment.status in [PaymentStatus.PENDING, PaymentStatus.PROCESSING]:
            bank_status = await bank_integration_service.get_payment_status(
                payment.bank_transaction_id
            )
            if bank_status and bank_status.status != payment.status:
                payment.status = bank_status.status
                payment.bank_reference = bank_status.reference
                payment.updated_at = datetime.utcnow()
                await db.commit()
        
        return {
            "payment_id": str(payment.id),
            "type": payment.payment_type.value,
            "status": payment.status.value,
            "amount": str(payment.amount),
            "currency": payment.currency,
            "payment_method": payment.payment_method,
            "bank_reference": payment.bank_reference,
            "created_at": payment.created_at.isoformat(),
            "updated_at": payment.updated_at.isoformat(),
            "completed_at": payment.completed_at.isoformat() if payment.completed_at else None,
            "failure_reason": payment.failure_reason,
            "estimated_completion": payment.estimated_completion.isoformat() if payment.estimated_completion else None
        }
        
    except Exception as e:
        logger.error(f"Failed to get payment status: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to retrieve payment status"
        )

@app.post("/api/v1/fees/calculate")
async def calculate_fees(
    request: FeeCalculationRequest,
    current_user = Depends(get_current_user)
):
    """Calculate fees for various transaction types"""
    try:
        if request.transaction_type == "disbursement":
            fee_result = await fee_calculation_service.calculate_disbursement_fee(
                amount=request.amount,
                currency=request.currency,
                payment_method=request.payment_method,
                urgency=request.urgency,
                user_tier=request.user_tier
            )
        elif request.transaction_type == "settlement":
            fee_result = await fee_calculation_service.calculate_settlement_fee(
                amount=request.amount,
                currency=request.currency,
                payment_method=request.payment_method
            )
        elif request.transaction_type == "transfer":
            fee_result = await fee_calculation_service.calculate_transfer_fee(
                amount=request.amount,
                currency=request.currency,
                payment_method=request.payment_method,
                urgency=request.urgency
            )
        else:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="Invalid transaction type"
            )
        
        return {
            "transaction_type": request.transaction_type,
            "amount": str(request.amount),
            "currency": request.currency,
            "payment_method": request.payment_method.value,
            "total_fee": str(fee_result.total_fee),
            "net_amount": str(request.amount - fee_result.total_fee),
            "fee_breakdown": fee_result.breakdown,
            "user_tier": request.user_tier,
            "calculation_time": datetime.utcnow().isoformat()
        }
        
    except Exception as e:
        logger.error(f"Fee calculation failed: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Fee calculation failed: {str(e)}"
        )

@app.post("/api/v1/reconciliation/run")
async def run_reconciliation(
    background_tasks: BackgroundTasks,
    request: ReconciliationRequest,
    current_user = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """Run payment reconciliation for specified period"""
    try:
        # Validate user has reconciliation permissions
        if current_user['role'] not in ['admin', 'finance', 'bank']:
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail="Insufficient permissions for reconciliation"
            )
        
        # Schedule background reconciliation
        reconciliation_job = await create_reconciliation_job(
            request, current_user['id'], db
        )
        
        background_tasks.add_task(
            process_reconciliation, reconciliation_job.id, request, db
        )
        
        logger.info(f"Reconciliation job started: {reconciliation_job.id}")
        
        return {
            "job_id": str(reconciliation_job.id),
            "status": "started",
            "period": {
                "from": request.date_from.isoformat(),
                "to": request.date_to.isoformat()
            },
            "filters": {
                "currency": request.currency,
                "payment_method": request.payment_method.value if request.payment_method else None,
                "include_fees": request.include_fees
            },
            "estimated_completion": datetime.utcnow() + timedelta(minutes=30)
        }
        
    except Exception as e:
        logger.error(f"Reconciliation failed: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Reconciliation failed: {str(e)}"
        )

@app.post("/api/v1/webhooks/bank-notification")
async def bank_webhook_notification(
    background_tasks: BackgroundTasks,
    notification: Dict[str, Any],
    db: AsyncSession = Depends(get_db)
):
    """Handle bank webhook notifications"""
    try:
        # Validate webhook signature
        if not await bank_integration_service.validate_webhook_signature(notification):
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="Invalid webhook signature"
            )
        
        # Process notification
        background_tasks.add_task(
            process_bank_notification, notification, db
        )
        
        return {"status": "accepted"}
        
    except Exception as e:
        logger.error(f"Webhook processing failed: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Webhook processing failed"
        )

@app.get("/api/v1/analytics/payment-trends")
async def get_payment_trends(
    days: int = 30,
    currency: Optional[str] = None,
    current_user = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """Get payment analytics and trends"""
    try:
        if current_user['role'] not in ['admin', 'finance', 'analyst']:
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail="Insufficient permissions for analytics"
            )
        
        trends = await get_payment_analytics(days, currency, db)
        
        return {
            "period_days": days,
            "currency_filter": currency,
            "trends": trends,
            "generated_at": datetime.utcnow().isoformat()
        }
        
    except Exception as e:
        logger.error(f"Payment trends analysis failed: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Payment trends analysis failed"
        )

# Background task functions
async def process_disbursement(disbursement_id: str, db: AsyncSession):
    """Background task to process disbursement"""
    try:
        disbursement = await db.get(Disbursement, disbursement_id)
        if not disbursement:
            return
        
        # Update status to processing
        disbursement.status = DisbursementStatus.PROCESSING
        await db.commit()
        
        # Submit to bank
        bank_result = await bank_integration_service.submit_disbursement(disbursement)
        
        # Update with bank reference
        disbursement.bank_transaction_id = bank_result.transaction_id
        disbursement.bank_reference = bank_result.reference
        disbursement.estimated_completion = bank_result.estimated_completion
        
        if bank_result.success:
            disbursement.status = DisbursementStatus.SUBMITTED
        else:
            disbursement.status = DisbursementStatus.FAILED
            disbursement.failure_reason = bank_result.error
        
        await db.commit()
        
        # Send notification
        await notification_service.send_disbursement_notification(
            disbursement, "status_update"
        )
        
        logger.info(f"Disbursement processed: {disbursement_id} - {disbursement.status}")
        
    except Exception as e:
        logger.error(f"Disbursement processing failed: {disbursement_id} - {str(e)}")

# Helper functions
async def validate_disbursement_permission(user: dict, sme_id: str) -> bool:
    """Validate if user can create disbursement for SME"""
    if user['role'] in ['admin', 'bank']:
        return True
    return user['id'] == sme_id or user.get('company_id') == sme_id

async def get_supported_currencies() -> List[str]:
    """Get list of supported currencies"""
    return ["USD", "EUR", "GBP", "CAD", "AUD", "JPY", "CHF", "SEK", "NOK", "DKK"]

def calculate_estimated_completion(payment_method: PaymentMethod, urgency: str) -> datetime:
    """Calculate estimated completion time"""
    base_minutes = {
        PaymentMethod.REAL_TIME_PAYMENT: 5,
        PaymentMethod.ACH: 1440,  # 24 hours
        PaymentMethod.WIRE_TRANSFER: 480,  # 8 hours
        PaymentMethod.SWIFT: 2880,  # 48 hours
        PaymentMethod.DOMESTIC_TRANSFER: 720,  # 12 hours
        PaymentMethod.CRYPTOCURRENCY: 60  # 1 hour
    }.get(payment_method, 1440)
    
    # Adjust for urgency
    if urgency == "urgent":
        base_minutes = int(base_minutes * 0.5)
    elif urgency == "high":
        base_minutes = int(base_minutes * 0.7)
    elif urgency == "low":
        base_minutes = int(base_minutes * 1.5)
    
    return datetime.utcnow() + timedelta(minutes=base_minutes)

if __name__ == "__main__":
    uvicorn.run(
        "main:app",
        host=settings.host,
        port=settings.port,
        reload=settings.debug,
        workers=1 if settings.debug else settings.workers
    )
