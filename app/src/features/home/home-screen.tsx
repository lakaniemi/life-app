import { StyleSheet, Text, View } from "react-native";

export function HomeScreen() {
  return (
    <View style={styles.container}>
      <Text>Life App</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    alignItems: "center",
    justifyContent: "center",
  },
});
