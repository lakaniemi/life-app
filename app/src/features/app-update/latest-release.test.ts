import {
  fetchLatestRelease,
  isNewer,
} from "@/features/app-update/latest-release";

const url = "https://releases.test/latest.json";

const fetchMock = jest.fn();

beforeEach(() => {
  globalThis.fetch = fetchMock;
});

afterEach(() => {
  fetchMock.mockReset();
});

const respondWith = (body: string, status = 200) =>
  fetchMock.mockResolvedValue({
    ok: status >= 200 && status < 300,
    status,
    // GitHub serves the file as application/octet-stream, so it has no
    // json() to lean on here, only text().
    text: () => Promise.resolve(body),
  });

describe("fetchLatestRelease", () => {
  test("parses the file", async () => {
    respondWith('{"version":"0.0.2","versionCode":2}');

    expect(await fetchLatestRelease(url)).toEqual({
      version: "0.0.2",
      versionCode: 2,
    });
    expect(fetchMock).toHaveBeenCalledWith(url, expect.anything());
  });

  test("is null without a release (404)", async () => {
    respondWith("Not Found", 404);

    expect(await fetchLatestRelease(url)).toBeNull();
  });

  test("is null for a server error", async () => {
    respondWith("", 500);

    expect(await fetchLatestRelease(url)).toBeNull();
  });

  test("is null for a body that isn't JSON", async () => {
    respondWith("<html>Sign in to the Wi-Fi</html>");

    expect(await fetchLatestRelease(url)).toBeNull();
  });

  test("is null for a body of the wrong shape", async () => {
    respondWith('{"version":"0.0.2","versionCode":"2"}');

    expect(await fetchLatestRelease(url)).toBeNull();
  });

  test("is null without a connection", async () => {
    fetchMock.mockRejectedValue(new TypeError("Network request failed"));

    expect(await fetchLatestRelease(url)).toBeNull();
  });
});

describe("isNewer", () => {
  const release = { version: "0.0.5", versionCode: 5 };

  test("is true for a higher versionCode", () => {
    expect(isNewer(release, 4)).toBe(true);
  });

  test("is false for the same versionCode", () => {
    expect(isNewer(release, 5)).toBe(false);
  });

  test("is false for a lower versionCode", () => {
    expect(isNewer(release, 6)).toBe(false);
  });

  test("is false when the running versionCode is unknown", () => {
    expect(isNewer(release, null)).toBe(false);
  });
});
