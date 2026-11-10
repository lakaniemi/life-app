import { act, renderHook, waitFor } from "@testing-library/react-native";
import { startActivityAsync } from "expo-intent-launcher";

import {
  downloadAndInstall,
  latestApkUrl,
  useApkInstall,
} from "@/features/app-update/install-apk";

const mockFile = {
  exists: false,
  delete: jest.fn(),
  contentUri: "content://life-app.test/life-app.apk",
};
const mockDownloadAsync = jest.fn();
const mockCreateDownloadTask = jest.fn((..._args: unknown[]) => ({
  downloadAsync: mockDownloadAsync,
}));

jest.mock("expo-file-system", () => ({
  File: Object.assign(
    jest.fn(() => mockFile),
    {
      createDownloadTask: (...args: unknown[]) => mockCreateDownloadTask(...args),
    },
  ),
  Paths: { cache: "file:///cache" },
}));
jest.mock("expo-intent-launcher", () => ({
  startActivityAsync: jest.fn(),
}));

type TaskOptions = {
  onProgress: (data: { bytesWritten: number; totalBytes: number }) => void;
  signal: AbortSignal;
};

// The options of the download task most recently created.
const taskOptions = () =>
  mockCreateDownloadTask.mock.lastCall?.[2] as TaskOptions;

beforeEach(() => {
  mockFile.exists = false;
  mockDownloadAsync.mockResolvedValue(mockFile);
  jest.mocked(startActivityAsync).mockResolvedValue({ resultCode: 0 });
});

afterEach(() => {
  jest.clearAllMocks();
});

describe("downloadAndInstall", () => {
  const run = () =>
    downloadAndInstall({
      onProgress: jest.fn(),
      signal: new AbortController().signal,
    });

  test("downloads the APK and opens it in the installer", async () => {
    await run();

    expect(mockCreateDownloadTask).toHaveBeenCalledWith(
      latestApkUrl,
      mockFile,
      expect.anything(),
    );
    expect(startActivityAsync).toHaveBeenCalledWith(
      "android.intent.action.VIEW",
      {
        data: mockFile.contentUri,
        type: "application/vnd.android.package-archive",
        flags: 1,
      },
    );
  });

  test("removes an APK left behind by an earlier attempt", async () => {
    mockFile.exists = true;

    await run();

    expect(mockFile.delete).toHaveBeenCalled();
  });

  test("reports the fraction downloaded, or null without a size", async () => {
    const onProgress = jest.fn();
    await downloadAndInstall({
      onProgress,
      signal: new AbortController().signal,
    });

    taskOptions().onProgress({ bytesWritten: 50, totalBytes: 200 });
    taskOptions().onProgress({ bytesWritten: 50, totalBytes: -1 });

    expect(onProgress.mock.calls).toEqual([[0.25], [null]]);
  });

  test("doesn't open the installer when the download didn't finish", async () => {
    mockDownloadAsync.mockResolvedValue(null);

    await expect(run()).rejects.toThrow();

    expect(startActivityAsync).not.toHaveBeenCalled();
  });
});

describe("useApkInstall", () => {
  test("starts idle, downloads, and is idle again after the installer", async () => {
    const { result } = await renderHook(() => useApkInstall());
    expect(result.current.state).toEqual({ status: "idle" });

    let finishDownload: (file: unknown) => void = () => {};
    mockDownloadAsync.mockReturnValue(
      new Promise((resolve) => {
        finishDownload = resolve;
      }),
    );
    await act(async () => {
      void result.current.install();
    });
    expect(result.current.state).toEqual({
      status: "downloading",
      progress: null,
    });

    await act(async () => {
      taskOptions().onProgress({ bytesWritten: 1, totalBytes: 4 });
    });
    expect(result.current.state).toEqual({
      status: "downloading",
      progress: 0.25,
    });

    await act(async () => {
      finishDownload(mockFile);
    });
    await waitFor(() =>
      expect(result.current.state).toEqual({ status: "idle" }),
    );
    expect(startActivityAsync).toHaveBeenCalledTimes(1);
  });

  test("fails when the download fails, and can be retried", async () => {
    mockDownloadAsync.mockRejectedValueOnce(new Error("offline"));
    const { result } = await renderHook(() => useApkInstall());

    await act(async () => {
      await result.current.install();
    });
    expect(result.current.state).toEqual({ status: "failed" });

    await act(async () => {
      await result.current.install();
    });
    expect(result.current.state).toEqual({ status: "idle" });
    expect(startActivityAsync).toHaveBeenCalledTimes(1);
  });

  test("ignores a second install while one is downloading", async () => {
    mockDownloadAsync.mockReturnValue(new Promise(() => {}));
    const { result } = await renderHook(() => useApkInstall());

    await act(async () => {
      void result.current.install();
      void result.current.install();
    });

    expect(mockCreateDownloadTask).toHaveBeenCalledTimes(1);
  });

  test("cancels a running download when it unmounts, without failing", async () => {
    mockDownloadAsync.mockReturnValue(new Promise(() => {}));
    const { result, unmount } = await renderHook(() => useApkInstall());
    await act(async () => {
      void result.current.install();
    });
    const { signal } = taskOptions();

    await unmount();

    expect(signal.aborted).toBe(true);
  });
});
