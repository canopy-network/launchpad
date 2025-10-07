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

type hashRequest struct {
	Hash string `json:"hash"`
}

type addressRequest struct {
	Address lib.HexBytes `json:"address"`
}

func (h *heightRequest) GetHeight() uint64 {
	return h.Height
}

type queryWithHeight interface {
	GetHeight() uint64
}
