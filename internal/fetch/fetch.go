package fetch

import (
	"context"
	"fmt"
	"sort"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"invest-tracker-tui/internal/config"
	"invest-tracker-tui/internal/utils"
)

// Quote holds a live-priced asset with its computed portfolio value.
type Quote struct {
	Name          string
	Symbol        string
	Type          config.AssetType
	Shares        float64
	Price         float64
	ChangePercent float64
	Value         float64
}

// QuotesMsg is the Bubbletea message returned after fetching all portfolio quotes.
type QuotesMsg struct {
	Quotes         []Quote
	PortfolioTotal float64
	AccountsTotal  float64
	Total          float64
	Err            error
}

// yahooMeta maps the fields we need from the Yahoo Finance v8 chart meta object.
type yahooMeta struct {
	Currency           string  `json:"currency"`
	RegularMarketPrice float64 `json:"regularMarketPrice"`
	PreviousClose      float64 `json:"previousClose"`
	ChartPreviousClose float64 `json:"chartPreviousClose"`
}

type yahooChartResponse struct {
	Chart struct {
		Result []struct {
			Meta yahooMeta `json:"meta"`
		} `json:"result"`
	} `json:"chart"`
}

// QuotesCmd returns a Bubbletea command that fetches live quotes for all portfolio assets.
func QuotesCmd(cfg config.Config) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		eurUSD, err := fetchEURUSD(ctx)
		if err != nil {
			return QuotesMsg{Err: err}
		}

		assets := cfg.Portfolio.AllAssets()
		quotes := make([]Quote, 0, len(assets))

		for _, asset := range assets {
			q, err := fetchYahooQuote(ctx, asset, eurUSD)
			if err != nil {
				return QuotesMsg{Err: fmt.Errorf("%s: %w", asset.Symbol, err)}
			}
			quotes = append(quotes, q)
		}

		sort.Slice(quotes, func(i, j int) bool {
			return quotes[i].Name < quotes[j].Name
		})

		var portfolioTotal float64
		for _, q := range quotes {
			portfolioTotal += q.Value
		}

		var accountsTotal float64
		for _, a := range cfg.Accounts {
			accountsTotal += a.Value
		}

		return QuotesMsg{
			Quotes:         quotes,
			PortfolioTotal: portfolioTotal,
			AccountsTotal:  accountsTotal,
			Total:          portfolioTotal + accountsTotal,
		}
	}
}

func fetchEURUSD(ctx context.Context) (float64, error) {
	meta, err := fetchYahooMeta(ctx, "EURUSD=X")
	if err != nil {
		return 0, err
	}
	if meta.RegularMarketPrice == 0 {
		return 0, fmt.Errorf("EURUSD=X returned zero")
	}
	return meta.RegularMarketPrice, nil
}

func fetchYahooQuote(ctx context.Context, asset config.PortfolioAsset, eurUSD float64) (Quote, error) {
	meta, err := fetchYahooMeta(ctx, asset.Symbol)
	if err != nil {
		return Quote{}, err
	}

	price := meta.RegularMarketPrice
	prev := meta.PreviousClose
	if prev == 0 {
		prev = meta.ChartPreviousClose
	}
	if prev == 0 {
		return Quote{}, fmt.Errorf("previous close unavailable")
	}

	if asset.Type == config.AssetTypeCrypto {
		price /= eurUSD
		prev /= eurUSD
	}

	changePct := ((price - prev) / prev) * 100

	return Quote{
		Name:          asset.Name,
		Symbol:        asset.Symbol,
		Type:          asset.Type,
		Shares:        asset.Shares,
		Price:         utils.Round(price, 2),
		ChangePercent: utils.Round(changePct, 2),
		Value:         utils.Round(price*asset.Shares, 2),
	}, nil
}

func fetchYahooMeta(ctx context.Context, symbol string) (*yahooMeta, error) {
	url := fmt.Sprintf(
		"https://query1.finance.yahoo.com/v8/finance/chart/%s?interval=1d&range=1d",
		symbol,
	)

	body, err := fetchRaw(ctx, url)
	if err != nil {
		return nil, err
	}

	var payload yahooChartResponse
	if err := decodeJSON(body, &payload); err != nil {
		return nil, err
	}

	if len(payload.Chart.Result) == 0 {
		return nil, fmt.Errorf("no result for %s", symbol)
	}

	return &payload.Chart.Result[0].Meta, nil
}
