// Frontend validation utilities with security considerations

export interface ValidationRule {
  required?: boolean;
  minLength?: number;
  maxLength?: number;
  pattern?: RegExp;
  custom?: (value: any) => boolean | string;
}

export interface ValidationRules {
  [key: string]: ValidationRule;
}

export interface ValidationResult {
  isValid: boolean;
  errors: { [key: string]: string };
}

// Sanitize input to prevent XSS
export const sanitizeInput = (input: string): string => {
  if (typeof input !== 'string') return '';
  
  return input
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#x27;')
    .replace(/\//g, '&#x2F;')
    .trim();
};

// Sanitize object recursively
export const sanitizeObject = (obj: any): any => {
  if (typeof obj === 'string') {
    return sanitizeInput(obj);
  }
  
  if (Array.isArray(obj)) {
    return obj.map(sanitizeObject);
  }
  
  if (obj && typeof obj === 'object') {
    const sanitized: any = {};
    for (const [key, value] of Object.entries(obj)) {
      sanitized[key] = sanitizeObject(value);
    }
    return sanitized;
  }
  
  return obj;
};

// Email validation
export const isValidEmail = (email: string): boolean => {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email) && email.length <= 255;
};

// Password validation
export const isValidPassword = (password: string): boolean => {
  // At least 8 characters, with uppercase, lowercase, digit, and special character
  const hasMinLength = password.length >= 8 && password.length <= 128;
  const hasUpperCase = /[A-Z]/.test(password);
  const hasLowerCase = /[a-z]/.test(password);
  const hasDigit = /[0-9]/.test(password);
  const hasSpecialChar = /[!@#$%^&*(),.?":{}|<>]/.test(password);
  
  return hasMinLength && hasUpperCase && hasLowerCase && hasDigit && hasSpecialChar;
};

// Phone validation
export const isValidPhone = (phone: string): boolean => {
  const phoneRegex = /^(\+\d{1,3}[- ]?)?\d{10}$/;
  return phoneRegex.test(phone);
};

// Company name validation
export const isValidCompanyName = (name: string): boolean => {
  const nameRegex = /^[a-zA-Z0-9\s\-&.,()]+$/;
  return name.length >= 2 && name.length <= 100 && nameRegex.test(name);
};

// Invoice number validation
export const isValidInvoiceNumber = (invoiceNum: string): boolean => {
  const invoiceRegex = /^[a-zA-Z0-9\-_]+$/;
  return invoiceNum.length >= 3 && invoiceNum.length <= 50 && invoiceRegex.test(invoiceNum);
};

// Amount validation
export const isValidAmount = (amount: number): boolean => {
  return amount > 0 && amount <= 10000000 && Number.isFinite(amount);
};

// Date validation
export const isValidFutureDate = (dateString: string): boolean => {
  const date = new Date(dateString);
  const now = new Date();
  return !isNaN(date.getTime()) && date > now;
};

// Generic validation function
export const validateField = (value: any, rule: ValidationRule): string | null => {
  // Required validation
  if (rule.required && (!value || (typeof value === 'string' && value.trim() === ''))) {
    return 'This field is required';
  }
  
  // Skip other validations if field is empty and not required
  if (!value && !rule.required) {
    return null;
  }
  
  // String validations
  if (typeof value === 'string') {
    const sanitizedValue = sanitizeInput(value);
    
    if (rule.minLength && sanitizedValue.length < rule.minLength) {
      return `Must be at least ${rule.minLength} characters`;
    }
    
    if (rule.maxLength && sanitizedValue.length > rule.maxLength) {
      return `Must be no more than ${rule.maxLength} characters`;
    }
    
    if (rule.pattern && !rule.pattern.test(sanitizedValue)) {
      return 'Invalid format';
    }
  }
  
  // Custom validation
  if (rule.custom) {
    const customResult = rule.custom(value);
    if (customResult !== true) {
      return typeof customResult === 'string' ? customResult : 'Invalid value';
    }
  }
  
  return null;
};

// Validate entire form
export const validateForm = (data: any, rules: ValidationRules): ValidationResult => {
  const errors: { [key: string]: string } = {};
  
  // Sanitize data first
  const sanitizedData = sanitizeObject(data);
  
  for (const [field, rule] of Object.entries(rules)) {
    const error = validateField(sanitizedData[field], rule);
    if (error) {
      errors[field] = error;
    }
  }
  
  return {
    isValid: Object.keys(errors).length === 0,
    errors
  };
};

// Predefined validation rules
export const validationRules = {
  email: {
    required: true,
    maxLength: 255,
    custom: (value: string) => isValidEmail(value) || 'Invalid email format'
  },
  password: {
    required: true,
    minLength: 8,
    maxLength: 128,
    custom: (value: string) => isValidPassword(value) || 
      'Password must contain uppercase, lowercase, digit, and special character'
  },
  confirmPassword: (originalPassword: string) => ({
    required: true,
    custom: (value: string) => value === originalPassword || 'Passwords do not match'
  }),
  companyName: {
    required: true,
    minLength: 2,
    maxLength: 100,
    custom: (value: string) => isValidCompanyName(value) || 
      'Company name can only contain letters, numbers, spaces, and common punctuation'
  },
  phone: {
    custom: (value: string) => !value || isValidPhone(value) || 'Invalid phone number format'
  },
  invoiceNumber: {
    required: true,
    minLength: 3,
    maxLength: 50,
    custom: (value: string) => isValidInvoiceNumber(value) || 
      'Invoice number can only contain letters, numbers, hyphens, and underscores'
  },
  customerName: {
    required: true,
    minLength: 2,
    maxLength: 100
  },
  amount: {
    required: true,
    custom: (value: number) => isValidAmount(Number(value)) || 
      'Amount must be positive and not exceed 10,000,000'
  },
  dueDate: {
    required: true,
    custom: (value: string) => isValidFutureDate(value) || 'Date must be in the future'
  },
  description: {
    maxLength: 1000
  },
  terms: {
    maxLength: 2000
  }
};

// Debounce function for real-time validation
export const debounce = <T extends (...args: any[]) => any>(
  func: T,
  wait: number
): ((...args: Parameters<T>) => void) => {
  let timeout: NodeJS.Timeout;
  
  return (...args: Parameters<T>) => {
    clearTimeout(timeout);
    timeout = setTimeout(() => func(...args), wait);
  };
};

// Rate limiting for client-side requests
class ClientRateLimiter {
  private requests: { [key: string]: number[] } = {};
  private limits: { [key: string]: { count: number; window: number } } = {
    auth: { count: 5, window: 60000 }, // 5 requests per minute for auth
    default: { count: 60, window: 60000 }, // 60 requests per minute for other endpoints
    upload: { count: 10, window: 60000 } // 10 uploads per minute
  };

  isAllowed(endpoint: string, type: string = 'default'): boolean {
    const now = Date.now();
    const limit = this.limits[type] || this.limits.default;
    
    if (!this.requests[endpoint]) {
      this.requests[endpoint] = [];
    }
    
    // Remove old requests outside the window
    this.requests[endpoint] = this.requests[endpoint].filter(
      timestamp => now - timestamp < limit.window
    );
    
    // Check if limit is exceeded
    if (this.requests[endpoint].length >= limit.count) {
      return false;
    }
    
    // Add current request
    this.requests[endpoint].push(now);
    return true;
  }

  getRemainingRequests(endpoint: string, type: string = 'default'): number {
    const now = Date.now();
    const limit = this.limits[type] || this.limits.default;
    
    if (!this.requests[endpoint]) {
      return limit.count;
    }
    
    // Remove old requests outside the window
    this.requests[endpoint] = this.requests[endpoint].filter(
      timestamp => now - timestamp < limit.window
    );
    
    return Math.max(0, limit.count - this.requests[endpoint].length);
  }
}

export const rateLimiter = new ClientRateLimiter();

// Content Security Policy helpers
export const isExternalUrl = (url: string): boolean => {
  try {
    const urlObj = new URL(url);
    return urlObj.origin !== window.location.origin;
  } catch {
    return false;
  }
};

export const sanitizeUrl = (url: string): string => {
  // Only allow http, https, and relative URLs
  if (url.startsWith('http://') || url.startsWith('https://') || url.startsWith('/')) {
    return url;
  }
  
  // Block javascript:, data:, and other potentially dangerous protocols
  return '#';
};

// File upload validation
export const validateFile = (file: File, options: {
  maxSize?: number;
  allowedTypes?: string[];
  allowedExtensions?: string[];
} = {}): string | null => {
  const {
    maxSize = 10 * 1024 * 1024, // 10MB default
    allowedTypes = ['application/pdf', 'image/jpeg', 'image/png', 'image/jpg'],
    allowedExtensions = ['.pdf', '.jpg', '.jpeg', '.png']
  } = options;
  
  // Check file size
  if (file.size > maxSize) {
    return `File size must be less than ${Math.round(maxSize / 1024 / 1024)}MB`;
  }
  
  // Check file type
  if (!allowedTypes.includes(file.type)) {
    return `File type ${file.type} is not allowed`;
  }
  
  // Check file extension
  const extension = '.' + file.name.split('.').pop()?.toLowerCase();
  if (!allowedExtensions.includes(extension)) {
    return `File extension ${extension} is not allowed`;
  }
  
  return null;
};

// Local storage security helpers
export const secureStorage = {
  setItem: (key: string, value: any): void => {
    try {
      // Don't store sensitive data in localStorage in production
      if (process.env.NODE_ENV === 'production' && key.includes('token')) {
        console.warn('Storing tokens in localStorage is not recommended in production');
      }
      
      localStorage.setItem(key, JSON.stringify(value));
    } catch (error) {
      console.error('Failed to store item:', error);
    }
  },
  
  getItem: (key: string): any => {
    try {
      const item = localStorage.getItem(key);
      return item ? JSON.parse(item) : null;
    } catch (error) {
      console.error('Failed to retrieve item:', error);
      return null;
    }
  },
  
  removeItem: (key: string): void => {
    try {
      localStorage.removeItem(key);
    } catch (error) {
      console.error('Failed to remove item:', error);
    }
  },
  
  clear: (): void => {
    try {
      localStorage.clear();
    } catch (error) {
      console.error('Failed to clear storage:', error);
    }
  }
};
