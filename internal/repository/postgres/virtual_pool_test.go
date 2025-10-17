package postgres

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/enielson/launchpad/internal/models"
	"github.com/enielson/launchpad/internal/repository/interfaces"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPoolByChainID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewVirtualPoolRepository(sqlxDB)

	chainID := uuid.New()
	poolID := uuid.New()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "chain_id", "cnpy_reserve", "token_reserve", "current_price_cnpy",
			"market_cap_usd", "total_volume_cnpy", "total_transactions", "unique_traders",
			"is_active", "price_24h_change_percent", "volume_24h_cnpy", "high_24h_cnpy",
			"low_24h_cnpy", "created_at", "updated_at",
		}).AddRow(
			poolID, chainID, 10000.0, 800000000, 0.0000125, 10000.0,
			5000.0, 10, 5, true, 2.5, 1000.0, 0.000015, 0.00001,
			time.Now(), time.Now(),
		)

		mock.ExpectQuery("SELECT (.+) FROM virtual_pools WHERE chain_id").
			WithArgs(chainID).
			WillReturnRows(rows)

		pool, err := repo.GetPoolByChainID(context.Background(), chainID)
		require.NoError(t, err)
		assert.NotNil(t, pool)
		assert.Equal(t, poolID, pool.ID)
		assert.Equal(t, chainID, pool.ChainID)
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT (.+) FROM virtual_pools WHERE chain_id").
			WithArgs(chainID).
			WillReturnError(sqlmock.ErrCancelled)

		pool, err := repo.GetPoolByChainID(context.Background(), chainID)
		assert.Error(t, err)
		assert.Nil(t, pool)
	})
}

func TestUpdatePoolState(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewVirtualPoolRepository(sqlxDB)

	chainID := uuid.New()

	t.Run("success with all fields", func(t *testing.T) {
		newReserve := big.NewFloat(12000.0)
		newTokenReserve := big.NewFloat(750000000)
		newPrice := big.NewFloat(0.000016)
		totalTx := 15
		uniqueTraders := 7

		update := &interfaces.PoolStateUpdate{
			CNPYReserve:       newReserve,
			TokenReserve:      newTokenReserve,
			CurrentPriceCNPY:  newPrice,
			TotalTransactions: &totalTx,
			UniqueTraders:     &uniqueTraders,
		}

		mock.ExpectExec("UPDATE virtual_pools SET").
			WithArgs(chainID, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.UpdatePoolState(context.Background(), chainID, update)
		require.NoError(t, err)
	})

	t.Run("pool not found", func(t *testing.T) {
		update := &interfaces.PoolStateUpdate{
			CNPYReserve: big.NewFloat(12000.0),
		}

		mock.ExpectExec("UPDATE virtual_pools SET").
			WithArgs(chainID, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.UpdatePoolState(context.Background(), chainID, update)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestCreateTransaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewVirtualPoolRepository(sqlxDB)

	transaction := &models.VirtualPoolTransaction{
		VirtualPoolID:         uuid.New(),
		ChainID:               uuid.New(),
		UserID:                uuid.New(),
		TransactionType:       "buy",
		CNPYAmount:            100.0,
		TokenAmount:           8000,
		PricePerTokenCNPY:     0.0125,
		TradingFeeCNPY:        1.0,
		PoolCNPYReserveAfter:  10100.0,
		PoolTokenReserveAfter: 792000,
		MarketCapAfterUSD:     10100.0,
	}

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "created_at"}).
			AddRow(uuid.New(), time.Now())

		mock.ExpectQuery("INSERT INTO virtual_pool_transactions").
			WithArgs(
				transaction.VirtualPoolID,
				transaction.ChainID,
				transaction.UserID,
				transaction.TransactionType,
				transaction.CNPYAmount,
				transaction.TokenAmount,
				transaction.PricePerTokenCNPY,
				transaction.TradingFeeCNPY,
				transaction.SlippagePercent,
				transaction.PoolCNPYReserveAfter,
				transaction.PoolTokenReserveAfter,
				transaction.MarketCapAfterUSD,
				transaction.TransactionHash,
				transaction.BlockHeight,
				transaction.GasUsed,
			).
			WillReturnRows(rows)

		err := repo.CreateTransaction(context.Background(), transaction)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, transaction.ID)
	})
}

func TestGetUserPosition(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewVirtualPoolRepository(sqlxDB)

	userID := uuid.New()
	chainID := uuid.New()
	positionID := uuid.New()
	poolID := uuid.New()

	t.Run("success", func(t *testing.T) {
		now := time.Now()
		rows := sqlmock.NewRows([]string{
			"id", "user_id", "chain_id", "virtual_pool_id", "token_balance",
			"total_cnpy_invested", "total_cnpy_withdrawn", "average_entry_price_cnpy",
			"unrealized_pnl_cnpy", "realized_pnl_cnpy", "total_return_percent",
			"is_active", "first_purchase_at", "last_activity_at", "created_at", "updated_at",
		}).AddRow(
			positionID, userID, chainID, poolID, 8000, 100.0, 0.0, 0.0125,
			5.0, 0.0, 5.0, true, now, now, now, now,
		)

		mock.ExpectQuery("SELECT (.+) FROM user_virtual_positions WHERE user_id").
			WithArgs(userID, chainID).
			WillReturnRows(rows)

		position, err := repo.GetUserPosition(context.Background(), userID, chainID)
		require.NoError(t, err)
		assert.NotNil(t, position)
		assert.Equal(t, positionID, position.ID)
		assert.Equal(t, userID, position.UserID)
		assert.Equal(t, int64(8000), position.TokenBalance)
	})

	t.Run("not found returns nil", func(t *testing.T) {
		mock.ExpectQuery("SELECT (.+) FROM user_virtual_positions WHERE user_id").
			WithArgs(userID, chainID).
			WillReturnError(sqlmock.ErrCancelled)

		position, err := repo.GetUserPosition(context.Background(), userID, chainID)
		assert.Error(t, err)
		assert.Nil(t, position)
	})
}

func TestUpsertUserPosition(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewVirtualPoolRepository(sqlxDB)

	now := time.Now()
	position := &models.UserVirtualLPPosition{
		UserID:                uuid.New(),
		ChainID:               uuid.New(),
		VirtualPoolID:         uuid.New(),
		TokenBalance:          8000,
		TotalCNPYInvested:     100.0,
		TotalCNPYWithdrawn:    0.0,
		AverageEntryPriceCNPY: 0.0125,
		UnrealizedPnlCNPY:     5.0,
		RealizedPnlCNPY:       0.0,
		TotalReturnPercent:    5.0,
		IsActive:              true,
		FirstPurchaseAt:       &now,
		LastActivityAt:        &now,
	}

	t.Run("success insert", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(uuid.New(), time.Now(), time.Now())

		mock.ExpectQuery("INSERT INTO user_virtual_positions").
			WithArgs(
				position.UserID,
				position.ChainID,
				position.VirtualPoolID,
				position.TokenBalance,
				position.TotalCNPYInvested,
				position.TotalCNPYWithdrawn,
				position.AverageEntryPriceCNPY,
				position.UnrealizedPnlCNPY,
				position.RealizedPnlCNPY,
				position.TotalReturnPercent,
				position.IsActive,
				position.FirstPurchaseAt,
				position.LastActivityAt,
			).
			WillReturnRows(rows)

		err := repo.UpsertUserPosition(context.Background(), position)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, position.ID)
	})
}

func TestGetTransactionsByPoolID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewVirtualPoolRepository(sqlxDB)

	poolID := uuid.New()
	pagination := interfaces.Pagination{Limit: 10, Offset: 0}

	t.Run("success", func(t *testing.T) {
		// Count query
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
		mock.ExpectQuery("SELECT COUNT").
			WithArgs(poolID).
			WillReturnRows(countRows)

		// Data query
		dataRows := sqlmock.NewRows([]string{
			"id", "virtual_pool_id", "chain_id", "user_id", "transaction_type",
			"cnpy_amount", "token_amount", "price_per_token_cnpy", "trading_fee_cnpy",
			"slippage_percent", "pool_cnpy_reserve_after", "pool_token_reserve_after",
			"market_cap_after_usd", "transaction_hash", "block_height", "gas_used", "created_at",
		}).
			AddRow(uuid.New(), poolID, uuid.New(), uuid.New(), "buy", 100.0, 8000, 0.0125,
				1.0, 0.5, 10100.0, 792000, 10100.0, nil, nil, nil, time.Now()).
			AddRow(uuid.New(), poolID, uuid.New(), uuid.New(), "sell", 50.0, 4000, 0.0125,
				0.5, 0.3, 10050.0, 796000, 10050.0, nil, nil, nil, time.Now())

		mock.ExpectQuery("SELECT (.+) FROM virtual_pool_transactions WHERE virtual_pool_id").
			WithArgs(poolID, pagination.Limit, pagination.Offset).
			WillReturnRows(dataRows)

		transactions, total, err := repo.GetTransactionsByPoolID(context.Background(), poolID, pagination)
		require.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, transactions, 2)
	})
}

func TestGetPositionsByChainID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewVirtualPoolRepository(sqlxDB)

	chainID := uuid.New()
	pagination := interfaces.Pagination{Limit: 10, Offset: 0}

	t.Run("success", func(t *testing.T) {
		// Count query
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
		mock.ExpectQuery("SELECT COUNT").
			WithArgs(chainID).
			WillReturnRows(countRows)

		// Data query
		now := time.Now()
		dataRows := sqlmock.NewRows([]string{
			"id", "user_id", "chain_id", "virtual_pool_id", "token_balance",
			"total_cnpy_invested", "total_cnpy_withdrawn", "average_entry_price_cnpy",
			"unrealized_pnl_cnpy", "realized_pnl_cnpy", "total_return_percent",
			"is_active", "first_purchase_at", "last_activity_at", "created_at", "updated_at",
		}).AddRow(
			uuid.New(), uuid.New(), chainID, uuid.New(), 8000, 100.0, 0.0, 0.0125,
			5.0, 0.0, 5.0, true, now, now, now, now,
		)

		mock.ExpectQuery("SELECT (.+) FROM user_virtual_positions WHERE chain_id").
			WithArgs(chainID, pagination.Limit, pagination.Offset).
			WillReturnRows(dataRows)

		positions, total, err := repo.GetPositionsByChainID(context.Background(), chainID, pagination)
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, positions, 1)
	})
}
