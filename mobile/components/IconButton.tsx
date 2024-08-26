import { Ionicons } from "@expo/vector-icons";
import { IconProps } from "@expo/vector-icons/build/createIconSet";
import { ComponentProps } from "react";
import { StyleProp, StyleSheet, TouchableOpacity, ViewStyle } from "react-native";

interface Props extends IconProps<ComponentProps<typeof Ionicons>["name"]> {
  onPress: () => void;
  touchableOpacityStyles: StyleProp<ViewStyle>;
}

export function IconButton({ touchableOpacityStyles, onPress, ...rest }: Props) {
  return (
    <TouchableOpacity onPress={onPress} style={[styles.touchable, touchableOpacityStyles]}>
      <Ionicons color="white" size={29} {...rest} />
    </TouchableOpacity>
  );
}

const styles = StyleSheet.create({
  touchable: {
    flexDirection: "row",
    borderRadius: 10,
    backgroundColor: "black",
    padding: 10
  }
});
