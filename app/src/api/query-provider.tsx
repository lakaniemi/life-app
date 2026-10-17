import AsyncStorage from "@react-native-async-storage/async-storage";
import { createAsyncStoragePersister } from "@tanstack/query-async-storage-persister";
import { PersistQueryClientProvider } from "@tanstack/react-query-persist-client";
import { nativeApplicationVersion, nativeBuildVersion } from "expo-application";
import * as SplashScreen from "expo-splash-screen";
import { useState, type ReactNode } from "react";

import { connectQueryManagers } from "@/api/managers";
import { cacheMaxAge, createQueryClient } from "@/api/query-client";

connectQueryManagers();

const persister = createAsyncStoragePersister({ storage: AsyncStorage });

// A new app version starts with an empty cache instead of reading data cached
// under older schemas. Until over-the-air updates exist, the version is the
// APK's; with them, it becomes the running update's id.
const buster = `${nativeApplicationVersion}+${nativeBuildVersion}`;

type QueryProviderProps = {
  children: ReactNode;
  // Called when a query fails while showing earlier data. Read once, when the
  // query client is created.
  onRefreshError: () => void;
};

// The query client, with its cache persisted on the device so the app starts
// with the last data it loaded, and shows it while offline.
export const QueryProvider = ({
  children,
  onRefreshError,
}: QueryProviderProps) => {
  const [queryClient] = useState(() => createQueryClient({ onRefreshError }));

  // The root layout keeps the splash screen up until the cache is restored,
  // so a screen never flashes its loading state before the cached data.
  // onSuccess also runs when there was nothing to restore.
  return (
    <PersistQueryClientProvider
      client={queryClient}
      persistOptions={{ persister, maxAge: cacheMaxAge, buster }}
      onSuccess={SplashScreen.hide}
      onError={SplashScreen.hide}
    >
      {children}
    </PersistQueryClientProvider>
  );
};
