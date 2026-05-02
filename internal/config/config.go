package config

import (
	"encoding/json"
	"os"
)


type AssetType string

type PortfolioAsset struct {
	Name   string    `json:"name"`
	Symbol string    `json:"symbol"`
	Shares float64   `json:"shares"`
	Type   AssetType `json:"-"`
}

type Account struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

type Portfolio struct {
	Stocks  []PortfolioAsset `json:"stocks"`
	Cryptos []PortfolioAsset `json:"cryptos"`
}

type Config struct {
	Portfolio Portfolio `json:"portfolio"`
	Accounts  []Account `json:"accounts"`
}


const (
	AssetTypeStock  AssetType = "stock"
	AssetTypeCrypto AssetType = "crypto"
)


func (p Portfolio) AllAssets() []PortfolioAsset {
	out := make([]PortfolioAsset, 0, len(p.Stocks)+len(p.Cryptos))
	for _, a := range p.Stocks {
		a.Type = AssetTypeStock
		out = append(out, a)
	}
	for _, a := range p.Cryptos {
		a.Type = AssetTypeCrypto
		out = append(out, a)
	}
	return out
}

func Load(path string) (Config, error) {
	var cfg Config
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	err = json.Unmarshal(data, &cfg)
	return cfg, err
}
