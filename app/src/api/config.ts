import Constants from "expo-constants";
import { z } from "zod";

// Set per variant in app.config.ts. Parsed when this module loads, so a build
// without a URL fails at launch instead of on its first request. A regex, not
// z.url(): React Native's URL class is a loose regex-based polyfill, so a URL
// check would mean less on the device than it does in tests. The trailing
// slash is dropped so that paths can be appended as "/health".
const urlSchema = z.string().regex(/^https?:\/\/[^\s/]+(\/\S*)?$/);

export const apiUrl = urlSchema
  .transform((url) => url.replace(/\/+$/, ""))
  .parse(Constants.expoConfig?.extra?.apiUrl);

// Where to find the latest release's `latest.json`, or null when this build
// doesn't check for updates (the development variant).
export const latestReleaseUrl = urlSchema
  .nullable()
  .parse(Constants.expoConfig?.extra?.latestReleaseUrl ?? null);
