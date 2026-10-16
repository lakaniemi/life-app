import "@/i18n/i18n";

import { Stack } from "expo-router";

import { AppProviders } from "@/ui/app-providers";

export default function RootLayout() {
  return (
    <AppProviders>
      <Stack />
    </AppProviders>
  );
}
