import "@/i18n/i18n";

import { Stack } from "expo-router";
import * as SplashScreen from "expo-splash-screen";

import { UpdateGate } from "@/features/app-update/update-gate";
import { AppProviders } from "@/ui/app-providers";

// Kept up until the persisted query cache is restored (QueryProvider hides
// it), so the first screen renders with its cached data.
void SplashScreen.preventAutoHideAsync();

const RootLayout = () => {
  return (
    <AppProviders>
      <UpdateGate>
        <Stack />
      </UpdateGate>
    </AppProviders>
  );
};

export default RootLayout;
