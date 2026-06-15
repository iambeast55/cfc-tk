## Context

Targets, credentials, Kerberos caches, and open-port findings already exist as team-scoped resources. The first version of the access matrix can derive likely access in the Svelte UI without adding backend persistence.

## Goals / Non-Goals

**Goals:**

- Provide a dense per-team access view.
- Infer likely methods from open ports and credential material.
- Keep the first implementation reversible and frontend-only.

**Non-Goals:**

- Credential validation or login attempts.
- Historical access state.
- Backend-calculated confidence.

## Decisions

1. Compute matrix rows client-side.

   Existing APIs already provide the source data. This avoids adding a backend contract before the usefulness of the view is proven.

2. Treat results as inferred, not verified.

   Badge confidence communicates strength of the guess. The UI should not imply that access has been tested.

3. Use row-oriented cards/table instead of a huge grid.

   A full credential-by-target grid can become unwieldy quickly. Rows grouped by credential with target method cells keep the view readable in the existing app shell.

## Risks / Trade-offs

- Inference can be wrong -> Use confidence labels and avoid saying access is confirmed.
- Large teams could create many cells -> Add filters and hide empty rows by default.
- No backend persistence -> Acceptable for v1; derived data updates as source data changes.
