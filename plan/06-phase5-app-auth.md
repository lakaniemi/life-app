# Phase 5 — App: sign-in (PR 5)

*Outline only. Detail it, and re-check library recommendations against current Expo docs, when starting the phase.*

- **Google sign-in:** `@react-native-google-signin/google-signin`. It needs native code, so it requires an Expo **development build**; it won't work in Expo Go. Configure it with `webClientId` = the Web client ID from phase 2, which makes the ID token's `aud` match what the API expects.
- **Google Cloud:** create iOS and Android OAuth client IDs. Android needs the signing certificate SHA-1.
- **Token storage:** the API session token goes in `expo-secure-store` (Keychain/Keystore), not AsyncStorage.
- **API client:** a small typed `fetch` wrapper that adds the Bearer header and treats 401 as signed out.
- **Auth gate:** expo-router layout that shows the sign-in screen when signed out and the app when signed in.
- **Sign out:** call `POST /auth/logout`, then clear the secure store and the Google session.
