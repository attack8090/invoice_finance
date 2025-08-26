"""
Configuration management for AI Service
"""
import os
from typing import List, Optional
from decouple import config
from pydantic import BaseSettings


class Settings(BaseSettings):
    # Application Settings
    app_name: str = config("APP_NAME", default="Invoice Financing AI Service")
    app_version: str = config("APP_VERSION", default="2.0.0")
    debug: bool = config("DEBUG", default=False, cast=bool)
    environment: str = config("ENVIRONMENT", default="development")
    
    # Server Configuration
    host: str = config("HOST", default="0.0.0.0")
    port: int = config("PORT", default=5000, cast=int)
    workers: int = config("WORKERS", default=1, cast=int)
    
    # Security
    secret_key: str = config("SECRET_KEY", default="your-secret-key-change-in-production")
    api_key: Optional[str] = config("API_KEY", default=None)
    allowed_hosts: List[str] = config("ALLOWED_HOSTS", default="*").split(",")
    
    # External API Keys
    openai_api_key: Optional[str] = config("OPENAI_API_KEY", default=None)
    anthropic_api_key: Optional[str] = config("ANTHROPIC_API_KEY", default=None)
    google_cloud_api_key: Optional[str] = config("GOOGLE_CLOUD_API_KEY", default=None)
    
    # Database Configuration
    database_url: Optional[str] = config("DATABASE_URL", default=None)
    redis_url: str = config("REDIS_URL", default="redis://localhost:6379")
    mongodb_url: Optional[str] = config("MONGODB_URL", default=None)
    
    # Model Configuration
    model_cache_dir: str = config("MODEL_CACHE_DIR", default="./models")
    model_update_interval: int = config("MODEL_UPDATE_INTERVAL", default=3600, cast=int)  # 1 hour
    enable_model_training: bool = config("ENABLE_MODEL_TRAINING", default=False, cast=bool)
    
    # OCR Configuration
    tesseract_path: Optional[str] = config("TESSERACT_PATH", default=None)
    tesseract_config: str = config("TESSERACT_CONFIG", default="--oem 3 --psm 6")
    
    # File Processing
    max_file_size: int = config("MAX_FILE_SIZE", default=10485760, cast=int)  # 10MB
    allowed_file_types: List[str] = config("ALLOWED_FILE_TYPES", default="pdf,png,jpg,jpeg").split(",")
    upload_path: str = config("UPLOAD_PATH", default="./uploads")
    
    # Rate Limiting
    rate_limit_requests: int = config("RATE_LIMIT_REQUESTS", default=100, cast=int)
    rate_limit_window: int = config("RATE_LIMIT_WINDOW", default=3600, cast=int)  # 1 hour
    
    # Monitoring & Logging
    log_level: str = config("LOG_LEVEL", default="INFO")
    enable_metrics: bool = config("ENABLE_METRICS", default=True, cast=bool)
    sentry_dsn: Optional[str] = config("SENTRY_DSN", default=None)
    
    # Feature Flags
    enable_credit_scoring: bool = config("ENABLE_CREDIT_SCORING", default=True, cast=bool)
    enable_risk_assessment: bool = config("ENABLE_RISK_ASSESSMENT", default=True, cast=bool)
    enable_fraud_detection: bool = config("ENABLE_FRAUD_DETECTION", default=True, cast=bool)
    enable_document_verification: bool = config("ENABLE_DOCUMENT_VERIFICATION", default=True, cast=bool)
    enable_market_analysis: bool = config("ENABLE_MARKET_ANALYSIS", default=True, cast=bool)
    enable_predictive_analytics: bool = config("ENABLE_PREDICTIVE_ANALYTICS", default=True, cast=bool)
    
    # AI Model Settings
    credit_score_model_type: str = config("CREDIT_SCORE_MODEL_TYPE", default="xgboost")
    risk_assessment_model_type: str = config("RISK_ASSESSMENT_MODEL_TYPE", default="lightgbm")
    fraud_detection_model_type: str = config("FRAUD_DETECTION_MODEL_TYPE", default="isolation_forest")
    
    # Threshold Configuration
    fraud_detection_threshold: float = config("FRAUD_DETECTION_THRESHOLD", default=0.7, cast=float)
    risk_assessment_high_threshold: float = config("RISK_HIGH_THRESHOLD", default=0.6, cast=float)
    risk_assessment_medium_threshold: float = config("RISK_MEDIUM_THRESHOLD", default=0.3, cast=float)
    document_verification_threshold: float = config("DOCUMENT_VERIFICATION_THRESHOLD", default=0.8, cast=float)
    
    # Performance Settings
    model_inference_timeout: int = config("MODEL_INFERENCE_TIMEOUT", default=30, cast=int)
    batch_processing_size: int = config("BATCH_PROCESSING_SIZE", default=10, cast=int)
    enable_parallel_processing: bool = config("ENABLE_PARALLEL_PROCESSING", default=True, cast=bool)
    
    class Config:
        env_file = ".env"
        case_sensitive = False


# Global settings instance
settings = Settings()


# Feature check functions
def is_feature_enabled(feature: str) -> bool:
    """Check if a specific feature is enabled"""
    feature_map = {
        "credit_scoring": settings.enable_credit_scoring,
        "risk_assessment": settings.enable_risk_assessment,
        "fraud_detection": settings.enable_fraud_detection,
        "document_verification": settings.enable_document_verification,
        "market_analysis": settings.enable_market_analysis,
        "predictive_analytics": settings.enable_predictive_analytics,
    }
    return feature_map.get(feature, False)


def get_model_config(model_type: str) -> dict:
    """Get configuration for specific model types"""
    config_map = {
        "credit_scoring": {
            "model_type": settings.credit_score_model_type,
            "cache_dir": os.path.join(settings.model_cache_dir, "credit_scoring"),
            "timeout": settings.model_inference_timeout,
        },
        "risk_assessment": {
            "model_type": settings.risk_assessment_model_type,
            "cache_dir": os.path.join(settings.model_cache_dir, "risk_assessment"),
            "timeout": settings.model_inference_timeout,
            "high_threshold": settings.risk_assessment_high_threshold,
            "medium_threshold": settings.risk_assessment_medium_threshold,
        },
        "fraud_detection": {
            "model_type": settings.fraud_detection_model_type,
            "cache_dir": os.path.join(settings.model_cache_dir, "fraud_detection"),
            "timeout": settings.model_inference_timeout,
            "threshold": settings.fraud_detection_threshold,
        },
        "document_verification": {
            "cache_dir": os.path.join(settings.model_cache_dir, "document_verification"),
            "timeout": settings.model_inference_timeout,
            "threshold": settings.document_verification_threshold,
            "tesseract_path": settings.tesseract_path,
            "tesseract_config": settings.tesseract_config,
        },
    }
    return config_map.get(model_type, {})
