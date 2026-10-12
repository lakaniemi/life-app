# Sign-in

Status: in progress (API built, app not)

## Why

Every family's data is private to its members, so the app needs to know who the user is. Google sign-in avoids passwords, and users expect to stay signed in on their phone.

## Behavior

- A signed-out user sees only the sign-in screen, with "Sign in with Google".
- After signing in, the user lands in the app. Without a family, that's the families first-run screen (see [families.md](families.md#ui)).
- The user stays signed in across app restarts. A session lasts 90 days and each use extends it, so only a phone left unused that long signs out by itself.
- Signing out returns to the sign-in screen.
- The user's name comes from their Google account at first sign-in. After that it's theirs to edit, and later sign-ins never overwrite it.

## Rules

- Only the user's identity is requested from Google (`openid profile`). No email is requested or stored.
- If the API rejects the session (401), the app treats the user as signed out.

## Acceptance criteria

- A signed-out user can't reach any screen other than sign-in.
- A successful sign-in shows the app. Reopening the app later skips the sign-in screen.
- Cancelling sign-in returns to the sign-in screen silently. A failure stays there and shows a message.
- A 401 from any request clears the stored session and shows the sign-in screen.
- Signing out ends the session in the API, clears the stored session and the Google session, and shows the sign-in screen.

## UI

- **Sign-in screen:** the app's name and "Sign in with Google". While signing in, the button shows progress and can't be pressed again.
- **Sign out:** in the avatar menu (see [app-shell.md](app-shell.md)).

## Non-goals

- Other sign-in providers. Sign in with Apple is in [ideas.md](ideas.md).
- Signing out everywhere, or listing sessions.
- Account deletion.
- Sign-in on iOS, for now. The app is Android-first ([`app/docs/ARCHITECTURE.md`](../app/docs/ARCHITECTURE.md)), and `docs/AUTH.md` covers Android (Credential Manager).

## Open questions

Resolve these when planning the slice:

- **Nonce library on Android.** The API requires a nonce. `@react-native-google-signin` supports one on Android only in its paid version. The alternatives are another Credential Manager library, or a small Expo module of our own. See `docs/AUTH.md`, "App side".
- **Editing the name:** `PATCH /me` exists. Where the app offers it is open.

## Implementation

- API: `POST /auth/nonce`, `POST /auth/google`, `POST /auth/logout`, `GET /me`, `PATCH /me`. The flow, decisions and sources are in `docs/AUTH.md`.
- Known from planning, to be re-checked then:
  - Native Google sign-in needs an Expo development build; it doesn't work in Expo Go.
  - Sign-in is configured with the Web client ID (`GOOGLE_CLIENT_IDS`), which makes it the ID token's audience.
  - An Android OAuth client needs the app's package name and signing certificate SHA-1.
  - The session token goes in `expo-secure-store`, never AsyncStorage.
- App: not started.
