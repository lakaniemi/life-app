# Phase 5 — App: sign-in (PR 5)

*Outline only. Detail it, and re-check library recommendations against current Expo docs, when starting the phase.*

- **Google sign-in:** it needs native code, so it requires an Expo **development build**; it won't work in Expo Go. Configure it with the Web client ID from phase 2 (`webClientId` / `serverClientId`), which makes the ID token's `aud` match what the API expects.
- **Nonce (required by the API):** before sign-in, call `POST /auth/nonce` and pass the result to Google sign-in. **Library decision needed:** in `@react-native-google-signin`, an Android nonce needs the paid "Universal sign-in" version; the free API supports a nonce only on iOS. The options are the paid license, another Credential Manager library, or our own small Expo module. See `docs/AUTH.md`, "App side".
- **Google Cloud:** create iOS and Android OAuth client IDs. Android needs the signing certificate SHA-1.
- **Token storage:** the API session token goes in `expo-secure-store` (Keychain/Keystore), not AsyncStorage.
- **API client:** a small typed `fetch` wrapper that adds the Bearer header and treats 401 as signed out.
- **Auth gate:** expo-router layout that shows the sign-in screen when signed out and the app when signed in.
- **Sign out:** call `POST /auth/logout`, then clear the secure store and the Google session.
