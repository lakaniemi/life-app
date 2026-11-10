import type { ParseKeys } from "i18next";
import { useTranslation } from "react-i18next";
import { StyleSheet, View } from "react-native";
import { Text } from "react-native-paper";

import { errorMessage } from "@/api/error-message";
import { isWaitingForConnection } from "@/api/query-state";
import { UpdateBanner } from "@/features/app-update/update-banner";
import { useHealth, type Health } from "@/features/home/api";
import { ErrorState } from "@/ui/error-state";
import { LoadingState } from "@/ui/loading-state";
import { Screen } from "@/ui/screen";

const statusKeys: Record<Health["status"], ParseKeys> = {
  ok: "home.apiStatus.ok",
  degraded: "home.apiStatus.degraded",
  unknown: "home.apiStatus.unknown",
};

// The banner is outside the states below, so it shows whatever the health
// query is doing, and keeps its "Later" while they change.
export const HomeScreen = () => (
  <Screen>
    <UpdateBanner />
    <HealthStatus />
  </Screen>
);

// Until features contribute their cards, Home shows the API's health, which
// exercises the whole data layer.
const HealthStatus = () => {
  const { t } = useTranslation();
  const health = useHealth();

  if (isWaitingForConnection(health)) {
    return <ErrorState message={t("errors.offline")} />;
  }

  if (health.isPending) {
    return <LoadingState />;
  }

  if (health.isLoadingError) {
    return (
      <ErrorState
        message={errorMessage(health.error)}
        onRetry={() => void health.refetch()}
      />
    );
  }

  return (
    <View style={styles.status}>
      <Text variant="headlineMedium">{t("home.title")}</Text>
      <Text variant="bodyLarge">{t(statusKeys[health.data.status])}</Text>
      <Text variant="bodySmall">
        {t("home.apiVersion", { version: health.data.version.slice(0, 7) })}
      </Text>
    </View>
  );
};

const styles = StyleSheet.create({
  status: {
    flex: 1,
    alignItems: "center",
    justifyContent: "center",
    gap: 8,
  },
});
