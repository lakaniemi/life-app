# Families, members and invites

Status: in progress (API built, app not)

## Why

Most app data belongs to a family, such as meal plans and shopping lists. A family is the group that shares it, and membership decides who can see and change it.

## Behavior

- A signed-in user can create a family and becomes its admin.
- A user can belong to several families, each with its own role: `admin` or `member`.
- Members see the family's name and its members with their roles.
- Admins can rename the family, delete it, change members' roles, remove members and invite new ones.
- Any member can leave.
- **Invites:** an admin taps "Invite a member", and the app shows an invite as a QR code and a shareable link. Anyone who opens it previews the family's name, then confirms to join as a `member`. An invite is never managed: it works for anyone, any number of times, for 24 hours, then expires. 24 hours rather than minutes, because a link sent over chat may be opened later in the day.

## Rules

1. Every family has at least one admin. The only admin can't be demoted, and can't leave while others remain. They promote someone first.
2. The only member can't leave. They delete the family instead.
3. Outsiders can't tell whether a family exists: every family route answers 404 to non-members.
4. Accepting an invite as an existing member changes nothing. An admin who scans their own QR code stays admin.
5. An unknown invite and an expired one look the same (404), so nobody can probe which invites ever existed.
6. Deleting a family deletes all of its data, for every member.

## Acceptance criteria

API (covered by `api/internal/server/families_test.go` and `invites_test.go`):

- Creating a family makes the caller its admin. The name is trimmed, required and at most 100 characters.
- Listing families returns the caller's families with the caller's role in each, or an empty list. `GET /me` includes the same families.
- A non-member gets 404 on every `/families/{familyID}/…` route.
- A member gets 403 for admin actions: renaming, deleting, changing roles, removing someone else and creating invites.
- An admin can demote themselves while another admin remains. The sole admin can't: 409 `sole_admin`.
- A member can leave, and so can an admin while another admin remains. The sole admin can't while others remain (409 `sole_admin`), and the sole member can't (409 `sole_member`).
- A role change applies at once: an admin demoted by a concurrent request can't complete an admin action.
- An invite token previews the family (`id`, `name`) and tells whether the caller is `alreadyMember`.
- Accepting joins as `member` (204), and the same token works for further users. Creating a new invite leaves earlier active ones valid.
- An unknown or expired token gives 404 on both preview and accept.

App: written when the app slices are planned.

## UI

- **Single family for now.** The app works in one active family (see [app-shell.md](app-shell.md)), which is the user's family. A user can't join a second family from the app yet. A family switcher can come later without API changes.
- **First run:** a signed-in user with no family chooses "Create a family" or "Join with an invite".
- **Family screen:** the name, and the members with their roles. Admins can rename the family, change roles and remove members.
- **Leave and delete:** "Leave" becomes "Delete family" when you're the sole member. If you're the sole admin and others remain, the app asks you to promote someone first.
- **Invite:** a single "Invite a member" button for admins. It shows the invite as a QR code with a button to share the link. Closing the screen leaves the invite to expire.
- **Join:** scan the QR code, or open the link from another app. Some chat apps don't make `lifeapp://` links tappable, so "Join with an invite" also accepts a pasted link. The preview shows the family's name before confirming.
- **Join preview states:**
  - Already a member: "You're already in X", with "Open family" instead of "Join".
  - Already in another family: "You're already in a family. Joining a second one isn't supported yet." No "Join" button.
  - Unknown or expired invite: "This invite has expired or isn't valid. Ask for a new one."

## Non-goals

- Listing, revoking or re-displaying invites.
- Typing an invite by hand. A QR code and a link cover every case.
- Several families in the UI, for now.
- Ownership separate from the admin role.

## Implementation

- API: `api/internal/family` (rules, authorization and locking in the package doc), with handlers in `api/internal/server/families.go` and `invites.go`.
- Invite links are `lifeapp://join?token=…`. The API deals only in tokens; the link format is the app's concern.
- Decisions: root `CLAUDE.md`, the Authorization and Invites rows.
- App: not started.
