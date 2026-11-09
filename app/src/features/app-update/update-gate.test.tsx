import { act, render, screen } from "@testing-library/react-native";
import { Text } from "react-native";

import { markAppOutdated } from "@/api/compat";
import { UpdateGate } from "@/features/app-update/update-gate";

// One test, because the outdated flag is never reset.
test("shows the app until the API turns it away, then only the update screen", async () => {
  await render(
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
