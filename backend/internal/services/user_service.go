package services

import (

	"invoice-financing-platform/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) Create(user *models.User) error {
	return s.db.Create(user).Error
}

func (s *UserService) GetByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := s.db.First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := s.db.First(&user, "email = ?", email).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) Update(user *models.User) error {
	return s.db.Save(user).Error
}

func (s *UserService) Delete(id uuid.UUID) error {
	return s.db.Delete(&models.User{}, "id = ?", id).Error
}

func (s *UserService) GetAll(limit, offset int) ([]models.User, error) {
	var users []models.User
	err := s.db.Limit(limit).Offset(offset).Find(&users).Error
	return users, err
}

func (s *UserService) GetByRole(role models.UserRole) ([]models.User, error) {
	var users []models.User
	err := s.db.Where("role = ?", role).Find(&users).Error
	return users, err
}

func (s *UserService) Verify(userID uuid.UUID) error {
	return s.db.Model(&models.User{}).
		Where("id = ?", userID).
		Update("is_verified", true).Error
}

func (s *UserService) UpdateCreditScore(userID uuid.UUID, score int) error {
	return s.db.Model(&models.User{}).
		Where("id = ?", userID).
		Update("credit_score", score).Error
}

func (s *UserService) GetStats(userID uuid.UUID) (map[string]interface{}, error) {
	user, err := s.GetByID(userID)
	if err != nil {
		return nil, err
	}

	stats := make(map[string]interface{})

	if user.Role == models.RoleSME {
		// SME stats
		var totalInvoices int64
		s.db.Model(&models.Invoice{}).Where("user_id = ?", userID).Count(&totalInvoices)

		var totalFinanced float64
		s.db.Model(&models.FinancingRequest{}).
			Where("user_id = ? AND status = ?", userID, models.FinancingStatusCompleted).
			Select("COALESCE(SUM(net_amount), 0)").Scan(&totalFinanced)

		var pendingRequests int64
		s.db.Model(&models.FinancingRequest{}).
			Where("user_id = ? AND status = ?", userID, models.FinancingStatusPending).
			Count(&pendingRequests)

		stats["total_invoices"] = totalInvoices
		stats["total_financed"] = totalFinanced
		stats["pending_requests"] = pendingRequests
		stats["credit_score"] = user.CreditScore

	} else if user.Role == models.RoleInvestor {
		// Investor stats
		var totalInvestments int64
		s.db.Model(&models.Investment{}).Where("investor_id = ?", userID).Count(&totalInvestments)

		var totalInvested float64
		s.db.Model(&models.Investment{}).
			Where("investor_id = ?", userID).
			Select("COALESCE(SUM(amount), 0)").Scan(&totalInvested)

		var totalReturns float64
		s.db.Model(&models.Investment{}).
			Where("investor_id = ? AND status = ?", userID, models.InvestmentStatusCompleted).
			Select("COALESCE(SUM(actual_return), 0)").Scan(&totalReturns)

		var activeInvestments int64
		s.db.Model(&models.Investment{}).
			Where("investor_id = ? AND status = ?", userID, models.InvestmentStatusActive).
			Count(&activeInvestments)

		stats["total_investments"] = totalInvestments
		stats["total_invested"] = totalInvested
		stats["total_returns"] = totalReturns
		stats["active_investments"] = activeInvestments
	}

	return stats, nil
}

func (s *UserService) UpdateWalletAddress(userID uuid.UUID, walletAddress string) error {
	return s.db.Model(&models.User{}).
		Where("id = ?", userID).
		Update("wallet_address", walletAddress).Error
}
