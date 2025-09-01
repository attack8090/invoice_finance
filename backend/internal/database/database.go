package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

func Initialize(databaseURL string) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(databaseURL))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Test the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	// Extract database name from URL or use default
	databaseName := "invoice_financing"
	db := client.Database(databaseName)

	log.Printf("Successfully connected to MongoDB database: %s", databaseName)

	return &MongoDB{
		Client:   client,
		Database: db,
	}, nil
}

// RunMigrations creates indexes for better performance
func RunMigrations(db *MongoDB) error {
	ctx := context.Background()

	// Create indexes for Users collection
	userCollection := db.Database.Collection("users")
	_, err := userCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: map[string]int{"email": 1},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		log.Printf("Warning: Could not create email index: %v", err)
	}

	// Create indexes for Invoices collection
	invoiceCollection := db.Database.Collection("invoices")
	_, err = invoiceCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: map[string]int{"user_id": 1, "status": 1},
	})
	if err != nil {
		log.Printf("Warning: Could not create invoice index: %v", err)
	}

	// Create indexes for FinancingRequests collection
	financingCollection := db.Database.Collection("financing_requests")
	_, err = financingCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: map[string]int{"invoice_id": 1, "status": 1},
	})
	if err != nil {
		log.Printf("Warning: Could not create financing request index: %v", err)
	}

	// Create indexes for Investments collection
	investmentCollection := db.Database.Collection("investments")
	_, err = investmentCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: map[string]int{"investor_id": 1, "status": 1},
	})
	if err != nil {
		log.Printf("Warning: Could not create investment index: %v", err)
	}

	// Create indexes for Transactions collection
	transactionCollection := db.Database.Collection("transactions")
	_, err = transactionCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: map[string]int{"user_id": 1, "created_at": -1},
	})
	if err != nil {
		log.Printf("Warning: Could not create transaction index: %v", err)
	}

	log.Println("MongoDB indexes created successfully")
	return nil
}

// Close closes the MongoDB connection
func (db *MongoDB) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return db.Client.Disconnect(ctx)
}
