import { File, Paths } from "expo-file-system";
import { startActivityAsync } from "expo-intent-launcher";
import { useCallback, useEffect, useRef, useState } from "react";

import { latestApkUrl } from "@/api/config";

const apkMimeType = "application/vnd.android.package-archive";

// Intent.FLAG_GRANT_READ_URI_PERMISSION. Without it the installer can't read
// our content:// URI, because the file belongs to this app.
const grantReadUriPermission = 1;

type DownloadAndInstallOptions = {
  // The fraction downloaded, or null when the server didn't say how big the
  // file is.
  onProgress: (progress: number | null) => void;
  signal: AbortSignal;
};

// Downloads the latest APK and hands it to Android's installer. Resolves when
// the user comes back from the installer, whether they installed or not.
export const downloadAndInstall = async ({
  onProgress,
  signal,
}: DownloadAndInstallOptions): Promise<void> => {
  const file = new File(Paths.cache, "life-app.apk");
  // A download onto an existing file fails, and a file may be left behind by an
  // earlier attempt.
  if (file.exists) {
    file.delete();
  }

  const task = File.createDownloadTask(latestApkUrl, file, {
    // totalBytes is -1 when the response has no Content-Length.
    onProgress: ({ bytesWritten, totalBytes }) =>
      onProgress(totalBytes > 0 ? bytesWritten / totalBytes : null),
    signal,
  });
  // null means the task was paused, which nothing here does.
  if (!(await task.downloadAsync())) {
    throw new Error("The download didn't finish");
  }

  await startActivityAsync("android.intent.action.VIEW", {
    data: file.contentUri,
    type: apkMimeType,
    flags: grantReadUriPermission,
  });
};

export type ApkInstallState =
  | { status: "idle" }
  | { status: "downloading"; progress: number | null }
  | { status: "failed" };

// The state of one download-and-install, for a screen to show. It's component
// state rather than a TanStack mutation: it's an action on this phone, not
// server data, so there's no cache to update and no retry or Retry snackbar to
// apply. Leaving the screen cancels a download that is still running.
export const useApkInstall = () => {
  const [state, setState] = useState<ApkInstallState>({ status: "idle" });
  const running = useRef<AbortController | null>(null);

  useEffect(() => {
    // Nothing to set up on mount: the download starts from install(). The
    // returned cleanup runs on unmount.
    return () => {
      running.current?.abort();
      running.current = null;
    };
  }, []);

  const install = useCallback(async () => {
    if (running.current) {
      return;
    }
    const controller = new AbortController();
    running.current = controller;
    setState({ status: "downloading", progress: null });
    try {
      await downloadAndInstall({
        onProgress: (progress) =>
          setState({ status: "downloading", progress }),
        signal: controller.signal,
      });
      setState({ status: "idle" });
    } catch {
      if (!controller.signal.aborted) {
        setState({ status: "failed" });
      }
    } finally {
      // Not if a remount already started a newer download.
      if (running.current === controller) {
        running.current = null;
      }
    }
  }, []);

  return { state, install };
};
