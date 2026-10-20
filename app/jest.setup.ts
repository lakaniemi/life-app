// Initializes i18next once, as the root layout does in the app.
import "@/i18n/i18n";

// Tests render without the native safe-area provider that Expo Router sets up
// in the app, so the library's own mock supplies fixed insets.
jest.mock(
  "react-native-safe-area-context",
  () => require("react-native-safe-area-context/jest/mock").default,
);

// Screen tests render a screen on its own, outside a route tree, where
// useIsFocused (used by useApiQuery) would throw. A test of focus behavior
// can render a route tree with expo-router/testing-library and requireActual.
jest.mock("expo-router", () => ({
  ...jest.requireActual("expo-router"),
  useIsFocused: () => true,
}));

// The real URL comes from app.config.ts through expo-constants, which tests
// don't load. No test reaches the network: they mock fetch or the client.
jest.mock("@/api/config", () => ({ apiUrl: "https://api.test" }));
