import { onlineManager } from "@tanstack/react-query";
import { act, render, screen } from "@testing-library/react-native";

import { OfflineBanner } from "@/ui/offline-banner";

afterEach(() => {
  onlineManager.setOnline(true);
});

test("is hidden while online", async () => {
  await render(<OfflineBanner />);

  expect(
    screen.queryByText("Offline · showing saved data"),
  ).not.toBeOnTheScreen();
});

test("shows while offline and hides when the connection returns", async () => {
  await render(<OfflineBanner />);

  await act(() => onlineManager.setOnline(false));
  expect(screen.getByText("Offline · showing saved data")).toBeOnTheScreen();

  await act(() => onlineManager.setOnline(true));
  expect(
    screen.queryByText("Offline · showing saved data"),
  ).not.toBeOnTheScreen();
});
