package framework

import (
	"time"

	"github.com/celo-org/celo-blockchain/core/types"
	"gitlab.com/stevealexrs/celo-client-lite-go/celoclient"
	"gitlab.com/stevealexrs/celo-client-lite-go/contractkit/celotoken"
)

type CeloHandler struct {
	Client celoclient.Client
}

func NewCeloHandler(url string) (*CeloHandler, error) {
	client, err := celoclient.New(url)
	if err != nil {
		return nil, err
	}

	return &CeloHandler{Client: client}, nil
}

func (handler *CeloHandler) Balance(address string) (*celotoken.Balance, error){
	return handler.Client.Balance(address)
}

func (handler *CeloHandler) SubmitSignedTransactionAndWait(signedTxnHex string, timeout time.Duration) (*types.Receipt, error) {
	hash, err := handler.Client.SendRawTransaction(signedTxnHex)
	if err != nil {
		return nil, err
	}

	return handler.Client.WaitForTransaction(hash, timeout)
}