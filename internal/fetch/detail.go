package fetch

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"invest-tracker-tui/internal/config"
	"invest-tracker-tui/internal/utils"
)

// Detail holds the full market data for a single asset fetched from Yahoo Finance.
type Detail struct {
	Name             string
	Symbol           string
	Currency         string
	Exchange         string
	ExchangeFull     string
	InstrumentType   string
	Timezone         string
	Price            float64
	PreviousClose    float64
	ChangeAbs        float64
	ChangePercent    float64
	DayHigh          float64
	DayLow           float64
	Volume           int64
	FiftyTwoWeekHigh float64
	FiftyTwoWeekLow  float64
	Open             float64
	Shares           float64
	Value            float64
	FetchedAt        time.Time
}

// DetailMsg is the Bubbletea message returned after fetching a single asset detail.
type DetailMsg struct {
	Detail Detail
	Err    error
}

// yahooDetailMeta extends yahooMeta with all additional fields available in the v8 API.
type yahooDetailMeta struct {
	Currency           string  `json:"currency"`
	Symbol             string  `json:"symbol"`
	ExchangeName       string  `json:"exchangeName"`
	FullExchangeName   string  `json:"fullExchangeName"`
	InstrumentType     string  `json:"instrumentType"`
	Timezone           string  `json:"timezone"`
	RegularMarketPrice float64 `json:"regularMarketPrice"`
	PreviousClose      float64 `json:"previousClose"`
	ChartPreviousClose float64 `json:"chartPreviousClose"`
	RegularMarketDayHigh float64 `json:"regularMarketDayHigh"`
	RegularMarketDayLow  float64 `json:"regularMarketDayLow"`
	RegularMarketVolume  int64   `json:"regularMarketVolume"`
	FiftyTwoWeekHigh     float64 `json:"fiftyTwoWeekHigh"`
	FiftyTwoWeekLow      float64 `json:"fiftyTwoWeekLow"`
}

type yahooDetailResponse struct {
	Chart struct {
		Result []struct {
			Meta       yahooDetailMeta `json:"meta"`
			Indicators struct {
				Quote []struct {
					Open []float64 `json:"open"`
				} `json:"quote"`
			} `json:"indicators"`
		} `json:"result"`
	} `json:"chart"`
}

// DetailCmd fetches full market detail for a single asset.
func DetailCmd(asset config.PortfolioAsset) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		eurUSD := 1.0
		if asset.Type == config.AssetTypeCrypto {
			rate, err := fetchEURUSD(ctx)
			if err != nil {
				return DetailMsg{Err: fmt.Errorf("EUR/USD: %w", err)}
			}
			eurUSD = rate
		}

		url := fmt.Sprintf(
			"https://query1.finance.yahoo.com/v8/finance/chart/%s?interval=1d&range=1d",
			asset.Symbol,
		)

		body, err := fetchRaw(ctx, url)
		if err != nil {
			return DetailMsg{Err: err}
		}

		var payload yahooDetailResponse
		if err := decodeJSON(body, &payload); err != nil {
			return DetailMsg{Err: err}
		}

		if len(payload.Chart.Result) == 0 {
			return DetailMsg{Err: fmt.Errorf("no result for %s", asset.Symbol)}
		}

		m := payload.Chart.Result[0].Meta

		prev := m.PreviousClose
		if prev == 0 {
			prev = m.ChartPreviousClose
		}

		price := m.RegularMarketPrice
		high := m.RegularMarketDayHigh
		low := m.RegularMarketDayLow
		wkHigh := m.FiftyTwoWeekHigh
		wkLow := m.FiftyTwoWeekLow

		var open float64
		if len(payload.Chart.Result[0].Indicators.Quote) > 0 {
			opens := payload.Chart.Result[0].Indicators.Quote[0].Open
			if len(opens) > 0 {
				open = opens[0]
			}
		}

		if asset.Type == config.AssetTypeCrypto {
			price /= eurUSD
			prev /= eurUSD
			high /= eurUSD
			low /= eurUSD
			wkHigh /= eurUSD
			wkLow /= eurUSD
			open /= eurUSD
		}

		changeAbs := price - prev
		changePct := 0.0
		if prev != 0 {
			changePct = (changeAbs / prev) * 100
		}

		return DetailMsg{Detail: Detail{
			Name:             asset.Name,
			Symbol:           asset.Symbol,
			Currency:         m.Currency,
			Exchange:         m.ExchangeName,
			ExchangeFull:     m.FullExchangeName,
			InstrumentType:   m.InstrumentType,
			Timezone:         m.Timezone,
			Price:            utils.Round(price, 2),
			PreviousClose:    utils.Round(prev, 2),
			ChangeAbs:        utils.Round(changeAbs, 2),
			ChangePercent:    utils.Round(changePct, 2),
			DayHigh:          utils.Round(high, 2),
			DayLow:           utils.Round(low, 2),
			Volume:           m.RegularMarketVolume,
			FiftyTwoWeekHigh: utils.Round(wkHigh, 2),
			FiftyTwoWeekLow:  utils.Round(wkLow, 2),
			Open:             utils.Round(open, 2),
			Shares:           asset.Shares,
			Value:            utils.Round(price*asset.Shares, 2),
			FetchedAt:        time.Now(),
		}}
	}
}
