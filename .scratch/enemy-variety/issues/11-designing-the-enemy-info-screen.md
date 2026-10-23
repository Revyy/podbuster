# Designing the enemy info screen

Type: prototype
Status: resolved
Blocked by: 08

## Question

What does the info screen look like, what does it say, and where does it sit in the menus? It explains each archetype: its look, what it does, and which Pod kind it comes from. Decide:
- menu placement: from the title screen, from mode select, or both, and how to get back
- content for each of the four archetypes: live-animated droid, name, behaviour, source owner kind, and the note that deleting a Job Pod may count against the Job's retry limit
- how tiers are explained (Light/Standard/Heavy cues, and that tier comes from resource requests relative to the namespace)
- how Bolts and Symptoms are explained. Per "Deciding live cosmetic overlays", the prototype must also preview and tune the Symptom visuals: eye flicker and antenna sparks for not-ready, patch plates for restarts (none, one or two for 0, 1–4 or 5+), and the ~20-tick reboot blink.
- the layout for the 800×600 screen

Build on the archetype drawings on branch `prototype/archetype-silhouettes`.

## Answer

- **Name**: the screen is the **Field guide** (added to `CONTEXT.md`). It reads no cluster data, so all its examples are illustrative.
- **Placement**: reachable **only from mode select**. It is a third entry below Practice and Chaos, set apart by a divider so it doesn't read as a mode.
  - Enter opens it.
  - Esc/Backspace returns to mode select with the cursor on the entry.
  - The entry is inactive while the chaos confirmation is showing.
  - The title screen has no key or hint for it.
- **Format**: a paged codex (Left/Right). It combines prototype variant B with the kubectl table from variant C as its last page. The grid-only variant A was rejected.
- **Pages**:
  1. **Probe**.
  2. **Turret**.
  3. **Sentinel**.
  4. **Dart**. Each archetype page shows the name, source owner kind, one or two lines of behaviour, the three tiers side by side, and a live demo against a dummy player. There are no "look" descriptions, because the player can see the droid. The Dart's page notes that deleting a Job Pod may count against the Job's retry limit.
  5. **Tiers & Bolts**: tiers come from requests relative to the namespace, with their cues and stat effects. Bolts cost 1 Health, can't be shot down, and pass through enemies.
  6. **Symptoms**: one intro line ("Symptoms show a Pod's live health. They never change behaviour."), a live row of all four archetypes, and a reference row with two-line captions: 0 / 1–4 / 5+ restarts, not ready, and restart in play.
  7. **How Pods become enemies**: a `kubectl get pods`-style table of example Pods, labelled "illustrative, not your cluster". Each row shows its live droid and the Archetype/Tier in the archetype's colour. There is one row per owner kind, plus an unknown kind that becomes a Probe and an unmanaged row that is never an enemy. A detail pane shows the selected Pod's mapping and Symptoms.
- **Layout rules**:
  - Every page has the header "FIELD GUIDE - <title>" at 2× scale, with "page n/7" right-aligned.
  - A footer key bar holds page-specific keys, then the common ones.
  - Margins are 24 px, with at most about three short lines per text block.
  - Droids are clipped to their panels.
- **Symptom visuals**: approved as built. Not ready is an eye flicker plus antenna sparks. Restarts are welded patch plates. The reboot blink lasts 20 ticks.
- **Interactive controls** (N/P/R/T/X, 1/2/3, Up/Down) are a prototype aid for tuning. Whether the shipped Field guide keeps any of them is left to implementation; the live demos alone are enough.

Prototype (primary source): branch `prototype/enemy-info-screen`, commits `acb111f`, `81dae5b` and `bb89118`, file `game/prototype_infoscreen.go`. Run it with `go run -tags prototype . --mock`, then press Enter and choose Field guide.
