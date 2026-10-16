import { StyleSheet } from "react-native";
import { Text } from "react-native-paper";

import { Screen } from "@/ui/screen";

export function HomeScreen() {
  return (
    <Screen style={styles.container}>
      <Text variant="headlineMedium">Life App</Text>
    </Screen>
  );
}

const styles = StyleSheet.create({
  container: {
    alignItems: "center",
    justifyContent: "center",
  },
});
