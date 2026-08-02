## Context

The coordinator agent prompt at `internal/agentkit/content/agents/coordinator.md` is a 22-line embedded markdown file that defines the coordinator's behavioral contract. Analysis of all four agent prompt files in the project reveals a clear robustness hierarchy:

| File | Structure | Compression Resilience |
|------|-----------|----------------------|
| `coordinator.md` | Single "Rules" section, 6 bullets | Fragile |
| `worker.md` | Checklist + Constraints sections | Moderate |
| `background-worker.md` | Constraints-first, capability negations | Moderate |
| `SKILL.md` (forge-coordination) | Multi-section, MUST/NEVER keywords, numbered protocols | Reference standard |

The coordinator prompt is the most fragile. Its single critical negative constraint ("Never reserve files") sits at line 14 of 22 — 64% through the file — where a 50% truncation would drop it entirely.

## Goals / Non-Goals

### Goals
- Restructure `coordinator.md` so critical constraints survive context compression (DCP, summary, truncation)
- Adopt patterns proven in `worker.md`, `background-worker.md`, and the forge coordination skill
- Enforce explicit ordering: `forge_review` MUST precede `forge_complete`
- Maintain behavioral parity for existing rules; make the implicit `forge_review` → `forge_complete` ordering explicit

### Non-Goals
- Changing the coordinator's actual behavioral contract (no new rules, no removed rules — the `forge_review` → `forge_complete` ordering is a codification of existing implicit behavior from the forge coordination skill)
- Hardening other agent prompts (`worker.md`, `background-worker.md`) — those are separate changes
- Adding runtime enforcement of constraints (e.g., tool-level guards blocking out-of-order calls)
- Modifying the forge coordination skill (`SKILL.md`) — it already follows the reference pattern

## Decisions

### 1. Identity-first opening with embedded constraints

The first sentence of the file will state who the coordinator is AND what it must not do. Compressors prioritize opening content — an identity statement like "You are a coordinator. You orchestrate workers but NEVER reserve files or edit code directly." survives any reasonable summarization.

**Rationale**: The forge coordination skill (`SKILL.md`) demonstrates this pattern at scale with its multi-section design. `background-worker.md` places constraints before capabilities, which is structurally sound though it does not embed constraints in the identity opening itself.

### 2. Dedicated "Critical Constraints" section before workflow

Negative constraints (NEVER reserve files, NEVER edit code directly) and mandatory ordering (MUST review before complete) move to a dedicated section with a strong header, positioned before the workflow protocol.

**Rationale**: Position matters for compression. Content appearing earlier in a document is more likely to survive truncation. The forge coordination skill (`SKILL.md`) uses this pattern with its "File Reservation Rules" section.

### 3. Uppercase severity keywords (MUST/NEVER/ALWAYS)

All constraints use RFC 2119-style uppercase keywords for severity signaling. This matches the convention established in the forge coordination skill and the project constitution.

**Rationale**: Uppercase keywords serve as compression-resistant markers. A compressor summarizing "NEVER reserve files" is more likely to preserve the prohibition than one summarizing "Never reserve files" in lowercase.

### 4. Numbered protocol replacing unordered bullet list

The current 6-bullet unordered list becomes a numbered checklist with explicit ordering. This makes the `forge_review` → `forge_complete` dependency visible and enforceable.

**Rationale**: `worker.md` uses a 7-step numbered checklist that clearly conveys ordering. The forge coordination skill uses numbered protocols for both coordinator and worker flows.

### 5. Preserve YAML front matter and "Available Tools" section

The existing YAML front matter (`name`, `description`, `mode`) and the "Available Tools" footer remain unchanged. These are structural elements consumed by the agent framework.

**Rationale**: Composability First — the file's interface with the agentkit embed system must not change.

## Risks / Trade-offs

### Longer prompt consumes more context window

The restructured prompt will be approximately 30-40 lines (up from 22). This adds ~18 lines of context to every coordinator session.

**Mitigation**: The added lines are structural (section headers, numbered steps) rather than new information. The actual constraint count stays at 6. The trade-off is acceptable: 18 lines of context is trivial compared to the risk of a coordinator that silently drops quality gates.

### Compression-resistance is heuristic, not guaranteed

No prompt structure can guarantee survival under all possible compression strategies. The patterns used here (identity-first, constraints-before-workflow, uppercase keywords) are empirically effective but not provably optimal.

**Mitigation**: This is a defense-in-depth measure. The forge coordination skill (`SKILL.md`) provides a redundant statement of the same constraints. Even if the coordinator prompt is compressed, the skill's constraints may survive in a separate part of the context.

### No runtime enforcement

This change relies on prompt engineering, not code-level guards. A sufficiently degraded context could still produce constraint violations.

**Mitigation**: Runtime enforcement (e.g., blocking `forge_complete` if `forge_review` wasn't called) is a valid follow-up but is explicitly out of scope for this change. The prompt hardening provides immediate value with zero runtime risk. If the restructured prompt causes behavioral regression, the fix is to revert `coordinator.md` to its previous version — a single-file modification with no runtime dependencies, making rollback trivial.
<!-- scaffolded by uf vdev -->
