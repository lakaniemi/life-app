import { act, fireEvent, render, screen } from "@testing-library/react-native";
import { Button } from "react-native";

import { SnackbarProvider, useSnackbar } from "@/ui/snackbar";

function ShowButton({ onRetry }: { onRetry: () => void }) {
  const { show } = useSnackbar();
  return (
    <Button
      title="Show"
      onPress={() => {
        show("Couldn't save the item.", {
          action: { label: "Retry", onPress: onRetry },
        });
        show("Couldn't refresh.");
      }}
    />
  );
}

async function renderAndShow(onRetry = jest.fn()) {
  await render(
    <SnackbarProvider>
      <ShowButton onRetry={onRetry} />
    </SnackbarProvider>,
  );
  await fireEvent.press(screen.getByRole("button", { name: "Show" }));
}

beforeEach(() => {
  jest.useFakeTimers();
});

afterEach(() => {
  jest.useRealTimers();
});

test("shows queued messages one at a time, in order", async () => {
  await renderAndShow();

  expect(screen.getByText("Couldn't save the item.")).toBeOnTheScreen();
  expect(screen.queryByText("Couldn't refresh.")).not.toBeOnTheScreen();

  await fireEvent.press(screen.getByRole("button", { name: "Retry" }));

  expect(screen.queryByText("Couldn't save the item.")).not.toBeOnTheScreen();
  expect(screen.getByText("Couldn't refresh.")).toBeOnTheScreen();
});

test("runs the action and dismisses the message", async () => {
  const onRetry = jest.fn();
  await renderAndShow(onRetry);

  await fireEvent.press(screen.getByRole("button", { name: "Retry" }));

  expect(onRetry).toHaveBeenCalledTimes(1);
  expect(screen.queryByText("Couldn't save the item.")).not.toBeOnTheScreen();
});

test("dismisses a message after its duration", async () => {
  await renderAndShow();

  // Paper's 200 ms entry animation, then its 7 s default duration.
  await act(() => jest.advanceTimersByTime(200 + 7000));

  expect(screen.queryByText("Couldn't save the item.")).not.toBeOnTheScreen();
  expect(screen.getByText("Couldn't refresh.")).toBeOnTheScreen();
});

test("useSnackbar throws outside the provider", async () => {
  jest.spyOn(console, "error").mockImplementation(() => {});

  await expect(render(<ShowButton onRetry={jest.fn()} />)).rejects.toThrow(
    "useSnackbar must be used inside a SnackbarProvider",
  );
});
