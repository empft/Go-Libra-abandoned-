package connector

type BCQuery struct {
	provider map[string]BCTransaction
}

func NewBCQuery() *BCQuery {
	return &BCQuery{
		provider: make(map[string]BCTransaction),
	}
}

func (bc *BCQuery) Register(name string, provider BCTransaction) {
	if provider == nil {
		panic("framework: BCQuery provider is nil")
	}
	if _, dup := bc.provider[name]; dup {
		panic("framework: BCQuery called twice for provider " + name)
	}
	bc.provider[name] = provider
}

func (bc *BCQuery) TransactionByWallet(chain string, address string) {

}

type BCTransaction interface {
	Transaction(address string)
}