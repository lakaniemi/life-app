import { nativeApplicationVersion } from "expo-application";
import { router } from "expo-router";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Banner, Icon, Text, useTheme } from "react-native-paper";

import { useNewerRelease } from "@/features/app-update/latest-release";
import type { AppPaperTheme } from "@/ui/theme";

// Offers a newer release at the top of Home, and nothing when there isn't
// one. It stays until acted on: unlike a snackbar, an available update stays
// true. "Later" only hides it, in state that lives as long as Home does, so
// the banner is back at the next launch.
export const UpdateBanner = () => {
  const { t } = useTranslation();
  const { colors } = useTheme<AppPaperTheme>();
  const { data: release } = useNewerRelease();
  const [isDismissed, setDismissed] = useState(false);

  if (!release || isDismissed) {
    return null;
  }

  return (
    <Banner
      visible
      // A function, so the icon takes the theme's colour. Banner draws a plain
      // string icon in a default colour.
      icon={({ size }) => (
        <Icon source="cellphone-arrow-down" size={size} color={colors.success} />
      )}
      actions={[
        { label: t("appUpdate.later"), onPress: () => setDismissed(true) },
        {
          label: t("appUpdate.install"),
          onPress: () => router.push("/update"),
        },
      ]}
    >
      {t("appUpdate.available")}
      {"\n"}
      {/* The versions are detail for the curious, so quieter than the
          question. The taller line gives them room below it. */}
      <Text
        variant="bodySmall"
        style={{ color: colors.onSurfaceVariant, lineHeight: 24 }}
      >
        {t("appUpdate.versions", {
          current: nativeApplicationVersion ?? "?",
          latest: release.version,
        })}
      </Text>
    </Banner>
  );
};
