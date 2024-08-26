package main

import "time"

// Candle struct represents a single OHLC (High, Low, Open, Close) candle
type Candle struct {
	Symbol    string    `json:"symbol"`
	Open      float64   `json:"open"`
	Close     float64   `json:"close"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Timestamp time.Time `json:"timestamp"`
}

// TempCandle represents an item from the temp candle slice building the candles
type TempCandle struct {
	Symbol     string
	OpenTime   time.Time
	CloseTime  time.Time
	OpenPrice  float64
	ClosePrice float64
	HighPrice  float64
	LowPrice   float64
	Volume     float64
}

// Structure of the data that comes from the Finnhub ws api
type FinnhubMessage struct {
	Data []TradeData `json:"data"`
	Type string      `json:"type"` // ping | trade
}
type TradeData struct {
	Close     []string `json:"c"`
	Price     float64  `json:"p"`
	Symbol    string   `json:"s"`
	Timestamp int64    `json:"t"`
	Volume    int      `json:"v"`
}

// Data to write to clients connected
type BroadcastMessage struct {
	UpdateType UpdateType `json:"updateType"` // "live" | "closed"
	Candle     *Candle    `json:"candle"`
}

type UpdateType string

const (
	Live   UpdateType = "live"   // Real time ongoing candle
	Closed UpdateType = "closed" // Past candle. Already closed
)

// Converts a tempCandle into a Candle
func (tc *TempCandle) toCandle() *Candle {
	return &Candle{
		Symbol:    tc.Symbol,
		Open:      tc.OpenPrice,
		Close:     tc.ClosePrice,
		High:      tc.HighPrice,
		Low:       tc.LowPrice,
		Timestamp: tc.CloseTime,
	}
}
