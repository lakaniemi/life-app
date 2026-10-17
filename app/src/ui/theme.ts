import { DarkTheme, DefaultTheme, type Theme } from "expo-router";
import { useColorScheme } from "react-native";
import { MD3DarkTheme, MD3LightTheme, type MD3Theme } from "react-native-paper";

// M3 "tonal spot" schemes generated from the seed #4A6A8F with
// @material/material-color-utilities, the library behind Material Theme
// Builder. They're static on purpose: the app keeps its own scheme rather than
// Android's dynamic color. The elevation levels are surface tinted with primary
// at Paper's opacities, as opaque colors so that Surface shadows render.
const lightColors: MD3Theme["colors"] = {
  primary: "#36618E",
  primaryContainer: "#D1E4FF",
  secondary: "#535F70",
  secondaryContainer: "#D7E3F7",
  tertiary: "#6B5778",
  tertiaryContainer: "#F2DAFF",
  surface: "#F8F9FF",
  surfaceVariant: "#DFE2EB",
  surfaceDisabled: "rgba(25, 28, 32, 0.12)",
  background: "#F8F9FF",
  error: "#BA1A1A",
  errorContainer: "#FFDAD6",
  onPrimary: "#FFFFFF",
  onPrimaryContainer: "#194975",
  onSecondary: "#FFFFFF",
  onSecondaryContainer: "#3B4858",
  onTertiary: "#FFFFFF",
  onTertiaryContainer: "#523F5F",
  onSurface: "#191C20",
  onSurfaceVariant: "#43474E",
  onSurfaceDisabled: "rgba(25, 28, 32, 0.38)",
  onError: "#FFFFFF",
  onErrorContainer: "#93000A",
  onBackground: "#191C20",
  outline: "#73777F",
  outlineVariant: "#C3C7CF",
  inverseSurface: "#2E3135",
  inverseOnSurface: "#EFF0F7",
  inversePrimary: "#A0CAFD",
  shadow: "#000000",
  scrim: "#000000",
  backdrop: "rgba(44, 49, 55, 0.4)",
  elevation: {
    level0: "transparent",
    level1: "rgb(238, 241, 249)",
    level2: "rgb(233, 237, 246)",
    level3: "rgb(227, 232, 243)",
    level4: "rgb(225, 231, 241)",
    level5: "rgb(221, 228, 239)",
  },
};

const darkColors: MD3Theme["colors"] = {
  primary: "#A0CAFD",
  primaryContainer: "#194975",
  secondary: "#BBC7DB",
  secondaryContainer: "#3B4858",
  tertiary: "#D7BEE4",
  tertiaryContainer: "#523F5F",
  surface: "#111418",
  surfaceVariant: "#43474E",
  surfaceDisabled: "rgba(225, 226, 232, 0.12)",
  background: "#111418",
  error: "#FFB4AB",
  errorContainer: "#93000A",
  onPrimary: "#003258",
  onPrimaryContainer: "#D1E4FF",
  onSecondary: "#253140",
  onSecondaryContainer: "#D7E3F7",
  onTertiary: "#3B2948",
  onTertiaryContainer: "#F2DAFF",
  onSurface: "#E1E2E8",
  onSurfaceVariant: "#C3C7CF",
  onSurfaceDisabled: "rgba(225, 226, 232, 0.38)",
  onError: "#690005",
  onErrorContainer: "#FFDAD6",
  onBackground: "#E1E2E8",
  outline: "#8D9199",
  outlineVariant: "#43474E",
  inverseSurface: "#E1E2E8",
  inverseOnSurface: "#2E3135",
  inversePrimary: "#36618E",
  shadow: "#000000",
  scrim: "#000000",
  backdrop: "rgba(44, 49, 55, 0.4)",
  elevation: {
    level0: "transparent",
    level1: "rgb(24, 29, 35)",
    level2: "rgb(28, 35, 42)",
    level3: "rgb(33, 40, 49)",
    level4: "rgb(34, 42, 52)",
    level5: "rgb(37, 46, 56)",
  },
};

export const lightTheme: MD3Theme = { ...MD3LightTheme, colors: lightColors };
export const darkTheme: MD3Theme = { ...MD3DarkTheme, colors: darkColors };

// The native stack's header and screen background are drawn by native code,
// themed through Expo Router's ThemeProvider rather than Paper's, so they get
// the same tokens mapped onto React Navigation's theme shape. Its fonts stay
// the platform's, which is what Paper uses too.
const navigationTheme = (base: Theme, { colors }: MD3Theme): Theme => {
  return {
    ...base,
    colors: {
      primary: colors.primary,
      background: colors.background,
      card: colors.surface,
      text: colors.onSurface,
      border: colors.outlineVariant,
      notification: colors.error,
    },
  };
};

const lightNavigationTheme = navigationTheme(DefaultTheme, lightTheme);
const darkNavigationTheme = navigationTheme(DarkTheme, darkTheme);

export type AppTheme = { paper: MD3Theme; navigation: Theme };

export const useAppTheme = (): AppTheme => {
  const scheme = useColorScheme();
  return scheme === "dark"
    ? { paper: darkTheme, navigation: darkNavigationTheme }
    : { paper: lightTheme, navigation: lightNavigationTheme };
};
