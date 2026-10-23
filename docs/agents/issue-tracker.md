# Issue tracker: Local Markdown

Issues and specs for this repo live as markdown files in `.scratch/`.

## Conventions

- One feature per directory: `.scratch/<feature-slug>/`
- The spec is `.scratch/<feature-slug>/spec.md`
- Implementation issues are one file per ticket at `.scratch/<feature-slug>/issues/<NN>-<slug>.md`, numbered from `01`, never a single combined tickets file
- Ticket status is recorded as a `Status:` line near the top of each issue file. New implementation tickets start as `ready-for-agent`; completed implementation tickets are `resolved`. Triage-only roles are defined in `triage-labels.md`.
- A ticket is unblocked when every ticket listed in `Blocked by:` has `Status: resolved`.
- When an implementation ticket is completed, check every satisfied acceptance criterion and add or update a compact `## Completion` section:

  ```md
  ## Completion

  Completed: YYYY-MM-DD
  Commit: <short-hash>
  Notes: <one sentence>
  ```

- When a local spec is fully implemented, append or update an `## Implementation` section in `spec.md`:

  ```md
  ## Implementation

  Completed: YYYY-MM-DD
  Branch or PR: <branch-or-pr>
  Tickets resolved: <ticket list>
  ```

- Comments and conversation history append to the bottom of the file under a `## Comments` heading

## When a skill says "publish to the issue tracker"

Create a new file under `.scratch/<feature-slug>/` (creating the directory if needed).

## When a skill says "fetch the relevant ticket"

Read the file at the referenced path. The user will normally pass the path or the issue number directly.

## Wayfinding operations

Used by `/wayfinder`. The `map` is a file with one `child` file per ticket.

- `Map`: `.scratch/<effort>/map.md` (the Notes / Decisions-so-far / Fog body).
- `Child ticket`: `.scratch/<effort>/issues/NN-<slug>.md`, numbered from `01`, with the question in the body. A `Type:` line records the ticket type (`research`/`prototype`/`grilling`/`task`); a `Status:` line records `claimed`/`resolved`.
- `Blocking`: a `Blocked by: NN, NN` line near the top. A ticket is unblocked when every file it lists is `resolved`.
- `Frontier`: scan `.scratch/<effort>/issues/` for files that are open, unblocked, and unclaimed; first by number wins. Implementation tickets use `ready-for-agent` for open and `resolved` for closed.
- `Claim`: set `Status: claimed` and save before any work.
- `Resolve`: append the answer under an `## Answer` heading, set `Status: resolved`, then append a context pointer (gist + link) to the map's Decisions-so-far in `map.md`.
