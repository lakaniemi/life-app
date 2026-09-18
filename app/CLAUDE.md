# CLAUDE.md (app/)

- **Expo SDK 57.** Its API surface changes fast between versions; check the versioned docs (https://docs.expo.dev/versions/v57.0.0/) rather than relying on training-data knowledge before writing Expo/RN code.
- **Layout:** routes/screens live under `src/app` (Expo Router, file-based routing). Everything else — components, hooks, constants — lives under `src/` and is imported via the `@/*` path alias (e.g. `@/components/themed-text`), not relative paths.
