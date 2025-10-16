package canopy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/canopy-network/canopy/fsm"
	"github.com/canopy-network/canopy/lib"
	"github.com/canopy-network/canopy/lib/crypto"
)

type Client struct {
	rpcURL string
	client http.Client
}

func NewClient(rpcURL string) *Client {
	return &Client{rpcURL: rpcURL, client: http.Client{}}
}

func (c *Client) Version() (version *string, err lib.ErrorI) {
	version = new(string)
	err = c.get(VersionRouteName, "", version)
	return
}

func (c *Client) Height() (p *uint64, err lib.ErrorI) {
	p = new(uint64)
	err = c.post(HeightRouteName, nil, p)
	return
}

func (c *Client) BlockByHeight(height uint64) (p *lib.BlockResult, err lib.ErrorI) {
	p = new(lib.BlockResult)
	err = c.heightRequest(BlockByHeightRouteName, height, p)
	return
}

func (c *Client) BlockByHash(hash string) (p *lib.BlockResult, err lib.ErrorI) {
	p = new(lib.BlockResult)
	err = c.hashRequest(BlockByHashRouteName, hash, p)
	return
}

func (c *Client) Blocks(params lib.PageParams) (p *lib.Page, err lib.ErrorI) {
	p = new(lib.Page)
	err = c.paginatedHeightRequest(BlocksRouteName, 0, params, p)
	return
}

func (c *Client) Pending(params lib.PageParams) (p *lib.Page, err lib.ErrorI) {
	p = new(lib.Page)
	err = c.paginatedAddrRequest(PendingRouteName, "", params, p)
	return
}

func (c *Client) Proposals() (p *fsm.GovProposals, err lib.ErrorI) {
	p = new(fsm.GovProposals)
	err = c.get(ProposalsRouteName, "", p)
	return
}

func (c *Client) Poll() (p *fsm.Poll, err lib.ErrorI) {
	p = new(fsm.Poll)
	err = c.get(PollRouteName, "", p)
	return
}

func (c *Client) CertByHeight(height uint64) (p *lib.QuorumCertificate, err lib.ErrorI) {
	p = new(lib.QuorumCertificate)
	err = c.heightRequest(CertByHeightRouteName, height, p)
	return
}

func (c *Client) TransactionByHash(hash string) (p *lib.TxResult, err lib.ErrorI) {
	p = new(lib.TxResult)
	err = c.hashRequest(TxByHashRouteName, hash, p)
	return
}

func (c *Client) TransactionsByHeight(height uint64, params lib.PageParams) (p *lib.Page, err lib.ErrorI) {
	p = new(lib.Page)
	err = c.paginatedHeightRequest(TxsByHeightRouteName, height, params, p)
	return
}

func (c *Client) TransactionsBySender(address string, params lib.PageParams) (p *lib.Page, err lib.ErrorI) {
	p = new(lib.Page)
	err = c.paginatedAddrRequest(TxsBySenderRouteName, address, params, p)
	return
}

func (c *Client) TransactionsByRecipient(address string, params lib.PageParams) (p *lib.Page, err lib.ErrorI) {
	p = new(lib.Page)
	err = c.paginatedAddrRequest(TxsByRecRouteName, address, params, p)
	return
}

func (c *Client) Account(height uint64, address string) (p *fsm.Account, err lib.ErrorI) {
	p = new(fsm.Account)
	err = c.heightAndAddressRequest(AccountRouteName, height, address, p)
	return
}

func (c *Client) Accounts(height uint64, params lib.PageParams) (p *lib.Page, err lib.ErrorI) {
	p = new(lib.Page)
	err = c.paginatedHeightRequest(AccountsRouteName, height, params, p)
	return
}

func (c *Client) Pool(height uint64, id uint64) (p *fsm.Pool, err lib.ErrorI) {
	p = new(fsm.Pool)
	err = c.heightAndIdRequest(PoolRouteName, height, id, p)
	return
}

func (c *Client) Pools(height uint64, params lib.PageParams) (p *lib.Page, err lib.ErrorI) {
	p = new(lib.Page)
	err = c.paginatedHeightRequest(PoolsRouteName, height, params, p)
	return
}

func (c *Client) Validator(height uint64, address string) (p *fsm.Validator, err lib.ErrorI) {
	p = new(fsm.Validator)
	err = c.heightAndAddressRequest(ValidatorRouteName, height, address, p)
	return
}

func (c *Client) Validators(height uint64, params lib.PageParams, filter lib.ValidatorFilters) (p *lib.Page, err lib.ErrorI) {
	p = new(lib.Page)
	err = c.paginatedHeightRequest(ValidatorsRouteName, height, params, p, filter)
	return
}

func (c *Client) Committee(height uint64, id uint64, params lib.PageParams) (p *lib.Page, err lib.ErrorI) {
	p = new(lib.Page)
	err = c.paginatedHeightRequest(CommitteeRouteName, height, params, p, lib.ValidatorFilters{Committee: id})
	return
}

func (c *Client) CommitteeData(height uint64, id uint64) (p *lib.CommitteeData, err lib.ErrorI) {
	p = new(lib.CommitteeData)
	err = c.heightAndIdRequest(CommitteeDataRouteName, height, id, p)
	return
}

func (c *Client) CommitteesData(height uint64) (p *lib.CommitteesData, err lib.ErrorI) {
	p = new(lib.CommitteesData)
	err = c.paginatedHeightRequest(CommitteesDataRouteName, height, lib.PageParams{}, p)
	return
}

func (c *Client) RootChainInfo(height, chainId uint64) (p *lib.RootChainInfo, err lib.ErrorI) {
	p = new(lib.RootChainInfo)
	err = c.heightAndIdRequest(RootChainInfoRouteName, height, chainId, p)
	return
}

func (c *Client) SubsidizedCommittees(height uint64) (p *[]uint64, err lib.ErrorI) {
	p = new([]uint64)
	err = c.heightRequest(SubsidizedCommitteesRouteName, height, p)
	return
}

func (c *Client) RetiredCommittees(height uint64) (p *[]uint64, err lib.ErrorI) {
	p = new([]uint64)
	err = c.heightRequest(RetiredCommitteesRouteName, height, p)
	return
}

func (c *Client) Order(height uint64, orderId string, chainId uint64) (p *lib.SellOrder, err lib.ErrorI) {
	p = new(lib.SellOrder)
	err = c.orderRequest(OrderRouteName, height, orderId, chainId, p)
	return
}

func (c *Client) Orders(height, chainId uint64) (p *lib.OrderBooks, err lib.ErrorI) {
	p = new(lib.OrderBooks)
	err = c.heightAndIdRequest(OrdersRouteName, height, chainId, p)
	return
}

func (c *Client) DexPrice(height, chainId uint64) (p *lib.DexPrice, err lib.ErrorI) {
	p = new(lib.DexPrice)
	err = c.heightAndIdRequest(DexPriceRouteName, height, chainId, p)
	return
}

func (c *Client) DexBatch(height, chainId uint64, withPoints bool) (p *lib.DexBatch, err lib.ErrorI) {
	p = new(lib.DexBatch)
	err = c.heightIdAndPointsRequest(DexBatchRouteName, height, chainId, withPoints, p)
	return
}

func (c *Client) NextDexBatch(height, chainId uint64, withPoints bool) (p *lib.DexBatch, err lib.ErrorI) {
	p = new(lib.DexBatch)
	err = c.heightIdAndPointsRequest(NextDexBatchRouteName, height, chainId, withPoints, p)
	return
}

func (c *Client) LastProposers(height uint64) (p *lib.Proposers, err lib.ErrorI) {
	p = new(lib.Proposers)
	err = c.heightRequest(LastProposersRouteName, height, p)
	return
}

func (c *Client) IsValidDoubleSigner(height uint64, address string) (p *bool, err lib.ErrorI) {
	p = new(bool)
	err = c.heightAndAddressRequest(IsValidDoubleSignerRouteName, height, address, p)
	return
}

func (c *Client) Checkpoint(height, id uint64) (p lib.HexBytes, err lib.ErrorI) {
	p = make(lib.HexBytes, 0)
	err = c.heightAndIdRequest(CheckpointRouteName, height, id, &p)
	return
}

func (c *Client) DoubleSigners(height uint64) (p *[]*lib.DoubleSigner, err lib.ErrorI) {
	p = new([]*lib.DoubleSigner)
	err = c.heightRequest(DoubleSignersRouteName, height, p)
	return
}

func (c *Client) ValidatorSet(height uint64, id uint64) (v lib.ValidatorSet, err lib.ErrorI) {
	p := new(lib.ConsensusValidators)
	err = c.heightAndIdRequest(ValidatorSetRouteName, height, id, p)
	if err != nil {
		return lib.ValidatorSet{}, err
	}
	return lib.NewValidatorSet(p)
}

func (c *Client) MinimumEvidenceHeight(height uint64) (p *uint64, err lib.ErrorI) {
	p = new(uint64)
	err = c.heightRequest(MinimumEvidenceHeightRouteName, height, p)
	return
}

func (c *Client) Lottery(height, id uint64) (p *lib.LotteryWinner, err lib.ErrorI) {
	p = new(lib.LotteryWinner)
	err = c.heightAndIdRequest(LotteryRouteName, height, id, p)
	return
}

func (c *Client) Supply(height uint64) (p *fsm.Supply, err lib.ErrorI) {
	p = new(fsm.Supply)
	err = c.heightRequest(SupplyRouteName, height, p)
	return
}

func (c *Client) NonSigners(height uint64) (p *fsm.NonSigners, err lib.ErrorI) {
	p = new(fsm.NonSigners)
	err = c.heightRequest(NonSignersRouteName, height, p)
	return
}

func (c *Client) Params(height uint64) (p *fsm.Params, err lib.ErrorI) {
	p = new(fsm.Params)
	err = c.heightRequest(ParamRouteName, height, p)
	return
}

func (c *Client) FeeParams(height uint64) (p *fsm.FeeParams, err lib.ErrorI) {
	p = new(fsm.FeeParams)
	err = c.heightRequest(FeeParamRouteName, height, p)
	return
}

func (c *Client) GovParams(height uint64) (p *fsm.GovernanceParams, err lib.ErrorI) {
	p = new(fsm.GovernanceParams)
	err = c.heightRequest(GovParamRouteName, height, p)
	return
}

func (c *Client) ConParams(height uint64) (p *fsm.ConsensusParams, err lib.ErrorI) {
	p = new(fsm.ConsensusParams)
	err = c.heightRequest(ConParamsRouteName, height, p)
	return
}

func (c *Client) ValParams(height uint64) (p *fsm.ValidatorParams, err lib.ErrorI) {
	p = new(fsm.ValidatorParams)
	err = c.heightRequest(ValParamRouteName, height, p)
	return
}

func (c *Client) State(height uint64) (p *fsm.GenesisState, err lib.ErrorI) {
	var param string
	if height != 0 {
		param = fmt.Sprintf("?height=%d", height)
	}
	p = new(fsm.GenesisState)
	err = c.get(StateRouteName, param, p)
	return
}

func (c *Client) StateDiff(height, startHeight uint64) (diff string, err lib.ErrorI) {
	bz, err := lib.MarshalJSON(heightsRequest{heightRequest: heightRequest{height}, StartHeight: startHeight})
	if err != nil {
		return
	}
	resp, e := c.client.Post(c.url(StateDiffRouteName, ""), ApplicationJSON, bytes.NewBuffer(bz))
	if e != nil {
		return "", lib.ErrPostRequest(e)
	}
	bz, e = io.ReadAll(resp.Body)
	if e != nil {
		return "", lib.ErrReadBody(e)
	}
	diff = string(bz)
	return
}

func (c *Client) TransactionJSON(tx json.RawMessage) (hash *string, err lib.ErrorI) {
	hash = new(string)
	err = c.post(TxRouteName, tx, hash)
	return
}

func (c *Client) Transaction(tx lib.TransactionI) (hash *string, err lib.ErrorI) {
	bz, err := lib.MarshalJSON(tx)
	if err != nil {
		return nil, err
	}
	hash = new(string)
	err = c.post(TxRouteName, bz, hash)
	return
}

func (c *Client) paginatedHeightRequest(routeName string, height uint64, p lib.PageParams, ptr any, filters ...lib.ValidatorFilters) (err lib.ErrorI) {
	var vf lib.ValidatorFilters
	if filters != nil {
		vf = filters[0]
	}
	bz, err := lib.MarshalJSON(paginatedHeightRequest{
		heightRequest:    heightRequest{height},
		PageParams:       p,
		ValidatorFilters: vf,
	})
	if err != nil {
		return
	}
	err = c.post(routeName, bz, ptr)
	return
}

func (c *Client) paginatedAddrRequest(routeName string, address string, p lib.PageParams, ptr any) (err lib.ErrorI) {
	addr, err := lib.StringToBytes(address)
	if err != nil {
		return err
	}
	bz, err := lib.MarshalJSON(paginatedAddressRequest{
		addressRequest: addressRequest{addr},
		PageParams:     p,
	})
	if err != nil {
		return
	}
	err = c.post(routeName, bz, ptr)
	return
}

func (c *Client) heightRequest(routeName string, height uint64, ptr any) (err lib.ErrorI) {
	bz, err := lib.MarshalJSON(heightRequest{Height: height})
	if err != nil {
		return
	}
	err = c.post(routeName, bz, ptr)
	return
}

func (c *Client) orderRequest(routeName string, height uint64, orderId string, chainId uint64, ptr any) (err lib.ErrorI) {
	bz, err := lib.MarshalJSON(orderRequest{
		ChainId: chainId,
		OrderId: orderId,
		heightRequest: heightRequest{
			Height: height,
		},
	})
	if err != nil {
		return
	}
	err = c.post(routeName, bz, ptr)
	return
}

func (c *Client) hashRequest(routeName string, hash string, ptr any) (err lib.ErrorI) {
	bz, err := lib.MarshalJSON(hashRequest{Hash: hash})
	if err != nil {
		return
	}
	err = c.post(routeName, bz, ptr)
	return
}

func (c *Client) heightAndAddressRequest(routeName string, height uint64, address string, ptr any) (err lib.ErrorI) {
	addr, err := lib.StringToBytes(address)
	if err != nil {
		return err
	}
	bz, err := lib.MarshalJSON(heightAndAddressRequest{
		heightRequest:  heightRequest{height},
		addressRequest: addressRequest{addr},
	})
	if err != nil {
		return
	}
	err = c.post(routeName, bz, ptr)
	return
}

func (c *Client) heightAndIdRequest(routeName string, height, id uint64, ptr any) (err lib.ErrorI) {
	bz, err := lib.MarshalJSON(heightAndIdRequest{
		heightRequest: heightRequest{height},
		idRequest:     idRequest{id},
	})
	if err != nil {
		return
	}
	err = c.post(routeName, bz, ptr)
	return
}

func (c *Client) heightIdAndPointsRequest(routeName string, height, id uint64, points bool, ptr any) (err lib.ErrorI) {
	bz, err := lib.MarshalJSON(heightIdAndPointsRequest{
		heightAndIdRequest: heightAndIdRequest{
			heightRequest: heightRequest{height},
			idRequest:     idRequest{id},
		},
		Points: points,
	})
	if err != nil {
		return
	}
	err = c.post(routeName, bz, ptr)
	return
}

func (c *Client) url(routeName, param string) string {
	return c.rpcURL + routePaths[routeName].Path + param
}

func (c *Client) post(routeName string, json []byte, ptr any) lib.ErrorI {
	resp, err := c.client.Post(c.url(routeName, ""), ApplicationJSON, bytes.NewBuffer(json))
	if err != nil {
		return lib.ErrPostRequest(err)
	}
	return c.unmarshal(resp, ptr)
}

func (c *Client) get(routeName, param string, ptr any) lib.ErrorI {
	resp, err := c.client.Get(c.url(routeName, param))
	if err != nil {
		return lib.ErrGetRequest(err)
	}
	return c.unmarshal(resp, ptr)
}

func (c *Client) unmarshal(resp *http.Response, ptr any) lib.ErrorI {
	bz, err := io.ReadAll(resp.Body)
	if err != nil {
		return lib.ErrReadBody(err)
	}
	if resp.StatusCode != http.StatusOK {
		return lib.ErrHttpStatus(resp.Status, resp.StatusCode, bz)
	}
	return lib.UnmarshalJSON(bz, ptr)
}

// Helper function for transaction requests
func getFrom(address, nickname string) (from fromFields, err lib.ErrorI) {
	if address != "" {
		from.Address, err = lib.NewHexBytesFromString(address)
		if err != nil {
			return
		}
	}
	from.Nickname = nickname
	return from, nil
}

// transactionRequest handles transaction submission or preview
func (c *Client) transactionRequest(routeName string, txRequest any, submit bool) (hash *string, tx json.RawMessage, e lib.ErrorI) {
	bz, e := lib.MarshalJSON(txRequest)
	if e != nil {
		return
	}
	if submit {
		hash = new(string)
		e = c.post(routeName, bz, hash)
	} else {
		tx = json.RawMessage{}
		e = c.post(routeName, bz, &tx)
	}
	return
}

// TxDexLimitOrder creates a DEX limit order transaction
func (c *Client) TxDexLimitOrder(from AddrOrNickname, amount, receiveAmount, chainId uint64,
	pwd string, submit bool, optFee uint64) (hash *string, tx json.RawMessage, e lib.ErrorI) {
	txReq := txDexLimitOrder{
		Fee:               optFee,
		Amount:            amount,
		ReceiveAmount:     receiveAmount,
		Submit:            submit,
		Password:          pwd,
		committeesRequest: committeesRequest{fmt.Sprintf("%d", chainId)},
	}
	var err lib.ErrorI
	txReq.fromFields, err = getFrom(from.Address, from.Nickname)
	if err != nil {
		return nil, nil, err
	}
	return c.transactionRequest(TxDexLimitOrderRouteName, txReq, submit)
}

// TxDexLiquidityDeposit creates a DEX liquidity deposit transaction
func (c *Client) TxDexLiquidityDeposit(from AddrOrNickname, amount, chainId uint64,
	pwd string, submit bool, optFee uint64) (hash *string, tx json.RawMessage, e lib.ErrorI) {
	txReq := txDexLiquidityDeposit{
		Fee:               optFee,
		Amount:            amount,
		Submit:            submit,
		Password:          pwd,
		committeesRequest: committeesRequest{fmt.Sprintf("%d", chainId)},
	}
	var err lib.ErrorI
	txReq.fromFields, err = getFrom(from.Address, from.Nickname)
	if err != nil {
		return nil, nil, err
	}
	return c.transactionRequest(TxDexLiquidityDepositRouteName, txReq, submit)
}

// TxDexLiquidityWithdraw creates a DEX liquidity withdrawal transaction
func (c *Client) TxDexLiquidityWithdraw(from AddrOrNickname, percent int, chainId uint64,
	pwd string, submit bool, optFee uint64) (hash *string, tx json.RawMessage, e lib.ErrorI) {
	txReq := txDexLiquidityWithdraw{
		Fee:               optFee,
		Percent:           percent,
		Submit:            submit,
		Password:          pwd,
		committeesRequest: committeesRequest{fmt.Sprintf("%d", chainId)},
	}
	var err lib.ErrorI
	txReq.fromFields, err = getFrom(from.Address, from.Nickname)
	if err != nil {
		return nil, nil, err
	}
	return c.transactionRequest(TxDexLiquidityWithdrawRouteName, txReq, submit)
}

// Keystore methods (optional - can be used if needed)
func (c *Client) Keystore() (keystore *crypto.Keystore, err lib.ErrorI) {
	keystore = new(crypto.Keystore)
	err = c.get(KeystoreRouteName, "", keystore)
	return
}
