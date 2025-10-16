package models

import (
	"net"
	"time"

	"github.com/google/uuid"
)

// SessionToken represents an authenticated user session
type SessionToken struct {
	ID                 uuid.UUID  `json:"id" db:"id"`
	UserID             uuid.UUID  `json:"user_id" db:"user_id"`
	TokenHash          string     `json:"-" db:"token_hash"` // Never expose in JSON
	TokenPrefix        string     `json:"token_prefix" db:"token_prefix"`
	UserAgent          *string    `json:"user_agent" db:"user_agent"`
	IPAddress          *net.IP    `json:"ip_address" db:"ip_address"`
	DeviceName         *string    `json:"device_name" db:"device_name"`
	ExpiresAt          time.Time  `json:"expires_at" db:"expires_at"`
	LastUsedAt         time.Time  `json:"last_used_at" db:"last_used_at"`
	IsRevoked          bool       `json:"is_revoked" db:"is_revoked"`
	RevokedAt          *time.Time `json:"revoked_at" db:"revoked_at"`
	RevocationReason   *string    `json:"revocation_reason" db:"revocation_reason"`
	JWTVersionSnapshot int        `json:"-" db:"jwt_version_snapshot"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
}

// LoginResponse is returned after successful email verification
type LoginResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

// SessionInfo represents session information for display
type SessionInfo struct {
	ID           uuid.UUID  `json:"id"`
	TokenPrefix  string     `json:"token_prefix"`
	UserAgent    *string    `json:"user_agent"`
	IPAddress    string     `json:"ip_address,omitempty"`
	DeviceName   *string    `json:"device_name"`
	ExpiresAt    time.Time  `json:"expires_at"`
	LastUsedAt   time.Time  `json:"last_used_at"`
	IsRevoked    bool       `json:"is_revoked"`
	RevokedAt    *time.Time `json:"revoked_at"`
	IsCurrent    bool       `json:"is_current"`
	CreatedAt    time.Time  `json:"created_at"`
}

// Revocation reason constants
const (
	RevocationReasonUserLogout    = "user_logout"
	RevocationReasonSecurityEvent = "security_event"
	RevocationReasonAdminAction   = "admin_action"
	RevocationReasonExpired       = "expired"
)

// IsValid checks if a session token is currently valid
func (st *SessionToken) IsValid(currentJWTVersion int) bool {
	if st.IsRevoked {
		return false
	}
	if time.Now().After(st.ExpiresAt) {
		return false
	}
	if st.JWTVersionSnapshot != currentJWTVersion {
		return false
	}
	return true
}

// ToSessionInfo converts SessionToken to SessionInfo for display
func (st *SessionToken) ToSessionInfo(isCurrent bool) *SessionInfo {
	info := &SessionInfo{
		ID:          st.ID,
		TokenPrefix: st.TokenPrefix,
		UserAgent:   st.UserAgent,
		DeviceName:  st.DeviceName,
		ExpiresAt:   st.ExpiresAt,
		LastUsedAt:  st.LastUsedAt,
		IsRevoked:   st.IsRevoked,
		RevokedAt:   st.RevokedAt,
		IsCurrent:   isCurrent,
		CreatedAt:   st.CreatedAt,
	}

	// Convert IP address to string if present
	if st.IPAddress != nil {
		info.IPAddress = st.IPAddress.String()
	}

	return info
}
