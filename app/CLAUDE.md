# CLAUDE.md (app/)

- **Expo SDK 57.** Its API surface changes fast between versions; check the versioned docs (https://docs.expo.dev/versions/v57.0.0/) rather than relying on training-data knowledge before writing Expo/RN code.
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
- **Copy:** all user-facing text goes through i18n. Errors are shown by the API's error `code`, never its `message`.
- **Runs as a development build**, not in Expo Go, because native Google sign-in needs our own native module and package id.
