from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from typing import Dict, Any
import random
import logging
from datetime import datetime

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI(title="Invoice Financing AI Service", version="1.0.0")

class CreditScoreRequest(BaseModel):
    user_id: str
    company_data: Dict[str, Any]
    financial_data: Dict[str, Any]
    transaction_history: Dict[str, Any]

class RiskAssessmentRequest(BaseModel):
    invoice_data: Dict[str, Any]
    customer_data: Dict[str, Any]
    historical_data: Dict[str, Any]

class FraudDetectionRequest(BaseModel):
    invoice_data: Dict[str, Any]
    user_data: Dict[str, Any]
    transaction_patterns: Dict[str, Any]

class DocumentVerificationRequest(BaseModel):
    document_url: str
    document_type: str
    expected_fields: Dict[str, Any]

@app.get("/")
def root():
    return {"message": "Invoice Financing AI Service", "status": "running"}

@app.post("/api/ml/credit-score")
def calculate_credit_score(request: CreditScoreRequest):
    """
    Calculate credit score for a user based on various factors.
    In production, this would use ML models trained on historical data.
    """
    try:
        logger.info(f"Calculating credit score for user: {request.user_id}")
        
        # Simulate AI-powered credit scoring
        base_score = 600
        
        # Factors that could influence credit score
        factors = {
            "company_age": request.company_data.get("years_in_business", 0) * 10,
            "annual_revenue": min(request.financial_data.get("annual_revenue", 0) / 100000, 50),
            "payment_history": request.transaction_history.get("on_time_payments", 0.8) * 100,
            "debt_to_income": max(0, 50 - request.financial_data.get("debt_ratio", 0.3) * 100),
            "cash_flow": min(request.financial_data.get("monthly_cash_flow", 0) / 10000, 30)
        }
        
        calculated_score = base_score + sum(factors.values())
        calculated_score = max(300, min(850, calculated_score))  # Cap between 300-850
        
        return {
            "credit_score": int(calculated_score),
            "score_factors": factors,
            "risk_category": "low" if calculated_score > 700 else "medium" if calculated_score > 600 else "high",
            "calculated_at": datetime.utcnow().isoformat()
        }
        
    except Exception as e:
        logger.error(f"Error calculating credit score: {str(e)}")
        raise HTTPException(status_code=500, detail="Credit score calculation failed")

@app.post("/api/ml/risk-assessment")
def assess_risk(request: RiskAssessmentRequest):
    """
    Assess risk for invoice financing.
    In production, this would use ML models to analyze various risk factors.
    """
    try:
        logger.info("Performing risk assessment")
        
        # Simulate AI-powered risk assessment
        invoice_amount = request.invoice_data.get("amount", 0)
        due_date_days = request.invoice_data.get("days_until_due", 30)
        customer_rating = request.customer_data.get("credit_rating", 3)
        
        # Calculate risk score (0-1, where 1 is highest risk)
        amount_risk = min(invoice_amount / 1000000, 0.3)  # Higher amounts = higher risk
        time_risk = max(0, (due_date_days - 30) / 365)  # Longer terms = higher risk
        customer_risk = (5 - customer_rating) / 10  # Lower ratings = higher risk
        
        total_risk_score = (amount_risk + time_risk + customer_risk) / 3
        
        # Determine risk level
        if total_risk_score < 0.3:
            risk_level = "low"
        elif total_risk_score < 0.6:
            risk_level = "medium"
        else:
            risk_level = "high"
        
        return {
            "risk_score": round(total_risk_score, 3),
            "risk_level": risk_level,
            "risk_factors": {
                "amount_risk": round(amount_risk, 3),
                "time_risk": round(time_risk, 3),
                "customer_risk": round(customer_risk, 3)
            },
            "recommended_interest_rate": round(5 + (total_risk_score * 10), 2),
            "assessed_at": datetime.utcnow().isoformat()
        }
        
    except Exception as e:
        logger.error(f"Error in risk assessment: {str(e)}")
        raise HTTPException(status_code=500, detail="Risk assessment failed")

@app.post("/api/ml/fraud-detection")
def detect_fraud(request: FraudDetectionRequest):
    """
    Detect potential fraud in invoice submissions.
    In production, this would use ML models trained on fraud patterns.
    """
    try:
        logger.info("Performing fraud detection")
        
        # Simulate AI-powered fraud detection
        fraud_indicators = []
        fraud_score = 0.0
        
        # Check various fraud indicators
        invoice_amount = request.invoice_data.get("amount", 0)
        if invoice_amount > 1000000:  # Unusually high amounts
            fraud_indicators.append("unusually_high_amount")
            fraud_score += 0.3
            
        customer_name = request.invoice_data.get("customer_name", "")
        if len(customer_name) < 3:  # Suspicious customer names
            fraud_indicators.append("suspicious_customer_name")
            fraud_score += 0.2
            
        # Check submission patterns
        submission_hour = datetime.now().hour
        if submission_hour < 6 or submission_hour > 22:  # Odd submission times
            fraud_indicators.append("unusual_submission_time")
            fraud_score += 0.1
            
        # Random factor to simulate AI uncertainty
        fraud_score += random.uniform(0, 0.2)
        fraud_score = min(fraud_score, 1.0)
        
        is_fraud = fraud_score > 0.7
        
        return {
            "is_fraud": is_fraud,
            "fraud_score": round(fraud_score, 3),
            "confidence": round(1 - abs(fraud_score - 0.5) * 2, 3),
            "fraud_indicators": fraud_indicators,
            "recommendation": "reject" if is_fraud else "approve" if fraud_score < 0.3 else "review",
            "detected_at": datetime.utcnow().isoformat()
        }
        
    except Exception as e:
        logger.error(f"Error in fraud detection: {str(e)}")
        raise HTTPException(status_code=500, detail="Fraud detection failed")

@app.post("/api/ml/verify-document")
def verify_document(request: DocumentVerificationRequest):
    """
    Verify document authenticity using OCR and AI.
    In production, this would use OCR and ML models for document verification.
    """
    try:
        logger.info(f"Verifying document: {request.document_type}")
        
        # Simulate document verification
        verification_score = random.uniform(0.7, 0.95)
        
        extracted_fields = {
            "invoice_number": f"INV-{random.randint(1000, 9999)}",
            "amount": round(random.uniform(1000, 50000), 2),
            "date": datetime.now().strftime("%Y-%m-%d"),
            "customer": "Acme Corp Ltd"
        }
        
        is_authentic = verification_score > 0.8
        
        return {
            "is_authentic": is_authentic,
            "verification_score": round(verification_score, 3),
            "extracted_fields": extracted_fields,
            "confidence": round(verification_score, 3),
            "issues": [] if is_authentic else ["low_image_quality", "partial_text_recognition"],
            "verified_at": datetime.utcnow().isoformat()
        }
        
    except Exception as e:
        logger.error(f"Error in document verification: {str(e)}")
        raise HTTPException(status_code=500, detail="Document verification failed")

@app.get("/health")
def health_check():
    return {"status": "healthy", "timestamp": datetime.utcnow().isoformat()}

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=5000)
