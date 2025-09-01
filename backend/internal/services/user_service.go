package services

import (
	"context"
	"time"

	"invoice-financing-platform/internal/database"
	"invoice-financing-platform/internal/models"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserService struct {
	db *database.MongoDB
}

func NewUserService(db *database.MongoDB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) Create(user *models.User) error {
	user.UUID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	
	collection := s.db.Database.Collection("users")
	_, err := collection.InsertOne(context.Background(), user)
	return err
}

func (s *UserService) GetByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	collection := s.db.Database.Collection("users")
	
	filter := bson.M{"uuid": id}
	err := collection.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) GetByEmail(email string) (*models.User, error) {
	var user models.User
	collection := s.db.Database.Collection("users")
	
	filter := bson.M{"email": email}
	err := collection.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) Update(user *models.User) error {
	user.UpdatedAt = time.Now()
	collection := s.db.Database.Collection("users")
	
	filter := bson.M{"uuid": user.UUID}
	update := bson.M{"$set": user}
	_, err := collection.UpdateOne(context.Background(), filter, update)
	return err
}

func (s *UserService) Delete(id uuid.UUID) error {
	collection := s.db.Database.Collection("users")
	
	// Soft delete by setting deleted_at
	filter := bson.M{"uuid": id}
	now := time.Now()
	update := bson.M{"$set": bson.M{"deleted_at": now}}
	_, err := collection.UpdateOne(context.Background(), filter, update)
	return err
}

func (s *UserService) GetAll(limit, offset int) ([]models.User, error) {
	var users []models.User
	collection := s.db.Database.Collection("users")
	
	opts := options.Find().SetLimit(int64(limit)).SetSkip(int64(offset))
	cursor, err := collection.Find(context.Background(), bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	
	err = cursor.All(context.Background(), &users)
	return users, err
}

func (s *UserService) GetByRole(role models.UserRole) ([]models.User, error) {
	var users []models.User
	collection := s.db.Database.Collection("users")
	
	filter := bson.M{"role": role}
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	
	err = cursor.All(context.Background(), &users)
	return users, err
}

func (s *UserService) Verify(userID uuid.UUID) error {
	collection := s.db.Database.Collection("users")
	
	filter := bson.M{"uuid": userID}
	update := bson.M{"$set": bson.M{
		"is_verified": true,
		"updated_at": time.Now(),
	}}
	_, err := collection.UpdateOne(context.Background(), filter, update)
	return err
}

func (s *UserService) UpdateCreditScore(userID uuid.UUID, score int) error {
	collection := s.db.Database.Collection("users")
	
	filter := bson.M{"uuid": userID}
	update := bson.M{"$set": bson.M{
		"credit_score": score,
		"updated_at": time.Now(),
	}}
	_, err := collection.UpdateOne(context.Background(), filter, update)
	return err
}

func (s *UserService) GetStats(userID uuid.UUID) (map[string]interface{}, error) {
	user, err := s.GetByID(userID)
	if err != nil {
		return nil, err
	}

	stats := make(map[string]interface{})

	if user.Role == models.RoleSME {
		// SME stats
		invoiceCollection := s.db.Database.Collection("invoices")
		totalInvoices, _ := invoiceCollection.CountDocuments(context.Background(), bson.M{"user_id": userID})

		financingCollection := s.db.Database.Collection("financing_requests")
		pendingRequests, _ := financingCollection.CountDocuments(context.Background(), bson.M{
			"user_id": userID,
			"status":  models.FinancingStatusPending,
		})

		stats["total_invoices"] = totalInvoices
		stats["total_financed"] = user.TotalFinanced
		stats["pending_requests"] = pendingRequests
		stats["credit_score"] = user.CreditScore

	} else if user.Role == models.RoleInvestor {
		// Investor stats
		investmentCollection := s.db.Database.Collection("investments")
		totalInvestments, _ := investmentCollection.CountDocuments(context.Background(), bson.M{"investor_id": userID})

		activeInvestments, _ := investmentCollection.CountDocuments(context.Background(), bson.M{
			"investor_id": userID,
			"status":      models.InvestmentStatusActive,
		})

		stats["total_investments"] = totalInvestments
		stats["total_invested"] = user.TotalInvestment
		stats["total_returns"] = 0 // Can be calculated from completed investments
		stats["active_investments"] = activeInvestments
	}

	return stats, nil
}

func (s *UserService) UpdateWalletAddress(userID uuid.UUID, walletAddress string) error {
	collection := s.db.Database.Collection("users")
	
	filter := bson.M{"uuid": userID}
	update := bson.M{"$set": bson.M{
		"wallet_address": walletAddress,
		"updated_at":     time.Now(),
	}}
	_, err := collection.UpdateOne(context.Background(), filter, update)
	return err
}
