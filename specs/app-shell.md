# App shell

Status: specified

## Why

Features share one frame: how the app is navigated, where personal settings live, and how loading, failures, being offline and app updates look. Specifying it once keeps every feature consistent, and lets each feature spec describe only what's specific to it.

## Behavior

- **Top app bar** on the main screens, with the user's avatar at the trailing end. Tapping it opens a menu with profile, settings and sign-out. When multi-family support comes, switching family goes there too, as in Google's account switcher.
- **Active family.** The app works in one family at a time, and everything family-related on screen belongs to it. With a single family, that's simply the user's family.
- **Home** is the first screen after sign-in: a dashboard of cards. Each feature can contribute a card, either for the family (e.g. the current shopping list) or for the user personally.
- **No tab bar** until there are at least three main destinations. Until then, a feature opens from its Home card.
- **The app needs a connection.** Without one, it still shows the last data it loaded, read-only, under an "Offline · showing saved data" banner.
- **Changes show immediately.** Edits appear as soon as they're made and are saved in the background. If saving fails after automatic retries, the change is undone and a snackbar says what failed, with **Retry**.
- **Data stays fresh** by reloading when a screen opens or the app returns from the background. Lists can also be pulled down to refresh.
- **App updates** download in the background. When one is ready, a snackbar "Update ready · Restart" offers to restart into it. The app never restarts on its own. This comes with the Over-the-air updates slice; until then, a new version means installing a new APK.
- **Update required.** If the app is too old for the API, a blocking screen asks the user to update. If an over-the-air update is available, the screen offers **Restart** to apply it. Otherwise the update needs a new app install, and the screen links to the latest APK on GitHub.

## Rules

- Heavy or irreversible actions (deleting a family, leaving one) always ask for confirmation in a dialog.
- Lighter destructive actions may instead happen at once with an **Undo** snackbar. Each feature's spec says which applies.
- No "Saved" messages: a change appearing is the confirmation.

## Acceptance criteria

- The avatar menu opens from the top app bar and offers profile, settings and sign-out.
- With no connection, previously loaded screens still show their data, and the offline banner is visible. It disappears when the connection returns.
- A failing change is reverted and shows a snackbar with Retry, and Retry repeats the change.
- Opening a screen, returning to the app, or pulling down a list shows fresh data.
- After an update has downloaded, the restart snackbar appears; tapping Restart starts the new version (Over-the-air updates slice).
- When the API turns the app away as too old, nothing but the "Update required" screen is reachable. Its Restart (over-the-air) or download link (new APK) leads to a version the API accepts.

## UI

Every screen uses the same states:

- **Loading:** a progress indicator, only when nothing has been loaded before. Reloads in the background show nothing.
- **Empty:** an illustration or icon, a title, a short explanation and, where it makes sense, the action that fills it ("Create your first list"). Each feature's spec words its own.
- **Error, with nothing to show:** a message and **Retry**.
- **Offline, with nothing to show:** a message saying so, without Retry. The data loads by itself once the connection is back.
- **Error while showing earlier data:** the data stays, with a snackbar "Couldn't refresh".
- **Form errors:** shown next to the field they concern.

## Non-goals

- Editing while offline.
- Seeing other people's changes live, without a reload.
- Switching between several families (later, see [ideas.md](ideas.md)).
- Languages other than English, for now.

## Open questions

- **Profile and settings contents.** What the profile and settings screens hold, beyond editing the name (see [sign-in.md](sign-in.md)).
- **Home without features.** What Home shows before any feature contributes a card.

## Implementation

- Technical design: [`app/docs/ARCHITECTURE.md`](../app/docs/ARCHITECTURE.md).
- App: most of this is built in the App baseline slice, and app updates in the Over-the-air updates slice.
  - The shared screen states are in `app/src/ui/` (`loading-state.tsx`, `empty-state.tsx`, `error-state.tsx`), and the snackbar host is `app/src/ui/snackbar.tsx` (#23).
  - The API client, error copy and query setup are in `app/src/api/`: the persisted cache and "Couldn't refresh" in `query-client.ts` and `query-provider.tsx`, refetching on focus and reconnect in `managers.ts` and `use-api-query.ts`, and the offline first load in `query-state.ts`. The offline banner is `app/src/ui/offline-banner.tsx`, shown by `Screen` (#24).
