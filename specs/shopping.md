# Shopping

Status: specified

## Why

Most groceries are the same from week to week. Making a list should be as quick as walking past the fridge and cupboards and tapping whatever is missing, with typing needed only rarely. An occasional one-off list, like one for the hardware store, should still be painless. Completed lists are kept as a history of what was bought.

## Behavior

- **Catalogue:** each family has one catalogue of the items it usually buys, including rarely bought ones like sauces. It starts empty, and any member can add, rename, recategorise and delete its items.
- **Categories:** a fixed set, in typical store order, ending with "Other":

  | Key | Label |
  | --- | ----- |
  | `produce` | Fruit & vegetables |
  | `bread` | Bread |
  | `dairy` | Dairy & eggs |
  | `meat_fish` | Meat & fish |
  | `frozen` | Frozen |
  | `dry_goods` | Pasta, rice & baking |
  | `condiments` | Sauces, spices & oils |
  | `snacks` | Snacks & sweets |
  | `drinks` | Drinks |
  | `baby` | Baby |
  | `household` | Household & cleaning |
  | `personal_care` | Personal care |
  | `other` | Other |

  The data stores keys, and the app translates them into labels, so making categories editable later is a data migration rather than a rewrite.
- **Lists:** a family can have any number of active lists. All lists are the same type: a grocery list and a hardware-store list differ only in how their items were added.
- A list has an optional name. A list without one shows its creation date instead ("List · 12 Oct"). Lists can be renamed.
- A list item has a name, a category, an optional quantity (a whole number, shown as "2× Milk"), an optional note ("lactose-free") and a ticked state.
- **Adding items:** "Add items" opens a screen with two tabs:
  - **Catalogue** (the default) shows the whole catalogue, grouped by category. Tap an item to put it on the list, and tap it again to take it off. Each category ends with "+ Add to ‹category›", which creates a catalogue item and puts it on the list. Where an item ends up depends on where it was typed, so the app never asks "also add to catalogue?".
  - **Type** has a text field that keeps focus: each Enter adds the text as an item on this list only, then clears the field for the next one. A category chip next to the field sets the category for typed items. It starts as "Other" and keeps its value between entries. If the typed name matches a catalogue item (ignoring case), the item gets that catalogue item's category instead. The items typed during this visit are listed below the field, and each can be removed.
  - Both tabs edit the same list directly. Every tap and every Enter is saved immediately, so switching tabs never needs merging, and "Done" just closes the screen. In the Catalogue tab, an item shows as selected when the list has an item with the same name (ignoring case), however that item was added. For example, typing "milk" marks Milk as selected, and tapping Milk then takes it off the list.
  - A list never holds two items with the same name. Adding a name that's already on the list highlights the existing item.
- **Creating a list** opens "Add items" straight away. For a hardware-store list, you create the list, switch to Type, type the items and tap Done.
- A list item copies its name and category from the catalogue and doesn't link back to it, so changing or deleting a catalogue item never changes a list.
- **Shopping:** on the list screen, tapping an item ticks it, and tapping again unticks it. Ticked items stay in place, struck through, so the store order doesn't shift. Each item's ⋮ menu edits its quantity and note, changes its category or removes it.
- **Finish shopping** completes the list. If every item is ticked, the list completes immediately. Otherwise a sheet says "You didn't tick these off. Select the ones to move to a new list." It lists the unticked items, none of them selected, and offers three actions:
  - **Move selected**, disabled while nothing is selected
  - **Move all**
  - **Don't move**

  Each action completes the list, and dismissing the sheet cancels finishing. Moved items go to a new active list with the same name. They're unticked, and they keep their quantity and note.
- **Completed lists** are read-only, kept indefinitely as history.
- Several members can edit and tick the same list. As described in [app-shell.md](app-shell.md), they see each other's changes when the screen reloads.

## Rules

1. Shopping data belongs to the family. Every route answers 404 to non-members, as in [families.md](families.md). All members have the same rights, and nothing is admin-only.
2. A category must be one of the fixed keys. Any other value gets 400.
3. Catalogue item names are unique within a family, ignoring case, and a duplicate gets 409. The picker's inline add avoids this: if the name already exists, it selects the existing item.
4. Item names are unique within a list, ignoring case, and a duplicate gets 409. The app avoids this by highlighting the existing item.
5. A completed list is read-only. Changing it or its items, or finishing it again, gets 409 `list_completed`.
6. Finishing a list completes it and creates the carry-over list in one transaction.
7. Each change adds, updates, ticks or removes a single item. A list is never replaced as a whole, so members editing the same list at once don't overwrite each other's changes.
8. Any member can delete any list, active or completed.
9. Names are trimmed and required: a list name (when given) and an item name can be at most 100 characters each. A note can be at most 200 characters, and a quantity, when given, is from 1 to 99.

## Acceptance criteria

API:

- Any member can create, list, rename, recategorise and delete catalogue items. A duplicate name, ignoring case, gets 409.
- An unknown category gets 400, both for catalogue items and for list items.
- Any member can create, list, rename and delete lists. Listing separates active lists from completed ones.
- Any member can add, update, tick, untick and remove list items. A duplicate name in the same list, ignoring case, gets 409.
- Renaming or deleting a catalogue item leaves existing list items unchanged.
- Finishing with no items selected completes the list and creates no new list. Finishing with items selected also creates a new active list with the same name, containing copies of those items, unticked, with their quantity and note.
- Any change to a completed list or its items gets 409 `list_completed`, and so does finishing it again.
- A non-member gets 404 on every shopping route.

App: written when the app slices are planned.

## UI

> **This is a baseline, not a final design.** "Add items" (both tabs) and the catalogue editor are the core of the feature. Their UX is refined further at the start of the app slice, in plan mode, before anything is built. Behavior and Rules are binding on the API slices regardless.

- **Home card ("Shopping"):** lists the active lists, with each one's name and item count. Tapping a list opens it. "New list" is always shown, even when lists already exist. Tapping the card's title opens the Shopping screen.
- **Shopping screen:** two equal tabs.
  - **Lists:** the active lists and "New list". A "Past lists" entry at the bottom opens the completed lists, newest first.
  - **Catalogue:** the catalogue editor.
- **Catalogue editor:** designed for building the catalogue quickly. It's a separate screen from the picker in "Add items", which is designed for choosing items. Items are grouped by category, and each category has its own add field. The field keeps focus after Enter, so a whole category can be typed in one go. Tapping an item's name renames it in place, and its ⋮ menu has "Move to category" and "Delete". Nothing in the editor changes any list.
- **List screen:** items grouped by category. If every item is in the same category, as is typical for a hardware-store list, the category headers are hidden. The main action is "Add items", next to "Finish shopping". The ⋮ menu has "Rename" and "Delete list".
- **Deleting:** removing a list item or a catalogue item happens immediately, with Undo. Deleting a list, active or completed, asks for confirmation first.
- **Empty states:**
  - With an empty catalogue, the Catalogue tab of "Add items" says "Your catalogue is empty. Add items your family usually buys", and points to the Type tab.
  - An empty list shows "Add items".
  - With no active lists, the Lists tab and the Home card offer "Create a list".

## Non-goals

- Searching the catalogue or the picker.
- Editable or family-specific categories.
- A category order per store.
- Units or prices.
- Reopening a completed list.
- Showing who ticked what.
- Integrating with meal planning. Whether and how is decided later.
- Editing offline, and seeing other members' changes live (see [app-shell.md](app-shell.md)).

## Open questions

- **"Add items" and catalogue editor UX:** refined when the app slice starts (see [UI](#ui)).
- **Category set:** the keys and labels above are a first proposal.

## Implementation

Not started.
