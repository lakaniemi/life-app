import { nativeApplicationVersion } from "expo-application";
import { useEffect } from "react";
import { useTranslation } from "react-i18next";
import { Linking, StyleSheet, View } from "react-native";
import { Button, Icon, ProgressBar, Text, useTheme } from "react-native-paper";

import {
  latestApkUrl,
  useApkInstall,
} from "@/features/app-update/install-apk";
import { useNewerRelease } from "@/features/app-update/latest-release";
import { Screen } from "@/ui/screen";

// Downloads the latest release and opens Android's installer. The download
// starts as soon as the screen opens, because the user got here by pressing
// Install.
export const InstallUpdateScreen = () => {
  const { t } = useTranslation();
  const { colors } = useTheme();
  const { data: release } = useNewerRelease();
  const { state, install } = useApkInstall();

  useEffect(() => {
    void install();
  }, [install]);

  return (
    <Screen style={styles.content}>
      <Icon source="cellphone-arrow-down" size={48} color={colors.primary} />
      {release && (
        <Text variant="bodyMedium" style={styles.text}>
          {t("appUpdate.versions", {
            current: nativeApplicationVersion ?? "?",
            latest: release.version,
          })}
        </Text>
      )}

      {state.status === "downloading" && (
        <View style={styles.progress}>
          <ProgressBar
            indeterminate={state.progress === null}
            progress={state.progress ?? 0}
          />
          <Text variant="bodyMedium" style={styles.text}>
            {state.progress === null
              ? t("appUpdate.downloading")
              : t("appUpdate.downloadingProgress", {
                  percent: Math.round(state.progress * 100),
                })}
          </Text>
        </View>
      )}

      {state.status === "idle" && (
        <Button mode="contained" onPress={() => void install()}>
          {t("appUpdate.install")}
        </Button>
      )}

      {state.status === "failed" && (
        <>
          <Text variant="bodyLarge" style={styles.text}>
            {t("appUpdate.downloadFailed")}
          </Text>
          <Button mode="contained" onPress={() => void install()}>
            {t("common.retry")}
          </Button>
          <Button onPress={() => void Linking.openURL(latestApkUrl)}>
            {t("appUpdate.downloadFromGithub")}
          </Button>
        </>
      )}
    </Screen>
  );
};

const styles = StyleSheet.create({
  content: {
    alignItems: "center",
    justifyContent: "center",
    gap: 16,
    padding: 24,
  },
  progress: {
    alignSelf: "stretch",
    gap: 8,
  },
  text: {
    textAlign: "center",
  },
});
