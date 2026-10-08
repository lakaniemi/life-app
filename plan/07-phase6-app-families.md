# Phase 6 — App: families (PR 6)

*Outline only. Detail it when starting the phase.*

- **First run:** a signed-in user with no family chooses "Create a family" or "Join with a code".
- **Single family in the UI for now.** The app uses the user's first family. The API already supports several, so a family switcher can come later without backend changes.
- **Family screen:** name, members and roles. Admins can rename the family, change roles and remove members.
- **Leave / delete:** "Leave" becomes "Delete family" when you're the sole member. If you're the sole admin with other members, the app asks you to promote someone first. The API enforces both rules; the UI reflects them.
- **Invite:** an admin creates an invite and the app shows it as a QR code (e.g. `react-native-qrcode-svg`), with the plain code as a typeable fallback.
- **Join:** scan with `expo-camera`'s barcode scanning, parse the `lifeapp://join?code=…` deep link, preview the family name and confirm. The deep link also works when opened from another app.
