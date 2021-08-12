package celo

import (

	"github.com/celo-org/celo-blockchain/ethclient"
	"github.com/celo-org/celo-blockchain/rpc"
	"gitlab.com/stevealexrs/celo-explorer-client-go/celoexplorer"
)

type Handler struct {
	rpc *rpc.Client
	Eth *ethclient.Client
	Explorer *celoexplorer.Client
}

func NewHandler(url string) (*Handler, error) {
	rpcClient, err := rpc.Dial(url)
	if err != nil {
		return nil, err
	}
	ethClient := ethclient.NewClient(rpcClient)

	explorer := celoexplorer.New(url)

	return &Handler{rpc: rpcClient, Eth: ethClient, Explorer: explorer}, nil
}