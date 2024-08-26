import { Platform } from "react-native";

export const baseUrl = (scheme: "http" | "ws") => {
  const PORT = 3000
  const HOST = Platform.OS === "android" ? "10.0.2.2" : "localhost"

  return `${scheme}://${HOST}:${PORT}`
}
