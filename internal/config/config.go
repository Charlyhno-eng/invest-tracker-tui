package config

import (
	"encoding/json"
	"os"
)


type AssetType string

const (
	AssetTypeStock  AssetType = "stock"
	AssetTypeCrypto AssetType = "crypto"
	AssetTypeETF    AssetType = "etf"
)

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


// AllAssets returns stocks and cryptos merged into a single slice with Type set.
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

type Config struct {
	Portfolio Portfolio `json:"portfolio"`
	Accounts  []Account `json:"accounts"`
}

// Load reads and parses the portfolio JSON at path.
func Load(path string) (Config, error) {
	var cfg Config
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	err = json.Unmarshal(data, &cfg)
	return cfg, err
}

// WatchItem is a single entry in a watchlist sector.
type WatchItem struct {
	Name   string    `json:"name"`
	Symbol string    `json:"symbol"`
	Type   AssetType `json:"type"`
}

// WatchSector groups watchlist items under a label (e.g. "Tech", "Crypto").
type WatchSector struct {
	Name  string      `json:"name"`
	Items []WatchItem `json:"items"`
}

// Watchlist is the top-level structure of watchlist.json.
type Watchlist struct {
	Sectors []WatchSector `json:"sectors"`
}

// LoadWatchlist reads and parses the watchlist JSON at path.
func LoadWatchlist(path string) (Watchlist, error) {
	var wl Watchlist
	data, err := os.ReadFile(path)
	if err != nil {
		return wl, err
	}
	err = json.Unmarshal(data, &wl)
	return wl, err
}
