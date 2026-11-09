import type { FetchStatus } from "@tanstack/react-query";

// Offline, TanStack Query doesn't start a query's first fetch but pauses it:
// it stays pending until the connection returns, and then loads by itself.
// Screens check this before showing their loading state, so that offline
// with nothing cached they explain why instead of spinning indefinitely.
export const isWaitingForConnection = (query: {
  isPending: boolean;
  fetchStatus: FetchStatus;
}): boolean => query.isPending && query.fetchStatus === "paused";
