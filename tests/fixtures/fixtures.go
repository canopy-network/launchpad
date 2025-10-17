package fixtures

import (
	"context"
	"time"

	"github.com/enielson/launchpad/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// UserFixture provides test user creation with sensible defaults
type UserFixture struct {
	ID               uuid.UUID
	WalletAddress    string
	Email            string
	Username         string
	DisplayName      string
	Bio              string
	GithubUsername   string
	IsVerified       bool
	VerificationTier string
}

// DefaultUser returns a user fixture with default values
func DefaultUser() *UserFixture {
	id := uuid.New()
	// UUID is 36 chars (8-4-4-4-12), use first 32 chars for 0x+40 char address
	idStr := id.String()
	walletAddr := "0x" + idStr[:8] + idStr[9:13] + idStr[14:18] + idStr[19:23] + idStr[24:36]
	return &UserFixture{
		ID:               id,
		WalletAddress:    walletAddr, // 0x + 40 hex chars
		Email:            "test-" + idStr[:8] + "@example.com",
		Username:         "user_" + idStr[:8],
		DisplayName:      "Test User",
		Bio:              "Test user for integration tests",
		GithubUsername:   "testuser",
		IsVerified:       true,
		VerificationTier: "verified",
	}
}

// WithEmail sets a custom email
func (u *UserFixture) WithEmail(email string) *UserFixture {
	u.Email = email
	return u
}

// WithWallet sets a custom wallet address
func (u *UserFixture) WithWallet(wallet string) *UserFixture {
	u.WalletAddress = wallet
	return u
}

// WithUsername sets a custom username
func (u *UserFixture) WithUsername(username string) *UserFixture {
	u.Username = username
	return u
}

// Create persists the user to the database (works with *sqlx.DB or *sqlx.Tx)
func (u *UserFixture) Create(ctx context.Context, db sqlx.ExtContext) (*models.User, error) {
	query := `
		INSERT INTO users (
			wallet_address, email, username, display_name, bio, github_username,
			is_verified, verification_tier
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`

	user := &models.User{
		WalletAddress:    u.WalletAddress,
		Email:            &u.Email,
		Username:         &u.Username,
		DisplayName:      &u.DisplayName,
		Bio:              &u.Bio,
		GithubUsername:   &u.GithubUsername,
		IsVerified:       u.IsVerified,
		VerificationTier: u.VerificationTier,
	}

	result := struct {
		ID        uuid.UUID `db:"id"`
		CreatedAt time.Time `db:"created_at"`
		UpdatedAt time.Time `db:"updated_at"`
	}{}

	err := sqlx.GetContext(ctx, db, &result, query,
		user.WalletAddress, user.Email, user.Username, user.DisplayName,
		user.Bio, user.GithubUsername, user.IsVerified, user.VerificationTier)

	if err != nil {
		return nil, err
	}

	user.ID = result.ID
	user.CreatedAt = result.CreatedAt
	user.UpdatedAt = result.UpdatedAt

	return user, nil
}

// ChainFixture provides test chain creation with sensible defaults
type ChainFixture struct {
	ChainName          string
	TokenSymbol        string
	ChainDescription   *string
	TemplateID         *uuid.UUID
	ConsensusMechanism string
	TokenTotalSupply   int64
	CreatedBy          uuid.UUID
	Status             string
	InitialCNPYReserve float64
	InitialTokenSupply int64
	BondingCurveSlope  float64
}

// DefaultChain returns a chain fixture with default values
func DefaultChain(creatorID uuid.UUID) *ChainFixture {
	desc := "Test chain for integration tests"
	return &ChainFixture{
		ChainName:          "Test Chain " + time.Now().Format("20060102-150405"),
		TokenSymbol:        "TEST",
		ChainDescription:   &desc,
		TemplateID:         nil,
		ConsensusMechanism: "nestbft",
		TokenTotalSupply:   1000000000,
		CreatedBy:          creatorID,
		Status:             models.ChainStatusDraft,
		InitialCNPYReserve: 100.0,
		InitialTokenSupply: 800000000,
		BondingCurveSlope:  0.00000001,
	}
}

// WithTemplate sets the template ID
func (c *ChainFixture) WithTemplate(templateID uuid.UUID) *ChainFixture {
	c.TemplateID = &templateID
	return c
}

// WithStatus sets the chain status
func (c *ChainFixture) WithStatus(status string) *ChainFixture {
	c.Status = status
	return c
}

// WithTokenSymbol sets the token symbol
func (c *ChainFixture) WithTokenSymbol(symbol string) *ChainFixture {
	c.TokenSymbol = symbol
	return c
}

// WithBondingCurve sets bonding curve parameters
func (c *ChainFixture) WithBondingCurve(cnpyReserve float64, tokenSupply int64, slope float64) *ChainFixture {
	c.InitialCNPYReserve = cnpyReserve
	c.InitialTokenSupply = tokenSupply
	c.BondingCurveSlope = slope
	return c
}

// Create persists the chain to the database (works with *sqlx.DB or *sqlx.Tx)
func (c *ChainFixture) Create(ctx context.Context, db sqlx.ExtContext) (*models.Chain, error) {
	query := `
		INSERT INTO chains (
			chain_name, token_symbol, chain_description, template_id, consensus_mechanism,
			token_total_supply, created_by, status, initial_cnpy_reserve,
			initial_token_supply, bonding_curve_slope
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at
	`

	chain := &models.Chain{
		ChainName:          c.ChainName,
		TokenSymbol:        c.TokenSymbol,
		ChainDescription:   c.ChainDescription,
		TemplateID:         c.TemplateID,
		ConsensusMechanism: c.ConsensusMechanism,
		TokenTotalSupply:   c.TokenTotalSupply,
		CreatedBy:          c.CreatedBy,
		Status:             c.Status,
		InitialCNPYReserve: c.InitialCNPYReserve,
		InitialTokenSupply: c.InitialTokenSupply,
		BondingCurveSlope:  c.BondingCurveSlope,
	}

	result := struct {
		ID        uuid.UUID `db:"id"`
		CreatedAt time.Time `db:"created_at"`
		UpdatedAt time.Time `db:"updated_at"`
	}{}

	err := sqlx.GetContext(ctx, db, &result, query,
		chain.ChainName, chain.TokenSymbol, chain.ChainDescription, chain.TemplateID,
		chain.ConsensusMechanism, chain.TokenTotalSupply, chain.CreatedBy, chain.Status,
		&chain.InitialCNPYReserve, &chain.InitialTokenSupply, &chain.BondingCurveSlope)

	if err != nil {
		return nil, err
	}

	chain.ID = result.ID
	chain.CreatedAt = result.CreatedAt
	chain.UpdatedAt = result.UpdatedAt

	return chain, nil
}

// ChainKeyFixture provides test chain key creation
type ChainKeyFixture struct {
	ChainID             uuid.UUID
	Address             string
	PublicKey           []byte
	EncryptedPrivateKey string
	Salt                []byte
	KeyPurpose          string
	IsActive            bool
}

// DefaultChainKey returns a chain key fixture with default values
func DefaultChainKey(chainID uuid.UUID) *ChainKeyFixture {
	return &ChainKeyFixture{
		ChainID:             chainID,
		Address:             chainID.String()[:16], // First 16 chars as hex address
		PublicKey:           []byte{0x01, 0x02, 0x03, 0x04},
		EncryptedPrivateKey: "encrypted_test_key_" + chainID.String()[:8],
		Salt:                []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10},
		KeyPurpose:          "treasury",
		IsActive:            true,
	}
}

// WithAddress sets a custom address
func (k *ChainKeyFixture) WithAddress(address string) *ChainKeyFixture {
	k.Address = address
	return k
}

// WithPurpose sets the key purpose
func (k *ChainKeyFixture) WithPurpose(purpose string) *ChainKeyFixture {
	k.KeyPurpose = purpose
	return k
}

// Create persists the chain key to the database (works with *sqlx.DB or *sqlx.Tx)
func (k *ChainKeyFixture) Create(ctx context.Context, db sqlx.ExtContext) (*models.ChainKey, error) {
	query := `
		INSERT INTO chain_keys (
			chain_id, address, public_key, encrypted_private_key, salt,
			key_purpose, is_active
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`

	key := &models.ChainKey{
		ChainID:             k.ChainID,
		Address:             k.Address,
		PublicKey:           k.PublicKey,
		EncryptedPrivateKey: k.EncryptedPrivateKey,
		Salt:                k.Salt,
		KeyPurpose:          k.KeyPurpose,
		IsActive:            k.IsActive,
	}

	result := struct {
		ID        uuid.UUID `db:"id"`
		CreatedAt time.Time `db:"created_at"`
		UpdatedAt time.Time `db:"updated_at"`
	}{}

	err := sqlx.GetContext(ctx, db, &result, query,
		key.ChainID, key.Address, key.PublicKey, key.EncryptedPrivateKey,
		key.Salt, key.KeyPurpose, key.IsActive)

	if err != nil {
		return nil, err
	}

	key.ID = result.ID
	key.CreatedAt = result.CreatedAt
	key.UpdatedAt = result.UpdatedAt

	return key, nil
}

// VirtualPoolFixture provides test virtual pool creation
type VirtualPoolFixture struct {
	ChainID           uuid.UUID
	CNPYReserve       float64
	TokenReserve      int64
	CurrentPriceCNPY  float64
	TotalTransactions int
	IsActive          bool
}

// DefaultVirtualPool returns a virtual pool fixture with default values
func DefaultVirtualPool(chainID uuid.UUID) *VirtualPoolFixture {
	cnpyReserve := 100.0
	tokenReserve := int64(800000000)
	currentPrice := cnpyReserve / float64(tokenReserve)

	return &VirtualPoolFixture{
		ChainID:           chainID,
		CNPYReserve:       cnpyReserve,
		TokenReserve:      tokenReserve,
		CurrentPriceCNPY:  currentPrice,
		TotalTransactions: 0,
		IsActive:          true,
	}
}

// WithReserves sets the pool reserves
func (p *VirtualPoolFixture) WithReserves(cnpy float64, tokens int64) *VirtualPoolFixture {
	p.CNPYReserve = cnpy
	p.TokenReserve = tokens
	p.CurrentPriceCNPY = cnpy / float64(tokens)
	return p
}

// Create persists the virtual pool to the database (works with *sqlx.DB or *sqlx.Tx)
func (p *VirtualPoolFixture) Create(ctx context.Context, db sqlx.ExtContext) (*models.VirtualPool, error) {
	query := `
		INSERT INTO virtual_pools (
			chain_id, cnpy_reserve, token_reserve, current_price_cnpy,
			total_transactions, is_active
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	pool := &models.VirtualPool{
		ChainID:           p.ChainID,
		CNPYReserve:       p.CNPYReserve,
		TokenReserve:      p.TokenReserve,
		CurrentPriceCNPY:  p.CurrentPriceCNPY,
		TotalTransactions: p.TotalTransactions,
		IsActive:          p.IsActive,
	}

	result := struct {
		ID        uuid.UUID `db:"id"`
		CreatedAt time.Time `db:"created_at"`
		UpdatedAt time.Time `db:"updated_at"`
	}{}

	err := sqlx.GetContext(ctx, db, &result, query,
		pool.ChainID, pool.CNPYReserve, pool.TokenReserve, pool.CurrentPriceCNPY,
		pool.TotalTransactions, pool.IsActive)

	if err != nil {
		return nil, err
	}

	pool.ID = result.ID
	pool.CreatedAt = result.CreatedAt
	pool.UpdatedAt = result.UpdatedAt

	return pool, nil
}

// UserPositionFixture provides test user position creation
type UserPositionFixture struct {
	UserID                uuid.UUID
	ChainID               uuid.UUID
	VirtualPoolID         uuid.UUID
	TokenBalance          int64
	TotalCNPYInvested     float64
	AverageEntryPriceCNPY float64
	UnrealizedPnlCNPY     float64
	RealizedPnlCNPY       float64
	IsActive              bool
}

// DefaultUserPosition returns a user position fixture with default values
func DefaultUserPosition(userID, chainID, poolID uuid.UUID) *UserPositionFixture {
	return &UserPositionFixture{
		UserID:                userID,
		ChainID:               chainID,
		VirtualPoolID:         poolID,
		TokenBalance:          0,
		TotalCNPYInvested:     0,
		AverageEntryPriceCNPY: 0,
		UnrealizedPnlCNPY:     0,
		RealizedPnlCNPY:       0,
		IsActive:              true,
	}
}

// WithPosition sets the position details
func (p *UserPositionFixture) WithPosition(tokenBalance int64, cnpyInvested, entryPrice float64) *UserPositionFixture {
	p.TokenBalance = tokenBalance
	p.TotalCNPYInvested = cnpyInvested
	p.AverageEntryPriceCNPY = entryPrice
	return p
}

// Create persists the user position to the database (works with *sqlx.DB or *sqlx.Tx)
func (p *UserPositionFixture) Create(ctx context.Context, db sqlx.ExtContext) (*models.UserVirtualLPPosition, error) {
	query := `
		INSERT INTO user_virtual_positions (
			user_id, chain_id, virtual_pool_id, token_balance, total_cnpy_invested,
			average_entry_price_cnpy, unrealized_pnl_cnpy, realized_pnl_cnpy, is_active
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`

	position := &models.UserVirtualLPPosition{
		UserID:                p.UserID,
		ChainID:               p.ChainID,
		VirtualPoolID:         p.VirtualPoolID,
		TokenBalance:          p.TokenBalance,
		TotalCNPYInvested:     p.TotalCNPYInvested,
		AverageEntryPriceCNPY: p.AverageEntryPriceCNPY,
		UnrealizedPnlCNPY:     p.UnrealizedPnlCNPY,
		RealizedPnlCNPY:       p.RealizedPnlCNPY,
		IsActive:              p.IsActive,
	}

	result := struct {
		ID        uuid.UUID `db:"id"`
		CreatedAt time.Time `db:"created_at"`
		UpdatedAt time.Time `db:"updated_at"`
	}{}

	err := sqlx.GetContext(ctx, db, &result, query,
		position.UserID, position.ChainID, position.VirtualPoolID,
		position.TokenBalance, position.TotalCNPYInvested, position.AverageEntryPriceCNPY,
		position.UnrealizedPnlCNPY, position.RealizedPnlCNPY, position.IsActive)

	if err != nil {
		return nil, err
	}

	position.ID = result.ID
	position.CreatedAt = result.CreatedAt
	position.UpdatedAt = result.UpdatedAt

	return position, nil
}

// ChainTemplateFixture provides test template creation
type ChainTemplateFixture struct {
	TemplateName        string
	TemplateDescription string
	TemplateCategory    string
	SupportedLanguage   string
	DefaultTokenSupply  int64
	Version             string
	IsActive            bool
}

// DefaultChainTemplate returns a template fixture with default values
func DefaultChainTemplate() *ChainTemplateFixture {
	return &ChainTemplateFixture{
		TemplateName:        "Test Template",
		TemplateDescription: "Template for integration tests",
		TemplateCategory:    "general",
		SupportedLanguage:   "go",
		DefaultTokenSupply:  1000000000,
		Version:             "1.0.0",
		IsActive:            true,
	}
}

// WithCategory sets the template category
func (t *ChainTemplateFixture) WithCategory(category string) *ChainTemplateFixture {
	t.TemplateCategory = category
	return t
}

// WithName sets the template name
func (t *ChainTemplateFixture) WithName(name string) *ChainTemplateFixture {
	t.TemplateName = name
	return t
}

// Create persists the template to the database (works with *sqlx.DB or *sqlx.Tx)
func (t *ChainTemplateFixture) Create(ctx context.Context, db sqlx.ExtContext) (*models.ChainTemplate, error) {
	query := `
		INSERT INTO chain_templates (
			template_name, template_description, template_category, supported_language,
			default_token_supply, version, is_active
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`

	template := &models.ChainTemplate{
		TemplateName:        t.TemplateName,
		TemplateDescription: t.TemplateDescription,
		TemplateCategory:    t.TemplateCategory,
		SupportedLanguage:   t.SupportedLanguage,
		DefaultTokenSupply:  t.DefaultTokenSupply,
		Version:             t.Version,
		IsActive:            t.IsActive,
	}

	result := struct {
		ID        uuid.UUID `db:"id"`
		CreatedAt time.Time `db:"created_at"`
		UpdatedAt time.Time `db:"updated_at"`
	}{}

	err := sqlx.GetContext(ctx, db, &result, query,
		template.TemplateName, template.TemplateDescription, template.TemplateCategory,
		template.SupportedLanguage, template.DefaultTokenSupply,
		template.Version, template.IsActive)

	if err != nil {
		return nil, err
	}

	template.ID = result.ID
	template.CreatedAt = result.CreatedAt
	template.UpdatedAt = result.UpdatedAt

	return template, nil
}

// VirtualPoolTransactionFixture provides test virtual pool transaction creation
type VirtualPoolTransactionFixture struct {
	VirtualPoolID         uuid.UUID
	ChainID               uuid.UUID
	UserID                uuid.UUID
	TransactionType       string
	CNPYAmount            float64
	TokenAmount           int64
	PricePerTokenCNPY     float64
	TradingFeeCNPY        float64
	SlippagePercent       float64
	TransactionHash       *string
	BlockHeight           *int64
	GasUsed               *int
	PoolCNPYReserveAfter  float64
	PoolTokenReserveAfter int64
	MarketCapAfterUSD     float64
}

// DefaultVirtualPoolTransaction returns a transaction fixture with default values
func DefaultVirtualPoolTransaction(chainID, userID uuid.UUID) *VirtualPoolTransactionFixture {
	txHash := "0x" + time.Now().Format("20060102150405")
	blockHeight := int64(100)
	gasUsed := 21000

	return &VirtualPoolTransactionFixture{
		VirtualPoolID:         uuid.New(), // Will be set by the test
		ChainID:               chainID,
		UserID:                userID,
		TransactionType:       "buy",
		CNPYAmount:            10.0,
		TokenAmount:           1000000,
		PricePerTokenCNPY:     0.00001,
		TradingFeeCNPY:        0.1,
		SlippagePercent:       0.5,
		TransactionHash:       &txHash,
		BlockHeight:           &blockHeight,
		GasUsed:               &gasUsed,
		PoolCNPYReserveAfter:  1010.0,
		PoolTokenReserveAfter: 799000000,
		MarketCapAfterUSD:     0,
	}
}

// WithTransactionType sets the transaction type
func (t *VirtualPoolTransactionFixture) WithTransactionType(txType string) *VirtualPoolTransactionFixture {
	t.TransactionType = txType
	return t
}

// WithCNPYAmount sets the CNPY amount
func (t *VirtualPoolTransactionFixture) WithCNPYAmount(amount float64) *VirtualPoolTransactionFixture {
	t.CNPYAmount = amount
	return t
}

// WithTokenAmount sets the token amount
func (t *VirtualPoolTransactionFixture) WithTokenAmount(amount int64) *VirtualPoolTransactionFixture {
	t.TokenAmount = amount
	return t
}

// WithVirtualPoolID sets the virtual pool ID
func (t *VirtualPoolTransactionFixture) WithVirtualPoolID(poolID uuid.UUID) *VirtualPoolTransactionFixture {
	t.VirtualPoolID = poolID
	return t
}

// Create persists the transaction to the database
func (t *VirtualPoolTransactionFixture) Create(ctx context.Context, db sqlx.ExtContext) (*models.VirtualPoolTransaction, error) {
	query := `
		INSERT INTO virtual_pool_transactions (
			virtual_pool_id, chain_id, user_id, transaction_type, cnpy_amount, token_amount,
			price_per_token_cnpy, trading_fee_cnpy, slippage_percent, transaction_hash,
			block_height, gas_used, pool_cnpy_reserve_after, pool_token_reserve_after,
			market_cap_after_usd
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id, created_at
	`

	result := struct {
		ID        uuid.UUID `db:"id"`
		CreatedAt time.Time `db:"created_at"`
	}{}

	err := sqlx.GetContext(ctx, db, &result, query,
		t.VirtualPoolID, t.ChainID, t.UserID, t.TransactionType, t.CNPYAmount,
		t.TokenAmount, t.PricePerTokenCNPY, t.TradingFeeCNPY, t.SlippagePercent,
		t.TransactionHash, t.BlockHeight, t.GasUsed, t.PoolCNPYReserveAfter,
		t.PoolTokenReserveAfter, t.MarketCapAfterUSD)

	if err != nil {
		return nil, err
	}

	transaction := &models.VirtualPoolTransaction{
		ID:                    result.ID,
		VirtualPoolID:         t.VirtualPoolID,
		ChainID:               t.ChainID,
		UserID:                t.UserID,
		TransactionType:       t.TransactionType,
		CNPYAmount:            t.CNPYAmount,
		TokenAmount:           t.TokenAmount,
		PricePerTokenCNPY:     t.PricePerTokenCNPY,
		TradingFeeCNPY:        t.TradingFeeCNPY,
		SlippagePercent:       t.SlippagePercent,
		TransactionHash:       t.TransactionHash,
		BlockHeight:           t.BlockHeight,
		GasUsed:               t.GasUsed,
		PoolCNPYReserveAfter:  t.PoolCNPYReserveAfter,
		PoolTokenReserveAfter: t.PoolTokenReserveAfter,
		MarketCapAfterUSD:     t.MarketCapAfterUSD,
		CreatedAt:             result.CreatedAt,
	}

	return transaction, nil
}

// SampleDataIDs contains UUIDs from sample_data.sql for reference
var SampleDataIDs = struct {
	Users struct {
		Alice uuid.UUID
		Bob   uuid.UUID
		Carol uuid.UUID
		Dave  uuid.UUID
	}
	Templates struct {
		DeFiStandard    uuid.UUID
		GamingHub       uuid.UUID
		EnterpriseGrade uuid.UUID
		SocialNetwork   uuid.UUID
		BasicChain      uuid.UUID
	}
	Chains struct {
		DeFiSwapPro           uuid.UUID
		MetaRealmGaming       uuid.UUID
		SupplyTraceEnterprise uuid.UUID
		SocialVerse           uuid.UUID
		TestChainAlpha        uuid.UUID
	}
}{
	Users: struct {
		Alice uuid.UUID
		Bob   uuid.UUID
		Carol uuid.UUID
		Dave  uuid.UUID
	}{
		Alice: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
		Bob:   uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
		Carol: uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
		Dave:  uuid.MustParse("550e8400-e29b-41d4-a716-446655440003"),
	},
	Templates: struct {
		DeFiStandard    uuid.UUID
		GamingHub       uuid.UUID
		EnterpriseGrade uuid.UUID
		SocialNetwork   uuid.UUID
		BasicChain      uuid.UUID
	}{
		DeFiStandard:    uuid.MustParse("550e8400-e29b-41d4-a716-446655441001"),
		GamingHub:       uuid.MustParse("550e8400-e29b-41d4-a716-446655441002"),
		EnterpriseGrade: uuid.MustParse("550e8400-e29b-41d4-a716-446655441003"),
		SocialNetwork:   uuid.MustParse("550e8400-e29b-41d4-a716-446655441004"),
		BasicChain:      uuid.MustParse("550e8400-e29b-41d4-a716-446655441005"),
	},
	Chains: struct {
		DeFiSwapPro           uuid.UUID
		MetaRealmGaming       uuid.UUID
		SupplyTraceEnterprise uuid.UUID
		SocialVerse           uuid.UUID
		TestChainAlpha        uuid.UUID
	}{
		DeFiSwapPro:           uuid.MustParse("550e8400-e29b-41d4-a716-446655442001"),
		MetaRealmGaming:       uuid.MustParse("550e8400-e29b-41d4-a716-446655442002"),
		SupplyTraceEnterprise: uuid.MustParse("550e8400-e29b-41d4-a716-446655442003"),
		SocialVerse:           uuid.MustParse("550e8400-e29b-41d4-a716-446655442004"),
		TestChainAlpha:        uuid.MustParse("550e8400-e29b-41d4-a716-446655442005"),
	},
}
