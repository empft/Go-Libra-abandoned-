package framework

import (
	"testing"
)

var logurl = ("https://alfajores-blockscout.celo-testnet.org/api")
var nodeurl = "https://alfajores-forno.celo-testnet.org"

var knownAddress = "5dB994389C7495996A313F830f21015B9F8127B0"

func TestBalance(t *testing.T) {
	var handler, err = NewCeloHandler(nodeurl)
	if err != nil {
		t.Error(err)
	}

	bal, err := handler.Balance(knownAddress)
	if err != nil {
		t.Error(err)
	}

	if bal.CELO == nil || bal.CEUR == nil || bal.CUSD == nil {
		t.Error("failed to get balance")
	}

	t.Logf("\nCelo Balance: %s", bal.CELO)
	t.Logf("\ncUSD Balance: %s", bal.CUSD)
	t.Logf("\ncEUR Balance: %s", bal.CEUR)
}