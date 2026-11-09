import { act, renderHook } from "@testing-library/react-native";

import { markAppOutdated, useIsAppOutdated } from "@/api/compat";

// One test, because the flag is never reset, which is the point of it.
test("tells subscribers once the app is marked outdated", async () => {
  const { result } = await renderHook(() => useIsAppOutdated());
  expect(result.current).toBe(false);

  await act(() => {
    markAppOutdated();
  });

  expect(result.current).toBe(true);
});
