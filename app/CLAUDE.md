# CLAUDE.md (app/)

- **Expo SDK 58.** Its API surface changes fast between versions; check the versioned docs (https://docs.expo.dev/versions/v58.0.0/) rather than relying on training-data knowledge before writing Expo/RN code.
- **Dependencies:** add Expo and React Native packages with `npx expo install`, which picks the SDK's version. Otherwise use the newest version, and pin an older one only for compatibility, with the reason in the commit. Packages deliberately off the SDK's versions are listed in `expo.install.exclude` in `package.json`.
- **`android/` and `ios/` are generated** by `expo prebuild` from `app.config.ts` and config plugins (Continuous Native Generation), and are gitignored. Never edit them by hand: a native change goes through `app.config.ts` or a config plugin. `APP_VARIANT` (`development`, the default, or `production`) selects the variant.
- **Architecture:** `docs/ARCHITECTURE.md` holds the agreed design and the reasons for it. Build on it. When a slice changes a decision, update that doc in the same PR.
- **Layout:** feature-first.
  - `src/app` holds routes only (Expo Router, file-based routing). Route files stay thin: they render a screen from `src/features/<feature>/` and set route options.
  - `src/api` holds the fetch client and the TanStack Query setup. `src/ui` holds the theme and shared components.
  - Features depend on `api` and `ui`, not on each other's internals. The two exceptions are in the architecture doc.
  - Imports use the `@/*` path alias (e.g. `@/ui/screen`), not relative paths. No barrel `index.ts` files.
- **UI:** Material 3 through React Native Paper, using the theme's color tokens, never hard-coded colors. Layout uses `StyleSheet` and flexbox. Every screen handles its loading, empty and error states explicitly, with the shared components in `src/ui`.
- **Server data:** TanStack Query only.
  - Each feature's `api.ts` declares zod schemas for responses (types are inferred from them), a query-key factory and named hooks. Screens use those hooks, never raw `useQuery` keys.
  - Mutations are optimistic, roll back on failure and show a snackbar with Retry.
  - Timestamps stay ISO strings in schemas. Enums from the API get an unknown-value fallback.
  - The API compatibility level (a constant in `src/api/`, sent as `App-Compat-Level`) is bumped in the update that starts depending on a breaking API change.
- **Copy:** all user-facing text goes through i18n. Errors are shown by the API's error `code`, never its `message`.
- **Runs as a development build**, not in Expo Go, because native Google sign-in needs our own native module and package id.
- **Tests** are Jest (`jest-expo` preset) and React Native Testing Library, next to the code they test (`home-screen.test.tsx`), never under `src/app/`, where Expo Router would treat them as routes.
