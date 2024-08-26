import { Candle } from "@/types/types";
import { useEffect, useMemo, useState } from "react";

const TRENDING_COLORS = {
  up: 'green',
  down: 'red',
  flat: 'black'
}

interface Props {
  candles: Candle[]
  visibleChart: "candlesticks" | "line"
}

export function useCandles({ candles, visibleChart }: Props) {
  const newest = candles[candles.length -1 ]
  const oldest = candles[0]

  const [trending, setTrending] = useState<"up" | "down" | "flat">("flat")
  const [startToEndDifference, setStartToEndDifference] = useState<{
    amount: number,
    percentage: number
  }>({
    amount: 0,
    percentage: 0
  })

  useEffect(()=>{
    if (candles.length < 2) return;

    const difference = newest.close - oldest.close;
    const percentage = difference / oldest.close * 100

    setTrending(difference > 0 ? 'up' : difference < 0 ? 'down' : 'flat')
    setStartToEndDifference({ amount: difference, percentage })
  }, [candles])

  const chartData = useMemo(() => candles.map(({ timestamp, ...rest }) => ({
    timestamp: new Date(timestamp).getTime(),
    ...(visibleChart  === 'candlesticks' ? rest : { value: rest.close }),
  })), [candles, visibleChart]);

  return {
    trendingColor: TRENDING_COLORS[trending],
    trendingSign: trending === "up" ? '+' : '',
    startToEndDifference,
    oldest,
    newest,
    chartData,
  }
}
