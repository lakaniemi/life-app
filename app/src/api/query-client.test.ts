import { ApiError, NetworkError } from "@/api/errors";
import { createQueryClient, shouldRetry } from "@/api/query-client";

describe("shouldRetry", () => {
  test("retries when there was no response", () => {
    expect(shouldRetry(0, new NetworkError("timed out"))).toBe(true);
  });

  test("retries a server error", () => {
    expect(shouldRetry(0, new ApiError(503, "unknown", "unavailable"))).toBe(
      true,
    );
  });

  test("doesn't retry a client error", () => {
    expect(shouldRetry(0, new ApiError(404, "not_found", "not found"))).toBe(
      false,
    );
  });

  test("doesn't retry other errors, such as a schema mismatch", () => {
    expect(shouldRetry(0, new Error("invalid response"))).toBe(false);
  });

  test("stops after two retries", () => {
    expect(shouldRetry(1, new NetworkError("timed out"))).toBe(true);
    expect(shouldRetry(2, new NetworkError("timed out"))).toBe(false);
  });
});

describe("onRefreshError", () => {
  const failingQuery = {
    queryKey: ["item"],
    queryFn: () => Promise.reject(new Error("failed")),
    retry: false,
    // Stale at once, so that cached data is refetched rather than returned.
    staleTime: 0,
  };

  // The client's gc timers would otherwise keep Jest from exiting.
  let queryClient: ReturnType<typeof createQueryClient>;
  afterEach(() => {
    queryClient.clear();
  });

  test("isn't called when a query fails with nothing cached", async () => {
    const onRefreshError = jest.fn();
    queryClient = createQueryClient({ onRefreshError });

    await queryClient.fetchQuery(failingQuery).catch(() => {});

    expect(onRefreshError).not.toHaveBeenCalled();
  });

  test("is called once when queries with cached data fail together", async () => {
    const onRefreshError = jest.fn();
    queryClient = createQueryClient({ onRefreshError });
    queryClient.setQueryData(["a"], "cached");
    queryClient.setQueryData(["b"], "cached");

    await Promise.all([
      queryClient
        .fetchQuery({ ...failingQuery, queryKey: ["a"] })
        .catch(() => {}),
      queryClient
        .fetchQuery({ ...failingQuery, queryKey: ["b"] })
        .catch(() => {}),
    ]);

    expect(onRefreshError).toHaveBeenCalledTimes(1);
  });
});
