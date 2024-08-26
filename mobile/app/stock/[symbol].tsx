import { IconButton } from "@/components/IconButton";
import { StockImage } from "@/components/StockImage";
import { useCandles } from "@/hooks/useCandles";
import { baseUrl } from "@/network";
import { Candle, WsCandleUpdate } from "@/types/types";
import { useFocusEffect, useLocalSearchParams, useNavigation } from "expo-router";
import { useCallback, useEffect, useState } from "react";
import { ActivityIndicator, Alert, Dimensions, StyleSheet, Text, View } from "react-native";
import { CandlestickChart, LineChart, TData, TLineChartDataProp } from "react-native-wagmi-charts";

export default function StockScreen() {
  const navigation = useNavigation();
  const { symbol } = useLocalSearchParams<{ symbol: string; }>();

  const [visibleChart, setVisibleChart] = useState<"candlesticks" | "line">("candlesticks");
  const [isLoading, setIsLoading] = useState(true);
  const [candles, setCandles] = useState<Candle[]>([]);

  const chartWidth = Dimensions.get("screen").width - 20;
  const chartHeight = Dimensions.get("screen").height / 2;

  const {
    chartData,
    newest,
    trendingColor,
    trendingSign,
    startToEndDifference
  } = useCandles({ candles, visibleChart });

  const fetchHistory = useCallback(async () => {
    try {
      const response = await fetch(`${baseUrl("http")}/stock-candles?symbol=${symbol}`);
      const data = await response.json();
      setCandles(data);
    } catch (error) {
      if (error instanceof Error) Alert.alert("Error", error.message);
    } finally {
      setIsLoading(false);
    }
  }, [symbol]);

  useEffect(() => {
    navigation.setOptions({ headerTitle: symbol });
    fetchHistory();
  }, [fetchHistory, symbol]);

  useFocusEffect(
    useCallback(() => {
      const ws = new WebSocket(`${baseUrl("ws")}/ws`);
      ws.onopen = () => ws.send(symbol);
      ws.onmessage = ({ data }) => {
        const { updateType, candle } = JSON.parse(data) as WsCandleUpdate;
        console.log("🚀 ~ useCallback ~ updateType:", updateType);

        if (updateType === "closed") {
          setCandles((candles) => [...candles, candle]);
        } else {
          setCandles((candles) => [...candles.slice(0, -1), candle]);
        }
      };

      return () => {
        ws.close();
      };
    }, [symbol])
  );

  if (isLoading) {
    return <View style={styles.loader}>
      <ActivityIndicator animating size="large" />
    </View>;
  }

  return (
    <View style={styles.container}>

      <View style={styles.innerContainer}>
        <View style={styles.imgContainer}>
          <StockImage style={styles.img} symbol={symbol} />
          <Text style={styles.symbol}>{symbol}</Text>
        </View>

        <View style={styles.priceContainer}>
          <Text style={styles.price}>
            {"$ " + newest.close.toFixed(2)}
          </Text>
          <Text
            style={[styles.priceStatus, { color: trendingColor }]}>
            {trendingSign}
            {startToEndDifference.amount.toFixed(2)}
            {" "}
            ({trendingSign}{startToEndDifference.percentage.toFixed(2) + "%"})
          </Text>
        </View>
      </View>

      <View style={styles.buttonsContainer}>
        <IconButton
          name="analytics"
          touchableOpacityStyles={{
            backgroundColor: visibleChart === "line" ? "black" : "gray"
          }}
          onPress={() => setVisibleChart("line")}
        />
        <IconButton
          name="stats-chart"
          touchableOpacityStyles={{
            backgroundColor: visibleChart === "candlesticks" ? "black" : "gray"
          }}
          onPress={() => setVisibleChart("candlesticks")}
        />
      </View>

      {visibleChart === "candlesticks" ? (
        <CandlestickChart.Provider data={chartData as TData}>
          <CandlestickChart width={chartWidth} height={chartHeight}>
            <CandlestickChart.Candles />
            <CandlestickChart.Crosshair>
              <CandlestickChart.Tooltip />
            </CandlestickChart.Crosshair>
          </CandlestickChart>

          <CandlestickChart.PriceText type="open" />
          <CandlestickChart.PriceText type="high" />
          <CandlestickChart.PriceText type="low" />
          <CandlestickChart.PriceText type="close" />
          <CandlestickChart.DatetimeText />
        </CandlestickChart.Provider>
      ) : (
        <LineChart.Provider data={chartData as TLineChartDataProp}>
          <LineChart width={chartWidth} height={chartHeight}>
            <LineChart.Path color={trendingColor}>
              <LineChart.Dot color={trendingColor} at={chartData.length - 1} hasPulse />
              <LineChart.Gradient />
              <LineChart.HorizontalLine color={trendingColor} at={{ index: 0 }} />
            </LineChart.Path>
          </LineChart>
        </LineChart.Provider>
      )}

    </View>
  );
}


const styles = StyleSheet.create({
  loader: {
    flex: 1,
    justifyContent: "center",
    alignItems: "center"
  },
  container: {
    flex: 1,
  },
  innerContainer: {
    flexDirection: 'row',
    justifyContent: "space-between",
    alignItems: "center",
    marginBottom: 20,
    padding: 20
  },
  imgContainer: {
    flexDirection: "row",
    justifyContent: "center",
    alignItems: "center",
    gap: 10
  },
  img: {
    width: 70,
    height: 70
  },
  symbol: {
    fontSize: 25,
    fontWeight: "bold"
  },
  priceContainer: {
    justifyContent: "center",
    alignItems: "flex-end",
    alignSelf: "center",
    gap: 5
  },
  buttonsContainer: {
    flexDirection: "row",
    justifyContent: "center",
    gap: 30,
    marginBottom: 20
  },
  price: {
    fontSize: 26,
    fontWeight: "bold"
  },
  priceStatus: {
    fontSize: 15,
    fontWeight: "semibold"
  }
});
