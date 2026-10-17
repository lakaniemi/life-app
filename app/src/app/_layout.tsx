import "@/i18n/i18n";

import { Stack } from "expo-router";

import { AppProviders } from "@/ui/app-providers";

const RootLayout = () => {
  return (
    <AppProviders>
      <Stack />
    </AppProviders>
  );
};

export default RootLayout;
