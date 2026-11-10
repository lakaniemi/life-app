import { useTranslation } from "react-i18next";
import { Linking, StyleSheet, View } from "react-native";
import { Button, Icon, Text, useTheme } from "react-native-paper";
import { useSafeAreaInsets } from "react-native-safe-area-context";

// The release workflow attaches the APK under this fixed name, so the URL
// always points to the newest one (docs/ARCHITECTURE.md, "Distribution").
export const latestApkUrl =
  "https://github.com/lakaniemi/life-app/releases/latest/download/life-app.apk";

// Blocks the whole app when the API has turned it away as too old. The only
// fix is a new APK.
export const UpdateRequiredScreen = () => {
  const { t } = useTranslation();
  const { colors } = useTheme();
  const insets = useSafeAreaInsets();

  return (
    <View
      style={[
        styles.container,
        {
          backgroundColor: colors.background,
          paddingTop: insets.top,
          paddingBottom: insets.bottom,
        },
      ]}
    >
      <Icon source="cellphone-arrow-down" size={48} color={colors.primary} />
      <Text variant="headlineSmall" style={styles.text}>
        {t("updateRequired.title")}
      </Text>
      <Text variant="bodyLarge" style={styles.text}>
        {t("updateRequired.body")}
      </Text>
      <Button
        mode="contained"
        onPress={() => void Linking.openURL(latestApkUrl)}
      >
        {t("updateRequired.download")}
      </Button>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    alignItems: "center",
    justifyContent: "center",
    gap: 16,
    padding: 24,
  },
  text: {
    textAlign: "center",
  },
});
