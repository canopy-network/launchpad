package models

import (
	"time"

	"github.com/google/uuid"
)

// Wallet represents an encrypted keypair stored in the database
type Wallet struct {
	ID                    uuid.UUID  `json:"id" db:"id"`
	UserID                *uuid.UUID `json:"user_id,omitempty" db:"user_id"`
	ChainID               *uuid.UUID `json:"chain_id,omitempty" db:"chain_id"`
	Address               string     `json:"address" db:"address"`
	PublicKey             string     `json:"public_key" db:"public_key"`
	EncryptedPrivateKey   string     `json:"-" db:"encrypted_private_key"` // Never expose in JSON
	Salt                  []byte     `json:"-" db:"salt"`                  // Never expose in JSON
	WalletName            *string    `json:"wallet_name,omitempty" db:"wallet_name"`
	WalletDescription     *string    `json:"wallet_description,omitempty" db:"wallet_description"`
	IsActive              bool       `json:"is_active" db:"is_active"`
	IsLocked              bool       `json:"is_locked" db:"is_locked"`
	LastUsedAt            *time.Time `json:"last_used_at,omitempty" db:"last_used_at"`
	PasswordChangedAt     *time.Time `json:"password_changed_at,omitempty" db:"password_changed_at"`
	FailedDecryptAttempts int        `json:"-" db:"failed_decrypt_attempts"` // Don't expose
	LockedUntil           *time.Time `json:"locked_until,omitempty" db:"locked_until"`
	CreatedBy             *uuid.UUID `json:"created_by,omitempty" db:"created_by"`
	CreatedAt             time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at" db:"updated_at"`
}

// WalletWithDecryptedKey contains the wallet with decrypted private key (use with extreme caution)
type WalletWithDecryptedKey struct {
	Wallet
	DecryptedPrivateKey string `json:"private_key"` // Only for immediate use, never persist
}
