export interface Candle {
  symbol: string;
  open: number
  close: number
  high: number
  low: number
  timestamp: string
}

export interface WsCandleUpdate {
  updateType: "live" | "closed"
  candle: Candle
}
