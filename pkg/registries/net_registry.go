package registries

type NativeAsset struct {
	Name     string
	Symbol   string
	Decimals int
}

var NetWorkRegistry = map[string]NativeAsset{
	"eth-mainnet":      {Name: "Ethereum", Symbol: "ETH", Decimals: 18},
	"bnb-mainnet":      {Name: "BNB", Symbol: "BNB", Decimals: 18},
	"solana-mainnet":   {Name: "Solana", Symbol: "SOL", Decimals: 9},
	"Arbitrum-mainnet": {Name: "Arbitrum", Symbol: "ARB", Decimals: 18},
}
