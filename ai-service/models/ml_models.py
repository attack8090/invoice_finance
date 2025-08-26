"""
Advanced Machine Learning Models for Invoice Financing AI Service
"""
import os
import joblib
import numpy as np
import pandas as pd
from typing import Dict, List, Tuple, Any, Optional
from datetime import datetime, timedelta
import logging

# ML Libraries
from sklearn.model_selection import train_test_split, cross_val_score, GridSearchCV
from sklearn.preprocessing import StandardScaler, LabelEncoder, RobustScaler
from sklearn.ensemble import IsolationForest, RandomForestClassifier, RandomForestRegressor
from sklearn.linear_model import LogisticRegression, LinearRegression
from sklearn.metrics import accuracy_score, precision_score, recall_score, f1_score, roc_auc_score
from sklearn.impute import SimpleImputer
import xgboost as xgb
import lightgbm as lgb

logger = logging.getLogger(__name__)


class BaseMLModel:
    """Base class for all ML models"""
    
    def __init__(self, model_name: str, model_dir: str):
        self.model_name = model_name
        self.model_dir = model_dir
        self.model = None
        self.scaler = None
        self.feature_names = []
        self.is_trained = False
        self.last_updated = None
        
        # Ensure model directory exists
        os.makedirs(model_dir, exist_ok=True)
    
    def save_model(self):
        """Save the trained model and scaler"""
        if self.model is not None:
            model_path = os.path.join(self.model_dir, f"{self.model_name}_model.pkl")
            joblib.dump(self.model, model_path)
            logger.info(f"Model saved to {model_path}")
        
        if self.scaler is not None:
            scaler_path = os.path.join(self.model_dir, f"{self.model_name}_scaler.pkl")
            joblib.dump(self.scaler, scaler_path)
            logger.info(f"Scaler saved to {scaler_path}")
        
        # Save metadata
        metadata = {
            'feature_names': self.feature_names,
            'is_trained': self.is_trained,
            'last_updated': datetime.now().isoformat()
        }
        metadata_path = os.path.join(self.model_dir, f"{self.model_name}_metadata.pkl")
        joblib.dump(metadata, metadata_path)
    
    def load_model(self):
        """Load the trained model and scaler"""
        try:
            model_path = os.path.join(self.model_dir, f"{self.model_name}_model.pkl")
            scaler_path = os.path.join(self.model_dir, f"{self.model_name}_scaler.pkl")
            metadata_path = os.path.join(self.model_dir, f"{self.model_name}_metadata.pkl")
            
            if os.path.exists(model_path):
                self.model = joblib.load(model_path)
                logger.info(f"Model loaded from {model_path}")
            
            if os.path.exists(scaler_path):
                self.scaler = joblib.load(scaler_path)
                logger.info(f"Scaler loaded from {scaler_path}")
            
            if os.path.exists(metadata_path):
                metadata = joblib.load(metadata_path)
                self.feature_names = metadata.get('feature_names', [])
                self.is_trained = metadata.get('is_trained', False)
                self.last_updated = metadata.get('last_updated')
                
        except Exception as e:
            logger.error(f"Error loading model: {str(e)}")
            self.is_trained = False
    
    def preprocess_features(self, data: Dict[str, Any]) -> np.ndarray:
        """Preprocess input features"""
        raise NotImplementedError("Subclasses must implement preprocess_features")
    
    def predict(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """Make prediction using the trained model"""
        raise NotImplementedError("Subclasses must implement predict")


class CreditScoringModel(BaseMLModel):
    """Advanced Credit Scoring Model using XGBoost"""
    
    def __init__(self, model_dir: str):
        super().__init__("credit_scoring", model_dir)
        self.load_model()
    
    def preprocess_features(self, data: Dict[str, Any]) -> np.ndarray:
        """Extract and preprocess features for credit scoring"""
        
        # Company data features
        company_data = data.get('company_data', {})
        years_in_business = company_data.get('years_in_business', 0)
        industry = company_data.get('industry', 'unknown')
        company_size = company_data.get('employee_count', 1)
        
        # Financial data features
        financial_data = data.get('financial_data', {})
        annual_revenue = financial_data.get('annual_revenue', 0)
        monthly_cash_flow = financial_data.get('monthly_cash_flow', 0)
        debt_ratio = financial_data.get('debt_ratio', 0.0)
        profit_margin = financial_data.get('profit_margin', 0.0)
        current_ratio = financial_data.get('current_ratio', 1.0)
        
        # Transaction history features
        transaction_history = data.get('transaction_history', {})
        on_time_payments = transaction_history.get('on_time_payments', 0.0)
        avg_payment_delay = transaction_history.get('avg_payment_delay', 0)
        total_transactions = transaction_history.get('total_transactions', 0)
        default_history = transaction_history.get('defaults', 0)
        
        # Derived features
        revenue_per_employee = annual_revenue / max(company_size, 1)
        cash_flow_stability = abs(monthly_cash_flow) / max(annual_revenue / 12, 1)
        payment_reliability = on_time_payments * (1 - default_history / max(total_transactions, 1))
        
        # Industry risk mapping (simplified)
        industry_risk_map = {
            'technology': 0.2,
            'healthcare': 0.1,
            'finance': 0.15,
            'retail': 0.4,
            'manufacturing': 0.25,
            'construction': 0.5,
            'hospitality': 0.6,
            'unknown': 0.3
        }
        industry_risk = industry_risk_map.get(industry.lower(), 0.3)
        
        features = np.array([
            years_in_business,
            np.log1p(annual_revenue),  # Log transform for better distribution
            np.log1p(company_size),
            monthly_cash_flow,
            debt_ratio,
            profit_margin,
            current_ratio,
            on_time_payments,
            avg_payment_delay,
            np.log1p(total_transactions),
            default_history,
            revenue_per_employee,
            cash_flow_stability,
            payment_reliability,
            industry_risk
        ]).reshape(1, -1)
        
        # Store feature names for reference
        self.feature_names = [
            'years_in_business', 'log_annual_revenue', 'log_company_size',
            'monthly_cash_flow', 'debt_ratio', 'profit_margin', 'current_ratio',
            'on_time_payments', 'avg_payment_delay', 'log_total_transactions',
            'default_history', 'revenue_per_employee', 'cash_flow_stability',
            'payment_reliability', 'industry_risk'
        ]
        
        return features
    
    def train_model(self, training_data: List[Dict[str, Any]], credit_scores: List[int]):
        """Train the credit scoring model"""
        logger.info("Training credit scoring model...")
        
        # Preprocess all training data
        features_list = []
        for data in training_data:
            features = self.preprocess_features(data)
            features_list.append(features.flatten())
        
        X = np.array(features_list)
        y = np.array(credit_scores)
        
        # Split data
        X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.2, random_state=42)
        
        # Scale features
        self.scaler = RobustScaler()
        X_train_scaled = self.scaler.fit_transform(X_train)
        X_test_scaled = self.scaler.transform(X_test)
        
        # Train XGBoost model
        self.model = xgb.XGBRegressor(
            n_estimators=100,
            max_depth=6,
            learning_rate=0.1,
            subsample=0.8,
            colsample_bytree=0.8,
            random_state=42
        )
        
        self.model.fit(X_train_scaled, y_train)
        
        # Evaluate model
        train_pred = self.model.predict(X_train_scaled)
        test_pred = self.model.predict(X_test_scaled)
        
        train_rmse = np.sqrt(np.mean((train_pred - y_train) ** 2))
        test_rmse = np.sqrt(np.mean((test_pred - y_test) ** 2))
        
        logger.info(f"Training RMSE: {train_rmse:.2f}")
        logger.info(f"Test RMSE: {test_rmse:.2f}")
        
        self.is_trained = True
        self.last_updated = datetime.now()
        self.save_model()
        
        return {
            'train_rmse': train_rmse,
            'test_rmse': test_rmse,
            'feature_importance': dict(zip(self.feature_names, self.model.feature_importances_))
        }
    
    def predict(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """Predict credit score"""
        if not self.is_trained or self.model is None:
            # Fallback to rule-based scoring
            return self._fallback_credit_scoring(data)
        
        try:
            features = self.preprocess_features(data)
            if self.scaler:
                features = self.scaler.transform(features)
            
            predicted_score = self.model.predict(features)[0]
            predicted_score = max(300, min(850, int(predicted_score)))  # Clamp to valid range
            
            # Calculate risk category
            if predicted_score >= 750:
                risk_category = "low"
            elif predicted_score >= 650:
                risk_category = "medium"
            else:
                risk_category = "high"
            
            # Feature importance (for explainability)
            if hasattr(self.model, 'feature_importances_'):
                feature_importance = dict(zip(self.feature_names, self.model.feature_importances_))
            else:
                feature_importance = {}
            
            return {
                'credit_score': predicted_score,
                'risk_category': risk_category,
                'confidence': 0.85,  # Model confidence
                'feature_importance': feature_importance,
                'model_version': '2.0',
                'calculated_at': datetime.utcnow().isoformat()
            }
            
        except Exception as e:
            logger.error(f"Error in credit scoring prediction: {str(e)}")
            return self._fallback_credit_scoring(data)
    
    def _fallback_credit_scoring(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """Fallback rule-based credit scoring"""
        company_data = data.get('company_data', {})
        financial_data = data.get('financial_data', {})
        transaction_history = data.get('transaction_history', {})
        
        base_score = 600
        
        # Company age factor
        years_factor = min(company_data.get('years_in_business', 0) * 10, 50)
        
        # Revenue factor
        revenue = financial_data.get('annual_revenue', 0)
        revenue_factor = min(revenue / 100000 * 5, 100)
        
        # Payment history factor
        payment_factor = transaction_history.get('on_time_payments', 0.8) * 80
        
        # Debt ratio factor
        debt_ratio = financial_data.get('debt_ratio', 0.3)
        debt_factor = max(0, 50 - debt_ratio * 100)
        
        calculated_score = base_score + years_factor + revenue_factor + payment_factor + debt_factor
        calculated_score = max(300, min(850, int(calculated_score)))
        
        risk_category = "low" if calculated_score > 700 else "medium" if calculated_score > 600 else "high"
        
        return {
            'credit_score': calculated_score,
            'risk_category': risk_category,
            'confidence': 0.65,
            'model_version': 'fallback',
            'calculated_at': datetime.utcnow().isoformat()
        }


class RiskAssessmentModel(BaseMLModel):
    """Advanced Risk Assessment Model using LightGBM"""
    
    def __init__(self, model_dir: str):
        super().__init__("risk_assessment", model_dir)
        self.load_model()
    
    def preprocess_features(self, data: Dict[str, Any]) -> np.ndarray:
        """Extract and preprocess features for risk assessment"""
        
        # Invoice data features
        invoice_data = data.get('invoice_data', {})
        invoice_amount = invoice_data.get('amount', 0)
        days_until_due = invoice_data.get('days_until_due', 30)
        invoice_currency = invoice_data.get('currency', 'USD')
        payment_terms = invoice_data.get('payment_terms', 30)
        
        # Customer data features
        customer_data = data.get('customer_data', {})
        customer_credit_rating = customer_data.get('credit_rating', 3)
        customer_payment_history = customer_data.get('payment_history_score', 0.8)
        customer_industry = customer_data.get('industry', 'unknown')
        customer_size = customer_data.get('company_size', 'medium')
        
        # Historical data features
        historical_data = data.get('historical_data', {})
        similar_invoices_paid = historical_data.get('similar_invoices_paid', 0)
        avg_payment_delay = historical_data.get('avg_payment_delay', 0)
        seasonal_factor = historical_data.get('seasonal_factor', 1.0)
        
        # Market data features (if available)
        market_volatility = data.get('market_volatility', 0.1)
        economic_indicator = data.get('economic_indicator', 1.0)
        
        # Derived features
        amount_risk = min(invoice_amount / 1000000, 1.0)
        time_risk = max(0, (days_until_due - 30) / 365)
        customer_risk = (5 - customer_credit_rating) / 5
        
        # Industry risk mapping
        industry_risk_map = {
            'technology': 0.15,
            'healthcare': 0.1,
            'finance': 0.12,
            'retail': 0.35,
            'manufacturing': 0.2,
            'construction': 0.4,
            'hospitality': 0.5,
            'energy': 0.25,
            'unknown': 0.3
        }
        industry_risk = industry_risk_map.get(customer_industry.lower(), 0.3)
        
        # Company size risk mapping
        size_risk_map = {
            'small': 0.4,
            'medium': 0.25,
            'large': 0.1,
            'enterprise': 0.05
        }
        size_risk = size_risk_map.get(customer_size.lower(), 0.25)
        
        features = np.array([
            np.log1p(invoice_amount),
            days_until_due,
            payment_terms,
            customer_credit_rating,
            customer_payment_history,
            similar_invoices_paid,
            avg_payment_delay,
            seasonal_factor,
            market_volatility,
            economic_indicator,
            amount_risk,
            time_risk,
            customer_risk,
            industry_risk,
            size_risk
        ]).reshape(1, -1)
        
        self.feature_names = [
            'log_invoice_amount', 'days_until_due', 'payment_terms',
            'customer_credit_rating', 'customer_payment_history',
            'similar_invoices_paid', 'avg_payment_delay', 'seasonal_factor',
            'market_volatility', 'economic_indicator', 'amount_risk',
            'time_risk', 'customer_risk', 'industry_risk', 'size_risk'
        ]
        
        return features
    
    def predict(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """Predict risk assessment"""
        if not self.is_trained or self.model is None:
            return self._fallback_risk_assessment(data)
        
        try:
            features = self.preprocess_features(data)
            if self.scaler:
                features = self.scaler.transform(features)
            
            risk_probability = self.model.predict_proba(features)[0]
            risk_score = risk_probability[1] if len(risk_probability) > 1 else risk_probability[0]
            
            # Determine risk level
            if risk_score < 0.3:
                risk_level = "low"
            elif risk_score < 0.6:
                risk_level = "medium"
            else:
                risk_level = "high"
            
            # Calculate recommended interest rate
            base_rate = 5.0
            risk_premium = risk_score * 10
            recommended_rate = base_rate + risk_premium
            
            return {
                'risk_score': round(float(risk_score), 4),
                'risk_level': risk_level,
                'confidence': 0.88,
                'recommended_interest_rate': round(recommended_rate, 2),
                'risk_factors': self._analyze_risk_factors(features[0]),
                'model_version': '2.0',
                'assessed_at': datetime.utcnow().isoformat()
            }
            
        except Exception as e:
            logger.error(f"Error in risk assessment prediction: {str(e)}")
            return self._fallback_risk_assessment(data)
    
    def _analyze_risk_factors(self, features: np.ndarray) -> Dict[str, float]:
        """Analyze individual risk factors"""
        risk_factors = {}
        
        if len(features) >= len(self.feature_names):
            for i, feature_name in enumerate(self.feature_names):
                if 'risk' in feature_name:
                    risk_factors[feature_name] = round(float(features[i]), 3)
        
        return risk_factors
    
    def _fallback_risk_assessment(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """Fallback rule-based risk assessment"""
        invoice_data = data.get('invoice_data', {})
        customer_data = data.get('customer_data', {})
        
        invoice_amount = invoice_data.get('amount', 0)
        due_date_days = invoice_data.get('days_until_due', 30)
        customer_rating = customer_data.get('credit_rating', 3)
        
        # Calculate risk score
        amount_risk = min(invoice_amount / 1000000, 0.4)
        time_risk = max(0, (due_date_days - 30) / 365 * 0.3)
        customer_risk = (5 - customer_rating) / 10
        
        total_risk_score = (amount_risk + time_risk + customer_risk) / 3
        
        if total_risk_score < 0.3:
            risk_level = "low"
        elif total_risk_score < 0.6:
            risk_level = "medium"
        else:
            risk_level = "high"
        
        return {
            'risk_score': round(total_risk_score, 3),
            'risk_level': risk_level,
            'confidence': 0.65,
            'recommended_interest_rate': round(5 + (total_risk_score * 10), 2),
            'risk_factors': {
                'amount_risk': round(amount_risk, 3),
                'time_risk': round(time_risk, 3),
                'customer_risk': round(customer_risk, 3)
            },
            'model_version': 'fallback',
            'assessed_at': datetime.utcnow().isoformat()
        }


class FraudDetectionModel(BaseMLModel):
    """Advanced Fraud Detection Model using Isolation Forest and XGBoost"""
    
    def __init__(self, model_dir: str):
        super().__init__("fraud_detection", model_dir)
        self.isolation_forest = None
        self.classification_model = None
        self.load_model()
    
    def preprocess_features(self, data: Dict[str, Any]) -> np.ndarray:
        """Extract and preprocess features for fraud detection"""
        
        # Invoice data features
        invoice_data = data.get('invoice_data', {})
        invoice_amount = invoice_data.get('amount', 0)
        customer_name = invoice_data.get('customer_name', '')
        invoice_date = invoice_data.get('date', datetime.now().isoformat())
        description_length = len(invoice_data.get('description', ''))
        
        # User data features
        user_data = data.get('user_data', {})
        user_registration_age = user_data.get('account_age_days', 0)
        user_submission_count = user_data.get('total_submissions', 0)
        user_success_rate = user_data.get('success_rate', 0.0)
        
        # Transaction patterns
        transaction_patterns = data.get('transaction_patterns', {})
        submission_hour = transaction_patterns.get('submission_hour', datetime.now().hour)
        submissions_today = transaction_patterns.get('submissions_today', 1)
        avg_amount_deviation = transaction_patterns.get('avg_amount_deviation', 0.0)
        
        # Behavioral features
        ip_changes = transaction_patterns.get('ip_changes_recent', 0)
        location_changes = transaction_patterns.get('location_changes_recent', 0)
        device_fingerprint_changes = transaction_patterns.get('device_changes_recent', 0)
        
        # Derived features
        amount_zscore = self._calculate_z_score(invoice_amount, user_data.get('avg_amount', 10000), 
                                               user_data.get('std_amount', 5000))
        name_entropy = self._calculate_string_entropy(customer_name)
        time_anomaly = 1 if (submission_hour < 6 or submission_hour > 22) else 0
        rapid_submissions = 1 if submissions_today > 5 else 0
        
        # Account behavior anomalies
        new_user_high_amount = 1 if (user_registration_age < 30 and invoice_amount > 50000) else 0
        inconsistent_patterns = 1 if (ip_changes > 2 or location_changes > 1) else 0
        
        features = np.array([
            np.log1p(invoice_amount),
            len(customer_name),
            description_length,
            user_registration_age,
            user_submission_count,
            user_success_rate,
            submission_hour,
            submissions_today,
            avg_amount_deviation,
            ip_changes,
            location_changes,
            device_fingerprint_changes,
            amount_zscore,
            name_entropy,
            time_anomaly,
            rapid_submissions,
            new_user_high_amount,
            inconsistent_patterns
        ]).reshape(1, -1)
        
        self.feature_names = [
            'log_invoice_amount', 'customer_name_length', 'description_length',
            'user_registration_age', 'user_submission_count', 'user_success_rate',
            'submission_hour', 'submissions_today', 'avg_amount_deviation',
            'ip_changes', 'location_changes', 'device_fingerprint_changes',
            'amount_zscore', 'name_entropy', 'time_anomaly', 'rapid_submissions',
            'new_user_high_amount', 'inconsistent_patterns'
        ]
        
        return features
    
    def predict(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """Predict fraud probability"""
        if not self.is_trained:
            return self._fallback_fraud_detection(data)
        
        try:
            features = self.preprocess_features(data)
            if self.scaler:
                features = self.scaler.transform(features)
            
            # Use isolation forest for anomaly detection
            anomaly_score = -1
            if self.isolation_forest:
                anomaly_score = self.isolation_forest.decision_function(features)[0]
                anomaly_prediction = self.isolation_forest.predict(features)[0]
            else:
                anomaly_prediction = 1
            
            # Use classification model for fraud probability
            fraud_probability = 0.1
            if self.classification_model:
                fraud_probability = self.classification_model.predict_proba(features)[0][1]
            
            # Combine anomaly detection and classification
            combined_score = (fraud_probability + (1 if anomaly_prediction == -1 else 0)) / 2
            
            is_fraud = combined_score > 0.7
            fraud_indicators = self._identify_fraud_indicators(features[0], data)
            
            # Determine recommendation
            if combined_score > 0.8:
                recommendation = "reject"
            elif combined_score > 0.4:
                recommendation = "review"
            else:
                recommendation = "approve"
            
            return {
                'is_fraud': is_fraud,
                'fraud_score': round(float(combined_score), 4),
                'confidence': 0.87,
                'anomaly_score': round(float(anomaly_score), 4),
                'fraud_indicators': fraud_indicators,
                'recommendation': recommendation,
                'model_version': '2.0',
                'detected_at': datetime.utcnow().isoformat()
            }
            
        except Exception as e:
            logger.error(f"Error in fraud detection prediction: {str(e)}")
            return self._fallback_fraud_detection(data)
    
    def _calculate_z_score(self, value: float, mean: float, std: float) -> float:
        """Calculate z-score for anomaly detection"""
        if std == 0:
            return 0
        return (value - mean) / std
    
    def _calculate_string_entropy(self, s: str) -> float:
        """Calculate entropy of a string"""
        if not s:
            return 0
        
        char_counts = {}
        for char in s:
            char_counts[char] = char_counts.get(char, 0) + 1
        
        entropy = 0
        for count in char_counts.values():
            prob = count / len(s)
            entropy -= prob * np.log2(prob)
        
        return entropy
    
    def _identify_fraud_indicators(self, features: np.ndarray, original_data: Dict[str, Any]) -> List[str]:
        """Identify specific fraud indicators"""
        indicators = []
        
        if len(features) >= len(self.feature_names):
            # Check for specific patterns
            if features[14] == 1:  # time_anomaly
                indicators.append("unusual_submission_time")
            
            if features[15] == 1:  # rapid_submissions
                indicators.append("rapid_multiple_submissions")
            
            if features[16] == 1:  # new_user_high_amount
                indicators.append("new_user_large_amount")
            
            if features[17] == 1:  # inconsistent_patterns
                indicators.append("inconsistent_user_behavior")
            
            if features[12] > 3:  # amount_zscore
                indicators.append("unusual_amount_pattern")
            
            # Check customer name patterns
            customer_name = original_data.get('invoice_data', {}).get('customer_name', '')
            if len(customer_name) < 3:
                indicators.append("suspicious_customer_name")
        
        return indicators
    
    def _fallback_fraud_detection(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """Fallback rule-based fraud detection"""
        fraud_indicators = []
        fraud_score = 0.0
        
        invoice_data = data.get('invoice_data', {})
        user_data = data.get('user_data', {})
        transaction_patterns = data.get('transaction_patterns', {})
        
        # Check various fraud indicators
        invoice_amount = invoice_data.get('amount', 0)
        if invoice_amount > 500000:  # Very high amounts
            fraud_indicators.append("extremely_high_amount")
            fraud_score += 0.4
        elif invoice_amount > 100000:
            fraud_indicators.append("unusually_high_amount")
            fraud_score += 0.2
        
        customer_name = invoice_data.get('customer_name', '')
        if len(customer_name) < 3:
            fraud_indicators.append("suspicious_customer_name")
            fraud_score += 0.3
        
        # Check submission patterns
        submission_hour = transaction_patterns.get('submission_hour', datetime.now().hour)
        if submission_hour < 6 or submission_hour > 22:
            fraud_indicators.append("unusual_submission_time")
            fraud_score += 0.1
        
        submissions_today = transaction_patterns.get('submissions_today', 1)
        if submissions_today > 5:
            fraud_indicators.append("rapid_multiple_submissions")
            fraud_score += 0.2
        
        # User behavior checks
        account_age = user_data.get('account_age_days', 100)
        if account_age < 7 and invoice_amount > 50000:
            fraud_indicators.append("new_user_large_amount")
            fraud_score += 0.3
        
        # Add some randomness for simulation
        import random
        fraud_score += random.uniform(0, 0.1)
        fraud_score = min(fraud_score, 1.0)
        
        is_fraud = fraud_score > 0.7
        
        if fraud_score > 0.8:
            recommendation = "reject"
        elif fraud_score > 0.4:
            recommendation = "review"
        else:
            recommendation = "approve"
        
        return {
            'is_fraud': is_fraud,
            'fraud_score': round(fraud_score, 3),
            'confidence': 0.65,
            'fraud_indicators': fraud_indicators,
            'recommendation': recommendation,
            'model_version': 'fallback',
            'detected_at': datetime.utcnow().isoformat()
        }
