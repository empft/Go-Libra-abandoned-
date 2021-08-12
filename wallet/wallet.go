package wallet

type Wallet struct {
	Address
	PublicKey string
}

type Address struct {
	Chain 	  string
	Hex   string
}