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
- **Specs drive the work.** What to build next is the roadmap in `specs/README.md`, and each feature's behavior is in its spec there (Markdown in the repo, not GitHub Issues). Start a slice from its spec, in plan mode. A slice can take several PRs. When behavior changes while building it, update the spec in the same PR, and update the roadmap and the spec's status when the slice's last PR lands.
- Don't re-litigate decisions already made (below, in `api/CLAUDE.md`, `app/docs/ARCHITECTURE.md` and `docs/`) — build on them. Do flag it if a new request seems to conflict with one.

## Pull requests

The user has to be able to review every PR, so PRs stay small and stand on their own.

- **Target ~300 lines of logic per PR**, counted as lines added plus removed. Tests, lockfiles, generated files, Markdown docs and specs, and translation JSON don't count. Measure with:
  `git diff --numstat main... -- . ':!**/package-lock.json' ':!**/go.sum' ':!**/*.test.ts' ':!**/*.test.tsx' ':!**/*_test.go' ':!**/*.md' ':!**/i18n/*.json'`
- **Going over needs the user's say-so.** If a PR is heading past the target, stop and ask whether to split or keep it whole, and propose the split you would make. A PR that stays over says why in its description.
- **Each PR stands on its own.** `main` stays green and deployable after the merge, and nothing a user can reach is half-built. Tested code that nothing uses yet (an endpoint without a screen, a screen without a navigation entry) is fine. No feature flags.
- **Mechanical changes get their own PR**, with no logic mixed in: SDK and dependency bumps, template regeneration, renames and file moves. They can be large because they're skimmed, not read line by line.
- **Size is planned up front.** A slice's plan estimates its logic lines and names the PR boundaries, cut by what each PR can ship alone rather than by concept.

## Decisions

- **Stack:** `app/` is React Native + Expo (TypeScript). `api/` is Go with stdlib `net/http` and no framework; revisit (e.g. chi) only if it feels thin. The database is PostgreSQL: Postgres 18 locally in Docker, with SQL kept portable to 17. API-specific decisions are in `api/CLAUDE.md`, and the app's architecture is in `app/docs/ARCHITECTURE.md`.
- **Auth:** native Google sign-in; the API verifies the ID token and issues its own opaque session. See `docs/AUTH.md`.
- **`app_outdated` is a 400.** An app below the API's minimum compatibility level gets 400 `app_outdated`, following Google's AIP-193 (FAILED_PRECONDITION is 400). 426 is meant for protocol upgrades, and the app acts on the `code` anyway.
- **Portability hedge:** keep the API coupled to nothing but a `DATABASE_URL` and a container image (multi-stage build on `distroless/static`). That keeps a provider switch to an afternoon. The image is the contract with hosting: `/api` as the entrypoint and `/migrate` beside it, configured only through env vars, built for `linux/amd64`.
- **Hosting:** a shared server configured in a separate repo, `lakaniemi/infra`; its design and open decisions live there. What this repo must keep working for it:
  - **Merging to `main` deploys.** CI publishes `ghcr.io/lakaniemi/life-app-api:<sha>`, moves `:production`, calls the server's deploy hook and waits for `/health` to report the commit (`.github/scripts/deploy-api.sh`). The hook only asks the server to check `:production`, so the server decides what runs. `DEPLOY_WEBHOOK_TOKEN` is used only by the publish and rollback jobs, never on pull requests.
  - **Rollbacks** go through the `api-rollback` workflow. It re-points `:production` at an already published `<sha>` and calls the same hook. Only this repo's workflows move `:production`, through `GITHUB_TOKEN`.
  - **Migrations run as a one-off `/migrate up` before the new API replaces the old one**, so the old API still serves during them (`api/CLAUDE.md`, Database).
  - **`GET /health`** stays public and cheap. Its `version` (the build's commit) is how a deploy is confirmed.
  - Postgres runs on the same host, reachable only on an internal network, so the API connects without TLS.
- **App releases:** the signed APK is a GitHub release (`life-app.apk`, built on GitHub Actions, no EAS). Bumping `app/version.json` (version and `versionCode`, by hand) on `main` triggers the release, and CI refuses an APK not signed with the release key. See `app/README.md`, Releasing.

## Open decisions

Don't present these as decided.

- **Over-the-air update server.** The app uses the stock `expo-updates` client against a self-hosted server implementing the Expo Updates protocol, not EAS Update. Which server (xprem, formerly expo-open-ota, or our own in Go) is decided when the Over-the-air updates slice is planned, the last on the roadmap.
- **Dependency maintenance.** How new versions (Expo SDK patches, npm and Go packages) get noticed and bumped. App CI checks versions only against the installed SDK, so nothing flags new patches yet.

Update these lists as decisions get made, so future sessions don't need to re-derive context from conversation history.

## Commands

### `app/` (run from `app/`)

```
npm start             # Metro, for the installed dev build
npm run android       # regenerate android/, build and install the dev variant, start Metro
npm run android:prod  # the same for the production variant, as a release build
npm run lint
npm run typecheck
npm test
```

Run `npm run lint`, `npm run typecheck` and `npm test` before committing.

### `api/`

Run from `api/`. Targets are in the `Makefile`; `make fmt lint test` before committing.
