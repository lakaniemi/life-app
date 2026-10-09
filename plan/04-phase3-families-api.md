# Phase 3 — Families API (PR 3)

All endpoints are behind `requireAuth`. The tables already exist from phase 1.

## Endpoints

| Method & path                                  | Who            | Notes                                                                 |
| ---------------------------------------------- | -------------- | --------------------------------------------------------------------- |
| `POST /families` `{name}`                      | any user       | The creator becomes `admin`. Family and membership are created in one transaction. |
| `GET /families`                                | any user       | Families I belong to, with my role                                    |
| `GET /families/{familyID}`                     | member         | The family and its members                                            |
| `PATCH /families/{familyID}` `{name}`          | admin          |                                                                       |
| `DELETE /families/{familyID}`                  | admin          | Cascades all family data                                              |
| `PATCH /families/{familyID}/members/{userID}` `{role}` | admin  | Demoting the last admin → 409                                         |
| `DELETE /families/{familyID}/members/{userID}` | admin, or self | Self = leave. The sole member → 409 (delete the family instead). The sole admin while others remain → 409 (promote someone first). |

`GET /me` also starts returning families.

## Authorization pattern

Every route under `/families/{familyID}/…`, including all future family data, goes through one helper that:
1. parses `familyID` from `r.PathValue` (malformed → 404)
2. loads the caller's membership row
3. returns **404 if the caller isn't a member**, not 403, so outsiders can't learn whether a family exists
4. hands the handler the membership (including role) for admin checks

Write it once here and reuse it everywhere.

## Membership rules

The rules from [01-data-model.md](01-data-model.md) are enforced in a transaction that first locks the family's member rows with `SELECT … FOR UPDATE`, then counts admins and members, then writes. Without the lock, two admins leaving at the same moment could both pass the "another admin remains" check.

Rule violations return **409 Conflict**: the request is valid, but the family's current state doesn't allow it. The JSON error body includes a machine-readable code so the app can show the right message.

**Where the code goes** (the layout rule in `api/CLAUDE.md`, following `internal/auth`):
- The membership rules and their transactions go in a new `internal/family` package, with sentinel errors for each rule. The handlers map those errors to 409 codes.
- The authorization helper and handlers stay in `internal/server`, in a new `families.go`.

## Teaching focus

- Transactions with pgx (`pool.Begin`, `defer tx.Rollback`, `tx.Commit`) and sqlc's `WithTx`.
- Row locking and why the check-then-write needs it.
- A consistent JSON error shape across handlers.
- Table-driven tests through `New(...)` with several users and families.

## Done when

- Tests cover each membership rule, including non-member → 404 and member-not-admin → 403 on admin actions.
- CI is green.
