# App architecture

The architectural vision for `app/`, agreed before the first feature was built. It is the baseline to build on, not a frozen contract: when a slice shows a decision is wrong, change it here in the same PR. Behavior a user can observe lives in the specs ([`specs/app-shell.md`](../../specs/app-shell.md) for what crosses features). This doc covers how the app is built.

Most of it was set up by the **App baseline** slice, now done; what remains is on the [roadmap](../../specs/README.md#roadmap). Library details (APIs, config keys) are re-checked against the Expo SDK 58 docs when each piece is built.

## 1. Design language

- **Material 3, one design for both platforms.** The app is Android-first; both current users are on Android. M3 is Android's native look, and on iOS a Material app is acceptable for now. One design means no per-platform components.
- **Native navigation chrome.** Expo Router's native stack (and native tabs, once there are tabs) renders the platform's real header, transitions and back gesture. Rebuilding these in JS is where apps start to feel off.
- **React Native Paper** for components and theming. It implements M3 in JS, so it looks the same on both platforms, and its theme is a set of M3 color tokens. Adopted at 5.15 in October 2026, when it had regular releases. 6.0 was in alpha then: an M3 overhaul that removes deprecated APIs, so moving to it is a planned upgrade of its own.
  - **Icons** are Material Design Icons from `@react-native-vector-icons/material-design-icons`, which Paper looks for first. Expo's icon guide marks `@expo/vector-icons` as headed for deprecation.
  - **The navigation theme is mapped from Paper's** in `ui/theme.ts`, because the native header is themed through Expo Router's `ThemeProvider`, not Paper's. Paper's `adaptNavigationTheme` isn't used: its types expect string colors, while React Navigation 7's are `ColorValue`.
  - Alternative: `@expo/ui`, whose universal components render Jetpack Compose on Android and SwiftUI on iOS. They are truly native, but the set is small (19 components), each tree needs a `<Host>`, and there's no documented theming API. Use it only where a native widget clearly beats Paper's, such as a date picker.
- **Layout** is plain `StyleSheet` and flexbox. Paper covers widgets, not layout.
- **Colors:** an M3 "tonal spot" scheme generated from a muted steel-blue seed (`#4A6A8F`) with `@material/material-color-utilities`, the library behind Material Theme Builder, with light and dark variants. The generated tokens are committed as static values in `ui/theme.ts`, and the system setting picks light or dark. No Android dynamic color: the app keeps its own scheme. There's no brand yet, so the seed is the identity for now.
  - **Extended colors.** M3 has no green, so a `success` set (`success`, `onSuccess`, `successContainer`, `onSuccessContainer`) is generated the way Material Theme Builder does for a custom color: the green `#2E8B3E` harmonized toward the seed, tones 40/100/90/10 in light and 80/20/30/90 in dark. It's for what is positive or on offer, such as an available update. Components read it with `useTheme<AppPaperTheme>()`.
- **Accessibility:** M3's default contrast, and the system font scale is respected. Icon-only buttons have an `accessibilityLabel`, touch targets are at least 48dp, and status is never conveyed by color alone (a checked item gets a checkmark, not just grey text).

## 2. App shell and navigation

- **Two data scopes, mirroring the API:** family (`/families/{familyID}/…`) and me (`/me`).
- **Active family.** One family at a time is active, held in a context and persisted on the device. Every family-scoped screen reads it from there, never from route params. With a single family it simply holds that family's id; a Google-style switcher behind the avatar can set it later without restructuring anything.
- **Top app bar with an avatar** that opens profile, settings and sign-out (and later the family switch). Profile is not a tab.
- **Home is a dashboard of cards.** Each feature provides its own card (e.g. `features/shopping/dashboard-card.tsx`), which fetches its own data; Home only arranges them. Cards can be family-scoped or personal.
- **Tabs are deferred** until there are at least three real destinations, the minimum for an M3 navigation bar. Until then, features are pushed onto the stack from their Home card. Because route files are thin (see [Code organization](#6-code-organization)), moving a feature into a tab, or under a broader domain tab, means moving route files, not refactoring the feature.
- **Deep links** use `lifeapp://` (e.g. `lifeapp://join?token=…`). Https App Links need a verified domain and are deferred; switching is cheap because invites expire after 24 hours, so no old link survives the change.

## 3. Data layer

- **TanStack Query for server state.** Almost all data is server data owned by a family or the user. TanStack Query provides the caching, refetching, optimistic updates with rollback, mutation retries and cache persistence the app needs.
  - Alternatives: SWR (no built-in mutation retry or persistence), RTK Query (brings Redux, for very little client state), and plain `fetch` hooks (we'd rebuild all of the above).
- **Client state is small** and lives in React context: the session (token in `expo-secure-store`) and the active family. No Redux or Zustand.
- **The app requires a connection.** The query cache is persisted to the device (AsyncStorage, kept for 7 days, with `gcTime` as long so restored queries survive), so the app starts instantly with the last-known data and shows it read-only when offline. It's cleared on sign-out. Offline writes are not supported.
- **Retries:** queries and mutations retry twice with backoff, but only failures that can change on a second try: a failed connection, or a 5xx. A 4xx or a schema mismatch fails at once, and so does a timeout, since each retry would wait the full timeout again. TanStack Query doesn't retry mutations by default, so this is set for them too.
- **Every mutation is optimistic:** the cache updates at once and the request runs in the background. On final failure the change rolls back and a snackbar offers **Retry**.
- **Freshness:** a 30-second `staleTime`, so moving between screens doesn't refetch every time. Stale queries refetch when the app returns to the foreground (`AppState`, wired into TanStack Query's `focusManager`) and when a screen regains focus. React Navigation keeps the screens below the top of the stack mounted, so feature hooks use `useApiQuery` (`src/api/`, enforced by lint) instead of `useQuery`: it passes `useIsFocused()` as TanStack Query's `subscribed` option, so a screen out of focus stops counting as active and regaining focus refetches its stale queries. Lists support pull-to-refresh (`RefreshControl`). No polling, and no real-time push.
- **Rules for new write endpoints**, so that retries and optimistic updates are safe:
  - **Client-generated UUIDs** for things the app creates, so an item can be edited (or its creation undone) before the server has answered.
  - **Idempotent writes:** "set `checked` to true", never "toggle". A replayed request gives the same result.
  - **Last write wins** on conflicts.

## 4. API contract

- **`src/api/`** holds a typed `fetch` wrapper and the TanStack Query setup. Each feature declares its response shapes as **hand-written zod schemas**, and the TS types are inferred from them. Responses are parsed at the boundary, so a mismatch with the API fails loudly in one place instead of as `undefined` deep in a screen.
  - Alternatives: plain TS interfaces (drift stays invisible), and an OpenAPI spec with generated types (a third artifact and toolchain to keep in sync; worth revisiting if the API grows much larger or gets a second client).
- **Errors.** The API's envelope is `{"error": {"code", "message", "fields"?}}`. The client turns it into one `ApiError` (status, `code`, `fields`); an error status without the envelope (e.g. a proxy's error page) gets the code `unknown`. A request that doesn't reach the API is a separate `NetworkError` ("Couldn't connect"): no response, a timeout after 15 seconds (React Native's `fetch` has no timeout), or a success that isn't JSON, such as a captive portal's login page.
- **The app owns user-facing copy**, keyed off `code` through i18n. The API's `message` is for developers and is never shown. Unknown codes get a generic message.
- **Any 401 on an authenticated request means signed out.** Sessions are opaque and slide over 90 days, with no refresh tokens (`docs/AUTH.md`), so there's nothing to refresh. The client handles it once, centrally: see [Session lifecycle](#5-session-lifecycle).

## 5. Session lifecycle

- **Cold start:** the splash screen stays up (`SplashScreen.preventAutoHideAsync`) while the app reads the token, restores the persisted cache and resolves the active family. Then it hides, and routing takes over. There is no separate loading screen.
- **Optimistic restore:** with a stored token, the app shows the app right away. It doesn't call `/me` first, which would make launch wait on the network and fail offline. If the session was revoked, the first request's 401 signs the user out.
- **Routing:** the root layout guards the signed-in routes with `Stack.Protected guard={isSignedIn}`. When the guard turns false, Expo Router removes those screens from history and falls back to the anchor route, so the sign-in screen is the unprotected group's `index`. SDK 58 added `redirectTo`, which picks the target explicitly; whether to use it is decided when sign-in is built.
- **`SessionProvider`** holds the token and a status: `loading`, `signedIn` or `signedOut`.
- **Resolving the active family** from `GET /families`:

  | Result | The app |
  | --- | --- |
  | No families | Shows first run: create a family, or join with an invite |
  | The persisted id is in the list | Uses it |
  | The persisted id is missing (e.g. the user was removed) | Uses the first family |

  A 404 or 403 on a family-scoped request means the user is no longer a member: the app refetches families, resolves again and shows a snackbar. This is handled centrally, like the 401.
- **Sign-out** calls `POST /auth/logout` best-effort (a failure or no connection doesn't block it), then runs the same local cleanup as a 401: delete the token, clear the query cache including the persisted copy, reset the active family. One cleanup function, two triggers.

## 6. Code organization

Feature-first, like the API's "packages by responsibility, not by layer":

```
src/
  app/              # routes only (Expo Router). Thin files that render a screen from features/
  features/
    session/        # SessionProvider, sign-in screen, sign-out cleanup
    families/       # active-family context, family screens, api.ts
    home/           # the dashboard, composing features' cards
    shopping/       # api.ts, screens, dashboard-card.tsx
  api/              # fetch wrapper, ApiError, QueryClient setup and persistence
  ui/               # Paper theme, shared components (Screen, EmptyState, ErrorState, snackbar host)
  i18n/             # i18next setup and the copy (en.json)
```

- **Route files only wire navigation:** they render a screen from `features/` plus route options such as the header title.
- **Each feature's `api.ts`** owns its zod schemas, a query-key factory (`shoppingKeys.list(familyId, listId)`) and named hooks (`useShoppingList`, `useCheckItem`). Screens use those hooks and never call `useQuery` with raw keys. Family-scoped keys start with the family id, so switching family can never show another family's cached data.
- **Dependencies point one way:** `features → api, ui`, and all three may use `i18n`, which imports none of them. `api` and `ui` don't import each other, except `ui/app-providers.tsx`, the composition root, which mounts the query provider; `api` takes callbacks instead (e.g. its refresh-error hook, which the root wires to the snackbar). Features don't import each other's internals, with three explicit exceptions: Home imports each feature's dashboard card, Home imports `app-update`'s banner, and family-scoped features read the active family from `families`.
- **No barrel `index.ts` files.** Import concrete files via `@/…`. Barrels hide circular imports, and Metro doesn't tree-shake.
- The template's `themed-*` components, `use-theme`, `constants/theme.ts` and `external-link` are replaced by the Paper theme and `ui/`.

## 7. Cross-cutting UX patterns

Shared components in `ui/`, used the same way on every screen:

- **Loading:** a centered M3 progress indicator only on a first load with nothing cached, which the persisted cache makes rare. Background refetches are silent. No skeleton screens.
- **Empty:** `EmptyState` (icon, title, a line of text, an optional primary action) renders the empty states the specs define.
- **Errors:**

  | Situation | Behavior |
  | --- | --- |
  | A query fails, nothing cached | `ErrorState` with Retry |
  | Offline, nothing cached | `ErrorState` saying so, without Retry (`isWaitingForConnection`): TanStack Query pauses the first fetch and runs it when back online |
  | A query fails, cached data shown | Keep the data, snackbar "Couldn't refresh. Showing cached data." (centrally, from the query cache; at most one per 10 seconds) |
  | A mutation fails after retries | Roll back, snackbar with Retry |
  | `400 invalid_body` from a form | Inline errors per field, from `fields` |
  | 401, or removed from the family | Handled centrally ([Session lifecycle](#5-session-lifecycle)) |

- **No success toasts:** with optimistic updates, the change itself is the feedback.
- **One global snackbar host with a queue:** a small provider plus `useSnackbar()`, because Paper's `Snackbar` is only a component.
- **Each screen handles its states explicitly** with these components (`if (isPending) … if (isError) …`). No generic `QueryBoundary` wrapper; a few more lines, but visible and easy to vary.
- **Destructive actions:** heavy or irreversible ones (delete a family, leave) always get a confirmation dialog. Whether a lighter action offers an Undo snackbar is decided per action, in its spec's UI section. With client-generated ids, undoing a delete re-creates the item with the same id.
- **Offline banner:** "Offline · showing cached data", at the top of every `Screen` while the query cache holds data (with nothing cached, the screen's own offline message says so instead). `expo-network` feeds TanStack Query's `onlineManager`, and the banner reads `onlineManager`, so it shows exactly while queries are paused. Online means `isConnected`, not `isInternetReachable`: on Android the latter needs Google's connectivity check to pass, which emulators and some networks fail, and being wrongly offline pauses every query while being wrongly online only fails them with "Couldn't connect". `expo-network` is Expo's own module, versioned with the SDK; `@react-native-community/netinfo` was the alternative, but it is barely maintained and isn't a New Architecture module.

## 8. Quality

- **Unit tests:** Jest with the `jest-expo` preset, for pure logic such as schemas, error-code mapping, active-family resolution and query keys.
- **Screen tests:** React Native Testing Library, rendering a screen with a real QueryClient and a mocked `src/api` client. They cover loading, empty and error states, and optimistic updates with rollback. Queries go by role and text, as with RTL on the web.
- Both are written within each slice, like the API's endpoint tests.
- **End-to-end tests are an option, not a decision.** If adopted, Maestro (YAML flows on a real emulator) is the candidate, being simpler than Detox. Google sign-in can't be automated, so it would need a way to start with a dev session (`make dev-session` in `api/`), which is a security design question of its own.
- **Language:** English only, but i18n-ready from day one: all copy goes through i18next (`react-i18next`), so a second language is translation work, not a refactor.
  - The copy is bundled (`i18n/en.json`), so i18next initializes synchronously and the first render has its text. `t()` keys are typed against the English copy, so a mistyped key fails typecheck.
  - The language is the first of the device's preferred languages that has copy (`expo-localization`), falling back to English. The native module is in from the start, so a second language won't need a new APK.
  - i18next's plural keys (`_one`, `_other`) need `Intl.PluralRules`, which Hermes doesn't document supporting. `intl-pluralrules` installs itself only where it's missing or incomplete, which it was on an SDK 58 dev build.

## 9. Build and release

- **A development build, not Expo Go.** Expo Go is a prebuilt shell from Expo with a fixed set of native modules, Expo's package id and Expo's signing key. Native Google sign-in needs our own native module and our package id and signing key, so the app runs as a development build: our own shell with `expo-dev-client`, still loading JS from Metro with fast refresh. It's rebuilt only when native dependencies change.
- **Continuous Native Generation:** `android/` and `ios/` are generated by `expo prebuild` from `app.config.ts` and config plugins, and gitignored. Native changes are made in config or a config plugin, never by hand. The build scripts regenerate the native project for their variant before building.
- **Variants:** `app.config.ts` selects the package id (`com.lakaniemi.lifeapp.dev` vs `com.lakaniemi.lifeapp`), name ("Life App Dev" vs "Life App"), URL scheme (`lifeapp-dev` vs `lifeapp`), icon and API URL from one `APP_VARIANT` variable, so dev and prod install side by side and don't claim each other's links. It defaults to `development`, so a production build is always explicit. Google's Android OAuth client is keyed by package name plus signing-certificate SHA-1, so each variant and signing key needs its own (`docs/AUTH.md`).
- **Local API:** the Android emulator reaches the Mac's `localhost` at `10.0.2.2`, the dev variant's default. A physical phone uses `adb reverse tcp:8080 tcp:8080` or the LAN IP, through an `API_URL` override read with the config, so in development changing it needs only a Metro restart.
- **Distribution: GitHub Releases.** Each native release is a GitHub release with the signed APK attached under a fixed file name. The phones install it directly (sideloading). Because the asset name is fixed, `https://github.com/lakaniemi/life-app/releases/latest/download/life-app.apk` always points to the newest APK; the "Update required" screen links there ([Version compatibility](#10-version-compatibility)).
  - **No paid Expo services.** EAS Build is a paid product beyond a small free tier, so APKs are built on GitHub Actions: `expo prebuild`, then `./gradlew assembleRelease`. The APK targets `arm64-v8a` only: every phone the app runs on is 64-bit ARM, and compiling the native code for all four ABIs took most of a 20+ minute build.
  - **Signed with our own key**, which keeps the SHA-1 for Google sign-in stable. A local config plugin (`plugins/with-release-signing.ts`) adds the release `signingConfig` to the generated `build.gradle`, reading the keystore, its password and the key alias from Gradle properties. It throws when the generated file isn't the shape it expects, so a template change in a later SDK fails the build instead of quietly signing with the debug key. The keystore is a GitHub secret, with a backup outside GitHub: Android installs an update only over an app signed with the same key, so losing it means reinstalling on every phone and registering a new SHA-1.
  - **Without those properties a release build is signed with the debug key**, so `npm run android:prod` works locally with no setup. Such an APK can't install over a real release without uninstalling it first. CI guards against publishing one: the release workflow compares the APK's signing certificate with the SHA-256 in the `ANDROID_RELEASE_CERT_SHA256` repository variable, and fails on a mismatch.
  - **The version lives in `version.json`**, which `app.config.ts` reads: the version name and `versionCode`, bumped by hand in a release PR. `versionCode` is the integer Android compares, so it must go up with every release; a hand-bumped file keeps it visible in review, and a check (`.github/scripts/check-app-version.sh`) fails the PR when the bump is inconsistent. A release isn't triggered by a button: merging a change to `version.json` into `main` runs `app-release.yml`, which builds and publishes `app-v<version>` once. Other merges release nothing.
  - The repo is public, so anyone can download the APK. That's acceptable: it holds no secrets (the Web client ID and API URL are public by nature), and everything is behind sign-in.
  - Alternative: the Google Play internal testing track, for installs and auto-updates through the Play Store (and Play's in-app update API). It needs a paid developer account, and Play App Signing re-signs the app, which adds another SHA-1 to register.
- **In-app updates, from the GitHub release.** Every change ships as a new APK, and the app updates itself from the releases above. There are no over-the-air updates, so a version number always identifies exactly what runs, and there's no update server to host.
  - **Finding a release:** the release workflow also attaches `latest.json` (`{ "version", "versionCode" }`, copied from `version.json`) once the APK is published. `releases/latest/download/latest.json` is a stable URL like the APK's, so the check is one plain `fetch` with no GitHub API and no rate limit. Because the same job publishes both files, the file never announces an APK that doesn't exist. The app compares `versionCode` with its own (`expo-application`'s `nativeBuildVersion`), parsing the file with a zod schema like any response. It checks at launch and when returning to the foreground, throttled, because Android keeps the process alive for days.
  - **The check is a TanStack Query query** (`features/app-update/latest-release.ts`) with a plain `fetch` rather than `apiFetch`, which adds our API's URL and headers. It's remote state with the same needs as the rest, and `managers.ts` already feeds it the foreground and reconnect events. Every failure (404, no connection, a timeout, a malformed body) resolves to "no release", so nothing is shown and nothing retries. The body is parsed as text, because GitHub serves release assets as `application/octet-stream`.
  - **The throttle is `staleTime`**: one hour. Coming back to the foreground refetches only a stale query, so no code of its own is needed. The query has `meta: { persist: false }`, which `shouldPersistQuery` honors: a restored answer would show the banner at launch before anything is checked, perhaps offline, where Install can't work. For the same reason the offline banner doesn't count it as cached data.
  - **The banner** (`UpdateBanner`) is rendered by Home above its content. "Later" is component state, so it lasts until the next launch and isn't persisted.
  - **Installing:** the app downloads the APK with `expo-file-system`, turns the file into a `content://` URI (Android doesn't let one app hand another a raw file path, the installer included) and starts the install intent with `expo-intent-launcher`. This needs the `REQUEST_INSTALL_PACKAGES` permission, added in `app.config.ts`, and the user's one-time "Allow installs from this app" setting. Android always shows its own confirmation, so the app never installs by itself. The "Update required" screen uses the same installer and keeps the APK link as a fallback.
  - **Only the production variant checks.** The dev variant is installed from Metro and has no release to compare with. `app.config.ts` puts the URL to check in `extra.latestReleaseUrl`: production's is the release's `latest.json`, and the dev variant leaves it out, which turns the check off (not `null`: Expo's config resolver turns that into `{}`). Setting `LATEST_RELEASE_URL` (dev only, like `API_URL`) points a dev build at a file of your own, to try the banner. Trying the installer needs a release build, and a debug-signed local build can't install over a release-signed one.
  - **The check isn't in our API.** An endpoint would tie the API's deploys to the app's releases, or make the API call GitHub, which breaks the "coupled to nothing but `DATABASE_URL`" hedge. Keeping the two checks apart also keeps their meanings apart: `app_outdated` means the API refuses this app, and `latest.json` means a newer app exists. Revisit only if we want rollout control, such as offering a version to one user first.
  - **Rejected: over-the-air updates with `expo-updates`**, against a self-hosted server implementing the [Expo Updates protocol](https://docs.expo.dev/technical-specs/expo-updates-1/). They would need a server to choose, host and sign for, and a `fingerprint` runtime version to keep JS off native shells it doesn't match. They would also make the same version number mean different code on different phones, and add a second way out of "Update required". The saving is a smaller download, which doesn't matter for two users on Wi-Fi. Revisit if the app gets many users, or a native rebuild per change becomes painful.
- **CI:** `expo install --check` (versions match the SDK), lint, `tsc --noEmit` and Jest in GitHub Actions (`.github/workflows/app.yml`), like the API's checks.

## 10. Version compatibility

The API is deployed once and changes for everyone at the same moment, while each phone runs whatever APK it last installed. Three boundaries can disagree:

| Boundary | Covered by |
| --- | --- |
| App vs API | Contract rules, a tolerant reader, and a minimum compatibility level (below) |
| Device cache vs app | ISO-string timestamps and a cache buster (below) |
| Database vs the API version still running during a deploy | Expand-then-contract migrations (`api/CLAUDE.md`, Database) |

These are the established practices, scaled to one first-party app with two users and an in-app update prompt. Sources: Google's [AIP-180](https://google.aip.dev/180) for the contract rules, Fowler's [parallel change](https://martinfowler.com/bliki/ParallelChange.html) for breaking changes.

### Most changes are never breaking

The API follows AIP-180's rules, listed in `api/CLAUDE.md` ("Compatibility"):
- Endpoints, response fields and enum values may be added.
- Nothing is removed, renamed, retyped or reformatted, and no default changes.
- **No new required request fields.** A new request field is allowed only if leaving it out means "behave as before". Inventing a default to make a field optional is not allowed. A field without that natural meaning is a breaking change, handled as described under "Breaking changes".

The app's half is the **tolerant reader**:
- zod objects drop unknown keys by default, so new response fields are harmless.
- **Enums are the trap:** a new value (say, a role `viewer`) would reject the whole response. So enums from the API parse unknown values to a fallback such as `'unknown'`, and the UI renders something neutral.
- Unknown error codes get a generic message.
- Updates send only the changed fields (`PATCH`), so an old app can't wipe a field it doesn't know about.

### Breaking changes

When a change can't be additive, there are two options, in order of preference:

1. **A new endpoint** next to the old one, when the new shape is naturally a separate operation. The old endpoint is removed once no supported app uses it (expand, migrate, contract).
2. **Raise the minimum compatibility level** in the same API deploy that makes the breaking change. Old apps are turned away with "Update required" instead of being served with guessed defaults. With two users and an in-app update prompt, that costs a minute's inconvenience, and the API stays clean.

There is no URL or date-based versioning (`/v1`, Stripe-style version transforms). Those keep every old version served in parallel, which is worth it for many external clients that can't be made to upgrade, and isn't for us.

### Compatibility level

- **The app declares an integer API compatibility level**: one constant in `src/api/`, starting at 1. It's bumped in the app update that starts depending on a breaking API change. An integer, rather than the version name or `versionCode`, because the level should change only when the app starts depending on a breaking API change, not with every release.
- **The app sends it on every request** in an `App-Compat-Level` header. Custom headers have no `X-` prefix (RFC 6648).
- **The API keeps its minimum as a constant in code**, so raising it ships atomically with the breaking change it protects. A middleware (`api/internal/server/compat.go`) answers requests below the minimum with 400 and the error code `app_outdated`, and a header that isn't an integer with 400 `invalid_compat_level`. Requests without the header (curl, tests) aren't gated: only the app sends it, and every build does. The API logs the level, which shows which levels are still in use before an old endpoint is removed.
- **The app handles `app_outdated` from its first build**, because an old app can only react to the code if the handling was already in it. It shows the blocking "Update required" screen ([app-shell.md](../../specs/app-shell.md)):
  1. The fix is always a new APK. The screen offers **Install**, which downloads the latest release and opens Android's installer ([Build and release](#9-build-and-release)). This comes with the App updates slice; until then, the screen only links to the APK.
  2. If the download fails, the screen falls back to a link to the latest GitHub release.

### Device cache

- Timestamps stay **ISO strings** in schemas and are formatted in the UI. The persisted cache round-trips through JSON, so a parsed `Date` would come back as a string while its type still says `Date`.
- The persister's `buster` identifies the running app version, so a new version starts with a fresh cache rather than reading data cached under older types. That's the APK's version name and `versionCode` (`expo-application`), so each release build needs a new `versionCode`.

### Not adopted, for now

- **Feature flags or remote config**, to ship dormant code and switch it on server-side. There are no staged rollouts with two users.
- **Automated breaking-change detection**: diffing OpenAPI specs in CI, or consumer-driven contract tests (Pact). There's no spec to diff yet; revisit together with OpenAPI.
- **Google Play's in-app update API.** It delivers updates through Google Play, so it doesn't apply to a sideloaded APK. The app's own installer ([Build and release](#9-build-and-release)) does that job instead.

## Deferred

Decided against for now, or left for when a slice needs it:

- **Tabs**, until there are three destinations.
- **The family switcher**, until multi-family is built.
- **Offline writes** (a queue of mutations replayed later). The write rules in [Data layer](#3-data-layer) keep this possible.
- **Real-time updates** (server-sent events or WebSockets).
- **End-to-end tests** (Maestro).
- **Google Play** distribution.
- **Crash reporting**, until someone outside the household uses the app. Check free-tier limits of Sentry or similar then.
- **OpenAPI** and generated types.
- **Https App Links** and any web fallback for links.
