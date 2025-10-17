package canopy

import "net/http"

const ApplicationJSON = "application/json"

// Canopy RPC Paths
const (
	VersionRoutePath                = "/v1/"
	TxRoutePath                     = "/v1/tx"
	HeightRoutePath                 = "/v1/query/height"
	AccountRoutePath                = "/v1/query/account"
	AccountsRoutePath               = "/v1/query/accounts"
	PoolRoutePath                   = "/v1/query/pool"
	PoolsRoutePath                  = "/v1/query/pools"
	ValidatorRoutePath              = "/v1/query/validator"
	ValidatorsRoutePath             = "/v1/query/validators"
	CommitteeRoutePath              = "/v1/query/committee"
	CommitteeDataRoutePath          = "/v1/query/committee-data"
	CommitteesDataRoutePath         = "/v1/query/committees-data"
	SubsidizedCommitteesRoutePath   = "/v1/query/subsidized-committees"
	RetiredCommitteesRoutePath      = "/v1/query/retired-committees"
	NonSignersRoutePath             = "/v1/query/non-signers"
	ParamRoutePath                  = "/v1/query/params"
	SupplyRoutePath                 = "/v1/query/supply"
	FeeParamRoutePath               = "/v1/query/fee-params"
	GovParamRoutePath               = "/v1/query/gov-params"
	ConParamsRoutePath              = "/v1/query/con-params"
	ValParamRoutePath               = "/v1/query/val-params"
	StateRoutePath                  = "/v1/query/state"
	StateDiffRoutePath              = "/v1/query/state-diff"
	CertByHeightRoutePath           = "/v1/query/cert-by-height"
	BlockByHeightRoutePath          = "/v1/query/block-by-height"
	BlocksRoutePath                 = "/v1/query/blocks"
	BlockByHashRoutePath            = "/v1/query/block-by-hash"
	TxsByHeightRoutePath            = "/v1/query/txs-by-height"
	TxsBySenderRoutePath            = "/v1/query/txs-by-sender"
	TxsByRecRoutePath               = "/v1/query/txs-by-rec"
	TxByHashRoutePath               = "/v1/query/tx-by-hash"
	OrderRoutePath                  = "/v1/query/order"
	OrdersRoutePath                 = "/v1/query/orders"
	DexPriceRoutePath               = "/v1/query/dex-price"
	DexBatchRoutePath               = "/v1/query/dex-batch"
	NextDexBatchRoutePath           = "/v1/query/next-dex-batch"
	LastProposersRoutePath          = "/v1/query/last-proposers"
	IsValidDoubleSignerRoutePath    = "/v1/query/valid-double-signer"
	DoubleSignersRoutePath          = "/v1/query/double-signers"
	MinimumEvidenceHeightRoutePath  = "/v1/query/minimum-evidence-height"
	LotteryRoutePath                = "/v1/query/lottery"
	PendingRoutePath                = "/v1/query/pending"
	ProposalsRoutePath              = "/v1/gov/proposals"
	PollRoutePath                   = "/v1/gov/poll"
	RootChainInfoRoutePath          = "/v1/query/root-chain-info"
	ValidatorSetRoutePath           = "/v1/query/validator-set"
	CheckpointRoutePath             = "/v1/query/checkpoint"
	KeystoreRoutePath               = "/v1/admin/keystore"
	TxDexLimitOrderRoutePath        = "/v1/admin/tx-dex-limit-order"
	TxDexLiquidityDepositRoutePath  = "/v1/admin/tx-dex-liquidity-deposit"
	TxDexLiquidityWithdrawRoutePath = "/v1/admin/tx-dex-liquidity-withdraw"
)

const (
	VersionRouteName                = "version"
	TxRouteName                     = "tx"
	HeightRouteName                 = "height"
	AccountRouteName                = "account"
	AccountsRouteName               = "accounts"
	PoolRouteName                   = "pool"
	PoolsRouteName                  = "pools"
	ValidatorRouteName              = "validator"
	ValidatorsRouteName             = "validators"
	ValidatorSetRouteName           = "validator-set"
	CommitteeRouteName              = "committee"
	CommitteeDataRouteName          = "committee-data"
	CommitteesDataRouteName         = "committees-data"
	SubsidizedCommitteesRouteName   = "subsidized-committees"
	RetiredCommitteesRouteName      = "retired-committees"
	NonSignersRouteName             = "non-signers"
	SupplyRouteName                 = "supply"
	ParamRouteName                  = "params"
	FeeParamRouteName               = "fee-params"
	GovParamRouteName               = "gov-params"
	ConParamsRouteName              = "con-params"
	ValParamRouteName               = "val-params"
	StateRouteName                  = "state"
	StateDiffRouteName              = "state-diff"
	CertByHeightRouteName           = "cert-by-height"
	BlocksRouteName                 = "blocks"
	BlockByHeightRouteName          = "block-by-height"
	BlockByHashRouteName            = "block-by-hash"
	TxsByHeightRouteName            = "txs-by-height"
	TxsBySenderRouteName            = "txs-by-sender"
	TxsByRecRouteName               = "txs-by-rec"
	TxByHashRouteName               = "tx-by-hash"
	PendingRouteName                = "pending"
	ProposalsRouteName              = "proposals"
	PollRouteName                   = "poll"
	OrderRouteName                  = "order"
	OrdersRouteName                 = "orders"
	DexPriceRouteName               = "dex-price"
	DexBatchRouteName               = "dex-batch"
	NextDexBatchRouteName           = "next-dex-batch"
	LastProposersRouteName          = "last-proposers"
	IsValidDoubleSignerRouteName    = "valid-double-signer"
	DoubleSignersRouteName          = "double-signers"
	MinimumEvidenceHeightRouteName  = "minimum-evidence-height"
	LotteryRouteName                = "lottery"
	RootChainInfoRouteName          = "root-chain-info"
	CheckpointRouteName             = "checkpoint"
	KeystoreRouteName               = "keystore"
	TxDexLimitOrderRouteName        = "tx-dex-limit-order"
	TxDexLiquidityDepositRouteName  = "tx-dex-liquidity-deposit"
	TxDexLiquidityWithdrawRouteName = "tx-dex-liquidity-withdraw"
)

// routes contains the method and path for a canopy command
type routes map[string]struct {
	Method string
	Path   string
}

// routePaths is a mapping from route names to their corresponding HTTP methods and paths.
var routePaths = routes{
	VersionRouteName:                {Method: http.MethodGet, Path: VersionRoutePath},
	TxRouteName:                     {Method: http.MethodPost, Path: TxRoutePath},
	HeightRouteName:                 {Method: http.MethodPost, Path: HeightRoutePath},
	AccountRouteName:                {Method: http.MethodPost, Path: AccountRoutePath},
	AccountsRouteName:               {Method: http.MethodPost, Path: AccountsRoutePath},
	PoolRouteName:                   {Method: http.MethodPost, Path: PoolRoutePath},
	PoolsRouteName:                  {Method: http.MethodPost, Path: PoolsRoutePath},
	ValidatorRouteName:              {Method: http.MethodPost, Path: ValidatorRoutePath},
	ValidatorsRouteName:             {Method: http.MethodPost, Path: ValidatorsRoutePath},
	CommitteeRouteName:              {Method: http.MethodPost, Path: CommitteeRoutePath},
	CommitteeDataRouteName:          {Method: http.MethodPost, Path: CommitteeDataRoutePath},
	CommitteesDataRouteName:         {Method: http.MethodPost, Path: CommitteesDataRoutePath},
	SubsidizedCommitteesRouteName:   {Method: http.MethodPost, Path: SubsidizedCommitteesRoutePath},
	RetiredCommitteesRouteName:      {Method: http.MethodPost, Path: RetiredCommitteesRoutePath},
	NonSignersRouteName:             {Method: http.MethodPost, Path: NonSignersRoutePath},
	ParamRouteName:                  {Method: http.MethodPost, Path: ParamRoutePath},
	SupplyRouteName:                 {Method: http.MethodPost, Path: SupplyRoutePath},
	FeeParamRouteName:               {Method: http.MethodPost, Path: FeeParamRoutePath},
	GovParamRouteName:               {Method: http.MethodPost, Path: GovParamRoutePath},
	ConParamsRouteName:              {Method: http.MethodPost, Path: ConParamsRoutePath},
	ValParamRouteName:               {Method: http.MethodPost, Path: ValParamRoutePath},
	StateRouteName:                  {Method: http.MethodGet, Path: StateRoutePath},
	StateDiffRouteName:              {Method: http.MethodPost, Path: StateDiffRoutePath},
	CertByHeightRouteName:           {Method: http.MethodPost, Path: CertByHeightRoutePath},
	BlockByHeightRouteName:          {Method: http.MethodPost, Path: BlockByHeightRoutePath},
	BlocksRouteName:                 {Method: http.MethodPost, Path: BlocksRoutePath},
	BlockByHashRouteName:            {Method: http.MethodPost, Path: BlockByHashRoutePath},
	TxsByHeightRouteName:            {Method: http.MethodPost, Path: TxsByHeightRoutePath},
	TxsBySenderRouteName:            {Method: http.MethodPost, Path: TxsBySenderRoutePath},
	TxsByRecRouteName:               {Method: http.MethodPost, Path: TxsByRecRoutePath},
	TxByHashRouteName:               {Method: http.MethodPost, Path: TxByHashRoutePath},
	OrderRouteName:                  {Method: http.MethodPost, Path: OrderRoutePath},
	OrdersRouteName:                 {Method: http.MethodPost, Path: OrdersRoutePath},
	DexPriceRouteName:               {Method: http.MethodPost, Path: DexPriceRoutePath},
	DexBatchRouteName:               {Method: http.MethodPost, Path: DexBatchRoutePath},
	NextDexBatchRouteName:           {Method: http.MethodPost, Path: NextDexBatchRoutePath},
	LastProposersRouteName:          {Method: http.MethodPost, Path: LastProposersRoutePath},
	IsValidDoubleSignerRouteName:    {Method: http.MethodPost, Path: IsValidDoubleSignerRoutePath},
	DoubleSignersRouteName:          {Method: http.MethodPost, Path: DoubleSignersRoutePath},
	MinimumEvidenceHeightRouteName:  {Method: http.MethodPost, Path: MinimumEvidenceHeightRoutePath},
	LotteryRouteName:                {Method: http.MethodPost, Path: LotteryRoutePath},
	PendingRouteName:                {Method: http.MethodPost, Path: PendingRoutePath},
	ProposalsRouteName:              {Method: http.MethodGet, Path: ProposalsRoutePath},
	PollRouteName:                   {Method: http.MethodGet, Path: PollRoutePath},
	RootChainInfoRouteName:          {Method: http.MethodPost, Path: RootChainInfoRoutePath},
	ValidatorSetRouteName:           {Method: http.MethodPost, Path: ValidatorSetRoutePath},
	CheckpointRouteName:             {Method: http.MethodPost, Path: CheckpointRoutePath},
	KeystoreRouteName:               {Method: http.MethodGet, Path: KeystoreRoutePath},
	TxDexLimitOrderRouteName:        {Method: http.MethodPost, Path: TxDexLimitOrderRoutePath},
	TxDexLiquidityDepositRouteName:  {Method: http.MethodPost, Path: TxDexLiquidityDepositRoutePath},
	TxDexLiquidityWithdrawRouteName: {Method: http.MethodPost, Path: TxDexLiquidityWithdrawRoutePath},
}
