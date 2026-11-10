import { act, screen } from "@testing-library/react-native";
import { Text } from "react-native";

import { markAppOutdated } from "@/api/compat";
import { renderWithQueryClient } from "@/api/test-utils";
import { UpdateGate } from "@/features/app-update/update-gate";

// The screen checks for the newest release, which this test doesn't care
// about, so the check finds none.
beforeEach(() => {
  globalThis.fetch = jest.fn().mockResolvedValue({ ok: false, status: 404 });
});

// One test, because the outdated flag is never reset.
test("shows the app until the API turns it away, then only the update screen", async () => {
  await renderWithQueryClient(
    <UpdateGate>
      <Text>The app</Text>
    </UpdateGate>,
  );
  expect(screen.getByText("The app")).toBeOnTheScreen();

  await act(() => {
    markAppOutdated();
  });

  expect(screen.queryByText("The app")).not.toBeOnTheScreen();
  expect(screen.getByText("Update required")).toBeOnTheScreen();
});
