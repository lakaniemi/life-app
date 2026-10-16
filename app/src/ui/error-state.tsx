import { useTranslation } from "react-i18next";
import { StyleSheet, View } from "react-native";
import { Button, Icon, Text, useTheme } from "react-native-paper";

type ErrorStateProps = {
  // Copy for what failed. Defaults to a generic message.
  message?: string;
  onRetry: () => void;
};

// For a failed load with nothing cached to show instead.
export function ErrorState({ message, onRetry }: ErrorStateProps) {
  const { t } = useTranslation();
  const { colors } = useTheme();

  return (
    <View style={styles.container}>
      <Icon source="alert-circle-outline" size={48} color={colors.error} />
      <Text variant="bodyLarge" style={styles.message}>
        {message ?? t("errors.generic")}
      </Text>
      <Button mode="contained-tonal" onPress={onRetry}>
        {t("common.retry")}
      </Button>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    alignItems: "center",
    justifyContent: "center",
    gap: 16,
    padding: 24,
  },
  message: {
    textAlign: "center",
  },
});
