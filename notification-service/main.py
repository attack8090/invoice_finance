"""
Notification Service
Manages user communications through email, SMS, push notifications for registration,
approval, payment reminders with UI Agent and Merchant Network integration, multi-language support
"""

import os
import asyncio
import logging
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Any, Union
from enum import Enum
import json

import uvicorn
from fastapi import FastAPI, HTTPException, Depends, status, BackgroundTasks
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from pydantic import BaseModel, validator, Field, EmailStr
from sqlalchemy.ext.asyncio import AsyncSession
import aioredis
from jinja2 import Environment, FileSystemLoader

# Internal imports
from config import settings
from database import get_db, init_db
from models.notification import Notification, NotificationStatus, NotificationType, NotificationChannel
from models.template import NotificationTemplate, TemplateType
from models.user_preference import UserNotificationPreference
from services.auth_service import AuthService
from services.email_service import EmailService
from services.sms_service import SMSService
from services.push_service import PushNotificationService
from services.ui_agent_service import UIAgentService
from services.merchant_network_service import MerchantNetworkService
from services.translation_service import TranslationService
from services.template_service import TemplateService
from middleware.security import SecurityMiddleware
from middleware.rate_limit import RateLimitMiddleware
from utils.logger import setup_logger

# Setup logging
logger = setup_logger(__name__)

# Initialize FastAPI app
app = FastAPI(
    title="Notification Service",
    version="1.0.0",
    description="Comprehensive notification service with multi-channel delivery and multi-language support",
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
email_service = EmailService()
sms_service = SMSService()
push_service = PushNotificationService()
ui_agent_service = UIAgentService()
merchant_network_service = MerchantNetworkService()
translation_service = TranslationService()
template_service = TemplateService()

# Redis for caching and rate limiting
redis_client = None

class NotificationPriority(str, Enum):
    LOW = "low"
    NORMAL = "normal"
    HIGH = "high"
    URGENT = "urgent"

class DeliveryMethod(str, Enum):
    EMAIL = "email"
    SMS = "sms"
    PUSH = "push"
    UI_AGENT = "ui_agent"
    MERCHANT_NETWORK = "merchant_network"
    IN_APP = "in_app"
    WEBHOOK = "webhook"

class NotificationEvent(str, Enum):
    USER_REGISTRATION = "user_registration"
    EMAIL_VERIFICATION = "email_verification"
    KYC_APPROVAL = "kyc_approval"
    KYC_REJECTION = "kyc_rejection"
    FINANCING_REQUEST_SUBMITTED = "financing_request_submitted"
    FINANCING_APPROVED = "financing_approved"
    FINANCING_REJECTED = "financing_rejected"
    PAYMENT_RECEIVED = "payment_received"
    PAYMENT_FAILED = "payment_failed"
    INVOICE_DUE_REMINDER = "invoice_due_reminder"
    INVOICE_OVERDUE = "invoice_overdue"
    DISPUTE_CREATED = "dispute_created"
    DISPUTE_RESOLVED = "dispute_resolved"
    SYSTEM_MAINTENANCE = "system_maintenance"
    SECURITY_ALERT = "security_alert"

# Request/Response models
class NotificationRequest(BaseModel):
    recipient_id: str
    notification_type: NotificationEvent
    channels: List[DeliveryMethod] = [DeliveryMethod.EMAIL]
    priority: NotificationPriority = NotificationPriority.NORMAL
    language: str = "en"
    data: Dict[str, Any] = {}
    scheduled_at: Optional[datetime] = None
    expires_at: Optional[datetime] = None
    custom_template: Optional[str] = None

class BulkNotificationRequest(BaseModel):
    recipient_ids: List[str]
    notification_type: NotificationEvent
    channels: List[DeliveryMethod] = [DeliveryMethod.EMAIL]
    priority: NotificationPriority = NotificationPriority.NORMAL
    language: str = "en"
    data: Dict[str, Any] = {}
    scheduled_at: Optional[datetime] = None
    batch_size: int = Field(default=100, ge=1, le=1000)

class TemplateRequest(BaseModel):
    template_type: TemplateType
    language: str = "en"
    channel: DeliveryMethod
    subject: Optional[str] = None
    content: str
    variables: List[str] = []
    is_active: bool = True

class UserPreferenceRequest(BaseModel):
    user_id: str
    email_enabled: bool = True
    sms_enabled: bool = True
    push_enabled: bool = True
    language: str = "en"
    timezone: str = "UTC"
    quiet_hours_start: Optional[str] = None  # HH:MM format
    quiet_hours_end: Optional[str] = None    # HH:MM format
    frequency_settings: Dict[str, str] = {}  # notification_type -> frequency

class BroadcastRequest(BaseModel):
    title: str
    message: str
    target_audience: Dict[str, Any] = {}  # filtering criteria
    channels: List[DeliveryMethod] = [DeliveryMethod.EMAIL, DeliveryMethod.PUSH]
    priority: NotificationPriority = NotificationPriority.NORMAL
    scheduled_at: Optional[datetime] = None
    languages: List[str] = ["en"]

@app.on_event("startup")
async def startup_event():
    """Initialize services on startup"""
    global redis_client
    
    await init_db()
    
    # Initialize Redis
    redis_client = aioredis.from_url(settings.redis_url)
    
    # Initialize services
    await email_service.initialize()
    await sms_service.initialize()
    await push_service.initialize()
    await ui_agent_service.initialize()
    await merchant_network_service.initialize()
    await translation_service.initialize()
    await template_service.initialize()
    
    logger.info("Notification Service started successfully")

@app.on_event("shutdown")
async def shutdown_event():
    """Cleanup on shutdown"""
    global redis_client
    
    if redis_client:
        await redis_client.close()
    
    await email_service.cleanup()
    await sms_service.cleanup()
    await push_service.cleanup()
    await ui_agent_service.cleanup()
    await merchant_network_service.cleanup()
    
    logger.info("Notification Service shut down")

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
    service_status = {
        "email": await email_service.check_health(),
        "sms": await sms_service.check_health(),
        "push": await push_service.check_health(),
        "ui_agent": await ui_agent_service.check_health(),
        "merchant_network": await merchant_network_service.check_health()
    }
    
    return {
        "service": "Notification Service",
        "version": "1.0.0",
        "status": "healthy",
        "features": {
            "multi_channel": True,
            "multi_language": True,
            "template_engine": True,
            "scheduled_delivery": True,
            "ui_agent_integration": True,
            "merchant_network": True,
            "batch_processing": True,
            "user_preferences": True
        },
        "supported_channels": [channel.value for channel in DeliveryMethod],
        "supported_languages": await translation_service.get_supported_languages(),
        "service_health": service_status
    }

@app.post("/api/v1/notifications/send")
async def send_notification(
    background_tasks: BackgroundTasks,
    request: NotificationRequest,
    current_user = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """Send single notification"""
    try:
        # Validate recipient access
        if not await validate_recipient_access(current_user, request.recipient_id):
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail="Access denied to send notification to recipient"
            )
        
        # Get recipient details and preferences
        recipient = await get_user_details(request.recipient_id, db)
        if not recipient:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail="Recipient not found"
            )
        
        # Get user notification preferences
        preferences = await get_user_preferences(request.recipient_id, db)
        
        # Filter channels based on user preferences
        allowed_channels = filter_channels_by_preferences(request.channels, preferences)
        if not allowed_channels:
            logger.warning(f"No allowed channels for user {request.recipient_id}")
            return {"status": "skipped", "reason": "No allowed channels based on user preferences"}
        
        # Create notification record
        notification = Notification(
            recipient_id=request.recipient_id,
            notification_type=request.notification_type.value,
            channels=allowed_channels,
            priority=request.priority.value,
            language=request.language,
            data=request.data,
            status=NotificationStatus.PENDING,
            scheduled_at=request.scheduled_at,
            expires_at=request.expires_at,
            created_by=current_user['id']
        )
        
        db.add(notification)
        await db.commit()
        await db.refresh(notification)
        
        # Schedule delivery
        if request.scheduled_at and request.scheduled_at > datetime.utcnow():
            await schedule_notification_delivery(notification.id, request.scheduled_at)
        else:
            background_tasks.add_task(
                process_notification_delivery, notification.id, db
            )
        
        logger.info(f"Notification created: {notification.id} for user {request.recipient_id}")
        
        return {
            "notification_id": str(notification.id),
            "status": "queued" if not request.scheduled_at else "scheduled",
            "channels": allowed_channels,
            "estimated_delivery": datetime.utcnow() + timedelta(minutes=5),
            "scheduled_at": request.scheduled_at.isoformat() if request.scheduled_at else None
        }
        
    except Exception as e:
        logger.error(f"Notification sending failed: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Notification sending failed: {str(e)}"
        )

@app.post("/api/v1/notifications/bulk-send")
async def bulk_send_notifications(
    background_tasks: BackgroundTasks,
    request: BulkNotificationRequest,
    current_user = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """Send bulk notifications"""
    try:
        # Validate bulk sending permissions
        if current_user['role'] not in ['admin', 'marketing', 'system']:
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail="Insufficient permissions for bulk notifications"
            )
        
        if len(request.recipient_ids) > settings.max_bulk_notification_count:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail=f"Too many recipients. Maximum {settings.max_bulk_notification_count} allowed"
            )
        
        # Create bulk job record
        bulk_job = await create_bulk_notification_job(
            request, current_user['id'], db
        )
        
        # Schedule background processing
        background_tasks.add_task(
            process_bulk_notifications, bulk_job.id, request, db
        )
        
        logger.info(f"Bulk notification job created: {bulk_job.id} for {len(request.recipient_ids)} recipients")
        
        return {
            "job_id": str(bulk_job.id),
            "recipient_count": len(request.recipient_ids),
            "status": "queued",
            "batch_size": request.batch_size,
            "estimated_completion": datetime.utcnow() + timedelta(
                minutes=len(request.recipient_ids) // request.batch_size * 5
            )
        }
        
    except Exception as e:
        logger.error(f"Bulk notification failed: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Bulk notification failed: {str(e)}"
        )

@app.post("/api/v1/templates/create")
async def create_template(
    request: TemplateRequest,
    current_user = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """Create notification template"""
    try:
        # Validate template creation permissions
        if current_user['role'] not in ['admin', 'marketing', 'template_editor']:
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail="Insufficient permissions to create templates"
            )
        
        # Validate template content
        validation_result = await template_service.validate_template(
            request.content, request.variables, request.channel
        )
        if not validation_result.valid:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail=f"Template validation failed: {validation_result.errors}"
            )
        
        # Create template record
        template = NotificationTemplate(
            template_type=request.template_type.value,
            language=request.language,
            channel=request.channel.value,
            subject=request.subject,
            content=request.content,
            variables=request.variables,
            is_active=request.is_active,
            created_by=current_user['id']
        )
        
        db.add(template)
        await db.commit()
        await db.refresh(template)
        
        # Cache template for faster access
        await cache_template(template)
        
        logger.info(f"Template created: {template.id} - {request.template_type}")
        
        return {
            "template_id": str(template.id),
            "template_type": template.template_type,
            "language": template.language,
            "channel": template.channel,
            "variables": template.variables,
            "is_active": template.is_active,
            "created_at": template.created_at.isoformat()
        }
        
    except Exception as e:
        logger.error(f"Template creation failed: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Template creation failed: {str(e)}"
        )

@app.put("/api/v1/users/{user_id}/preferences")
async def update_user_preferences(
    user_id: str,
    request: UserPreferenceRequest,
    current_user = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """Update user notification preferences"""
    try:
        # Validate access permissions
        if current_user['id'] != user_id and current_user['role'] not in ['admin']:
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail="Access denied to update user preferences"
            )
        
        # Get or create user preferences
        preferences = await get_user_preferences(user_id, db)
        if not preferences:
            preferences = UserNotificationPreference(user_id=user_id)
            db.add(preferences)
        
        # Update preferences
        preferences.email_enabled = request.email_enabled
        preferences.sms_enabled = request.sms_enabled
        preferences.push_enabled = request.push_enabled
        preferences.language = request.language
        preferences.timezone = request.timezone
        preferences.quiet_hours_start = request.quiet_hours_start
        preferences.quiet_hours_end = request.quiet_hours_end
        preferences.frequency_settings = request.frequency_settings
        preferences.updated_at = datetime.utcnow()
        
        await db.commit()
        
        # Cache updated preferences
        await cache_user_preferences(preferences)
        
        logger.info(f"User preferences updated: {user_id}")
        
        return {
            "user_id": user_id,
            "email_enabled": preferences.email_enabled,
            "sms_enabled": preferences.sms_enabled,
            "push_enabled": preferences.push_enabled,
            "language": preferences.language,
            "timezone": preferences.timezone,
            "quiet_hours": {
                "start": preferences.quiet_hours_start,
                "end": preferences.quiet_hours_end
            },
            "frequency_settings": preferences.frequency_settings,
            "updated_at": preferences.updated_at.isoformat()
        }
        
    except Exception as e:
        logger.error(f"User preferences update failed: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"User preferences update failed: {str(e)}"
        )

@app.post("/api/v1/broadcast/send")
async def send_broadcast(
    background_tasks: BackgroundTasks,
    request: BroadcastRequest,
    current_user = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """Send broadcast notification to target audience"""
    try:
        # Validate broadcast permissions
        if current_user['role'] not in ['admin', 'marketing', 'communications']:
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail="Insufficient permissions for broadcast notifications"
            )
        
        # Create broadcast job
        broadcast_job = await create_broadcast_job(
            request, current_user['id'], db
        )
        
        # Schedule background processing
        background_tasks.add_task(
            process_broadcast, broadcast_job.id, request, db
        )
        
        logger.info(f"Broadcast job created: {broadcast_job.id}")
        
        return {
            "broadcast_id": str(broadcast_job.id),
            "title": request.title,
            "channels": request.channels,
            "priority": request.priority.value,
            "status": "queued",
            "scheduled_at": request.scheduled_at.isoformat() if request.scheduled_at else None,
            "languages": request.languages
        }
        
    except Exception as e:
        logger.error(f"Broadcast failed: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Broadcast failed: {str(e)}"
        )

@app.get("/api/v1/notifications/{notification_id}/status")
async def get_notification_status(
    notification_id: str,
    current_user = Depends(get_current_user),
    db: AsyncSession = Depends(get_db)
):
    """Get notification delivery status"""
    try:
        notification = await db.get(Notification, notification_id)
        if not notification:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail="Notification not found"
            )
        
        # Validate access permissions
        if (notification.recipient_id != current_user['id'] and 
            current_user['id'] != notification.created_by and
            current_user['role'] not in ['admin']):
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail="Access denied to notification status"
            )
        
        # Get delivery status for each channel
        delivery_status = await get_delivery_status_details(notification_id, db)
        
        return {
            "notification_id": str(notification.id),
            "recipient_id": notification.recipient_id,
            "notification_type": notification.notification_type,
            "status": notification.status.value,
            "channels": notification.channels,
            "priority": notification.priority,
            "language": notification.language,
            "created_at": notification.created_at.isoformat(),
            "sent_at": notification.sent_at.isoformat() if notification.sent_at else None,
            "delivery_status": delivery_status,
            "failure_reason": notification.failure_reason
        }
        
    except Exception as e:
        logger.error(f"Failed to get notification status: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to retrieve notification status"
        )

@app.post("/api/v1/webhooks/delivery-status")
async def delivery_status_webhook(
    background_tasks: BackgroundTasks,
    webhook_data: Dict[str, Any],
    db: AsyncSession = Depends(get_db)
):
    """Handle delivery status webhooks from external services"""
    try:
        # Validate webhook signature (implementation depends on service)
        if not await validate_webhook_signature(webhook_data):
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="Invalid webhook signature"
            )
        
        # Process webhook
        background_tasks.add_task(
            process_delivery_webhook, webhook_data, db
        )
        
        return {"status": "accepted"}
        
    except Exception as e:
        logger.error(f"Webhook processing failed: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Webhook processing failed"
        )

# Background task functions
async def process_notification_delivery(notification_id: str, db: AsyncSession):
    """Background task to process notification delivery"""
    try:
        notification = await db.get(Notification, notification_id)
        if not notification:
            return
        
        # Update status to processing
        notification.status = NotificationStatus.PROCESSING
        notification.sent_at = datetime.utcnow()
        await db.commit()
        
        # Get recipient details
        recipient = await get_user_details(notification.recipient_id, db)
        
        # Process each channel
        delivery_results = {}
        
        for channel in notification.channels:
            try:
                if channel == "email" and recipient.email:
                    result = await email_service.send_email(
                        notification, recipient.email, recipient
                    )
                elif channel == "sms" and recipient.phone:
                    result = await sms_service.send_sms(
                        notification, recipient.phone, recipient
                    )
                elif channel == "push":
                    result = await push_service.send_push_notification(
                        notification, recipient
                    )
                elif channel == "ui_agent":
                    result = await ui_agent_service.send_notification(
                        notification, recipient
                    )
                elif channel == "merchant_network":
                    result = await merchant_network_service.send_notification(
                        notification, recipient
                    )
                else:
                    result = {"success": False, "error": "Unsupported channel"}
                
                delivery_results[channel] = result
                
            except Exception as e:
                logger.error(f"Channel {channel} delivery failed: {str(e)}")
                delivery_results[channel] = {"success": False, "error": str(e)}
        
        # Update notification status based on results
        successful_channels = [ch for ch, result in delivery_results.items() if result.get('success')]
        
        if successful_channels:
            notification.status = NotificationStatus.SENT
        else:
            notification.status = NotificationStatus.FAILED
            notification.failure_reason = "All channels failed"
        
        notification.delivery_results = delivery_results
        notification.delivered_at = datetime.utcnow()
        
        await db.commit()
        
        logger.info(f"Notification processed: {notification_id} - {notification.status}")
        
    except Exception as e:
        logger.error(f"Notification processing failed: {notification_id} - {str(e)}")

# Helper functions
async def validate_recipient_access(user: dict, recipient_id: str) -> bool:
    """Validate if user can send notifications to recipient"""
    if user['role'] in ['admin', 'system']:
        return True
    # Add additional logic based on business rules
    return True

async def get_user_details(user_id: str, db: AsyncSession):
    """Get user details for notification delivery"""
    # Implementation would fetch from user service or database
    return {
        "id": user_id,
        "email": f"user{user_id}@example.com",
        "phone": "+1234567890",
        "language": "en",
        "timezone": "UTC"
    }

async def get_user_preferences(user_id: str, db: AsyncSession):
    """Get user notification preferences"""
    # Implementation would fetch from database with caching
    return UserNotificationPreference(
        user_id=user_id,
        email_enabled=True,
        sms_enabled=True,
        push_enabled=True,
        language="en"
    )

def filter_channels_by_preferences(channels: List[str], preferences) -> List[str]:
    """Filter channels based on user preferences"""
    if not preferences:
        return channels
    
    allowed_channels = []
    for channel in channels:
        if channel == "email" and preferences.email_enabled:
            allowed_channels.append(channel)
        elif channel == "sms" and preferences.sms_enabled:
            allowed_channels.append(channel)
        elif channel == "push" and preferences.push_enabled:
            allowed_channels.append(channel)
        elif channel not in ["email", "sms", "push"]:
            allowed_channels.append(channel)  # Always allow non-preference channels
    
    return allowed_channels

if __name__ == "__main__":
    uvicorn.run(
        "main:app",
        host=settings.host,
        port=settings.port,
        reload=settings.debug,
        workers=1 if settings.debug else settings.workers
    )
