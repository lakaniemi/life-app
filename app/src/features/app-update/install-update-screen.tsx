import { nativeApplicationVersion } from "expo-application";
import { useEffect, type ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { Linking, StyleSheet, View } from "react-native";
import { Button, Icon, ProgressBar, Text, useTheme } from "react-native-paper";
import { useSafeAreaInsets } from "react-native-safe-area-context";

import { latestApkUrl } from "@/api/config";
import {
  useApkInstall,
  type ApkInstallState,
} from "@/features/app-update/install-apk";
import { useNewerRelease } from "@/features/app-update/latest-release";
import { Screen } from "@/ui/screen";

type InstallUpdateScreenProps = {
  // The API has turned this version away, and this screen replaces the whole
  // app (UpdateGate). Otherwise the user got here from the Home banner.
  required?: boolean;
};

// Downloads the latest release and opens Android's installer. Coming from the
// banner, the download starts as soon as the screen opens, because the user
// got here by pressing Install. When an update is required, it waits for
// Install instead, so nothing is downloaded, perhaps over mobile data, before
// the user has asked.
export const InstallUpdateScreen = ({
  required = false,
}: InstallUpdateScreenProps) => {
  const { t } = useTranslation();
  const { colors } = useTheme();
  const { data: release } = useNewerRelease();
  const { state, install } = useApkInstall();

  useEffect(() => {
    if (!required) {
      void install();
    }
  }, [required, install]);

  const content = (
    <>
      <Icon source="cellphone-arrow-down" size={48} color={colors.primary} />
      {required && (
        <>
          <Text variant="headlineSmall" style={styles.text}>
            {t("updateRequired.title")}
          </Text>
          <Text variant="bodyLarge" style={styles.text}>
            {t("updateRequired.body")}
          </Text>
        </>
      )}
      {release && (
        <Text variant="bodyMedium" style={styles.text}>
          {t("appUpdate.versions", {
            current: nativeApplicationVersion ?? "?",
            latest: release.version,
          })}
        </Text>
      )}

      <InstallStateView state={state} onInstall={() => void install()} />
    </>
  );

  return required ? (
    <FullScreen>{content}</FullScreen>
  ) : (
    <Screen style={styles.content}>{content}</Screen>
  );
};

// What replaces the app when an update is required. It's rendered outside any
// navigator, so there's no native header or offline banner, and the insets
// on both sides are its own to handle.
const FullScreen = ({ children }: { children: ReactNode }) => {
  const { colors } = useTheme();
  const insets = useSafeAreaInsets();

  return (
    <View
      style={[
        styles.content,
        styles.fullScreen,
        {
          backgroundColor: colors.background,
          paddingTop: insets.top,
          paddingBottom: insets.bottom,
        },
      ]}
    >
      {children}
    </View>
  );
};

type InstallStateViewProps = {
  state: ApkInstallState;
  onInstall: () => void;
};

const InstallStateView = ({ state, onInstall }: InstallStateViewProps) => {
  const { t } = useTranslation();

  switch (state.status) {
    case "idle":
      return (
        <Button mode="contained" onPress={onInstall}>
          {t("appUpdate.install")}
        </Button>
      );
    case "downloading":
      return (
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
      );
    case "failed":
      return (
        <>
          <Text variant="bodyLarge" style={styles.text}>
            {t("appUpdate.downloadFailed")}
          </Text>
          <Button mode="contained" onPress={onInstall}>
            {t("common.retry")}
          </Button>
          <Button onPress={() => void Linking.openURL(latestApkUrl)}>
            {t("appUpdate.downloadFromGithub")}
          </Button>
        </>
      );
  }
};

const styles = StyleSheet.create({
  content: {
    alignItems: "center",
    justifyContent: "center",
    gap: 16,
    padding: 24,
  },
  fullScreen: {
    flex: 1,
  },
  progress: {
    alignSelf: "stretch",
    gap: 8,
  },
  text: {
    textAlign: "center",
  },
});
