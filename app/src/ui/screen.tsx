import type { ReactNode } from "react";
import { StyleSheet, View, type StyleProp, type ViewStyle } from "react-native";
import { useTheme } from "react-native-paper";
import { useSafeAreaInsets } from "react-native-safe-area-context";

import { OfflineBanner } from "@/ui/offline-banner";

type ScreenProps = {
  children: ReactNode;
  // Applies to the content, below the offline banner.
  style?: StyleProp<ViewStyle>;
};

// The root view of every screen, with the offline banner at its top. The
// native header already sits below the status bar, so only the bottom inset
// (the gesture or button bar, which the app draws behind edge to edge) needs
// padding here.
export const Screen = ({ children, style }: ScreenProps) => {
  const { colors } = useTheme();
  const insets = useSafeAreaInsets();

  return (
    <View
      style={[
        styles.container,
        { backgroundColor: colors.background, paddingBottom: insets.bottom },
      ]}
    >
      <OfflineBanner />
      <View style={[styles.content, style]}>{children}</View>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  content: {
    flex: 1,
  },
});
