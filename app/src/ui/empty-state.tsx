import { StyleSheet, View } from "react-native";
import { Button, Icon, Text, useTheme } from "react-native-paper";

type EmptyStateProps = {
  // A Material Design Icons name, e.g. "cart-outline".
  icon: string;
  title: string;
  description: string;
  action?: { label: string; onPress: () => void };
};

export const EmptyState = ({
  icon,
  title,
  description,
  action,
}: EmptyStateProps) => {
  const { colors } = useTheme();

  return (
    <View style={styles.container}>
      <Icon source={icon} size={48} color={colors.onSurfaceVariant} />
      <Text variant="titleLarge" style={styles.centered}>
        {title}
      </Text>
      <Text
        variant="bodyMedium"
        style={[styles.centered, { color: colors.onSurfaceVariant }]}
      >
        {description}
      </Text>
      {action && (
        <Button mode="contained" onPress={action.onPress} style={styles.action}>
          {action.label}
        </Button>
      )}
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    alignItems: "center",
    justifyContent: "center",
    gap: 8,
    padding: 24,
  },
  centered: {
    textAlign: "center",
  },
  action: {
    marginTop: 16,
  },
});
