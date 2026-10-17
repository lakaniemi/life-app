import { ThemeProvider } from "expo-router";
import { StatusBar } from "expo-status-bar";
import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { PaperProvider } from "react-native-paper";

import { QueryProvider } from "@/api/query-provider";
import { SnackbarProvider, useSnackbar } from "@/ui/snackbar";
import { useAppTheme } from "@/ui/theme";

// The composition root, and so the one place in ui/ that imports api/.
export const AppProviders = ({ children }: { children: ReactNode }) => {
  const theme = useAppTheme();

  return (
    <PaperProvider theme={theme.paper}>
      <ThemeProvider value={theme.navigation}>
        <SnackbarProvider>
          <QueryProviderWithSnackbar>{children}</QueryProviderWithSnackbar>
        </SnackbarProvider>
        <StatusBar style="auto" />
      </ThemeProvider>
    </PaperProvider>
  );
};

// Inside SnackbarProvider, so that a failed background refresh can show a
// snackbar.
const QueryProviderWithSnackbar = ({ children }: { children: ReactNode }) => {
  const { t } = useTranslation();
  const snackbar = useSnackbar();

  return (
    <QueryProvider
      onRefreshError={() => snackbar.show(t("errors.refreshFailed"))}
    >
      {children}
    </QueryProvider>
  );
};
