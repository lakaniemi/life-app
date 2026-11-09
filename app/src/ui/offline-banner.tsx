import { onlineManager, useQueryClient } from "@tanstack/react-query";
import { useSyncExternalStore } from "react";
import { useTranslation } from "react-i18next";
import { StyleSheet, View } from "react-native";
import { Icon, Text, useTheme } from "react-native-paper";

const subscribeToOnline = (onChange: () => void) =>
  onlineManager.subscribe(onChange);
const isOnline = () => onlineManager.isOnline();

// Shown while offline, when screens show their cached data read-only. It
// follows TanStack Query's online state, so the banner is up exactly while
// queries are paused. With nothing cached it stays hidden: there is no data
// to point at, and the screen's own offline message says so instead.
export const OfflineBanner = () => {
  const { t } = useTranslation();
  const { colors } = useTheme();
  const queryCache = useQueryClient().getQueryCache();
  const online = useSyncExternalStore(subscribeToOnline, isOnline);
  const hasCachedData = useSyncExternalStore(
    (onChange) => queryCache.subscribe(onChange),
    () => queryCache.findAll().some((query) => query.state.data !== undefined),
  );

  if (online || !hasCachedData) {
    return null;
  }

  return (
    <View
      style={[styles.container, { backgroundColor: colors.surfaceVariant }]}
      accessibilityLiveRegion="polite"
    >
      <Icon
        source="cloud-off-outline"
        size={18}
        color={colors.onSurfaceVariant}
      />
      <Text variant="labelLarge" style={{ color: colors.onSurfaceVariant }}>
        {t("offline.banner")}
      </Text>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "center",
    gap: 8,
    paddingHorizontal: 16,
    paddingVertical: 8,
  },
});
