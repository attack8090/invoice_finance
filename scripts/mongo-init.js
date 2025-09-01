// MongoDB initialization script for Invoice Financing Platform
// This script creates multiple databases and users for different services

// Switch to admin database for user creation
db = db.getSiblingDB('admin');

// Create databases and collections
const databases = [
  'invoice_financing',
  'blockchain_ledger', 
  'payment_service',
  'notification_service',
  'bank_integration',
  'integration_external_data'
];

print('Creating databases and initial collections...');

databases.forEach(dbName => {
  print(`Creating database: ${dbName}`);
  
  // Switch to the database
  db = db.getSiblingDB(dbName);
  
  // Create initial collections based on database purpose
  if (dbName === 'invoice_financing') {
    // Main application collections
    db.createCollection('users');
    db.createCollection('invoices');
    db.createCollection('financing_requests');
    db.createCollection('investments');
    db.createCollection('transactions');
    
    // Create indexes for better performance
    db.users.createIndex({ "email": 1 }, { unique: true });
    db.users.createIndex({ "uuid": 1 }, { unique: true });
    db.invoices.createIndex({ "user_id": 1, "status": 1 });
    db.invoices.createIndex({ "uuid": 1 }, { unique: true });
    db.financing_requests.createIndex({ "invoice_id": 1, "status": 1 });
    db.investments.createIndex({ "investor_id": 1, "status": 1 });
    db.transactions.createIndex({ "user_id": 1, "created_at": -1 });
    
  } else if (dbName === 'blockchain_ledger') {
    // Blockchain ledger collections
    db.createCollection('blocks');
    db.createCollection('transactions');
    db.createCollection('chaincode_data');
    
    db.blocks.createIndex({ "block_number": 1 }, { unique: true });
    db.transactions.createIndex({ "tx_id": 1 }, { unique: true });
    
  } else if (dbName === 'payment_service') {
    // Payment service collections
    db.createCollection('payments');
    db.createCollection('payment_methods');
    db.createCollection('bank_accounts');
    
    db.payments.createIndex({ "user_id": 1, "status": 1 });
    db.payment_methods.createIndex({ "user_id": 1 });
    
  } else if (dbName === 'notification_service') {
    // Notification service collections
    db.createCollection('notifications');
    db.createCollection('email_templates');
    db.createCollection('notification_preferences');
    
    db.notifications.createIndex({ "user_id": 1, "created_at": -1 });
    db.notification_preferences.createIndex({ "user_id": 1 }, { unique: true });
    
  } else if (dbName === 'bank_integration') {
    // Bank integration collections
    db.createCollection('bank_connections');
    db.createCollection('bank_transactions');
    db.createCollection('account_mappings');
    
    db.bank_connections.createIndex({ "user_id": 1 });
    db.bank_transactions.createIndex({ "account_id": 1, "date": -1 });
    
  } else if (dbName === 'integration_external_data') {
    // External data integration collections
    db.createCollection('external_apis');
    db.createCollection('data_cache');
    db.createCollection('api_logs');
    
    db.external_apis.createIndex({ "api_name": 1 });
    db.data_cache.createIndex({ "cache_key": 1 }, { unique: true });
    db.api_logs.createIndex({ "created_at": -1 });
  }
  
  print(`✅ Database ${dbName} created with collections and indexes`);
});

print('✅ MongoDB initialization completed successfully!');
print('All databases and collections have been created with appropriate indexes.');
