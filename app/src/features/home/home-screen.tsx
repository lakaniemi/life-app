import type { ParseKeys } from "i18next";
import { useTranslation } from "react-i18next";
import { StyleSheet } from "react-native";
import { Text } from "react-native-paper";

import { errorMessage } from "@/api/error-message";
import { useHealth, type Health } from "@/features/home/api";
import { ErrorState } from "@/ui/error-state";
import { LoadingState } from "@/ui/loading-state";
import { Screen } from "@/ui/screen";

const statusKeys: Record<Health["status"], ParseKeys> = {
  ok: "home.apiStatus.ok",
  degraded: "home.apiStatus.degraded",
  unknown: "home.apiStatus.unknown",
};

// Until features contribute their cards, Home shows the API's health, which
// exercises the whole data layer.
export const HomeScreen = () => {
  const { t } = useTranslation();
  const health = useHealth();

  if (health.isPending) {
    return (
      <Screen>
        <LoadingState />
      </Screen>
    );
  }

  if (health.isLoadingError) {
    return (
      <Screen>
        <ErrorState
          message={errorMessage(health.error)}
          onRetry={() => void health.refetch()}
        />
      </Screen>
    );
  }

  return (
    <Screen style={styles.container}>
      <Text variant="headlineMedium">{t("home.title")}</Text>
      <Text variant="bodyLarge">{t(statusKeys[health.data.status])}</Text>
      <Text variant="bodySmall">
        {t("home.apiVersion", { version: health.data.version.slice(0, 7) })}
      </Text>
    </Screen>
  );
};

const styles = StyleSheet.create({
  container: {
    alignItems: "center",
    justifyContent: "center",
    gap: 8,
  },
});
