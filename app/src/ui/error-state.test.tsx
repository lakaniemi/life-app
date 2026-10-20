import { fireEvent, render, screen } from "@testing-library/react-native";

import { ErrorState } from "@/ui/error-state";

test("shows a generic message by default", async () => {
  await render(<ErrorState onRetry={jest.fn()} />);

  expect(screen.getByText("Something went wrong.")).toBeOnTheScreen();
});

test("shows the given message", async () => {
  await render(
    <ErrorState message="Couldn't load the list." onRetry={jest.fn()} />,
  );

  expect(screen.getByText("Couldn't load the list.")).toBeOnTheScreen();
});

test("has no Retry without onRetry", async () => {
  await render(<ErrorState />);

  expect(screen.queryByRole("button", { name: "Retry" })).not.toBeOnTheScreen();
});

test("retries when Retry is pressed", async () => {
  const onRetry = jest.fn();
  await render(<ErrorState onRetry={onRetry} />);

  await fireEvent.press(screen.getByRole("button", { name: "Retry" }));

  expect(onRetry).toHaveBeenCalledTimes(1);
});
