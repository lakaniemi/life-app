# Phase 6 — App: families (PR 6)

*Outline only. Detail it when starting the phase.*

- **First run:** a signed-in user with no family chooses "Create a family" or "Join with an invite".
- **Single family in the UI for now.** The app uses the user's first family. The API already supports several, so a family switcher can come later without backend changes.
- **Family screen:** name, members and roles. Admins can rename the family, change roles and remove members.
- **Leave / delete:** "Leave" becomes "Delete family" when you're the sole member. If you're the sole admin with other members, the app asks you to promote someone first. The API enforces both rules; the UI reflects them.
- **Invite:** a single "Invite a member" button for admins. It creates an invite and shows it as a QR code (e.g. `react-native-qrcode-svg`) with a share button for the link. Closing the screen just leaves the invite to expire.
- **Join:** scan with `expo-camera`'s barcode scanning, parse the `lifeapp://join?token=…` deep link, preview the family name and confirm. The deep link also works when opened from another app. Some chat apps don't make `lifeapp://` links tappable, so "Join with an invite" also accepts a pasted link.
- **Join preview states:**
  - Already a member (`alreadyMember`): "You're already in X" with "Open family" instead of "Join".
  - Invalid or expired invite (404): "This invite has expired or isn't valid. Ask for a new one."
  - **Open question:** what happens when the user already belongs to another family? The API allows it, but the single-family UI would show only the first one. Options include warning before joining, or switching to the new family.
