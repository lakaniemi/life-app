# App architecture

The architectural vision for `app/`, agreed before the first feature was built. It is the baseline to build on, not a frozen contract: when a slice shows a decision is wrong, change it here in the same PR. Behavior a user can observe lives in the specs ([`specs/app-shell.md`](../../specs/app-shell.md) for what crosses features). This doc covers how the app is built.

Most of it is set up by the **App baseline** slice in the [roadmap](../../specs/README.md#roadmap). Library details (APIs, config keys) are re-checked against the Expo SDK 57 docs when each piece is built.

## 1. Design language

- **Material 3, one design for both platforms.** The app is Android-first; both current users are on Android. M3 is Android's native look, and on iOS a Material app is acceptable for now. One design means no per-platform components.
- **Native navigation chrome.** Expo Router's native stack (and native tabs, once there are tabs) renders the platform's real header, transitions and back gesture. Rebuilding these in JS is where apps start to feel off.
- **React Native Paper** for components and theming. It implements M3 in JS, so it looks the same on both platforms, and its theme is a set of M3 color tokens. Check that it is still actively maintained (GitHub releases) before adopting it.
  - Alternative: `@expo/ui`, whose universal components render Jetpack Compose on Android and SwiftUI on iOS. They are truly native, but the set is small (19 components), each tree needs a `<Host>`, and there's no documented theming API. Use it only where a native widget clearly beats Paper's, such as a date picker.
- **Layout** is plain `StyleSheet` and flexbox. Paper covers widgets, not layout.
- **Colors:** an M3 scheme generated from a muted steel-blue seed (around `#4A6A8F`), previewed in Material Theme Builder, with light and dark variants. No Android dynamic color: the app keeps its own scheme. There's no brand yet, so the seed is the identity for now.
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
- **The app requires a connection.** The query cache is persisted to the device, so the app starts instantly with the last-known data and shows it read-only when offline. It's cleared on sign-out. Offline writes are not supported.
- **Every mutation is optimistic:** the cache updates at once and the request runs in the background. Mutations retry once or twice with backoff (TanStack Query doesn't retry mutations by default). On final failure the change rolls back and a snackbar offers **Retry**.
- **Freshness:** refetch when a screen gains focus (`useFocusEffect`) and when the app returns to the foreground (`AppState`, wired into TanStack Query's `focusManager`), with a short `staleTime` so moving between screens doesn't refetch every time. Lists support pull-to-refresh (`RefreshControl`). No polling, and no real-time push.
- **Rules for new write endpoints**, so that retries and optimistic updates are safe:
  - **Client-generated UUIDs** for things the app creates, so an item can be edited (or its creation undone) before the server has answered.
  - **Idempotent writes:** "set `checked` to true", never "toggle". A replayed request gives the same result.
  - **Last write wins** on conflicts.

## 4. API contract

- **`src/api/`** holds a typed `fetch` wrapper and the TanStack Query setup. Each feature declares its response shapes as **hand-written zod schemas**, and the TS types are inferred from them. Responses are parsed at the boundary, so a mismatch with the API fails loudly in one place instead of as `undefined` deep in a screen.
  - Alternatives: plain TS interfaces (drift stays invisible), and an OpenAPI spec with generated types (a third artifact and toolchain to keep in sync; worth revisiting if the API grows much larger or gets a second client).
- **Errors.** The API's envelope is `{"error": {"code", "message", "fields"?}}`. The client turns it into one `ApiError` (status, `code`, `fields`). A request that gets no response at all is a separate network error ("Couldn't connect").
- **The app owns user-facing copy**, keyed off `code` through i18n. The API's `message` is for developers and is never shown. Unknown codes get a generic message.
- **Any 401 on an authenticated request means signed out.** Sessions are opaque and slide over 90 days, with no refresh tokens (`docs/AUTH.md`), so there's nothing to refresh. The client handles it once, centrally: see [Session lifecycle](#5-session-lifecycle).

## 5. Session lifecycle

- **Cold start:** the splash screen stays up (`SplashScreen.preventAutoHideAsync`) while the app reads the token, restores the persisted cache and resolves the active family. Then it hides, and routing takes over. There is no separate loading screen.
- **Optimistic restore:** with a stored token, the app shows the app right away. It doesn't call `/me` first, which would make launch wait on the network and fail offline. If the session was revoked, the first request's 401 signs the user out.
- **Routing:** the root layout guards the signed-in routes with `Stack.Protected guard={isSignedIn}`. When the guard turns false, Expo Router removes those screens from history and falls back to the anchor route, so the sign-in screen is the unprotected group's `index`. (`redirectTo` would pick the target explicitly, but it needs SDK 58.)
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
```

- **Route files only wire navigation:** they render a screen from `features/` plus route options such as the header title.
- **Each feature's `api.ts`** owns its zod schemas, a query-key factory (`shoppingKeys.list(familyId, listId)`) and named hooks (`useShoppingList`, `useCheckItem`). Screens use those hooks and never call `useQuery` with raw keys. Family-scoped keys start with the family id, so switching family can never show another family's cached data.
- **Dependencies point one way:** `features → api, ui`. Features don't import each other's internals, with two explicit exceptions: Home imports each feature's dashboard card, and family-scoped features read the active family from `families`.
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
  | A query fails, cached data shown | Keep the data, snackbar "Couldn't refresh" |
  | A mutation fails after retries | Roll back, snackbar with Retry |
  | `400 invalid_body` from a form | Inline errors per field, from `fields` |
  | 401, or removed from the family | Handled centrally ([Session lifecycle](#5-session-lifecycle)) |

- **No success toasts:** with optimistic updates, the change itself is the feedback.
- **One global snackbar host with a queue:** a small provider plus `useSnackbar()`, because Paper's `Snackbar` is only a component.
- **Each screen handles its states explicitly** with these components (`if (isPending) … if (isError) …`). No generic `QueryBoundary` wrapper; a few more lines, but visible and easy to vary.
- **Destructive actions:** heavy or irreversible ones (delete a family, leave) always get a confirmation dialog. Whether a lighter action offers an Undo snackbar is decided per action, in its spec's UI section. With client-generated ids, undoing a delete re-creates the item with the same id.
- **Offline banner:** "Offline · showing saved data", driven by `@react-native-community/netinfo`, which also feeds TanStack Query's `onlineManager`.

## 8. Quality

- **Unit tests:** Jest with the `jest-expo` preset, for pure logic such as schemas, error-code mapping, active-family resolution and query keys.
- **Screen tests:** React Native Testing Library, rendering a screen with a real QueryClient and a mocked `src/api` client. They cover loading, empty and error states, and optimistic updates with rollback. Queries go by role and text, as with RTL on the web.
- Both are written within each slice, like the API's endpoint tests.
- **End-to-end tests are an option, not a decision.** If adopted, Maestro (YAML flows on a real emulator) is the candidate, being simpler than Detox. Google sign-in can't be automated, so it would need a way to start with a dev session (`make dev-session` in `api/`), which is a security design question of its own.
- **Language:** English only, but i18n-ready from day one: all copy goes through an i18n layer (e.g. i18next with `expo-localization`), so a second language is translation work, not a refactor.

## 9. Build and release

- **A development build, not Expo Go.** Expo Go is a prebuilt shell from Expo with a fixed set of native modules, Expo's package id and Expo's signing key. Native Google sign-in needs our own native module and our package id and signing key, so the app runs as a development build: our own shell with `expo-dev-client`, still loading JS from Metro with fast refresh. It's rebuilt only when native dependencies change.
- **Variants:** `app.config.ts` selects the package id (`…lifeapp.dev` vs `…lifeapp`), name, icon and API URL from one `APP_VARIANT` variable, so dev and prod install side by side. Google's Android OAuth client is keyed by package name plus signing-certificate SHA-1, so each variant and signing key needs its own (`docs/AUTH.md`).
- **Local API:** the Android emulator reaches the Mac's `localhost` at `10.0.2.2`. A physical phone uses `adb reverse tcp:8080 tcp:8080` or the LAN IP.
- **Distribution:** a release APK installed directly (sideloaded), built with EAS Build or locally. We hold the signing key, which keeps the SHA-1 for Google sign-in stable. The Google Play internal testing track is the alternative: installs and updates through the Play Store, but a paid developer account, and Play App Signing re-signs the app, which adds another SHA-1 to register.
- **Over-the-air updates with EAS Update.** JS-only changes reach installed apps without a new APK. An update is a JS bundle plus assets, published to a branch that the build's channel points to. The app checks at launch, downloads only the files it doesn't have, and runs the new bundle from the next cold start; if loading it fails, it falls back to the bundle embedded in the APK.
  - **`fingerprint` runtime version policy:** a hash of everything that affects the native layer. An update reaches only builds with the same runtime version, so JS never runs on a native shell it doesn't match; a native change means a new APK.
  - **Restart prompt:** besides the check at launch, the app checks again when it returns to the foreground (throttled), because Android keeps the process alive for days. When an update has downloaded (`useUpdates()`), a snackbar "Update ready · Restart" reloads the app (`Updates.reloadAsync()`). The app never restarts without asking, since that could interrupt an edit.
  - The Expo Updates protocol is an open spec and `expo-updates` works with a self-hosted server, so this isn't lock-in.
- **Free plans, checked 2026-10:** EAS Build 15 Android builds a month (low-priority queue); EAS Update 1,000 monthly active users; EAS Workflows 60 CI minutes a month.
- **CI:** lint, `tsc --noEmit` and Jest in GitHub Actions, like the API's checks. Builds and updates are published by hand.

## 10. Version compatibility

The API is deployed once and changes for everyone at the same moment, while each phone runs whatever version it last loaded. Three boundaries can disagree:

- **Native shell vs JS:** covered by the `fingerprint` runtime version.
- **App vs API:** the API changes only in backward-compatible ways, by expand then contract. The rules are in `api/CLAUDE.md` ("Compatibility"). The app's half:
  - **Read responses tolerantly.** zod objects drop unknown keys by default, so new fields are harmless. Enums are the trap: a new value (say, a role `viewer`) would reject the whole response, so enums from the API parse unknown values to a fallback such as `'unknown'`, and the UI renders something neutral.
  - **Send only changed fields** in updates (`PATCH`), so an old app can't wipe a field it doesn't know about.
  - Unknown error codes get a generic message.
- **Device cache vs app:**
  - Timestamps stay **ISO strings** in schemas and are formatted in the UI. The persisted cache round-trips through JSON, so a parsed `Date` would come back as a string while its type still says `Date`.
  - The persister's `buster` is the running update's id, so a new app version starts with a fresh cache rather than reading data cached under older types.
- **Escape hatch for an unavoidable breaking change:** the app sends its version in a header on every request, and shows a blocking "Update required" screen when the API answers with the code `app_outdated`. The app side ships in the first build, because an old app can only react to that code if the handling was already in it. The API-side check is added only when a breaking change needs it.
- **No URL versioning** (`/v1`). With one first-party client and over-the-air updates, expand-then-contract gives the same guarantee per field. Versioning would mean serving every old version in parallel, and retiring one still needs the gate above. If it's ever needed, the version header lets the API branch per endpoint.

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
