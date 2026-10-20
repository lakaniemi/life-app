import { QueryClient } from "@tanstack/react-query";

import { ApiError, NetworkError } from "@/api/errors";

// How long the persisted cache is kept on the device. Queries must stay in
// memory at least as long (gcTime), or a restored query would be garbage
// collected before anything uses it.
export const cacheMaxAge = 7 * 24 * 60 * 60 * 1000;

// Short, so that data is refetched often, but long enough that moving
// between screens doesn't refetch every time.
const staleTime = 30 * 1000;

const maxRetries = 2;

// Longer than a snackbar stays up, so that several queries failing at once
// (the API down) show one "Couldn't refresh", not a queue of them.
const refreshErrorThrottleMs = 10 * 1000;

// Retries only what may succeed on a second try: a request that failed to
// connect, or a server error. A 4xx or a response that doesn't match its
// schema won't change. Nor is a timeout retried: each attempt would wait the
// full timeout again, keeping a first load on its spinner for most of a
// minute before the error shows.
export const shouldRetry = (failureCount: number, error: unknown): boolean => {
  if (failureCount >= maxRetries) {
    return false;
  }
  if (error instanceof NetworkError) {
    return !error.timedOut;
  }
  return error instanceof ApiError && error.status >= 500;
};

export const createQueryClient = () =>
  new QueryClient({
    defaultOptions: {
      queries: { staleTime, gcTime: cacheMaxAge, retry: shouldRetry },
      // TanStack Query doesn't retry mutations by default.
      mutations: { retry: shouldRetry },
    },
  });

// Calls onRefreshError when a query fails (after its retries) while it still
// has earlier data to show. With nothing cached, the screen shows ErrorState
// itself. Returns the unsubscribe function, for use in an effect.
export const subscribeToRefreshErrors = (
  queryClient: QueryClient,
  onRefreshError: () => void,
) => {
  let lastRefreshErrorAt = -Infinity;

  return queryClient.getQueryCache().subscribe((event) => {
    if (event.type !== "updated" || event.action.type !== "error") {
      return;
    }
    if (event.query.state.data === undefined) {
      return;
    }
    const now = Date.now();
    if (now - lastRefreshErrorAt < refreshErrorThrottleMs) {
      return;
    }
    lastRefreshErrorAt = now;
    onRefreshError();
  });
};
