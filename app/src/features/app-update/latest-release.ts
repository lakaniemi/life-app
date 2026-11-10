import { nativeBuildVersion } from "expo-application";
import { z } from "zod";

import { latestReleaseUrl } from "@/api/config";
import { useApiQuery } from "@/api/use-api-query";

// How often the check may hit GitHub. TanStack Query refetches a stale query
// when the app returns to the foreground (managers.ts), so this staleTime is
// also the throttle: coming back sooner than this reuses the earlier answer.
const checkIntervalMs = 60 * 60 * 1000;

// React Native's fetch has no timeout of its own, and a request that hangs
// would keep the query "fetching" for the rest of the process.
const requestTimeoutMs = 10_000;

// The release workflow publishes this next to the APK, copied from
// version.json.
const latestReleaseSchema = z.object({
  version: z.string(),
  versionCode: z.number().int().positive(),
});

export type LatestRelease = z.infer<typeof latestReleaseSchema>;

export const appUpdateKeys = {
  latestRelease: () => ["app-update", "latest-release"] as const,
};

// The latest release, or null when there's nothing to act on. Every failure
// ends up here as null, with no error: a missing release (404), no connection
// and a malformed file all mean the same to the user, which is nothing to see.
// This is why it doesn't use apiFetch, which also sends our API's headers and
// turns failures into errors for screens to show.
export const fetchLatestRelease = async (
  url: string,
): Promise<LatestRelease | null> => {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), requestTimeoutMs);
  try {
    const response = await fetch(url, { signal: controller.signal });
    if (!response.ok) {
      return null;
    }
    // GitHub serves release assets as application/octet-stream, so the body
    // is parsed as text and the content type is ignored.
    const parsed = latestReleaseSchema.safeParse(
      JSON.parse(await response.text()),
    );
    return parsed.success ? parsed.data : null;
  } catch {
    return null;
  } finally {
    clearTimeout(timeout);
  }
};

// The running build's versionCode, or null when it isn't a usable number.
const ownVersionCode = (): number | null => {
  const code = Number(nativeBuildVersion);
  return Number.isInteger(code) && code > 0 ? code : null;
};

export const isNewer = (
  latest: LatestRelease,
  ownCode: number | null = ownVersionCode(),
): boolean => ownCode !== null && latest.versionCode > ownCode;

// The release to offer, or null: nothing newer, no release, a failed check,
// or a build that doesn't check at all (the development variant).
export const useNewerRelease = () =>
  useApiQuery({
    queryKey: appUpdateKeys.latestRelease(),
    queryFn: () => fetchLatestRelease(latestReleaseUrl!),
    enabled: latestReleaseUrl !== undefined,
    staleTime: checkIntervalMs,
    // What is newer depends on the running build, and an old answer would
    // show a banner offline at launch, where Install can't work.
    meta: { persist: false },
    select: (latest) => (latest && isNewer(latest) ? latest : null),
  });
