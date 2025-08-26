"""
Advanced Document Processing and OCR Service
"""
import os
import io
import cv2
import numpy as np
import pytesseract
from PIL import Image, ImageEnhance, ImageFilter
from pdf2image import convert_from_path, convert_from_bytes
import PyPDF2
import re
import logging
from typing import Dict, List, Tuple, Any, Optional, Union
from datetime import datetime
import hashlib
import magic

from config import settings

logger = logging.getLogger(__name__)


class DocumentProcessor:
    """Advanced document processing with OCR capabilities"""
    
    def __init__(self):
        self.setup_tesseract()
        self.supported_formats = ['pdf', 'png', 'jpg', 'jpeg', 'tiff', 'bmp']
        self.max_file_size = settings.max_file_size
    
    def setup_tesseract(self):
        """Setup Tesseract OCR configuration"""
        if settings.tesseract_path:
            pytesseract.pytesseract.tesseract_cmd = settings.tesseract_path
        
        self.tesseract_config = settings.tesseract_config
    
    def process_document(self, file_data: bytes, filename: str, document_type: str) -> Dict[str, Any]:
        """Process document and extract information"""
        try:
            # Validate file
            validation_result = self.validate_document(file_data, filename)
            if not validation_result['is_valid']:
                return {
                    'success': False,
                    'error': validation_result['error'],
                    'extracted_data': {},
                    'confidence': 0.0
                }
            
            # Detect file type
            file_type = self.detect_file_type(file_data, filename)
            
            # Extract text based on file type
            if file_type == 'pdf':
                extracted_data = self.process_pdf(file_data, document_type)
            else:
                extracted_data = self.process_image(file_data, document_type)
            
            # Apply document-specific parsing
            parsed_data = self.parse_document_content(extracted_data['text'], document_type)
            
            # Calculate overall confidence
            overall_confidence = self.calculate_confidence(extracted_data, parsed_data)
            
            return {
                'success': True,
                'extracted_data': {
                    'raw_text': extracted_data['text'],
                    'parsed_fields': parsed_data,
                    'metadata': extracted_data.get('metadata', {}),
                    'processing_info': {
                        'file_type': file_type,
                        'document_type': document_type,
                        'processing_method': extracted_data.get('method', 'unknown')
                    }
                },
                'confidence': overall_confidence,
                'issues': extracted_data.get('issues', []),
                'processed_at': datetime.utcnow().isoformat()
            }
            
        except Exception as e:
            logger.error(f"Error processing document: {str(e)}")
            return {
                'success': False,
                'error': f"Document processing failed: {str(e)}",
                'extracted_data': {},
                'confidence': 0.0
            }
    
    def validate_document(self, file_data: bytes, filename: str) -> Dict[str, Any]:
        """Validate document before processing"""
        # Check file size
        if len(file_data) > self.max_file_size:
            return {
                'is_valid': False,
                'error': f"File size exceeds maximum limit of {self.max_file_size / 1024 / 1024:.1f}MB"
            }
        
        # Check file extension
        file_ext = filename.split('.')[-1].lower()
        if file_ext not in self.supported_formats:
            return {
                'is_valid': False,
                'error': f"Unsupported file format. Supported formats: {', '.join(self.supported_formats)}"
            }
        
        # Check file magic bytes (MIME type)
        try:
            mime_type = magic.from_buffer(file_data, mime=True)
            expected_mimes = {
                'pdf': 'application/pdf',
                'png': 'image/png',
                'jpg': 'image/jpeg',
                'jpeg': 'image/jpeg',
                'tiff': 'image/tiff',
                'bmp': 'image/bmp'
            }
            
            if expected_mimes.get(file_ext) and expected_mimes[file_ext] not in mime_type:
                return {
                    'is_valid': False,
                    'error': f"File content doesn't match extension. Expected {expected_mimes[file_ext]}, got {mime_type}"
                }
        except Exception:
            logger.warning("Could not validate file MIME type")
        
        return {'is_valid': True}
    
    def detect_file_type(self, file_data: bytes, filename: str) -> str:
        """Detect the actual file type"""
        file_ext = filename.split('.')[-1].lower()
        
        # For now, trust the extension but validate with magic bytes if available
        try:
            mime_type = magic.from_buffer(file_data, mime=True)
            if 'pdf' in mime_type:
                return 'pdf'
            elif 'image' in mime_type:
                return file_ext
        except Exception:
            pass
        
        return file_ext
    
    def process_pdf(self, file_data: bytes, document_type: str) -> Dict[str, Any]:
        """Process PDF document"""
        try:
            # Try text extraction first (for digital PDFs)
            text_data = self.extract_text_from_pdf(file_data)
            
            if text_data['text'].strip() and len(text_data['text']) > 50:
                # Digital PDF with extractable text
                return {
                    'text': text_data['text'],
                    'method': 'text_extraction',
                    'metadata': text_data['metadata'],
                    'issues': []
                }
            else:
                # Scanned PDF - convert to images and use OCR
                return self.process_scanned_pdf(file_data)
                
        except Exception as e:
            logger.error(f"Error processing PDF: {str(e)}")
            return {
                'text': '',
                'method': 'error',
                'metadata': {},
                'issues': [f"PDF processing error: {str(e)}"]
            }
    
    def extract_text_from_pdf(self, file_data: bytes) -> Dict[str, Any]:
        """Extract text from digital PDF"""
        text_content = []
        metadata = {}
        
        try:
            pdf_file = io.BytesIO(file_data)
            pdf_reader = PyPDF2.PdfReader(pdf_file)
            
            # Extract metadata
            if pdf_reader.metadata:
                metadata = {
                    'title': pdf_reader.metadata.get('/Title', ''),
                    'author': pdf_reader.metadata.get('/Author', ''),
                    'creator': pdf_reader.metadata.get('/Creator', ''),
                    'producer': pdf_reader.metadata.get('/Producer', ''),
                    'creation_date': pdf_reader.metadata.get('/CreationDate', ''),
                }
            
            metadata['page_count'] = len(pdf_reader.pages)
            
            # Extract text from all pages
            for page_num, page in enumerate(pdf_reader.pages):
                try:
                    page_text = page.extract_text()
                    if page_text.strip():
                        text_content.append(f"--- Page {page_num + 1} ---\n{page_text}")
                except Exception as e:
                    logger.warning(f"Error extracting text from page {page_num + 1}: {str(e)}")
            
        except Exception as e:
            logger.error(f"Error reading PDF: {str(e)}")
            return {'text': '', 'metadata': {}}
        
        return {
            'text': '\n\n'.join(text_content),
            'metadata': metadata
        }
    
    def process_scanned_pdf(self, file_data: bytes) -> Dict[str, Any]:
        """Process scanned PDF using OCR"""
        try:
            # Convert PDF pages to images
            images = convert_from_bytes(file_data, dpi=300)
            
            all_text = []
            issues = []
            
            for page_num, image in enumerate(images):
                try:
                    # Enhance image for better OCR
                    enhanced_image = self.enhance_image_for_ocr(image)
                    
                    # Perform OCR
                    page_text = pytesseract.image_to_string(enhanced_image, config=self.tesseract_config)
                    
                    if page_text.strip():
                        all_text.append(f"--- Page {page_num + 1} ---\n{page_text}")
                    else:
                        issues.append(f"No text detected on page {page_num + 1}")
                        
                except Exception as e:
                    issues.append(f"OCR error on page {page_num + 1}: {str(e)}")
            
            return {
                'text': '\n\n'.join(all_text),
                'method': 'ocr_pdf',
                'metadata': {'page_count': len(images)},
                'issues': issues
            }
            
        except Exception as e:
            logger.error(f"Error processing scanned PDF: {str(e)}")
            return {
                'text': '',
                'method': 'error',
                'metadata': {},
                'issues': [f"Scanned PDF processing error: {str(e)}"]
            }
    
    def process_image(self, file_data: bytes, document_type: str) -> Dict[str, Any]:
        """Process image document using OCR"""
        try:
            # Load image
            image = Image.open(io.BytesIO(file_data))
            
            # Get image metadata
            metadata = {
                'format': image.format,
                'mode': image.mode,
                'size': image.size,
                'dpi': image.info.get('dpi', (72, 72))
            }
            
            # Enhance image for OCR
            enhanced_image = self.enhance_image_for_ocr(image)
            
            # Perform OCR with detailed output
            ocr_data = pytesseract.image_to_data(enhanced_image, output_type=pytesseract.Output.DICT, config=self.tesseract_config)
            
            # Extract text and confidence scores
            text_parts = []
            confidences = []
            
            for i, conf in enumerate(ocr_data['conf']):
                if int(conf) > 30:  # Only include text with confidence > 30%
                    text = ocr_data['text'][i].strip()
                    if text:
                        text_parts.append(text)
                        confidences.append(int(conf))
            
            full_text = ' '.join(text_parts)
            avg_confidence = np.mean(confidences) if confidences else 0
            
            # Identify issues
            issues = []
            if avg_confidence < 70:
                issues.append("Low OCR confidence - image quality may be poor")
            if len(full_text) < 50:
                issues.append("Very little text detected - document may be blank or illegible")
            
            return {
                'text': full_text,
                'method': 'ocr_image',
                'metadata': {
                    **metadata,
                    'avg_confidence': round(avg_confidence, 1),
                    'total_words': len(text_parts)
                },
                'issues': issues
            }
            
        except Exception as e:
            logger.error(f"Error processing image: {str(e)}")
            return {
                'text': '',
                'method': 'error',
                'metadata': {},
                'issues': [f"Image processing error: {str(e)}"]
            }
    
    def enhance_image_for_ocr(self, image: Image.Image) -> Image.Image:
        """Enhance image quality for better OCR results"""
        try:
            # Convert to grayscale if not already
            if image.mode != 'L':
                image = image.convert('L')
            
            # Increase contrast
            enhancer = ImageEnhance.Contrast(image)
            image = enhancer.enhance(1.5)
            
            # Increase sharpness
            enhancer = ImageEnhance.Sharpness(image)
            image = enhancer.enhance(2.0)
            
            # Apply noise reduction
            image = image.filter(ImageFilter.MedianFilter(size=3))
            
            # Scale up if image is too small (improves OCR accuracy)
            width, height = image.size
            if width < 1000 or height < 1000:
                scale_factor = max(1000 / width, 1000 / height)
                new_size = (int(width * scale_factor), int(height * scale_factor))
                image = image.resize(new_size, Image.Resampling.LANCZOS)
            
            return image
            
        except Exception as e:
            logger.warning(f"Error enhancing image: {str(e)}")
            return image  # Return original if enhancement fails
    
    def parse_document_content(self, text: str, document_type: str) -> Dict[str, Any]:
        """Parse extracted text based on document type"""
        if document_type.lower() == 'invoice':
            return self.parse_invoice_content(text)
        elif document_type.lower() == 'contract':
            return self.parse_contract_content(text)
        elif document_type.lower() == 'identity':
            return self.parse_identity_content(text)
        elif document_type.lower() == 'bank_statement':
            return self.parse_bank_statement_content(text)
        else:
            return self.parse_generic_content(text)
    
    def parse_invoice_content(self, text: str) -> Dict[str, Any]:
        """Parse invoice-specific information"""
        parsed_data = {}
        
        try:
            # Invoice number patterns
            invoice_patterns = [
                r'invoice\s*(?:number|#|no\.?)\s*:?\s*([A-Z0-9\-_]+)',
                r'inv\s*(?:number|#|no\.?)\s*:?\s*([A-Z0-9\-_]+)',
                r'bill\s*(?:number|#|no\.?)\s*:?\s*([A-Z0-9\-_]+)'
            ]
            
            for pattern in invoice_patterns:
                match = re.search(pattern, text, re.IGNORECASE)
                if match:
                    parsed_data['invoice_number'] = match.group(1)
                    break
            
            # Amount patterns
            amount_patterns = [
                r'total\s*:?\s*\$?([0-9,]+\.?[0-9]*)',
                r'amount\s*due\s*:?\s*\$?([0-9,]+\.?[0-9]*)',
                r'balance\s*:?\s*\$?([0-9,]+\.?[0-9]*)',
                r'\$([0-9,]+\.?[0-9]*)'
            ]
            
            amounts = []
            for pattern in amount_patterns:
                matches = re.findall(pattern, text, re.IGNORECASE)
                for match in matches:
                    try:
                        amount = float(match.replace(',', ''))
                        amounts.append(amount)
                    except ValueError:
                        continue
            
            if amounts:
                parsed_data['amount'] = max(amounts)  # Usually the largest amount is the total
            
            # Date patterns
            date_patterns = [
                r'date\s*:?\s*(\d{1,2}[-/]\d{1,2}[-/]\d{2,4})',
                r'invoice\s*date\s*:?\s*(\d{1,2}[-/]\d{1,2}[-/]\d{2,4})',
                r'(\d{1,2}[-/]\d{1,2}[-/]\d{2,4})'
            ]
            
            for pattern in date_patterns:
                match = re.search(pattern, text, re.IGNORECASE)
                if match:
                    parsed_data['date'] = match.group(1)
                    break
            
            # Customer/Company patterns
            company_patterns = [
                r'bill\s*to\s*:?\s*([A-Za-z0-9\s&.,()-]+?)(?:\n|$)',
                r'customer\s*:?\s*([A-Za-z0-9\s&.,()-]+?)(?:\n|$)',
                r'to\s*:?\s*([A-Za-z0-9\s&.,()-]+?)(?:\n|$)'
            ]
            
            for pattern in company_patterns:
                match = re.search(pattern, text, re.IGNORECASE)
                if match:
                    customer_name = match.group(1).strip()
                    if len(customer_name) > 3:  # Avoid capturing single words
                        parsed_data['customer_name'] = customer_name
                        break
            
            # Due date patterns
            due_date_patterns = [
                r'due\s*date\s*:?\s*(\d{1,2}[-/]\d{1,2}[-/]\d{2,4})',
                r'payment\s*due\s*:?\s*(\d{1,2}[-/]\d{1,2}[-/]\d{2,4})'
            ]
            
            for pattern in due_date_patterns:
                match = re.search(pattern, text, re.IGNORECASE)
                if match:
                    parsed_data['due_date'] = match.group(1)
                    break
            
        except Exception as e:
            logger.error(f"Error parsing invoice content: {str(e)}")
        
        return parsed_data
    
    def parse_contract_content(self, text: str) -> Dict[str, Any]:
        """Parse contract-specific information"""
        parsed_data = {}
        
        try:
            # Contract parties
            party_patterns = [
                r'between\s+([^,\n]+)\s+and\s+([^,\n]+)',
                r'party\s*(?:1|one)\s*:?\s*([A-Za-z0-9\s&.,()-]+?)(?:\n|party)',
                r'party\s*(?:2|two)\s*:?\s*([A-Za-z0-9\s&.,()-]+?)(?:\n|$)'
            ]
            
            parties = []
            for pattern in party_patterns:
                matches = re.findall(pattern, text, re.IGNORECASE)
                for match in matches:
                    if isinstance(match, tuple):
                        parties.extend([p.strip() for p in match if p.strip()])
                    else:
                        parties.append(match.strip())
            
            if parties:
                parsed_data['parties'] = parties[:2]  # Usually just two main parties
            
            # Contract value/amount
            value_patterns = [
                r'amount\s*of\s*\$?([0-9,]+\.?[0-9]*)',
                r'value\s*:?\s*\$?([0-9,]+\.?[0-9]*)',
                r'consideration\s*:?\s*\$?([0-9,]+\.?[0-9]*)'
            ]
            
            for pattern in value_patterns:
                match = re.search(pattern, text, re.IGNORECASE)
                if match:
                    try:
                        parsed_data['contract_value'] = float(match.group(1).replace(',', ''))
                        break
                    except ValueError:
                        continue
            
            # Contract dates
            date_patterns = [
                r'effective\s*date\s*:?\s*(\d{1,2}[-/]\d{1,2}[-/]\d{2,4})',
                r'start\s*date\s*:?\s*(\d{1,2}[-/]\d{1,2}[-/]\d{2,4})',
                r'end\s*date\s*:?\s*(\d{1,2}[-/]\d{1,2}[-/]\d{2,4})'
            ]
            
            for pattern in date_patterns:
                match = re.search(pattern, text, re.IGNORECASE)
                if match:
                    field_name = pattern.split('\\')[0].replace('s*', '_')
                    parsed_data[field_name] = match.group(1)
            
        except Exception as e:
            logger.error(f"Error parsing contract content: {str(e)}")
        
        return parsed_data
    
    def parse_identity_content(self, text: str) -> Dict[str, Any]:
        """Parse identity document information"""
        parsed_data = {}
        
        try:
            # Name patterns
            name_patterns = [
                r'name\s*:?\s*([A-Za-z\s]+?)(?:\n|$)',
                r'full\s*name\s*:?\s*([A-Za-z\s]+?)(?:\n|$)'
            ]
            
            for pattern in name_patterns:
                match = re.search(pattern, text, re.IGNORECASE)
                if match:
                    parsed_data['name'] = match.group(1).strip()
                    break
            
            # ID number patterns
            id_patterns = [
                r'id\s*(?:number|no\.?)\s*:?\s*([A-Z0-9]+)',
                r'license\s*(?:number|no\.?)\s*:?\s*([A-Z0-9]+)',
                r'passport\s*(?:number|no\.?)\s*:?\s*([A-Z0-9]+)'
            ]
            
            for pattern in id_patterns:
                match = re.search(pattern, text, re.IGNORECASE)
                if match:
                    parsed_data['id_number'] = match.group(1)
                    break
            
            # Date of birth
            dob_patterns = [
                r'(?:dob|date\s*of\s*birth)\s*:?\s*(\d{1,2}[-/]\d{1,2}[-/]\d{2,4})',
                r'born\s*:?\s*(\d{1,2}[-/]\d{1,2}[-/]\d{2,4})'
            ]
            
            for pattern in dob_patterns:
                match = re.search(pattern, text, re.IGNORECASE)
                if match:
                    parsed_data['date_of_birth'] = match.group(1)
                    break
            
        except Exception as e:
            logger.error(f"Error parsing identity content: {str(e)}")
        
        return parsed_data
    
    def parse_bank_statement_content(self, text: str) -> Dict[str, Any]:
        """Parse bank statement information"""
        parsed_data = {}
        
        try:
            # Account number
            account_patterns = [
                r'account\s*(?:number|no\.?)\s*:?\s*([0-9\-]+)',
                r'acct\s*(?:number|no\.?)\s*:?\s*([0-9\-]+)'
            ]
            
            for pattern in account_patterns:
                match = re.search(pattern, text, re.IGNORECASE)
                if match:
                    parsed_data['account_number'] = match.group(1)
                    break
            
            # Balance patterns
            balance_patterns = [
                r'balance\s*:?\s*\$?([0-9,]+\.?[0-9]*)',
                r'current\s*balance\s*:?\s*\$?([0-9,]+\.?[0-9]*)',
                r'ending\s*balance\s*:?\s*\$?([0-9,]+\.?[0-9]*)'
            ]
            
            for pattern in balance_patterns:
                match = re.search(pattern, text, re.IGNORECASE)
                if match:
                    try:
                        parsed_data['balance'] = float(match.group(1).replace(',', ''))
                        break
                    except ValueError:
                        continue
            
            # Statement period
            period_patterns = [
                r'statement\s*period\s*:?\s*(\d{1,2}[-/]\d{1,2}[-/]\d{2,4})\s*(?:to|-)\s*(\d{1,2}[-/]\d{1,2}[-/]\d{2,4})',
                r'from\s*(\d{1,2}[-/]\d{1,2}[-/]\d{2,4})\s*to\s*(\d{1,2}[-/]\d{1,2}[-/]\d{2,4})'
            ]
            
            for pattern in period_patterns:
                match = re.search(pattern, text, re.IGNORECASE)
                if match:
                    parsed_data['period_start'] = match.group(1)
                    parsed_data['period_end'] = match.group(2)
                    break
            
        except Exception as e:
            logger.error(f"Error parsing bank statement content: {str(e)}")
        
        return parsed_data
    
    def parse_generic_content(self, text: str) -> Dict[str, Any]:
        """Parse generic document content"""
        parsed_data = {}
        
        try:
            # Extract all dates
            dates = re.findall(r'\d{1,2}[-/]\d{1,2}[-/]\d{2,4}', text)
            if dates:
                parsed_data['dates_found'] = dates
            
            # Extract all amounts
            amounts = re.findall(r'\$?([0-9,]+\.?[0-9]*)', text)
            if amounts:
                numeric_amounts = []
                for amount in amounts:
                    try:
                        numeric_amounts.append(float(amount.replace(',', '')))
                    except ValueError:
                        continue
                if numeric_amounts:
                    parsed_data['amounts_found'] = numeric_amounts
            
            # Extract email addresses
            emails = re.findall(r'\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b', text)
            if emails:
                parsed_data['emails_found'] = emails
            
            # Extract phone numbers
            phones = re.findall(r'(?:\+?1[-.\s]?)?\(?([0-9]{3})\)?[-.\s]?([0-9]{3})[-.\s]?([0-9]{4})', text)
            if phones:
                parsed_data['phones_found'] = [f"({p[0]}) {p[1]}-{p[2]}" for p in phones]
            
        except Exception as e:
            logger.error(f"Error parsing generic content: {str(e)}")
        
        return parsed_data
    
    def calculate_confidence(self, extracted_data: Dict[str, Any], parsed_data: Dict[str, Any]) -> float:
        """Calculate overall confidence score"""
        confidence = 0.5  # Base confidence
        
        try:
            # OCR confidence (if available)
            if 'avg_confidence' in extracted_data.get('metadata', {}):
                ocr_conf = extracted_data['metadata']['avg_confidence']
                confidence = (confidence + ocr_conf / 100) / 2
            
            # Text extraction success
            text_length = len(extracted_data.get('text', ''))
            if text_length > 100:
                confidence += 0.1
            if text_length > 500:
                confidence += 0.1
            
            # Parsed fields success
            parsed_fields = len(parsed_data)
            if parsed_fields > 2:
                confidence += 0.1
            if parsed_fields > 5:
                confidence += 0.1
            
            # Issues penalty
            issues_count = len(extracted_data.get('issues', []))
            confidence -= issues_count * 0.05
            
            # Clamp confidence between 0 and 1
            confidence = max(0.1, min(1.0, confidence))
            
        except Exception as e:
            logger.warning(f"Error calculating confidence: {str(e)}")
            confidence = 0.5
        
        return round(confidence, 3)
    
    def generate_document_hash(self, file_data: bytes) -> str:
        """Generate hash for document verification"""
        return hashlib.sha256(file_data).hexdigest()
    
    def verify_document_integrity(self, file_data: bytes, expected_hash: str) -> bool:
        """Verify document integrity using hash"""
        actual_hash = self.generate_document_hash(file_data)
        return actual_hash == expected_hash
