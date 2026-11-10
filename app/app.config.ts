import type { ExpoConfig } from "expo/config";

import withReleaseSigning from "./plugins/with-release-signing.ts";
import appVersion from "./version.json";

// Evaluated in Node whenever Expo reads the config (start, prebuild, builds),
// never on the device. APP_VARIANT picks the variant; unset means development,
// so a production build is always asked for explicitly.
const variants = {
  development: {
    name: "Life App Dev",
    androidPackage: "com.lakaniemi.lifeapp.dev",
    scheme: "lifeapp-dev",
    // The Android emulator reaches the Mac's localhost at 10.0.2.2.
    apiUrl: "http://10.0.2.2:8080",
  },
  production: {
    name: "Life App",
    androidPackage: "com.lakaniemi.lifeapp",
    scheme: "lifeapp",
    apiUrl: "https://life-app-api.lakaniemi.com",
  },
} as const;

type Variant = keyof typeof variants;

const selectedVariant = (): Variant => {
  const value = process.env.APP_VARIANT ?? "development";
  if (!(value in variants)) {
    throw new Error(
      `Unknown APP_VARIANT "${value}"; expected one of: ${Object.keys(variants).join(", ")}`,
    );
  }
  return value as Variant;
};

const variantName = selectedVariant();
const variant = variants[variantName];

const config: ExpoConfig = {
  name: variant.name,
  slug: "life-app",
  // version is the label people see; versionCode is the integer Android
  // compares, and it refuses to install a lower one over a higher one. Both
  // are bumped by hand in version.json, which the release workflow watches.
  version: appVersion.version,
  orientation: "portrait",
  icon: "./assets/images/icon.png",
  scheme: variant.scheme,
  userInterfaceStyle: "automatic",
  ios: {
    bundleIdentifier: variant.androidPackage,
  },
  android: {
    package: variant.androidPackage,
    versionCode: appVersion.versionCode,
    adaptiveIcon: {
      backgroundColor: "#E6F4FE",
      foregroundImage: "./assets/images/android-icon-foreground.png",
      backgroundImage: "./assets/images/android-icon-background.png",
      monochromeImage: "./assets/images/android-icon-monochrome.png",
    },
    predictiveBackGestureEnabled: false,
  },
  plugins: [
    "expo-router",
    "expo-localization",
    [
      "expo-splash-screen",
      {
        backgroundColor: "#208AEF",
        image: "./assets/images/splash-icon.png",
        imageWidth: 76,
      },
    ],
  ],
  extra: {
    // API_URL overrides the development variant's URL, e.g. with a LAN IP for
    // a physical phone. In a development build the config comes from Metro, so
    // restarting Metro picks up a new value without rebuilding. Production
    // ignores it, so a stray API_URL can't ship a build pointing at a dev
    // address. || rather than ??, so that an empty API_URL falls back too.
    apiUrl:
      (variantName === "development" && process.env.API_URL) || variant.apiUrl,
    // Where the app looks for a newer release. The development variant has no
    // release to compare with, so it leaves this out and doesn't check, unless
    // LATEST_RELEASE_URL points it at a file to try the banner with. Same
    // restart-only rule as API_URL; production ignores it. Left out rather
    // than null: Expo's config resolver turns a null into {}.
    latestReleaseUrl:
      variantName === "production"
        ? "https://github.com/lakaniemi/life-app/releases/latest/download/latest.json"
        : process.env.LATEST_RELEASE_URL || undefined,
  },
  experiments: {
    typedRoutes: true,
    reactCompiler: true,
  },
};

// A local plugin is applied by wrapping the config; the `plugins` array only
// takes package names and paths.
export default withReleaseSigning(config);
