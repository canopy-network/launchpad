package services

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/enielson/launchpad/internal/models"
	"github.com/enielson/launchpad/internal/repository/interfaces"
	"github.com/enielson/launchpad/pkg/keygen"
	"github.com/google/uuid"
)

const (
	MaxFailedAttempts   = 5
	LockDurationMinutes = 15
)

var (
	ErrWalletLocked    = fmt.Errorf("wallet is locked due to too many failed attempts")
	ErrInvalidPassword = fmt.Errorf("invalid password")
	ErrWalletNotFound  = fmt.Errorf("wallet not found")
)

// WalletService handles wallet business logic
type WalletService struct {
	walletRepo interfaces.WalletRepository
}

// NewWalletService creates a new wallet service
func NewWalletService(walletRepo interfaces.WalletRepository) *WalletService {
	return &WalletService{
		walletRepo: walletRepo,
	}
}

// CreateWallet creates a new encrypted wallet
func (s *WalletService) CreateWallet(ctx context.Context, password string, userID *uuid.UUID, chainID *uuid.UUID, walletName, walletDescription *string, createdBy *uuid.UUID) (*models.Wallet, error) {
	// Validate that at least one association exists
	if userID == nil && chainID == nil && createdBy == nil {
		return nil, fmt.Errorf("wallet must be associated with at least one entity (user, chain, or creator)")
	}

	// Generate encrypted keypair
	_, encryptedKeyPair, err := keygen.GenerateEncryptedKeyPair(password)
	if err != nil {
		return nil, fmt.Errorf("failed to generate keypair: %w", err)
	}

	// Decode salt from hex
	salt, err := hex.DecodeString(encryptedKeyPair.Salt)
	if err != nil {
		return nil, fmt.Errorf("failed to decode salt: %w", err)
	}

	// Create wallet model
	wallet := &models.Wallet{
		UserID:                userID,
		ChainID:               chainID,
		Address:               encryptedKeyPair.Address,
		PublicKey:             encryptedKeyPair.PublicKey,
		EncryptedPrivateKey:   encryptedKeyPair.EncryptedPrivateKey,
		Salt:                  salt,
		WalletName:            walletName,
		WalletDescription:     walletDescription,
		IsActive:              true,
		IsLocked:              false,
		FailedDecryptAttempts: 0,
		CreatedBy:             createdBy,
	}

	// Save to database
	createdWallet, err := s.walletRepo.Create(ctx, wallet)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	return createdWallet, nil
}

// DecryptWallet decrypts a wallet's private key with password (with rate limiting)
func (s *WalletService) DecryptWallet(ctx context.Context, walletID uuid.UUID, password string) (*models.WalletWithDecryptedKey, error) {
	// Get wallet
	wallet, err := s.walletRepo.GetByID(ctx, walletID)
	if err != nil {
		return nil, ErrWalletNotFound
	}

	// Check if wallet is locked
	if wallet.IsLocked {
		if wallet.LockedUntil != nil && time.Now().After(*wallet.LockedUntil) {
			// Lock expired, unlock the wallet
			if err := s.walletRepo.UnlockWallet(ctx, walletID); err != nil {
				return nil, fmt.Errorf("failed to unlock wallet: %w", err)
			}
			wallet.IsLocked = false
			if err := s.walletRepo.ResetFailedAttempts(ctx, walletID); err != nil {
				return nil, fmt.Errorf("failed to reset failed attempts: %w", err)
			}
		} else {
			return nil, ErrWalletLocked
		}
	}

	// Prepare encrypted keypair for decryption
	encryptedKeyPair := &keygen.EncryptedKeyPair{
		PublicKey:           wallet.PublicKey,
		EncryptedPrivateKey: wallet.EncryptedPrivateKey,
		Salt:                hex.EncodeToString(wallet.Salt),
		Address:             wallet.Address,
	}

	// Attempt decryption
	privateKeyBytes, err := keygen.DecryptPrivateKey(encryptedKeyPair, []byte(password))
	if err != nil {
		// Increment failed attempts
		if err := s.walletRepo.IncrementFailedAttempts(ctx, walletID); err != nil {
			return nil, fmt.Errorf("failed to increment failed attempts: %w", err)
		}

		// Check if we should lock the wallet
		wallet.FailedDecryptAttempts++
		if wallet.FailedDecryptAttempts >= MaxFailedAttempts {
			if err := s.walletRepo.LockWallet(ctx, walletID, LockDurationMinutes); err != nil {
				return nil, fmt.Errorf("failed to lock wallet: %w", err)
			}
			return nil, ErrWalletLocked
		}

		return nil, ErrInvalidPassword
	}

	// Successful decryption - reset failed attempts and update last used
	if err := s.walletRepo.ResetFailedAttempts(ctx, walletID); err != nil {
		return nil, fmt.Errorf("failed to reset failed attempts: %w", err)
	}

	if err := s.walletRepo.UpdateLastUsed(ctx, walletID); err != nil {
		return nil, fmt.Errorf("failed to update last used: %w", err)
	}

	// Return wallet with decrypted private key
	return &models.WalletWithDecryptedKey{
		Wallet:              *wallet,
		DecryptedPrivateKey: hex.EncodeToString(privateKeyBytes),
	}, nil
}

// GetWallet retrieves a wallet by ID
func (s *WalletService) GetWallet(ctx context.Context, walletID uuid.UUID) (*models.Wallet, error) {
	wallet, err := s.walletRepo.GetByID(ctx, walletID)
	if err != nil {
		return nil, ErrWalletNotFound
	}
	return wallet, nil
}

// GetWalletByAddress retrieves a wallet by address
func (s *WalletService) GetWalletByAddress(ctx context.Context, address string) (*models.Wallet, error) {
	wallet, err := s.walletRepo.GetByAddress(ctx, address)
	if err != nil {
		return nil, ErrWalletNotFound
	}
	return wallet, nil
}

// ListWallets lists wallets with filters and pagination
func (s *WalletService) ListWallets(ctx context.Context, filters interfaces.WalletFilters, pagination interfaces.Pagination) ([]models.Wallet, int, error) {
	return s.walletRepo.List(ctx, filters, pagination)
}

// ListUserWallets lists all wallets for a specific user
func (s *WalletService) ListUserWallets(ctx context.Context, userID uuid.UUID, pagination interfaces.Pagination) ([]models.Wallet, int, error) {
	return s.walletRepo.ListByUserID(ctx, userID, pagination)
}

// ListChainWallets lists all wallets for a specific chain
func (s *WalletService) ListChainWallets(ctx context.Context, chainID uuid.UUID, pagination interfaces.Pagination) ([]models.Wallet, int, error) {
	return s.walletRepo.ListByChainID(ctx, chainID, pagination)
}

// UpdateWallet updates wallet metadata (name, description, active status)
func (s *WalletService) UpdateWallet(ctx context.Context, walletID uuid.UUID, walletName, walletDescription *string, isActive *bool) (*models.Wallet, error) {
	wallet, err := s.walletRepo.GetByID(ctx, walletID)
	if err != nil {
		return nil, ErrWalletNotFound
	}

	// Update fields if provided
	if walletName != nil {
		wallet.WalletName = walletName
	}
	if walletDescription != nil {
		wallet.WalletDescription = walletDescription
	}
	if isActive != nil {
		wallet.IsActive = *isActive
	}

	updatedWallet, err := s.walletRepo.Update(ctx, wallet)
	if err != nil {
		return nil, fmt.Errorf("failed to update wallet: %w", err)
	}

	return updatedWallet, nil
}

// DeleteWallet deletes a wallet
func (s *WalletService) DeleteWallet(ctx context.Context, walletID uuid.UUID) error {
	err := s.walletRepo.Delete(ctx, walletID)
	if err != nil {
		return ErrWalletNotFound
	}
	return nil
}

// UnlockWallet manually unlocks a wallet (admin/support function)
func (s *WalletService) UnlockWallet(ctx context.Context, walletID uuid.UUID) error {
	if err := s.walletRepo.UnlockWallet(ctx, walletID); err != nil {
		return ErrWalletNotFound
	}
	if err := s.walletRepo.ResetFailedAttempts(ctx, walletID); err != nil {
		return fmt.Errorf("failed to reset failed attempts: %w", err)
	}
	return nil
}
