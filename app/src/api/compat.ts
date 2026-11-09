import { useSyncExternalStore } from "react";

// The API compatibility level this build speaks, sent with every request. The
// API turns away levels below its minimum. Bump it in the update that starts
// depending on a breaking API change (docs/ARCHITECTURE.md, "Version
// compatibility").
export const appCompatLevel = 1;

export const appCompatLevelHeader = "App-Compat-Level";

// The error code of a request the API turned away for its level being too low.
export const appOutdatedCode = "app_outdated";

// Whether the API has turned this build away. It lives outside React because
// the API client, which notices it, isn't a component, like TanStack Query's
// onlineManager. It is never reset: only a new version of the app fixes it.
let isOutdated = false;
const listeners = new Set<() => void>();

export const markAppOutdated = () => {
  if (isOutdated) {
    return;
  }
  isOutdated = true;
  listeners.forEach((listener) => listener());
};

const subscribe = (listener: () => void) => {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
};

export const useIsAppOutdated = () =>
  useSyncExternalStore(subscribe, () => isOutdated);
