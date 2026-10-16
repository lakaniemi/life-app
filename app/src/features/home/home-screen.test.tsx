import { render, screen } from "@testing-library/react-native";

import { HomeScreen } from "@/features/home/home-screen";

test("renders the app name", async () => {
  await render(<HomeScreen />);

  expect(screen.getByText("Life App")).toBeOnTheScreen();
});
