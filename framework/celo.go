package framework

import (
	"math/big"
	"time"

	"github.com/celo-org/celo-blockchain/core/types"
	"gitlab.com/stevealexrs/celo-client-lite-go/celoclient"
	"gitlab.com/stevealexrs/celo-client-lite-go/contractkit/celotoken"
	"gitlab.com/stevealexrs/celo-explorer-client-go/celoexplorer"
	"gitlab.com/stevealexrs/celo-explorer-client-go/request"
)

// TODO: Update celoexplorer dependencies
type CeloHandler struct {
	Client celoclient.Client
	Explorer celoexplorer.Client
}

func NewCeloHandler(url string) (*CeloHandler, error) {
	client, err := celoclient.New(url)
	if err != nil {
		return nil, err
	}

	return &CeloHandler{Client: client}, nil
}

func (h *CeloHandler) Balance(address string) (*celotoken.Balance, error){
	return h.Client.Balance(address)
}

func (h *CeloHandler) SubmitSignedTransactionAndWait(signedTxnHex string, timeout time.Duration) (*types.Receipt, error) {
	hash, err := h.Client.SendRawTransaction(signedTxnHex)
	if err != nil {
		return nil, err
	}

	return h.Client.WaitForTransaction(hash, timeout)
}

func (h *CeloHandler) LatestBlock() (*big.Int, error) {
	return h.Client.LatestBlock()
}

func (h *CeloHandler) Transactions(address, tokenAddress string, blockRange *request.BlockRange) ([]celoexplorer.TokenTransfer, error) {
	return h.Explorer.TokenTx(address, &tokenAddress, nil, blockRange, nil)
}