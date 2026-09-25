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
| Backend       | Go, stdlib `net/http` (`api/`), no framework           | Deliberate second learning track. Go 1.22+ ServeMux pattern routing; revisit (e.g. chi) only if it feels thin. |
| API tooling   | golangci-lint v2, Makefile, GitHub Actions CI          | CI runs lint + `go test -race` + docker build on PRs touching `api/`.                            |
| Repo layout   | Monorepo: `app/`, `api/`, `infra/`                     | `infra/` created once hosting is decided                                                        |
| Hosting/infra | Google Cloud Run, `europe-north1` (Hamina, Finland)    | Always-free tier confirmed in console for this region: 2M req, 180k vCPU-s, 360k GiB-s / month. Same country as the DB. |
| Deploy artifact | Docker image, multi-stage build on `distroless/static` | `CGO_ENABLED=0` for a static binary; ~15–25MB image. Gotchas: distroless has no shell, and needs `import _ "time/tzdata"` for `time.LoadLocation`. |
| Database      | Existing Postgres in Oulu, Finland (hobby association, no extra cost) | **Pending verification**: must be reachable over public internet with `sslmode=verify-full`. If it's firewall/VPN-only, Cloud Run can't reach it and this choice collapses. |
| DB backups    | Deferred                                               | Revisit once there's real family data in the DB — someone else operates that server, so their backup practices are currently an unknown we've accepted. |
| Auth          | Not decided yet                                        |                                                                                                 |

Portability hedge: keep the API coupled to nothing but a `DATABASE_URL` and a container image. That keeps a provider switch to an afternoon — which matters because the database is a social arrangement, not a contract.

Update this table as decisions get made, so future sessions don't need to re-derive context from conversation history.

## Commands

### `app/` (run from `app/`)

```
npm run start   # expo start
npm run ios
npm run android
npm run lint    # expo lint
```

### `api/` (run from `api/`; see `api/README.md` for prerequisites)

```
make run            # go run ./cmd/api, on $PORT (default 8080)
make test           # go test -race ./...
make lint           # golangci-lint run
make fmt            # golangci-lint fmt
make docker-build
make docker-run
```
