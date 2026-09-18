# CLAUDE.md

Guidance for Claude Code in this repo. Product description: README.md.

## Learning context

- `app/` (React Native + Expo, TypeScript) — brand new to the user. Primary learning goal.
- `api/` (Go) — also new, deliberately chosen over Node/TS (which the user has 10 years of experience with) as a second, parallel learning track.
- Everything else — general software engineering, API design, standard React, TypeScript — the user has 10+ years of experience with. Don't explain fundamentals they already know.

## How to work in this repo

- **Act as a teacher, not an oracle.** Before or while implementing something non-trivial in `app/` or `api/`, explain the _why_: what the relevant RN/Expo or Go concept is, why this approach fits, and the trade-offs — especially where it differs from the React or Node/TS equivalent the user already knows. Analogies to standard React (for RN) and to Node/TS API design (for Go) are usually the fastest way to land an explanation.
- **You write the code.** The user wants to fully understand it — well enough to modify it by hand — but isn't asking to hand-type it themselves. Implement it, but make sure the explanation alongside it is enough that the code isn't a black box.
- **Keep code idiomatic and legible over clever.** This is explicitly not a "vibe coded" project. Prefer explicit, conventional patterns for the ecosystem (idiomatic Go, idiomatic Expo/RN) over abstractions that are hard to follow, even if more concise.
- Don't re-litigate decisions already made below — build on them. Do flag it if a new request seems to conflict with one.

## Decisions made so far

| Area          | Decision                                               | Notes                                                                                           |
| ------------- | ------------------------------------------------------ | ----------------------------------------------------------------------------------------------- |
| Mobile app    | React Native + Expo, TypeScript (`app/`)               | Primary learning goal                                                                           |
| Backend       | Go, custom HTTP API (`api/`), framework not chosen yet | Deliberate second learning track                                                                |
| Repo layout   | Monorepo: `app/`, `api/`, `infra/`                     | `infra/` created once hosting is decided                                                        |
| Hosting/infra | Not decided yet                                        | Needs to be free or cheap — personal/family-scale traffic. Revisit once the API has real shape. |
| Database      | Not decided yet                                        |                                                                                                 |
| Auth          | Not decided yet                                        |                                                                                                 |

Update this table as decisions get made, so future sessions don't need to re-derive context from conversation history.

## Commands

### `app/` (run from `app/`)

```
npm run start   # expo start
npm run ios
npm run android
npm run lint    # expo lint
```

### `api/`

Not scaffolded yet. Populate with the actual commands (e.g. `go test ./...`) once it exists.
