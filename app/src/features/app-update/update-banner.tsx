import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Linking } from "react-native";
import { Banner } from "react-native-paper";

import { useNewerRelease } from "@/features/app-update/latest-release";
import { latestApkUrl } from "@/features/app-update/update-required-screen";

// Offers a newer release at the top of Home, and nothing when there isn't
// one. It stays until acted on: unlike a snackbar, an available update stays
// true. "Later" only hides it, in state that lives as long as Home does, so
// the banner is back at the next launch.
export const UpdateBanner = () => {
  const { t } = useTranslation();
  const { data: release } = useNewerRelease();
  const [isDismissed, setDismissed] = useState(false);

  if (!release || isDismissed) {
    return null;
  }

  return (
    <Banner
      visible
      icon="cellphone-arrow-down"
      actions={[
        { label: t("appUpdate.later"), onPress: () => setDismissed(true) },
        {
          label: t("appUpdate.install"),
          // Until the app downloads and installs the APK itself, the browser
          // does it.
          onPress: () => void Linking.openURL(latestApkUrl),
        },
      ]}
    >
      {t("appUpdate.available", { version: release.version })}
    </Banner>
  );
};
