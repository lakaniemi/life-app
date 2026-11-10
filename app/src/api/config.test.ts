// The real module, not the mock from jest.setup.ts that the other tests use,
// read against a config shaped like Expo's.
const loadConfigWith = (extra: Record<string, unknown>) => {
  jest.resetModules();
  jest.doMock("expo-constants", () => ({
    __esModule: true,
    default: { expoConfig: { extra } },
  }));
  return jest.requireActual<typeof import("@/api/config")>("@/api/config");
};

describe("latestReleaseUrl", () => {
  test("is the configured URL", () => {
    const config = loadConfigWith({
      apiUrl: "http://10.0.2.2:8080",
      latestReleaseUrl: "http://10.0.2.2:8000/latest.json",
    });

    expect(config.latestReleaseUrl).toBe("http://10.0.2.2:8000/latest.json");
  });

  // The development variant's config leaves it out. Expo's config resolver
  // would turn a null into {}, which crashed the app at launch.
  test("is undefined when the config leaves it out", () => {
    const config = loadConfigWith({ apiUrl: "http://10.0.2.2:8080" });

    expect(config.latestReleaseUrl).toBeUndefined();
  });
});
