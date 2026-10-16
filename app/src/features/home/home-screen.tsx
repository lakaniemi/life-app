import { useTranslation } from "react-i18next";
import { StyleSheet } from "react-native";
import { Text } from "react-native-paper";

import { Screen } from "@/ui/screen";

export function HomeScreen() {
  const { t } = useTranslation();

  return (
    <Screen style={styles.container}>
      <Text variant="headlineMedium">{t("home.title")}</Text>
    </Screen>
  );
}

const styles = StyleSheet.create({
  container: {
    alignItems: "center",
    justifyContent: "center",
  },
});
