import { focusManager } from "@tanstack/react-query";
import {
  act,
  fireEvent,
  screen,
  waitFor,
} from "@testing-library/react-native";
import { router } from "expo-router";

import { renderWithQueryClient } from "@/api/test-utils";
import { UpdateBanner } from "@/features/app-update/update-banner";

// The running build is 0.0.5, with versionCode 5.
jest.mock("expo-application", () => ({
  nativeApplicationVersion: "0.0.5",
  nativeBuildVersion: "5",
}));

const fetchMock = jest.fn();

beforeEach(() => {
  globalThis.fetch = fetchMock;
});

afterEach(() => {
  fetchMock.mockReset();
  jest.restoreAllMocks();
});

const respondWith = (body: string, status = 200) =>
  fetchMock.mockResolvedValue({
    ok: status >= 200 && status < 300,
    status,
    text: () => Promise.resolve(body),
  });

const respondWithRelease = (versionCode: number) =>
  respondWith(JSON.stringify({ version: `0.0.${versionCode}`, versionCode }));

const banner = () => screen.queryByText(/A new version is available/);
const findBanner = () => screen.findByText(/A new version is available/);

test("offers a newer release, with the versions on a row of their own", async () => {
  respondWithRelease(6);

  await renderWithQueryClient(<UpdateBanner />);

  expect(await findBanner()).toHaveTextContent(
    "A new version is available. Update now?\n" +
      "Current: 0.0.5 · New: 0.0.6",
  );
  expect(screen.getByText("Current: 0.0.5 · New: 0.0.6")).toBeOnTheScreen();
});

test.each([
  ["the same", 5],
  ["an older", 4],
])("shows nothing for %s release", async (_, versionCode) => {
  respondWithRelease(versionCode);

  await renderWithQueryClient(<UpdateBanner />);

  await waitFor(() => expect(fetchMock).toHaveBeenCalled());
  expect(banner()).not.toBeOnTheScreen();
});

test("shows nothing without a release", async () => {
  respondWith("Not Found", 404);

  await renderWithQueryClient(<UpdateBanner />);

  await waitFor(() => expect(fetchMock).toHaveBeenCalled());
  expect(banner()).not.toBeOnTheScreen();
});

test("shows nothing for a bad file", async () => {
  respondWith("not json");

  await renderWithQueryClient(<UpdateBanner />);

  await waitFor(() => expect(fetchMock).toHaveBeenCalled());
  expect(banner()).not.toBeOnTheScreen();
});

test("shows nothing when the request fails", async () => {
  fetchMock.mockRejectedValue(new TypeError("Network request failed"));

  await renderWithQueryClient(<UpdateBanner />);

  await waitFor(() => expect(fetchMock).toHaveBeenCalled());
  expect(banner()).not.toBeOnTheScreen();
});

test("doesn't check in a build without a release URL", async () => {
  // The mock module from jest.setup.ts, not an import of it, which Babel's
  // interop would copy.
  jest.replaceProperty(
    jest.requireMock("@/api/config"),
    "latestReleaseUrl",
    undefined,
  );
  respondWithRelease(6);

  await renderWithQueryClient(<UpdateBanner />);

  expect(fetchMock).not.toHaveBeenCalled();
  expect(banner()).not.toBeOnTheScreen();
});

test("Later hides the banner", async () => {
  respondWithRelease(6);
  await renderWithQueryClient(<UpdateBanner />);
  await findBanner();

  await fireEvent.press(screen.getByRole("button", { name: "Later" }));

  expect(banner()).not.toBeOnTheScreen();
});

test("Install opens the install screen", async () => {
  const push = jest.spyOn(router, "push").mockImplementation(() => {});
  respondWithRelease(6);
  await renderWithQueryClient(<UpdateBanner />);
  await findBanner();

  await fireEvent.press(screen.getByRole("button", { name: "Install" }));

  expect(push).toHaveBeenCalledWith("/update");
});

describe("the throttle", () => {
  const returnToForeground = () =>
    act(() => {
      focusManager.setFocused(false);
      focusManager.setFocused(true);
    });

  test("checks again on returning to the foreground only after an hour", async () => {
    respondWithRelease(6);
    await renderWithQueryClient(<UpdateBanner />);
    await findBanner();
    expect(fetchMock).toHaveBeenCalledTimes(1);

    const now = Date.now();
    jest.spyOn(Date, "now").mockReturnValue(now + 59 * 60 * 1000);
    await returnToForeground();
    expect(fetchMock).toHaveBeenCalledTimes(1);

    jest.spyOn(Date, "now").mockReturnValue(now + 61 * 60 * 1000);
    await returnToForeground();
    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(2));
  });
});
