// Tests render without the native safe-area provider that Expo Router sets up
// in the app, so the library's own mock supplies fixed insets.
jest.mock(
  "react-native-safe-area-context",
  () => require("react-native-safe-area-context/jest/mock").default,
);
