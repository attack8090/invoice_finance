"""
Enhanced Credit Scoring Service
Supports IDD Core 5.1 integration, Epic 5 compliance, alternative data sources,
risk mitigation, and fraud prevention for SMEs and buyers
"""

import os
import asyncio
import logging
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Any, Union
import json
from decimal import Decimal

import uvicorn
from fastapi import FastAPI, HTTPException, Depends, status, BackgroundTasks
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from pydantic import BaseModel, validator, Field
import numpy as np
import pandas as pd
from sqlalchemy.ext.asyncio import AsyncSession

# ML and AI imports
from sklearn.ensemble import RandomForestClassifier, GradientBoostingClassifier
from sklearn.linear_model import LogisticRegression
from sklearn.preprocessing import StandardScaler
from sklearn.model_selection import cross_val_score
import xgboost as xgb
import lightgbm as lgb
import joblib

# Internal imports
from config import settings
from database import get_db, init_db
from models.credit_score import CreditScore, CreditScoreHistory, RiskAssessment
from models.alternative_data import AlternativeDataSource, DataPoint
from services.auth_service import AuthService
from services.idd_core_service import IDDCoreService
from services.external_data_service import ExternalDataService
from services.fraud_detection_service import FraudDetectionService
from services.risk_mitigation_service import RiskMitigationService
from services.compliance_service import ComplianceService
from middleware.security import SecurityMiddleware
from middleware.rate_limit import RateLimitMiddleware
from utils.model_manager import ModelManager
from utils.feature_engineering import FeatureEngineering
from utils.logger import setup_logger

# Setup logging
logger = setup_logger(__name__)

# Initialize FastAPI app
app = FastAPI(
    title="Credit Scoring Service",
    version="2.0.0",
    description="Advanced credit scoring with IDD Core 5.1 and Epic 5 compliance",
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
idd_core_service = IDDCoreService()
external_data_service = ExternalDataService()
fraud_detection_service = FraudDetectionService()
risk_mitigation_service = RiskMitigationService()
compliance_service = ComplianceService()

# Utilities
model_manager = ModelManager()
feature_engineering = FeatureEngineering()

# Credit scoring models
SCORING_MODELS = {
    "traditional": {
        "name": "Traditional Financial Model",
        "description": "Based on financial statements, payment history, and traditional metrics",
        "features": [
            "annual_revenue", "profit_margin", "current_ratio", "debt_to_equity",
            "payment_history_score", "years_in_business", "industry_risk_score"
        ]
    },
    "alternative": {
        "name": "Alternative Data Model", 
        "description": "Uses social media, web presence, transaction patterns, and behavioral data",
        "features": [
            "social_media_score", "web_presence_score", "transaction_velocity",
            "supplier_diversity", "customer_concentration", "digital_footprint"
        ]
    },
    "hybrid": {
        "name": "Hybrid AI Model",
        "description": "Combines traditional and alternative data with advanced ML",
        "features": [
            "financial_health_score", "behavioral_score", "market_position_score",
            "operational_efficiency", "growth_trajectory", "risk_indicators"
        ]
    },
    "idd_core": {
        "name": "IDD Core 5.1 Integrated Model",
        "description": "Integrates with IDD Core 5.1 for enhanced risk assessment",
        "features": [
            "idd_credit_score", "idd_risk_rating", "idd_payment_behavior",
            "idd_financial_stress_score", "idd_market_comparison"
        ]
    }
}

# Request/Response models
class CreditScoreRequest(BaseModel):
    user_id: str
    company_id: Optional[str] = None
    scoring_model: str = "hybrid"
    include_alternative_data: bool = True
    include_idd_core: bool = True
    risk_tolerance: str = Field(default="medium", regex="^(low|medium|high)$")
    purpose: str = Field(..., description="Scoring purpose: lending, insurance, partnership, etc.")
    
    @validator('scoring_model')
    def validate_scoring_model(cls, v):
        if v not in SCORING_MODELS:
            raise ValueError(f'Invalid scoring model: {v}')
        return v

class CreditScoreResult(BaseModel):
    user_id: str
    company_id: Optional[str]
    credit_score: int = Field(..., ge=300, le=850)
    score_band: str  # excellent, good, fair, poor
    confidence_level: float = Field(..., ge=0.0, le=1.0)
    risk_rating: str  # low, medium, high, very_high
    scoring_model: str
    factors: Dict[str, Any]
    recommendations: List[str]
    next_review_date: datetime
    compliance_flags: List[str]
    alternative_data_used: bool
    idd_core_integrated: bool
    processing_time: float
    created_at: datetime

class RiskAssessmentRequest(BaseModel):
    user_id: str
    company_id: Optional[str] = None
    assessment_type: str = Field(default="comprehensive", regex="^(basic|comprehensive|enhanced)$")
    loan_amount: Optional[Decimal] = None
    loan_purpose: Optional[str] = None
    collateral_value: Optional[Decimal] = None
    include_fraud_check: bool = True
    include_market_analysis: bool = True

class RiskAssessmentResult(BaseModel):
    user_id: str
    company_id: Optional[str]
    overall_risk_score: float = Field(..., ge=0.0, le=1.0)
    risk_category: str
    probability_of_default: float
    loss_given_default: float
    exposure_at_default: Optional[float] = None
    expected_loss: float
    risk_factors: Dict[str, Any]
    mitigation_strategies: List[str]
    monitoring_requirements: List[str]
    approval_recommendation: str
    conditions: List[str]
    valid_until: datetime
    compliance_status: str
    created_at: datetime

class AlternativeDataRequest(BaseModel):
    user_id: str
    company_id: Optional[str] = None
    data_sources: List[str] = ["social_media", "web_presence", "transaction_data", "behavioral_data"]
    consent_provided: bool = True
    data_retention_period: int = 365  # days

class FraudDetectionRequest(BaseModel):
    user_id: str
    company_id: Optional[str] = None
    transaction_data: Optional[Dict[str, Any]] = None
    behavioral_data: Optional[Dict[str, Any]] = None
    document_data: Optional[Dict[str, Any]] = None
    real_time_check: bool = True

class BulkScoringRequest(BaseModel):
    user_ids: List[str]
    scoring_model: str = "hybrid"
    priority: str = Field(default="normal", regex="^(low|normal|high|urgent)$")
    callback_url: Optional[str] = None

@app.on_event("startup")
async def startup_event():
    """Initialize services on startup"""
    await init_db()
    await model_manager.initialize()
    await idd_core_service.initialize()
    await external_data_service.initialize()
    await fraud_detection_service.initialize()
    
    logger.info("Credit Scoring Service started successfully")

@app.on_event("shutdown") 
async def shutdown_event():
    """Cleanup on shutdown"""
    await model_manager.cleanup()
    await idd_core_service.cleanup()
    await external_data_service.cleanup()
    
    logger.info("Credit Scoring Service shut down")

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
        "service": "Credit Scoring Service",
        "version": "2.0.0",
        "status": "healthy",
        "features": {
            "scoring_models": list(SCORING_MODELS.keys()),
            "idd_core_integration": "5.1",
            "alternative_data": True,
            "fraud_detection": True,
            "risk_mitigation": True,
            "compliance": "Epic-5",
            "ml_models": ["XGBoost", "LightGBM", "Random Forest", "Logistic Regression"]
        }
    }

@app.post("/api/v1/credit-score/calculate", response_model=CreditScoreResult)
async def calculate_credit_score(
    background_tasks: BackgroundTasks,
    request: CreditScoreRequest,
    current_user = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """Calculate comprehensive credit score using advanced ML models"""
    try:
        start_time = datetime.utcnow()
        
        # Validate access permissions
        if not await validate_user_access(request.user_id, current_user):
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail="Access denied to user's credit information"
            )
        
        # Collect traditional financial data
        financial_data = await collect_financial_data(request.user_id, request.company_id, db)
        
        # Collect alternative data if requested
        alternative_data = {}
        if request.include_alternative_data:
            alternative_data = await external_data_service.collect_alternative_data(
                request.user_id, request.company_id
            )
        
        # Integrate with IDD Core 5.1 if requested
        idd_data = {}
        if request.include_idd_core:
            idd_data = await idd_core_service.get_enhanced_credit_data(
                request.user_id, request.company_id
            )
        
        # Combine all data sources
        combined_data = {
            **financial_data,
            **alternative_data,
            **idd_data
        }
        
        # Feature engineering
        features = await feature_engineering.create_features(
            combined_data, request.scoring_model
        )
        
        # Load and apply scoring model
        model = await model_manager.load_model(request.scoring_model)
        
        # Calculate credit score
        raw_score = model.predict_proba([features])[0]
        credit_score = int(300 + (raw_score * 550))  # Scale to 300-850 range
        
        # Determine score band
        score_band = get_score_band(credit_score)
        
        # Calculate confidence level
        confidence_level = calculate_confidence_level(features, model, request.scoring_model)
        
        # Determine risk rating
        risk_rating = determine_risk_rating(credit_score, combined_data)
        
        # Extract key factors
        factors = await extract_key_factors(features, model, combined_data)
        
        # Generate recommendations
        recommendations = await generate_recommendations(
            credit_score, factors, request.purpose, combined_data
        )
        
        # Check compliance
        compliance_flags = await compliance_service.check_epic5_compliance(
            request.user_id, combined_data, credit_score
        )
        
        # Calculate next review date
        next_review_date = calculate_next_review_date(credit_score, risk_rating)
        
        processing_time = (datetime.utcnow() - start_time).total_seconds()
        
        # Store result in database
        credit_score_record = CreditScore(
            user_id=request.user_id,
            company_id=request.company_id,
            score=credit_score,
            score_band=score_band,
            confidence_level=confidence_level,
            risk_rating=risk_rating,
            scoring_model=request.scoring_model,
            factors=factors,
            recommendations=recommendations,
            next_review_date=next_review_date,
            compliance_flags=compliance_flags,
            alternative_data_used=request.include_alternative_data,
            idd_core_integrated=request.include_idd_core,
            processing_time=processing_time,
            purpose=request.purpose,
            created_by=current_user['id']
        )
        
        db.add(credit_score_record)
        await db.commit()
        await db.refresh(credit_score_record)
        
        # Schedule background tasks
        background_tasks.add_task(
            update_credit_history, request.user_id, credit_score, db
        )
        background_tasks.add_task(
            trigger_monitoring_alerts, credit_score_record, db
        )
        
        logger.info(f"Credit score calculated: {credit_score} for user {request.user_id} "
                   f"using {request.scoring_model} model")
        
        return CreditScoreResult(
            user_id=request.user_id,
            company_id=request.company_id,
            credit_score=credit_score,
            score_band=score_band,
            confidence_level=confidence_level,
            risk_rating=risk_rating,
            scoring_model=request.scoring_model,
            factors=factors,
            recommendations=recommendations,
            next_review_date=next_review_date,
            compliance_flags=compliance_flags,
            alternative_data_used=request.include_alternative_data,
            idd_core_integrated=request.include_idd_core,
            processing_time=processing_time,
            created_at=credit_score_record.created_at
        )
        
    except Exception as e:
        logger.error(f"Credit score calculation failed: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Credit score calculation failed: {str(e)}"
        )

@app.post("/api/v1/risk-assessment/comprehensive", response_model=RiskAssessmentResult)
async def comprehensive_risk_assessment(
    background_tasks: BackgroundTasks,
    request: RiskAssessmentRequest,
    current_user = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """Perform comprehensive risk assessment with fraud detection and mitigation strategies"""
    try:
        start_time = datetime.utcnow()
        
        # Validate access permissions
        if not await validate_user_access(request.user_id, current_user):
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail="Access denied to user's risk information"
            )
        
        # Get latest credit score
        latest_score = await get_latest_credit_score(request.user_id, request.company_id, db)
        if not latest_score:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="Credit score required for risk assessment. Please calculate credit score first."
            )
        
        # Collect comprehensive risk data
        risk_data = await collect_risk_data(
            request.user_id, request.company_id, request.assessment_type, db
        )
        
        # Perform fraud detection if requested
        fraud_score = 0.0
        if request.include_fraud_check:
            fraud_result = await fraud_detection_service.assess_fraud_risk(
                request.user_id, request.company_id, risk_data
            )
            fraud_score = fraud_result.get('fraud_score', 0.0)
            risk_data['fraud_indicators'] = fraud_result.get('indicators', [])
        
        # Market analysis if requested
        market_data = {}
        if request.include_market_analysis:
            market_data = await external_data_service.get_market_analysis(
                risk_data.get('industry'), risk_data.get('region')
            )
            risk_data.update(market_data)
        
        # Calculate risk metrics
        overall_risk_score = calculate_overall_risk_score(
            latest_score.score, fraud_score, risk_data
        )
        
        risk_category = determine_risk_category(overall_risk_score)
        
        # Calculate probability of default using advanced models
        probability_of_default = await calculate_probability_of_default(
            latest_score.score, risk_data, request.loan_amount
        )
        
        # Calculate loss given default
        loss_given_default = calculate_loss_given_default(
            request.loan_amount, request.collateral_value, risk_data
        )
        
        # Calculate exposure at default
        exposure_at_default = float(request.loan_amount) if request.loan_amount else None
        
        # Calculate expected loss
        expected_loss = probability_of_default * loss_given_default
        if exposure_at_default:
            expected_loss *= exposure_at_default
        
        # Extract key risk factors
        risk_factors = await extract_risk_factors(risk_data, fraud_score, market_data)
        
        # Generate mitigation strategies
        mitigation_strategies = await risk_mitigation_service.generate_strategies(
            overall_risk_score, risk_factors, request.loan_amount
        )
        
        # Define monitoring requirements
        monitoring_requirements = generate_monitoring_requirements(
            risk_category, risk_factors
        )
        
        # Make approval recommendation
        approval_recommendation, conditions = make_approval_recommendation(
            overall_risk_score, expected_loss, risk_factors, request.loan_amount
        )
        
        # Check compliance status
        compliance_status = await compliance_service.check_risk_compliance(
            overall_risk_score, risk_factors, request.loan_amount
        )
        
        # Set validity period
        valid_until = datetime.utcnow() + timedelta(days=30)  # Risk assessment valid for 30 days
        
        processing_time = (datetime.utcnow() - start_time).total_seconds()
        
        # Store risk assessment
        risk_assessment = RiskAssessment(
            user_id=request.user_id,
            company_id=request.company_id,
            credit_score_id=latest_score.id,
            overall_risk_score=overall_risk_score,
            risk_category=risk_category,
            probability_of_default=probability_of_default,
            loss_given_default=loss_given_default,
            exposure_at_default=exposure_at_default,
            expected_loss=expected_loss,
            risk_factors=risk_factors,
            mitigation_strategies=mitigation_strategies,
            monitoring_requirements=monitoring_requirements,
            approval_recommendation=approval_recommendation,
            conditions=conditions,
            valid_until=valid_until,
            compliance_status=compliance_status,
            assessment_type=request.assessment_type,
            loan_amount=request.loan_amount,
            loan_purpose=request.loan_purpose,
            fraud_score=fraud_score,
            processing_time=processing_time,
            created_by=current_user['id']
        )
        
        db.add(risk_assessment)
        await db.commit()
        await db.refresh(risk_assessment)
        
        logger.info(f"Risk assessment completed: {risk_category} for user {request.user_id}")
        
        return RiskAssessmentResult(
            user_id=request.user_id,
            company_id=request.company_id,
            overall_risk_score=overall_risk_score,
            risk_category=risk_category,
            probability_of_default=probability_of_default,
            loss_given_default=loss_given_default,
            exposure_at_default=exposure_at_default,
            expected_loss=expected_loss,
            risk_factors=risk_factors,
            mitigation_strategies=mitigation_strategies,
            monitoring_requirements=monitoring_requirements,
            approval_recommendation=approval_recommendation,
            conditions=conditions,
            valid_until=valid_until,
            compliance_status=compliance_status,
            created_at=risk_assessment.created_at
        )
        
    except Exception as e:
        logger.error(f"Risk assessment failed: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Risk assessment failed: {str(e)}"
        )

@app.post("/api/v1/alternative-data/collect")
async def collect_alternative_data(
    request: AlternativeDataRequest,
    current_user = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """Collect and analyze alternative data sources for credit scoring"""
    try:
        # Validate consent
        if not request.consent_provided:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="User consent required for alternative data collection"
            )
        
        # Validate access permissions
        if not await validate_user_access(request.user_id, current_user):
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail="Access denied to collect user's alternative data"
            )
        
        collected_data = {}
        
        for source in request.data_sources:
            try:
                if source == "social_media":
                    data = await external_data_service.collect_social_media_data(
                        request.user_id, request.company_id
                    )
                elif source == "web_presence":
                    data = await external_data_service.collect_web_presence_data(
                        request.user_id, request.company_id
                    )
                elif source == "transaction_data":
                    data = await external_data_service.collect_transaction_data(
                        request.user_id, request.company_id
                    )
                elif source == "behavioral_data":
                    data = await external_data_service.collect_behavioral_data(
                        request.user_id, request.company_id
                    )
                else:
                    logger.warning(f"Unknown data source: {source}")
                    continue
                
                collected_data[source] = data
                
            except Exception as e:
                logger.warning(f"Failed to collect {source} data: {str(e)}")
                collected_data[source] = {"error": str(e)}
        
        # Store alternative data with retention policy
        alt_data_record = AlternativeDataSource(
            user_id=request.user_id,
            company_id=request.company_id,
            data_sources=request.data_sources,
            collected_data=collected_data,
            consent_provided=request.consent_provided,
            retention_until=datetime.utcnow() + timedelta(days=request.data_retention_period),
            created_by=current_user['id']
        )
        
        db.add(alt_data_record)
        await db.commit()
        await db.refresh(alt_data_record)
        
        logger.info(f"Alternative data collected for user {request.user_id}: {request.data_sources}")
        
        return {
            "message": "Alternative data collected successfully",
            "user_id": request.user_id,
            "data_sources": request.data_sources,
            "collected_count": len([v for v in collected_data.values() if "error" not in v]),
            "failed_count": len([v for v in collected_data.values() if "error" in v]),
            "retention_until": alt_data_record.retention_until.isoformat(),
            "data_summary": {source: len(data) if isinstance(data, dict) and "error" not in data else "failed" 
                           for source, data in collected_data.items()}
        }
        
    except Exception as e:
        logger.error(f"Alternative data collection failed: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Alternative data collection failed: {str(e)}"
        )

@app.post("/api/v1/fraud-detection/assess")
async def assess_fraud_risk(
    request: FraudDetectionRequest,
    current_user = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """Assess fraud risk using advanced ML models and behavioral analysis"""
    try:
        # Validate access permissions
        if not await validate_user_access(request.user_id, current_user):
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail="Access denied to user's fraud assessment"
            )
        
        # Collect comprehensive fraud data
        fraud_data = {
            "transaction_data": request.transaction_data or {},
            "behavioral_data": request.behavioral_data or {},
            "document_data": request.document_data or {},
        }
        
        # Get historical data for comparison
        if request.user_id:
            historical_data = await get_user_historical_data(request.user_id, db)
            fraud_data["historical_patterns"] = historical_data
        
        # Perform fraud detection
        fraud_result = await fraud_detection_service.comprehensive_fraud_check(
            request.user_id, request.company_id, fraud_data, request.real_time_check
        )
        
        logger.info(f"Fraud assessment completed for user {request.user_id}: "
                   f"Risk level {fraud_result.get('risk_level')}")
        
        return {
            "user_id": request.user_id,
            "company_id": request.company_id,
            "fraud_score": fraud_result.get("fraud_score", 0.0),
            "risk_level": fraud_result.get("risk_level", "low"),
            "indicators": fraud_result.get("indicators", []),
            "confidence": fraud_result.get("confidence", 0.0),
            "recommendations": fraud_result.get("recommendations", []),
            "requires_manual_review": fraud_result.get("requires_manual_review", False),
            "real_time_check": request.real_time_check,
            "assessment_time": datetime.utcnow().isoformat()
        }
        
    except Exception as e:
        logger.error(f"Fraud detection failed: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Fraud detection failed: {str(e)}"
        )

@app.post("/api/v1/bulk-scoring/submit")
async def submit_bulk_scoring(
    background_tasks: BackgroundTasks,
    request: BulkScoringRequest,
    current_user = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """Submit bulk credit scoring request for multiple users"""
    try:
        if len(request.user_ids) > settings.max_bulk_scoring_count:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail=f"Too many users. Maximum {settings.max_bulk_scoring_count} allowed"
            )
        
        # Validate user has permission for bulk operations
        if current_user['role'] not in ['admin', 'bank', 'analyst']:
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail="Insufficient permissions for bulk scoring"
            )
        
        # Create bulk job record
        bulk_job = await create_bulk_scoring_job(
            request.user_ids, request.scoring_model, request.priority,
            current_user['id'], db
        )
        
        # Schedule background processing
        background_tasks.add_task(
            process_bulk_scoring, bulk_job.id, request.user_ids,
            request.scoring_model, request.callback_url, db
        )
        
        logger.info(f"Bulk scoring job submitted: {bulk_job.id} for {len(request.user_ids)} users")
        
        return {
            "job_id": str(bulk_job.id),
            "user_count": len(request.user_ids),
            "scoring_model": request.scoring_model,
            "priority": request.priority,
            "estimated_completion": datetime.utcnow() + timedelta(minutes=len(request.user_ids) * 2),
            "status": "queued",
            "callback_url": request.callback_url
        }
        
    except Exception as e:
        logger.error(f"Bulk scoring submission failed: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Bulk scoring submission failed: {str(e)}"
        )

@app.get("/api/v1/models/available")
async def get_available_models():
    """Get list of available credit scoring models"""
    return {
        "models": SCORING_MODELS,
        "default_model": "hybrid",
        "recommended_model": "idd_core",
        "model_performance": await model_manager.get_model_performance_metrics()
    }

@app.get("/api/v1/credit-score/{user_id}/history")
async def get_credit_score_history(
    user_id: str,
    limit: int = 10,
    current_user = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """Get credit score history for a user"""
    try:
        # Validate access permissions
        if not await validate_user_access(user_id, current_user):
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail="Access denied to user's credit history"
            )
        
        history = await get_user_credit_history(user_id, limit, db)
        
        return {
            "user_id": user_id,
            "history": history,
            "trend_analysis": analyze_credit_trend(history),
            "total_records": len(history)
        }
        
    except Exception as e:
        logger.error(f"Failed to get credit history: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to retrieve credit history"
        )

# Helper functions (implementations would be in separate modules)
async def validate_user_access(user_id: str, current_user: dict) -> bool:
    """Validate if current user has access to target user's data"""
    if current_user['role'] in ['admin', 'bank']:
        return True
    return current_user['id'] == user_id

async def collect_financial_data(user_id: str, company_id: Optional[str], db: AsyncSession) -> Dict[str, Any]:
    """Collect traditional financial data"""
    # Implementation would fetch from database and external sources
    return {
        "annual_revenue": 1000000,
        "profit_margin": 0.15,
        "current_ratio": 2.5,
        "debt_to_equity": 0.3,
        "years_in_business": 5,
        "payment_history_score": 85
    }

def get_score_band(credit_score: int) -> str:
    """Determine score band based on credit score"""
    if credit_score >= 750:
        return "excellent"
    elif credit_score >= 700:
        return "good"
    elif credit_score >= 650:
        return "fair"
    else:
        return "poor"

def calculate_confidence_level(features: List[float], model, scoring_model: str) -> float:
    """Calculate confidence level for the credit score"""
    # Implementation would use model uncertainty estimation
    return 0.85

def determine_risk_rating(credit_score: int, data: Dict[str, Any]) -> str:
    """Determine risk rating based on credit score and other factors"""
    if credit_score >= 750:
        return "low"
    elif credit_score >= 700:
        return "medium"
    elif credit_score >= 650:
        return "high"
    else:
        return "very_high"

async def extract_key_factors(features: List[float], model, data: Dict[str, Any]) -> Dict[str, Any]:
    """Extract key factors influencing the credit score"""
    return {
        "payment_history": 35,
        "credit_utilization": 30,
        "length_of_credit_history": 15,
        "types_of_credit": 10,
        "new_credit_inquiries": 10
    }

async def generate_recommendations(credit_score: int, factors: Dict[str, Any], purpose: str, data: Dict[str, Any]) -> List[str]:
    """Generate personalized recommendations for credit improvement"""
    recommendations = []
    
    if credit_score < 650:
        recommendations.append("Improve payment history by making all payments on time")
        recommendations.append("Reduce credit utilization below 30%")
    
    if credit_score < 700:
        recommendations.append("Consider diversifying credit types")
        recommendations.append("Avoid opening too many new credit accounts")
    
    return recommendations

def calculate_next_review_date(credit_score: int, risk_rating: str) -> datetime:
    """Calculate next review date based on score and risk"""
    if risk_rating == "very_high":
        return datetime.utcnow() + timedelta(days=30)
    elif risk_rating == "high":
        return datetime.utcnow() + timedelta(days=60)
    elif risk_rating == "medium":
        return datetime.utcnow() + timedelta(days=90)
    else:
        return datetime.utcnow() + timedelta(days=180)

if __name__ == "__main__":
    uvicorn.run(
        "main:app",
        host=settings.host,
        port=settings.port,
        reload=settings.debug,
        workers=1 if settings.debug else settings.workers
    )
