import { addReleaseSigning } from "./with-release-signing";

// The relevant part of the build.gradle that `expo prebuild` generates.
const template = `
    signingConfigs {
        debug {
            storeFile file('debug.keystore')
            storePassword 'android'
            keyAlias 'androiddebugkey'
            keyPassword 'android'
        }
    }
    buildTypes {
        debug {
            signingConfig signingConfigs.debug
        }
        release {
            // Caution! In production, you need to generate your own keystore file.
            signingConfig signingConfigs.debug
            minifyEnabled enableMinifyInReleaseBuilds
        }
    }
`;

describe("addReleaseSigning", () => {
  it("adds a release signing config that reads the Gradle properties", () => {
    const result = addReleaseSigning(template);

    expect(result).toContain("release {\n            if (findProperty(");
    expect(result).toContain("storeFile file(findProperty('lifeAppUploadStoreFile'))");
    expect(result).toContain("keyAlias findProperty('lifeAppUploadKeyAlias')");
  });

  it("signs release builds with the release key only when it is configured", () => {
    const result = addReleaseSigning(template);

    expect(result).toContain(
      "signingConfig findProperty('lifeAppUploadStoreFile') ? signingConfigs.release : signingConfigs.debug",
    );
  });

  it("leaves debug builds signed with the debug key", () => {
    const result = addReleaseSigning(template);

    expect(result).toMatch(/debug \{\s+signingConfig signingConfigs\.debug\s+\}/);
  });

  it("is idempotent", () => {
    const once = addReleaseSigning(template);

    expect(addReleaseSigning(once)).toBe(once);
  });

  it("throws on a template it doesn't recognise", () => {
    expect(() => addReleaseSigning("android { }")).toThrow(
      /doesn't look like the expected template/,
    );
  });
});
