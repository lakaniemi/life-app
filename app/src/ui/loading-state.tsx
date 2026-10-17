import { useTranslation } from "react-i18next";
import { StyleSheet, View } from "react-native";
import { ActivityIndicator } from "react-native-paper";

// Only for a first load with nothing to show. Background refetches stay silent.
export const LoadingState = () => {
  const { t } = useTranslation();

  return (
    <View style={styles.container}>
      <ActivityIndicator
        size="large"
        accessibilityLabel={t("common.loading")}
      />
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    alignItems: "center",
    justifyContent: "center",
  },
});
