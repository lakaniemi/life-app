import { fireEvent, screen } from "@testing-library/react-native";
import { Linking } from "react-native";

import { latestApkUrl } from "@/api/config";
import { renderWithQueryClient } from "@/api/test-utils";
import {
  useApkInstall,
  type ApkInstallState,
} from "@/features/app-update/install-apk";
import { InstallUpdateScreen } from "@/features/app-update/install-update-screen";

// The running build is 0.0.5, with versionCode 5.
jest.mock("expo-application", () => ({
  nativeApplicationVersion: "0.0.5",
  nativeBuildVersion: "5",
}));

// The hook has its own tests. Here it sets the state the screen shows.
jest.mock("@/features/app-update/install-apk", () => ({
  ...jest.requireActual("@/features/app-update/install-apk"),
  useApkInstall: jest.fn(),
}));

const install = jest.fn();

const withState = (state: ApkInstallState) =>
  jest.mocked(useApkInstall).mockReturnValue({ state, install });

beforeEach(() => {
  globalThis.fetch = jest.fn().mockResolvedValue({
    ok: true,
    status: 200,
    text: () =>
      Promise.resolve(JSON.stringify({ version: "0.0.6", versionCode: 6 })),
  });
});

afterEach(() => {
  jest.clearAllMocks();
  jest.restoreAllMocks();
});

test("starts the download when it opens", async () => {
  withState({ status: "downloading", progress: null });

  await renderWithQueryClient(<InstallUpdateScreen />);

  expect(install).toHaveBeenCalledTimes(1);
});

describe("when the update is required", () => {
  test("explains that the app is too old", async () => {
    withState({ status: "idle" });

    await renderWithQueryClient(<InstallUpdateScreen required />);

    expect(screen.getByText("Update required")).toBeOnTheScreen();
    expect(
      screen.getByText(
        "This version of Life App is too old. Install the latest version to continue.",
      ),
    ).toBeOnTheScreen();
  });

  test("waits for Install instead of downloading at once", async () => {
    withState({ status: "idle" });
    await renderWithQueryClient(<InstallUpdateScreen required />);
    expect(install).not.toHaveBeenCalled();

    await fireEvent.press(screen.getByRole("button", { name: "Install" }));

    expect(install).toHaveBeenCalledTimes(1);
  });

  test("shows the progress", async () => {
    withState({ status: "downloading", progress: 0.5 });

    await renderWithQueryClient(<InstallUpdateScreen required />);

    expect(screen.getByText("Downloading the update… 50%")).toBeOnTheScreen();
  });

  test("falls back to the APK on GitHub when the download fails", async () => {
    const openURL = jest
      .spyOn(Linking, "openURL")
      .mockResolvedValue(undefined);
    withState({ status: "failed" });
    await renderWithQueryClient(<InstallUpdateScreen required />);

    await fireEvent.press(
      screen.getByRole("button", { name: "Download from GitHub" }),
    );

    expect(openURL).toHaveBeenCalledWith(latestApkUrl);
  });
});

test("shows the versions", async () => {
  withState({ status: "downloading", progress: null });

  await renderWithQueryClient(<InstallUpdateScreen />);

  expect(
    await screen.findByText("Current: 0.0.5 · New: 0.0.6"),
  ).toBeOnTheScreen();
});

test("shows the progress as a percentage", async () => {
  withState({ status: "downloading", progress: 0.424 });

  await renderWithQueryClient(<InstallUpdateScreen />);

  expect(screen.getByText("Downloading the update… 42%")).toBeOnTheScreen();
});

test("shows progress without a percentage when the size is unknown", async () => {
  withState({ status: "downloading", progress: null });

  await renderWithQueryClient(<InstallUpdateScreen />);

  expect(screen.getByText("Downloading the update…")).toBeOnTheScreen();
});

test("offers Install again once the installer is closed", async () => {
  withState({ status: "idle" });
  await renderWithQueryClient(<InstallUpdateScreen />);
  install.mockClear();

  await fireEvent.press(screen.getByRole("button", { name: "Install" }));

  expect(install).toHaveBeenCalledTimes(1);
});

describe("when the download fails", () => {
  beforeEach(() => withState({ status: "failed" }));

  test("says so and retries", async () => {
    await renderWithQueryClient(<InstallUpdateScreen />);
    install.mockClear();

    expect(screen.getByText("Couldn't download the update.")).toBeOnTheScreen();
    await fireEvent.press(screen.getByRole("button", { name: "Retry" }));

    expect(install).toHaveBeenCalledTimes(1);
  });

  test("falls back to the APK on GitHub", async () => {
    const openURL = jest
      .spyOn(Linking, "openURL")
      .mockResolvedValue(undefined);
    await renderWithQueryClient(<InstallUpdateScreen />);

    await fireEvent.press(
      screen.getByRole("button", { name: "Download from GitHub" }),
    );

    expect(openURL).toHaveBeenCalledWith(latestApkUrl);
  });
});
