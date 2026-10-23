# Defining tier banding

Type: grilling
Status: resolved
Blocked by: 01

## Question

How do Pod resource traits become a Tier relative to the namespace? Decide:
- which traits feed the tier
- how many bands there are and how they are cut
- which levers each tier scales (Hits, speed, contact damage, size)
- where Pods without requests go
- what happens to banding when Pods arrive or leave mid-session, given a tier is fixed at spawn

## Answer

- **Inputs**: effective CPU and memory **requests**, summed across containers and including sidecars. Limits are ignored, since Kubernetes already copies a lone limit into the request.
- **Score**: CPU and memory are each divided by the namespace **median** for that resource, taken over eligible Pods that have requests. The score is the **larger** of the two ratios.
- **Bands** (three tiers):
  - **Light**: score under ½.
  - **Standard**: score from ½ to 2.
  - **Heavy**: score over 2.
  - A uniform namespace is all Standard.
  - A Pod with no requests is **Light** and is excluded from the median. If no Pod has requests, everything is Light.
  - A lone Pod with requests is Standard.
- **Timing**:
  - A spawning Pod is compared against all eligible Pods currently known, itself included, and its tier is never recomputed.
  - A practice-mode respawn keeps its tier.
  - A replacement Pod in chaos mode gets a fresh tier.
- **Shared levers**, the same for every archetype. These are starting values, to be tuned after trying them:

  | Lever | Light | Standard | Heavy |
  |---|---|---|---|
  | Hits | 2 | 3 | 5 |
  | Size (hitbox scales too) | 0.85× | 1× | 1.3× |
  | Speed | 1.1× | 1× | 0.85× |
  | Contact damage | 1 | 1 | 2 |

  Separation must account for the sizes of each pair of enemies instead of one constant.
- **Readability**: size is the main cue. Heavy also gets armour bands and an extra tentacle; Light gets a thinner hull.
- Escalation specific to each archetype, such as Turret fire patterns, is separate.
