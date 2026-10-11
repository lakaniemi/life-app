# Life App

Family life-management app — meal planning, shopping lists, and health tracking to start. Not a calendar replacement, but may integrate with one later. Scope is open; features get added as needed.

Users sign in with Google and belong to one or more families. Most app data belongs to a family, and users reach it through their membership. Feature specs and the roadmap are in [`specs/`](specs/README.md).

## Stack

- **`app/`** — React Native + Expo, TypeScript. Primary learning goal.
- **`api/`** — Go. Second learning track (not Node/TS, which is already familiar). See [`api/README.md`](api/README.md).
- **`infra/`** — TBD. Cheap/free hosting, sized for personal-scale traffic.
