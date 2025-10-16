package canopy

import (
	"github.com/canopy-network/canopy/lib"
)

// =====================================================
// Query Request Types
// =====================================================
type heightRequest struct {
	Height uint64 `json:"height"`
}

type orderRequest struct {
	ChainId uint64 `json:"chainId"`
	OrderId string `json:"orderId"`
	heightRequest
}

type heightsRequest struct {
	heightRequest
	StartHeight uint64 `json:"startHeight"`
}

type idRequest struct {
	ID uint64 `json:"id"`
}

type paginatedAddressRequest struct {
	addressRequest
	lib.PageParams
}

type paginatedHeightRequest struct {
	heightRequest
	lib.PageParams
	lib.ValidatorFilters
}

type heightAndAddressRequest struct {
	heightRequest
	addressRequest
}

type heightAndIdRequest struct {
	heightRequest
	idRequest
}

type heightIdAndPointsRequest struct {
	heightAndIdRequest
	Points bool `json:"points"`
}

type hashRequest struct {
	Hash string `json:"hash"`
}

type addressRequest struct {
	Address lib.HexBytes `json:"address"`
}

type committeesRequest struct {
	Committees string `json:"committees"`
}

type fromFields struct {
	Address  lib.HexBytes `json:"address,omitempty"`
	Nickname string       `json:"nickname,omitempty"`
}

func (h *heightRequest) GetHeight() uint64 {
	return h.Height
}

type queryWithHeight interface {
	GetHeight() uint64
}

// =====================================================
// Transaction Request Types
// =====================================================

type AddrOrNickname struct {
	Address  string
	Nickname string
}

type txDexLimitOrder struct {
	Fee           uint64 `json:"fee"`
	Amount        uint64 `json:"amount"`
	ReceiveAmount uint64 `json:"receiveAmount"`
	Submit        bool   `json:"submit"`
	Password      string `json:"password"`
	fromFields
	committeesRequest
}

type txDexLiquidityDeposit struct {
	Fee      uint64 `json:"fee"`
	Amount   uint64 `json:"amount"`
	Submit   bool   `json:"submit"`
	Password string `json:"password"`
	fromFields
	committeesRequest
}

type txDexLiquidityWithdraw struct {
	Fee      uint64 `json:"fee"`
	Percent  int    `json:"percent"`
	Submit   bool   `json:"submit"`
	Password string `json:"password"`
	fromFields
	committeesRequest
}
