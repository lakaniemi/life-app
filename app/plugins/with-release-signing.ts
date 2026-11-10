import { type ConfigPlugin, withAppBuildGradle } from "expo/config-plugins";

// Gradle properties that carry the release key. CI sets them through
// ORG_GRADLE_PROJECT_<name> environment variables; locally they are normally
// unset, and release builds then fall back to the debug key.
const STORE_FILE = "lifeAppUploadStoreFile";
const STORE_PASSWORD = "lifeAppUploadStorePassword";
const KEY_ALIAS = "lifeAppUploadKeyAlias";

const RELEASE_SIGNING_CONFIG = `
        release {
            if (findProperty('${STORE_FILE}')) {
                storeFile file(findProperty('${STORE_FILE}'))
                storePassword findProperty('${STORE_PASSWORD}')
                keyAlias findProperty('${KEY_ALIAS}')
                // A PKCS12 keystore has one password for the file and the key.
                keyPassword findProperty('${STORE_PASSWORD}')
            }
        }`;

const DEBUG_SIGNING_CONFIG = /signingConfigs \{\s*debug \{[^}]*\}/;
const RELEASE_BUILD_TYPE_SIGNING =
  /(buildTypes \{[\s\S]*?\brelease \{[\s\S]*?)signingConfig signingConfigs\.debug/;

// Pure string transform, so it can be tested without running a prebuild. It
// throws when the template isn't what it expects: a silent no-op would ship
// release APKs signed with the debug key.
export function addReleaseSigning(gradle: string): string {
  if (gradle.includes(STORE_FILE)) {
    return gradle;
  }
  if (
    !DEBUG_SIGNING_CONFIG.test(gradle) ||
    !RELEASE_BUILD_TYPE_SIGNING.test(gradle)
  ) {
    throw new Error(
      "with-release-signing: android/app/build.gradle doesn't look like the " +
        "expected template (signingConfigs.debug and buildTypes.release). " +
        "Update plugins/with-release-signing.ts for the new template.",
    );
  }
  return gradle
    .replace(
      DEBUG_SIGNING_CONFIG,
      (debugConfig) => debugConfig + RELEASE_SIGNING_CONFIG,
    )
    .replace(
      RELEASE_BUILD_TYPE_SIGNING,
      (_match, beforeSigning: string) =>
        `${beforeSigning}signingConfig findProperty('${STORE_FILE}') ? signingConfigs.release : signingConfigs.debug`,
    );
}

const withReleaseSigning: ConfigPlugin = (config) =>
  withAppBuildGradle(config, (config) => {
    config.modResults.contents = addReleaseSigning(config.modResults.contents);
    return config;
  });

export default withReleaseSigning;
