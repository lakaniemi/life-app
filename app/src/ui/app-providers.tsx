import { ThemeProvider } from "expo-router";
import { StatusBar } from "expo-status-bar";
import type { ReactNode } from "react";
import { PaperProvider } from "react-native-paper";

import { SnackbarProvider } from "@/ui/snackbar";
import { useAppTheme } from "@/ui/theme";

export const AppProviders = ({ children }: { children: ReactNode }) => {
  const theme = useAppTheme();

  return (
    <PaperProvider theme={theme.paper}>
      <ThemeProvider value={theme.navigation}>
        <SnackbarProvider>{children}</SnackbarProvider>
        <StatusBar style="auto" />
      </ThemeProvider>
    </PaperProvider>
  );
};
