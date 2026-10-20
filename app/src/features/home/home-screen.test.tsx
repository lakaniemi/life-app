import { onlineManager } from "@tanstack/react-query";
import { act, fireEvent, screen } from "@testing-library/react-native";

import { apiFetch } from "@/api/client";
import { NetworkError } from "@/api/errors";
import { renderWithQueryClient } from "@/api/test-utils";
import { HomeScreen } from "@/features/home/home-screen";

jest.mock("@/api/client");

const apiFetchMock = jest.mocked(apiFetch);

afterEach(() => {
  apiFetchMock.mockReset();
  onlineManager.setOnline(true);
});

test("shows the API's health", async () => {
  apiFetchMock.mockResolvedValue({
    status: "ok",
    database: "ok",
    version: "0123456789abcdef",
  });

  await renderWithQueryClient(<HomeScreen />);

  expect(await screen.findByText("The API is up.")).toBeOnTheScreen();
  expect(screen.getByText("API version 0123456")).toBeOnTheScreen();
  expect(apiFetchMock).toHaveBeenCalledWith(
    "/health",
    expect.anything(),
    expect.anything(),
  );
});

test("shows loading until the API answers", async () => {
  apiFetchMock.mockReturnValue(new Promise(() => {}));

  await renderWithQueryClient(<HomeScreen />);

  expect(screen.getByLabelText("Loading")).toBeOnTheScreen();
});

test("shows the error with Retry, and Retry loads again", async () => {
  apiFetchMock.mockRejectedValueOnce(new NetworkError("no response"));
  apiFetchMock.mockResolvedValueOnce({
    status: "ok",
    database: "ok",
    version: "0123456",
  });

  await renderWithQueryClient(<HomeScreen />);

  expect(await screen.findByText("Couldn't connect.")).toBeOnTheScreen();

  await fireEvent.press(screen.getByRole("button", { name: "Retry" }));

  expect(await screen.findByText("The API is up.")).toBeOnTheScreen();
});

test("says it's offline with nothing cached, and loads once back online", async () => {
  apiFetchMock.mockResolvedValue({
    status: "ok",
    database: "ok",
    version: "0123456",
  });
  onlineManager.setOnline(false);

  await renderWithQueryClient(<HomeScreen />);

  expect(
    screen.getByText("You're offline. This will load once you're back online."),
  ).toBeOnTheScreen();
  expect(apiFetchMock).not.toHaveBeenCalled();

  await act(() => onlineManager.setOnline(true));

  expect(await screen.findByText("The API is up.")).toBeOnTheScreen();
});
