import type { ExpoConfig } from "expo/config";

// Evaluated in Node whenever Expo reads the config (start, prebuild, builds),
// never on the device. APP_VARIANT picks the variant; unset means development,
// so a production build is always asked for explicitly.
const variants = {
  development: {
    name: "Life App Dev",
    androidPackage: "com.lakaniemi.lifeapp.dev",
    scheme: "lifeapp-dev",
  },
  production: {
    name: "Life App",
    androidPackage: "com.lakaniemi.lifeapp",
    scheme: "lifeapp",
  },
} as const;

type Variant = keyof typeof variants;

function selectedVariant(): Variant {
  const value = process.env.APP_VARIANT ?? "development";
  if (!(value in variants)) {
    throw new Error(
      `Unknown APP_VARIANT "${value}"; expected one of: ${Object.keys(variants).join(", ")}`,
    );
  }
  return value as Variant;
}

const variant = variants[selectedVariant()];

const config: ExpoConfig = {
  name: variant.name,
  slug: "life-app",
  version: "1.0.0",
  orientation: "portrait",
  icon: "./assets/images/icon.png",
  scheme: variant.scheme,
  userInterfaceStyle: "automatic",
  ios: {
    bundleIdentifier: variant.androidPackage,
  },
  android: {
    package: variant.androidPackage,
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
  experiments: {
    typedRoutes: true,
    reactCompiler: true,
  },
};

export default config;
