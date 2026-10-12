# Specs

What the app does and why, feature by feature, and what gets built next. Specs describe behavior a user can observe. How it's built lives in code and technical docs (`docs/`, `api/docs/`, package docs), and the coding rules live in the `CLAUDE.md` files.

## Product

A family life-management app. Users sign in with Google and belong to one or more families. Most app data (meals, shopping lists, health) belongs to a family, and users reach it through their membership. See the root [README](../README.md).

## Features

| Feature | Spec | Status |
| ------- | ---- | ------ |
| Sign-in | [sign-in.md](sign-in.md) | in progress: API built, app not |
| Families, members and invites | [families.md](families.md) | in progress: API built, app not |
| App shell | [app-shell.md](app-shell.md) | specified |
| Shopping | [shopping.md](shopping.md) | specified |

Unrefined ideas go in [ideas.md](ideas.md).

## Roadmap

The next slices, in order. Each one is a single PR unless noted.

1. **App baseline**: the foundation from [`app/docs/ARCHITECTURE.md`](../app/docs/ARCHITECTURE.md), with no features yet ([app-shell.md](app-shell.md)). It covers:
   - a development build and the dev and prod variants
   - the folder structure, replacing the template's leftovers
   - the Paper theme
   - the query client, with its persisted cache and the focus and online wiring
   - the API client and `ApiError`
   - the snackbar host and the offline banner
   - `expo-updates` with the restart prompt
   - the API compatibility level: the app's header and "Update required" screen, and the API's minimum-level middleware
   - release APKs published as GitHub releases
   - the shared loading, empty and error states
   - i18n setup
   - Jest and React Native Testing Library
   - app CI

   It may split into several PRs when planned.
2. **App sign-in**: sign in with Google, stay signed in, sign out ([sign-in.md](sign-in.md))
3. **App families**: first run (create a family), family screen, roles, leave and delete ([families.md](families.md#ui))
4. **App invites**: show an invite as a QR code and a link, and join by scanning or pasting one ([families.md](families.md#ui))
5. **Shopping API: catalogue**: catalogue items and the category keys ([shopping.md](shopping.md))
6. **Shopping API: lists**: lists and items, ticking, and finishing with carry-over ([shopping.md](shopping.md))
7. **App shopping**: the Home card, the Shopping screen, the list screen, "Add items", finishing, past lists and the catalogue editor ([shopping.md](shopping.md#ui)). The spec's UI is only a baseline, so this slice starts by refining the UX of "Add items" and of the catalogue editor. It may split into several PRs when planned.

## Workflow

1. **Capture:** an idea goes into `ideas.md` as a line or two.
2. **Specify:** in a session without code, the idea becomes a feature spec. Open questions are listed explicitly rather than guessed.
3. **Roadmap:** the spec's next PR-sized slices are added to the roadmap above, in order.
4. **Plan:** a slice starts in plan mode with a technical plan, re-checking library and vendor docs then. The plan isn't committed. Its outcome lands in code, docs and the PR description.
5. **Implement:** one slice per PR. If behavior changes while building it, the spec changes in the same PR. The slice leaves the roadmap, and the spec's status and Implementation section are updated.

Open questions are resolved, or explicitly deferred, before the slice that needs them.

## Spec template

An `idea` spec can be a few bullets. Detail grows as it nears the front of the roadmap.

```md
# <Feature>

Status: idea | specified | in progress | built

## Why

## Behavior
What the user can do, as flows or bullets.

## Rules
Business rules the API enforces.

## Acceptance criteria
Testable statements. They map to endpoint tests, and later to app tests.

## UI
Screens and their states: empty, loading, error, edge cases.

## Non-goals

## Open questions

## Implementation
Links to the code and technical docs, filled in as slices land.
```
