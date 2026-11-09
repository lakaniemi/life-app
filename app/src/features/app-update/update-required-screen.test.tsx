import { fireEvent, render, screen } from "@testing-library/react-native";
import { Linking } from "react-native";

import {
  latestApkUrl,
  UpdateRequiredScreen,
} from "@/features/app-update/update-required-screen";

test("explains that the app is too old", async () => {
  await render(<UpdateRequiredScreen />);

  expect(screen.getByText("Update required")).toBeOnTheScreen();
  expect(
    screen.getByText(
      "This version of Life App is too old. Install the latest version to continue.",
    ),
  ).toBeOnTheScreen();
});

test("opens the latest APK when Download update is pressed", async () => {
  const openURL = jest.spyOn(Linking, "openURL").mockResolvedValue(undefined);
  await render(<UpdateRequiredScreen />);

  await fireEvent.press(
    screen.getByRole("button", { name: "Download update" }),
  );

  expect(openURL).toHaveBeenCalledWith(latestApkUrl);
});
