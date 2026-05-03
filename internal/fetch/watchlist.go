package fetch

import (
	"context"
	"fmt"
	"math"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"invest-tracker-tui/internal/config"
	"invest-tracker-tui/internal/utils"
)

// WatchPerf holds all computed performance metrics for a single watchlist item.
type WatchPerf struct {
	Name     string
	Symbol   string
	Type     config.AssetType
	Price    float64
	Currency string
	Perf1d   float64
	Perf7d   float64
	Perf30d  float64
	Perf90d  float64
	Perf365d float64
	ATH      float64
	ATHDiff  float64 // negative means price is below ATH
}

// WatchSection mirrors a config.WatchSector but holds computed WatchPerf values.
type WatchSection struct {
	Name  string
	Items []WatchPerf
}

// WatchlistMsg is the Bubbletea message returned after fetching the full watchlist.
type WatchlistMsg struct {
	Sections []WatchSection
	Err      error
}

// yahooHistoryResponse maps the fields needed from a range=1y chart call.
type yahooHistoryResponse struct {
	Chart struct {
		Result []struct {
			Meta struct {
				Currency           string  `json:"currency"`
				RegularMarketPrice float64 `json:"regularMarketPrice"`
				PreviousClose      float64 `json:"previousClose"`
				ChartPreviousClose float64 `json:"chartPreviousClose"`
			} `json:"meta"`
			Timestamp  []int64 `json:"timestamp"`
			Indicators struct {
				Quote []struct {
					Close []float64 `json:"close"`
				} `json:"quote"`
			} `json:"indicators"`
		} `json:"result"`
	} `json:"chart"`
}

// WatchlistCmd fetches performance data for every item in the watchlist.
func WatchlistCmd(wl config.Watchlist) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		eurUSD, err := fetchEURUSD(ctx)
		if err != nil {
			return WatchlistMsg{Err: fmt.Errorf("EUR/USD: %w", err)}
		}

		var sections []WatchSection
		for _, sector := range wl.Sectors {
			sec := WatchSection{Name: sector.Name}
			for _, item := range sector.Items {
				perf, err := fetchWatchPerf(ctx, item, eurUSD)
				if err != nil {
					return WatchlistMsg{Err: fmt.Errorf("%s: %w", item.Symbol, err)}
				}
				sec.Items = append(sec.Items, perf)
			}
			sections = append(sections, sec)
		}

		return WatchlistMsg{Sections: sections}
	}
}

func fetchWatchPerf(ctx context.Context, item config.WatchItem, eurUSD float64) (WatchPerf, error) {
	url := fmt.Sprintf(
		"https://query1.finance.yahoo.com/v8/finance/chart/%s?interval=1d&range=1y",
		item.Symbol,
	)

	body, err := fetchRaw(ctx, url)
	if err != nil {
		return WatchPerf{}, err
	}

	var payload yahooHistoryResponse
	if err := decodeJSON(body, &payload); err != nil {
		return WatchPerf{}, err
	}

	if len(payload.Chart.Result) == 0 || len(payload.Chart.Result[0].Indicators.Quote) == 0 {
		return WatchPerf{}, fmt.Errorf("no result for %s", item.Symbol)
	}

	res := payload.Chart.Result[0]
	closes := res.Indicators.Quote[0].Close
	n := len(closes)
	if n == 0 {
		return WatchPerf{}, fmt.Errorf("no close data for %s", item.Symbol)
	}

	isCrypto := item.Type == config.AssetTypeCrypto

	// convert divides a USD price to EUR for crypto assets.
	convert := func(v float64) float64 {
		if isCrypto && eurUSD != 0 {
			return v / eurUSD
		}
		return v
	}

	current := convert(res.Meta.RegularMarketPrice)

	pct := func(from, to float64) float64 {
		if from == 0 {
			return 0
		}
		return utils.Round(((to-from)/from)*100, 2)
	}

	// lastClose returns the last non-zero close going back from index idx.
	lastClose := func(idx int) float64 {
		if idx >= n {
			idx = n - 1
		}
		for idx > 0 && closes[idx] == 0 {
			idx--
		}
		return convert(closes[idx])
	}

	// closes[n-1] = today's close (current session, same as live price).
	// closes[n-2] = yesterday's close = correct 1D reference.
	// All lookbacks are shifted by 1 so they don't include today's bar.
	prev1d := lastClose(n - 2)
	prev7d := lastClose(n - 2 - 5)   // ~5 trading days ≈ 7 calendar days
	prev30d := lastClose(n - 2 - 21) // ~21 trading days ≈ 30 calendar days
	prev90d := lastClose(n - 2 - 63) // ~63 trading days ≈ 90 calendar days

	// ATH is the highest close over the full 1-year series.
	ath := current
	for _, c := range closes {
		if c == 0 {
			continue
		}
		if cv := convert(c); cv > ath {
			ath = cv
		}
	}

	return WatchPerf{
		Name:     item.Name,
		Symbol:   item.Symbol,
		Type:     item.Type,
		Price:    utils.Round(current, 2),
		Currency: res.Meta.Currency,
		Perf1d:   pct(prev1d, current),
		Perf7d:   pct(prev7d, current),
		Perf30d:  pct(prev30d, current),
		Perf90d:  pct(prev90d, current),
		Perf365d: pct(convert(closes[0]), current),
		ATH:      utils.Round(ath, 2),
		ATHDiff:  utils.Round(math.Min(pct(ath, current), 0), 2),
	}, nil
}
