import "@/i18n/i18n";

import { Stack } from "expo-router";
import * as SplashScreen from "expo-splash-screen";

import { AppProviders } from "@/ui/app-providers";

// Kept up until the persisted query cache is restored (QueryProvider hides
// it), so the first screen renders with its cached data.
void SplashScreen.preventAutoHideAsync();

const RootLayout = () => {
  return (
    <AppProviders>
      <Stack />
    </AppProviders>
  );
};

export default RootLayout;
