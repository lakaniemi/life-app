import { fireEvent, render, screen } from "@testing-library/react-native";

import { EmptyState } from "@/ui/empty-state";

test("renders the title and description", async () => {
  await render(
    <EmptyState
      icon="cart-outline"
      title="No lists yet"
      description="Lists you create show up here."
    />,
  );

  expect(screen.getByText("No lists yet")).toBeOnTheScreen();
  expect(screen.getByText("Lists you create show up here.")).toBeOnTheScreen();
  expect(screen.queryByRole("button")).not.toBeOnTheScreen();
});

test("runs the action when its button is pressed", async () => {
  const onPress = jest.fn();
  await render(
    <EmptyState
      icon="cart-outline"
      title="No lists yet"
      description="Lists you create show up here."
      action={{ label: "Create a list", onPress }}
    />,
  );

  await fireEvent.press(screen.getByRole("button", { name: "Create a list" }));

  expect(onPress).toHaveBeenCalledTimes(1);
});
