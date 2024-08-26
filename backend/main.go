package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

var (
	symbols = []string{"AAPL", "AMZN", "TSLA", "GOOGL", "NFLX", "PYPL"}

	// Broadcast messages to all connected clients
	broadcast = make(chan *BroadcastMessage)

	// Map all connected clients and symbol they're subscribed to
	clientConns = make(map[*websocket.Conn]string)

	// Map of all ongoing live candles for each symbol
	tempCandles = make(map[string]*TempCandle)

	mu sync.Mutex
)

func main() {
	// Env config
	env := EnvConfig()

	// Db connection
	db := DBConnection(env)

	// Connect to Finnhub WebSockets
	finnhubWSConn := connectToFinnhub(env)
	defer finnhubWSConn.Close()

	// Handle Finnhub's WebSockets incoming messages
	go handleFinnhubMessages(finnhubWSConn, db)

	// Broadcast candle updates to all clients connected
	go broadcastUpdates()

	// --- Endpoints ----
	// Connect to the WebSocket
	http.HandleFunc("/ws", WSHandler)

	// Fetch all past candles for all of the symbols
	http.HandleFunc("/stocks-history", func(w http.ResponseWriter, r *http.Request) {
		StocksHistoryHandler(w, r, db)
	})

	// Fetch all past candles from a specific symbol
	http.HandleFunc("/stock-candles", func(w http.ResponseWriter, r *http.Request) {
		CandlesHandler(w, r, db)
	})

	// Serve the endpoints
	http.ListenAndServe(fmt.Sprintf(":%s", env.SERVER_PORT), nil)
}

// Websocket endpoint to connect clients to the latest updates on the symbol they're subscribed to
func WSHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade incoming GET request into a Websocket connection
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrate connection:", err)
	}

	// Close ws connection & unregister the client when they disconnect
	defer conn.Close()
	defer func() {
		delete(clientConns, conn)
		log.Println("Client disconnected!")
	}()

	// Register the new client to the symbol they're subscribing to
	for {
		_, symbol, err := conn.ReadMessage()
		clientConns[conn] = string(symbol)
		log.Println("New Client Connected!")

		if err != nil {
			log.Println("Error reading from the client:", err)
			break
		}
	}
}

// Fetch all past candles for all of the symbols
func StocksHistoryHandler(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	// Query the db for all candle data from all symbols
	var candles []Candle
	db.Order("timestamp asc").Find(&candles)

	// Create a map to group data by symbol
	groupedData := make(map[string][]Candle)

	// Group the candles by symbol
	for _, candle := range candles {
		symbol := candle.Symbol
		groupedData[symbol] = append(groupedData[symbol], candle)
	}

	// Marshal the grouped data into JSON and send over http
	jsonResponse, _ := json.Marshal(groupedData)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

// Fetch all past candles from a specific symbol
func CandlesHandler(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	// Get the symbol from the query
	symbol := r.URL.Query().Get("symbol")

	// Query the db for all candle data for that symbol
	var candles []Candle
	db.Where("symbol = ?", symbol).Order("timestamp asc").Find(&candles)

	// Marshal the candles data into JSON and send over http
	jsonCandles, _ := json.Marshal(candles)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonCandles)
}

// Connect to Finnhub WebSockets
func connectToFinnhub(env *Env) *websocket.Conn {
	ws, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("wss://ws.finnhub.io?token=%s", env.API_KEY), nil)
	if err != nil {
		panic(err)
	}

	for _, s := range symbols {
		msg, _ := json.Marshal(map[string]interface{}{"type": "subscribe", "symbol": s})
		ws.WriteMessage(websocket.TextMessage, msg)
	}

	return ws
}

// Handle Finnhub's WebSockets incoming messages
func handleFinnhubMessages(ws *websocket.Conn, db *gorm.DB) {
	finnhubMessage := &FinnhubMessage{}

	for {
		if err := ws.ReadJSON(finnhubMessage); err != nil {
			fmt.Println("Error reading the message: ", err)
			continue
		}

		// Only try to process the message data if it's a trade operation
		if finnhubMessage.Type == "trade" {
			for _, trade := range finnhubMessage.Data {
				// Process the trade data
				processTradeData(&trade, db)
			}
		}

		// Clean up old trades older than 20 minutes
		cutoffTime := time.Now().Add(-20 * time.Minute)
		db.Where("timestamp < ?", cutoffTime).Delete(&Candle{})
	}
}

// Process each trade and update or create temporary candles
func processTradeData(trade *TradeData, db *gorm.DB) {
	// Protect the goroutine from data races
	mu.Lock()
	defer mu.Unlock()

	// Extract trade data
	symbol := trade.Symbol
	price := trade.Price
	volume := float64(trade.Volume)
	timestamp := time.UnixMilli(trade.Timestamp)

	// Retrieve or create a tempCandle for the symbol
	tempCandle, exists := tempCandles[symbol]

	// If the tempCandle does not exist or should be already closed
	if !exists || timestamp.After(tempCandle.CloseTime) {
		// Finalize and save the previous candle, start a new one
		if exists {
			// Convert the tempCandle to a Candle
			candle := tempCandle.toCandle()

			// Save the candle to the db
			if err := db.Create(candle).Error; err != nil {
				fmt.Println("Error saving the candle to the DB: ", err)
			}

			// Broadcast the closed candle
			broadcast <- &BroadcastMessage{
				UpdateType: Closed,
				Candle:     candle,
			}
		}

		// Initialize a new candle
		tempCandle = &TempCandle{
			Symbol:     symbol,
			OpenTime:   timestamp,
			CloseTime:  timestamp.Add(time.Minute),
			OpenPrice:  price,
			ClosePrice: price,
			HighPrice:  price,
			LowPrice:   price,
			Volume:     volume,
		}
	}

	// Update current tempCandle with new trade data
	tempCandle.ClosePrice = price
	tempCandle.Volume += volume
	if price > tempCandle.HighPrice {
		tempCandle.HighPrice = price
	}
	if price < tempCandle.LowPrice {
		tempCandle.LowPrice = price
	}

	// Store the tempCandle for the symbol
	tempCandles[symbol] = tempCandle

	// Write to the broadcast channel live ongoing channel
	broadcast <- &BroadcastMessage{
		UpdateType: Live,
		Candle:     tempCandle.toCandle(),
	}
}

// Send candle updates to clients connected every 1 second at maximum, unless it's a closed candle
func broadcastUpdates() {
	// Set the broadcast interval to 1 second
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	var latestUpdate *BroadcastMessage

	for {
		select {
		// Watch for new updates from the broadcast channel
		case update := <-broadcast:
			// If the update is a closed candle, broadcast it immediatly
			if update.UpdateType == Closed {
				// Broadcast it
				broadcastToClients(update)
			} else {
				// replace temp updates
				latestUpdate = update
			}

		case <-ticker.C:
			// Broadcast the latest update
			if latestUpdate != nil {
				// Broadcast it
				broadcastToClients(latestUpdate)
			}
			latestUpdate = nil
		}
	}
}

// Broadcast updates to clients
func broadcastToClients(update *BroadcastMessage) {
	// Marshal the update struct into json
	jsonUpdate, _ := json.Marshal(update)

	// Send the update to all connected clients subscribed to the symbol
	for clientConn, symbol := range clientConns {
		// If the client is subscribed to the symbol of the update
		if update.Candle.Symbol == symbol {
			// Send the update to the client
			err := clientConn.WriteMessage(websocket.TextMessage, jsonUpdate)
			if err != nil {
				log.Println("Error sending message to client: ", err)
				clientConn.Close()
				delete(clientConns, clientConn)
			}
		}
	}
}
