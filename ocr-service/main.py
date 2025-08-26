"""
OCR Service with Gemini AI Integration
Extracts structured data from documents with field recognition and manual correction capabilities
Supports Epic 2 compliance and integration with document workflows
"""

import os
import io
import asyncio
import logging
from datetime import datetime
from typing import Dict, List, Optional, Any, Union
from pathlib import Path
import base64
import json

import uvicorn
from fastapi import FastAPI, HTTPException, UploadFile, File, Depends, status, Form, BackgroundTasks
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from pydantic import BaseModel, validator
import google.generativeai as genai
from PIL import Image
import cv2
import numpy as np
import pytesseract
from pdf2image import convert_from_bytes

# Internal imports
from config import settings
from services.auth_service import AuthService
from services.document_classifier import DocumentClassifier
from services.field_extractor import FieldExtractor
from services.validation_service import ValidationService
from middleware.security import SecurityMiddleware
from middleware.rate_limit import RateLimitMiddleware
from utils.image_processor import ImageProcessor
from utils.logger import setup_logger

# Setup logging
logger = setup_logger(__name__)

# Initialize FastAPI app
app = FastAPI(
    title="OCR Service",
    version="1.0.0",
    description="Advanced OCR with Gemini AI for structured data extraction",
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
document_classifier = DocumentClassifier()
field_extractor = FieldExtractor()
validation_service = ValidationService()
image_processor = ImageProcessor()

# Document types and their expected fields
DOCUMENT_SCHEMAS = {
    "invoice": {
        "required_fields": [
            "invoice_number", "invoice_date", "due_date", "total_amount",
            "supplier_name", "supplier_address", "buyer_name", "buyer_address"
        ],
        "optional_fields": [
            "tax_amount", "discount", "payment_terms", "currency",
            "line_items", "subtotal", "tax_rate", "po_number"
        ],
        "validation_rules": {
            "invoice_number": "string",
            "total_amount": "number",
            "invoice_date": "date",
            "due_date": "date"
        }
    },
    "contract": {
        "required_fields": [
            "contract_number", "contract_date", "parties", "contract_value",
            "start_date", "end_date", "scope_of_work"
        ],
        "optional_fields": [
            "payment_terms", "penalties", "termination_clause",
            "governing_law", "signatures", "witness"
        ],
        "validation_rules": {
            "contract_number": "string",
            "contract_value": "number",
            "contract_date": "date",
            "start_date": "date",
            "end_date": "date"
        }
    },
    "identity": {
        "required_fields": [
            "full_name", "document_number", "date_of_birth",
            "issue_date", "expiry_date", "nationality"
        ],
        "optional_fields": [
            "address", "place_of_birth", "issuing_authority",
            "gender", "height", "photo"
        ],
        "validation_rules": {
            "document_number": "string",
            "date_of_birth": "date",
            "issue_date": "date",
            "expiry_date": "date"
        }
    },
    "bank_statement": {
        "required_fields": [
            "account_number", "account_holder", "statement_period",
            "opening_balance", "closing_balance", "bank_name"
        ],
        "optional_fields": [
            "transactions", "total_credits", "total_debits",
            "average_balance", "statement_date", "branch_code"
        ],
        "validation_rules": {
            "account_number": "string",
            "opening_balance": "number",
            "closing_balance": "number",
            "statement_period": "string"
        }
    }
}

# Request/Response models
class OCRRequest(BaseModel):
    document_type: str
    language: str = "en"
    extract_tables: bool = True
    extract_signatures: bool = False
    confidence_threshold: float = 0.7
    manual_review_required: bool = False
    
    @validator('document_type')
    def validate_document_type(cls, v):
        if v not in DOCUMENT_SCHEMAS:
            raise ValueError(f'Unsupported document type: {v}')
        return v

class OCRResult(BaseModel):
    document_id: Optional[str] = None
    raw_text: str
    structured_data: Dict[str, Any]
    confidence_score: float
    processing_time: float
    detected_language: str
    page_count: int
    field_confidences: Dict[str, float]
    extracted_tables: List[Dict[str, Any]] = []
    extracted_signatures: List[Dict[str, Any]] = []
    validation_errors: List[str] = []
    requires_manual_review: bool = False

class FieldCorrectionRequest(BaseModel):
    document_id: str
    field_name: str
    corrected_value: str
    confidence_override: Optional[float] = None

class BulkOCRRequest(BaseModel):
    document_type: str
    language: str = "en"
    extract_tables: bool = True
    confidence_threshold: float = 0.7

class ManualReviewRequest(BaseModel):
    document_id: str
    reviewer_id: str
    corrections: Dict[str, str]
    approval_status: str  # approved, rejected, needs_revision
    comments: Optional[str] = None

@app.on_event("startup")
async def startup_event():
    """Initialize services on startup"""
    # Configure Gemini AI
    genai.configure(api_key=settings.gemini_api_key)
    
    # Initialize services
    await document_classifier.initialize()
    await field_extractor.initialize()
    
    logger.info("OCR Service started successfully")

@app.on_event("shutdown")
async def shutdown_event():
    """Cleanup on shutdown"""
    await document_classifier.cleanup()
    await field_extractor.cleanup()
    logger.info("OCR Service shut down")

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
        "service": "OCR Service",
        "version": "1.0.0",
        "status": "healthy",
        "features": {
            "ai_engine": "Gemini AI",
            "ocr_engine": "Tesseract + Gemini Vision",
            "document_types": list(DOCUMENT_SCHEMAS.keys()),
            "languages": ["en", "es", "fr", "de", "zh", "ja", "ar"],
            "compliance": "Epic-2"
        }
    }

@app.post("/api/v1/ocr/extract", response_model=OCRResult)
async def extract_document_data(
    background_tasks: BackgroundTasks,
    file: UploadFile = File(...),
    document_type: str = Form(...),
    language: str = Form("en"),
    extract_tables: bool = Form(True),
    extract_signatures: bool = Form(False),
    confidence_threshold: float = Form(0.7),
    manual_review_required: bool = Form(False),
    current_user = Depends(get_current_user)
):
    """Extract structured data from document using OCR and AI"""
    try:
        start_time = datetime.utcnow()
        
        # Validate document type
        if document_type not in DOCUMENT_SCHEMAS:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail=f"Unsupported document type: {document_type}"
            )
        
        # Read file content
        content = await file.read()
        
        # Process document based on file type
        if file.content_type.startswith('image/'):
            images = [Image.open(io.BytesIO(content))]
            page_count = 1
        elif file.content_type == 'application/pdf':
            images = convert_from_bytes(content, dpi=300)
            page_count = len(images)
        else:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="Unsupported file type. Please upload PDF or image files."
            )
        
        # Preprocess images
        processed_images = []
        for image in images:
            processed_image = await image_processor.preprocess(image)
            processed_images.append(processed_image)
        
        # Perform OCR using Tesseract for raw text
        raw_text = ""
        for image in processed_images:
            # Convert PIL image to numpy array for Tesseract
            image_np = np.array(image)
            text = pytesseract.image_to_string(image_np, lang=language)
            raw_text += text + "\n"
        
        # Use Gemini AI for structured extraction
        structured_data = await extract_with_gemini(
            processed_images, document_type, language, raw_text
        )
        
        # Extract tables if requested
        extracted_tables = []
        if extract_tables:
            extracted_tables = await extract_tables_with_gemini(processed_images)
        
        # Extract signatures if requested
        extracted_signatures = []
        if extract_signatures:
            extracted_signatures = await extract_signatures(processed_images)
        
        # Validate extracted data
        validation_errors = await validation_service.validate_document_data(
            structured_data, document_type
        )
        
        # Calculate confidence scores
        field_confidences = await calculate_field_confidences(
            structured_data, raw_text, document_type
        )
        
        # Overall confidence score
        overall_confidence = sum(field_confidences.values()) / len(field_confidences) if field_confidences else 0
        
        # Determine if manual review is required
        requires_manual_review = (
            manual_review_required or 
            overall_confidence < confidence_threshold or
            len(validation_errors) > 0
        )
        
        # Detect language
        detected_language = await detect_language(raw_text)
        
        processing_time = (datetime.utcnow() - start_time).total_seconds()
        
        # Log successful processing
        logger.info(f"OCR processing completed for {document_type} document by user {current_user['id']}")
        
        result = OCRResult(
            raw_text=raw_text,
            structured_data=structured_data,
            confidence_score=overall_confidence,
            processing_time=processing_time,
            detected_language=detected_language,
            page_count=page_count,
            field_confidences=field_confidences,
            extracted_tables=extracted_tables,
            extracted_signatures=extracted_signatures,
            validation_errors=validation_errors,
            requires_manual_review=requires_manual_review
        )
        
        return result
        
    except Exception as e:
        logger.error(f"OCR processing failed: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"OCR processing failed: {str(e)}"
        )

@app.post("/api/v1/ocr/bulk-extract")
async def bulk_extract_documents(
    background_tasks: BackgroundTasks,
    files: List[UploadFile] = File(...),
    request: BulkOCRRequest = Depends(),
    current_user = Depends(get_current_user)
):
    """Process multiple documents in bulk"""
    try:
        if len(files) > settings.max_bulk_ocr_count:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail=f"Too many files. Maximum {settings.max_bulk_ocr_count} allowed"
            )
        
        results = []
        failed_files = []
        
        for file in files:
            try:
                # Process each file
                content = await file.read()
                
                # Reset file pointer
                await file.seek(0)
                
                # Extract data (simplified version of main extraction)
                result = await process_single_file(
                    content, file.content_type, request.document_type,
                    request.language, request.extract_tables,
                    request.confidence_threshold
                )
                
                results.append({
                    "filename": file.filename,
                    "result": result,
                    "status": "success"
                })
                
            except Exception as e:
                failed_files.append({
                    "filename": file.filename,
                    "error": str(e),
                    "status": "failed"
                })
        
        logger.info(f"Bulk OCR completed: {len(results)} successful, {len(failed_files)} failed")
        
        return {
            "processed_files": results,
            "failed_files": failed_files,
            "total_processed": len(files),
            "success_count": len(results),
            "failure_count": len(failed_files)
        }
        
    except Exception as e:
        logger.error(f"Bulk OCR processing failed: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Bulk OCR processing failed: {str(e)}"
        )

@app.post("/api/v1/ocr/correct-field")
async def correct_field(
    request: FieldCorrectionRequest,
    current_user = Depends(get_current_user)
):
    """Manually correct a specific field value"""
    try:
        # Log the correction
        logger.info(f"Field correction: {request.field_name} = {request.corrected_value} "
                   f"for document {request.document_id} by user {current_user['id']}")
        
        # Here you would typically update the document in your database
        # and potentially retrain your models with the corrected data
        
        # Update field confidence if provided
        updated_confidence = request.confidence_override if request.confidence_override else 1.0
        
        return {
            "message": "Field corrected successfully",
            "field_name": request.field_name,
            "old_value": "previous_value",  # You'd fetch this from your database
            "new_value": request.corrected_value,
            "updated_confidence": updated_confidence,
            "corrected_by": current_user['id'],
            "corrected_at": datetime.utcnow().isoformat()
        }
        
    except Exception as e:
        logger.error(f"Field correction failed: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Field correction failed"
        )

@app.post("/api/v1/ocr/manual-review")
async def submit_manual_review(
    request: ManualReviewRequest,
    current_user = Depends(get_current_user)
):
    """Submit manual review results"""
    try:
        # Validate reviewer permissions
        if current_user['role'] not in ['admin', 'reviewer', 'bank']:
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail="Insufficient permissions for manual review"
            )
        
        # Process review
        review_result = {
            "document_id": request.document_id,
            "reviewer_id": request.reviewer_id,
            "review_status": request.approval_status,
            "corrections_count": len(request.corrections),
            "corrections": request.corrections,
            "comments": request.comments,
            "reviewed_at": datetime.utcnow().isoformat(),
            "reviewed_by": current_user['id']
        }
        
        # Log the review
        logger.info(f"Manual review completed for document {request.document_id}: "
                   f"{request.approval_status} by {current_user['id']}")
        
        return {
            "message": "Manual review submitted successfully",
            "review_result": review_result
        }
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Manual review submission failed: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Manual review submission failed"
        )

@app.get("/api/v1/ocr/document-schema/{document_type}")
async def get_document_schema(
    document_type: str,
    current_user = Depends(get_current_user)
):
    """Get the expected schema for a document type"""
    try:
        if document_type not in DOCUMENT_SCHEMAS:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail=f"Schema not found for document type: {document_type}"
            )
        
        schema = DOCUMENT_SCHEMAS[document_type]
        return {
            "document_type": document_type,
            "schema": schema,
            "total_fields": len(schema["required_fields"]) + len(schema["optional_fields"])
        }
        
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Failed to get document schema: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to retrieve document schema"
        )

@app.get("/api/v1/ocr/supported-languages")
async def get_supported_languages():
    """Get list of supported OCR languages"""
    return {
        "languages": [
            {"code": "en", "name": "English"},
            {"code": "es", "name": "Spanish"},
            {"code": "fr", "name": "French"},
            {"code": "de", "name": "German"},
            {"code": "zh", "name": "Chinese"},
            {"code": "ja", "name": "Japanese"},
            {"code": "ar", "name": "Arabic"},
            {"code": "hi", "name": "Hindi"},
            {"code": "pt", "name": "Portuguese"},
            {"code": "ru", "name": "Russian"}
        ]
    }

# Helper functions
async def extract_with_gemini(images: List[Image.Image], document_type: str, language: str, raw_text: str) -> Dict[str, Any]:
    """Use Gemini AI for structured data extraction"""
    try:
        # Get document schema
        schema = DOCUMENT_SCHEMAS[document_type]
        
        # Prepare prompt for Gemini
        prompt = f"""
        Extract structured data from this {document_type} document in {language}.
        
        Required fields: {', '.join(schema['required_fields'])}
        Optional fields: {', '.join(schema['optional_fields'])}
        
        Raw OCR text: {raw_text[:2000]}...  # Limit text for context
        
        Return the data as a JSON object with the field names as keys.
        For dates, use ISO format (YYYY-MM-DD).
        For amounts, return as numbers without currency symbols.
        If a field is not found, set it to null.
        Include a confidence score for each field (0-1).
        """
        
        # Configure Gemini model
        model = genai.GenerativeModel('gemini-1.5-pro-vision-latest')
        
        # Convert first image to base64 for Gemini
        img_byte_arr = io.BytesIO()
        images[0].save(img_byte_arr, format='PNG')
        img_byte_arr = img_byte_arr.getvalue()
        
        # Generate content with Gemini
        response = model.generate_content([prompt, images[0]])
        
        # Parse JSON response
        try:
            structured_data = json.loads(response.text)
        except json.JSONDecodeError:
            # Fallback: extract data using regex patterns
            structured_data = await extract_with_patterns(raw_text, document_type)
        
        return structured_data
        
    except Exception as e:
        logger.error(f"Gemini extraction failed: {str(e)}")
        # Fallback to pattern-based extraction
        return await extract_with_patterns(raw_text, document_type)

async def extract_with_patterns(text: str, document_type: str) -> Dict[str, Any]:
    """Fallback extraction using regex patterns"""
    import re
    
    extracted_data = {}
    
    # Common patterns for different document types
    patterns = {
        "invoice": {
            "invoice_number": r"(?:Invoice|Invoice No|Invoice #)\s*:?\s*([A-Z0-9-]+)",
            "total_amount": r"(?:Total|Amount Due|Grand Total)\s*:?\s*\$?(\d+\.?\d*)",
            "invoice_date": r"(?:Invoice Date|Date)\s*:?\s*(\d{1,2}[/-]\d{1,2}[/-]\d{4})",
            "due_date": r"(?:Due Date|Payment Due)\s*:?\s*(\d{1,2}[/-]\d{1,2}[/-]\d{4})"
        },
        "contract": {
            "contract_number": r"(?:Contract|Contract No|Agreement No)\s*:?\s*([A-Z0-9-]+)",
            "contract_value": r"(?:Contract Value|Total Value|Amount)\s*:?\s*\$?(\d+\.?\d*)",
            "contract_date": r"(?:Contract Date|Agreement Date|Date)\s*:?\s*(\d{1,2}[/-]\d{1,2}[/-]\d{4})"
        }
    }
    
    if document_type in patterns:
        for field, pattern in patterns[document_type].items():
            match = re.search(pattern, text, re.IGNORECASE)
            if match:
                extracted_data[field] = match.group(1)
            else:
                extracted_data[field] = None
    
    return extracted_data

async def extract_tables_with_gemini(images: List[Image.Image]) -> List[Dict[str, Any]]:
    """Extract tables using Gemini Vision"""
    try:
        tables = []
        
        for idx, image in enumerate(images):
            prompt = """
            Extract any tables from this document image.
            Return the tables as JSON arrays with headers and data rows.
            Include table metadata like position and size if possible.
            """
            
            model = genai.GenerativeModel('gemini-1.5-pro-vision-latest')
            response = model.generate_content([prompt, image])
            
            try:
                table_data = json.loads(response.text)
                if isinstance(table_data, list):
                    for table in table_data:
                        table['page'] = idx + 1
                    tables.extend(table_data)
                elif isinstance(table_data, dict):
                    table_data['page'] = idx + 1
                    tables.append(table_data)
            except json.JSONDecodeError:
                logger.warning(f"Failed to parse table JSON for page {idx + 1}")
        
        return tables
        
    except Exception as e:
        logger.error(f"Table extraction failed: {str(e)}")
        return []

async def extract_signatures(images: List[Image.Image]) -> List[Dict[str, Any]]:
    """Extract signatures using image processing"""
    try:
        signatures = []
        
        for idx, image in enumerate(images):
            # Convert to numpy array for OpenCV
            image_np = np.array(image)
            gray = cv2.cvtColor(image_np, cv2.COLOR_RGB2GRAY)
            
            # Apply signature detection algorithms
            # This is a simplified example - you'd use more sophisticated methods
            edges = cv2.Canny(gray, 50, 150)
            contours, _ = cv2.findContours(edges, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
            
            for contour in contours:
                area = cv2.contourArea(contour)
                if 1000 < area < 50000:  # Filter by size
                    x, y, w, h = cv2.boundingRect(contour)
                    
                    signatures.append({
                        "page": idx + 1,
                        "bounds": {"x": int(x), "y": int(y), "width": int(w), "height": int(h)},
                        "area": int(area),
                        "confidence": 0.8  # Placeholder confidence
                    })
        
        return signatures[:5]  # Limit to 5 potential signatures
        
    except Exception as e:
        logger.error(f"Signature extraction failed: {str(e)}")
        return []

async def calculate_field_confidences(structured_data: Dict[str, Any], raw_text: str, document_type: str) -> Dict[str, float]:
    """Calculate confidence scores for extracted fields"""
    confidences = {}
    
    for field, value in structured_data.items():
        if value is None:
            confidences[field] = 0.0
        elif isinstance(value, str) and value.strip() == "":
            confidences[field] = 0.0
        else:
            # Simple confidence calculation based on presence in raw text
            if str(value) in raw_text:
                confidences[field] = 0.9
            else:
                confidences[field] = 0.5  # Medium confidence for derived values
    
    return confidences

async def detect_language(text: str) -> str:
    """Detect the primary language of the text"""
    try:
        # Simple language detection based on character patterns
        # In production, you'd use a proper language detection library
        
        # Count English words vs other patterns
        import re
        english_words = len(re.findall(r'\b[a-zA-Z]+\b', text))
        total_words = len(text.split())
        
        if total_words > 0 and english_words / total_words > 0.7:
            return "en"
        else:
            return "auto"  # Auto-detect
            
    except Exception:
        return "en"  # Default to English

async def process_single_file(content: bytes, content_type: str, document_type: str, language: str, extract_tables: bool, confidence_threshold: float) -> Dict[str, Any]:
    """Process a single file for bulk operations"""
    # Simplified version of the main extraction logic
    # This would be extracted to a common service in a real implementation
    
    if content_type.startswith('image/'):
        images = [Image.open(io.BytesIO(content))]
    elif content_type == 'application/pdf':
        images = convert_from_bytes(content, dpi=300)
    else:
        raise ValueError("Unsupported file type")
    
    # Basic OCR
    raw_text = ""
    for image in images:
        image_np = np.array(image)
        text = pytesseract.image_to_string(image_np, lang=language)
        raw_text += text + "\n"
    
    # Extract structured data
    structured_data = await extract_with_patterns(raw_text, document_type)
    
    # Calculate basic confidence
    field_count = len([v for v in structured_data.values() if v is not None])
    total_fields = len(DOCUMENT_SCHEMAS[document_type]["required_fields"])
    confidence = field_count / total_fields if total_fields > 0 else 0
    
    return {
        "raw_text": raw_text,
        "structured_data": structured_data,
        "confidence_score": confidence,
        "page_count": len(images)
    }

if __name__ == "__main__":
    uvicorn.run(
        "main:app",
        host=settings.host,
        port=settings.port,
        reload=settings.debug,
        workers=1 if settings.debug else settings.workers
    )
