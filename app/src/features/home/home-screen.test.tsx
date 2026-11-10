import { onlineManager } from "@tanstack/react-query";
import { act, fireEvent, screen } from "@testing-library/react-native";

import { apiFetch } from "@/api/client";
import { NetworkError } from "@/api/errors";
import { renderWithQueryClient } from "@/api/test-utils";
import { HomeScreen } from "@/features/home/home-screen";

jest.mock("@/api/client");
// The running build's versionCode is 5.
jest.mock("expo-application", () => ({
  nativeApplicationVersion: "0.0.5",
  nativeBuildVersion: "5",
}));

const apiFetchMock = jest.mocked(apiFetch);
const fetchMock = jest.fn();

beforeEach(() => {
  // The update banner's check. No release unless a test says otherwise.
  globalThis.fetch = fetchMock;
  fetchMock.mockResolvedValue({
    ok: false,
    status: 404,
    text: () => Promise.resolve("Not Found"),
  });
});

afterEach(() => {
  apiFetchMock.mockReset();
  fetchMock.mockReset();
  onlineManager.setOnline(true);
});

test("shows the update banner above whatever state Home is in", async () => {
  fetchMock.mockResolvedValue({
    ok: true,
    status: 200,
    text: () => Promise.resolve('{"version":"0.0.6","versionCode":6}'),
  });
  apiFetchMock.mockReturnValue(new Promise(() => {}));

  await renderWithQueryClient(<HomeScreen />);

  expect(
    await screen.findByText(/A new version is available/),
  ).toBeOnTheScreen();
  expect(screen.getByLabelText("Loading")).toBeOnTheScreen();
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

  expect(
    await screen.findByText(
      "Couldn't connect. Check your connection and try again.",
    ),
  ).toBeOnTheScreen();

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
