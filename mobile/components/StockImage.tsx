import { useEffect, useState } from "react";
import { Image, ImageProps } from "react-native";

interface Props extends ImageProps {
  symbol: string;
}

const getURIPath = (symbol: string) => `https://eodhd.com/img/logos/US/${symbol}.png`;

export function StockImage({ symbol, ...rest }: Props) {
  const [uri, setUri] = useState(getURIPath(symbol));

  useEffect(() => {
    Image.getSize(uri, () => { }, () => {
      setUri(getURIPath(symbol.toLowerCase()));
    });
  }, [uri]);

  return <Image source={{ uri }} {...rest} />;
}
