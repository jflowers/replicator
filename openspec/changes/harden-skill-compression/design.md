## Context

Two embedded skill files -- `forge-global` and `always-on-guidance` -- use flat bullet lists and parallel "do/don't" structures that DCP context compression is known to weaken. The issue ([#51](https://github.com/unbound-force/replicator/issues/51)) documents six specific vulnerabilities where critical constraints sit in low-survival positions.

The proposal (constitution alignment: all PASS/N/A) calls for structural reorganization without semantic changes to any rule.

## Goals / Non-Goals

### Goals
- Restructure `forge-global` to use compression-resilient patterns (decision table, inline parameters, ordered steps)
- Restructure `always-on-guidance` to use compression-resilient patterns (priority positioning, dedicated safety section, smaller lists)
- Preserve exact semantic meaning of all existing rules
- Keep both files concise -- compression resilience should not mean verbosity

### Non-Goals
- Adding new rules or behavioral constraints
- Changing the skill loading mechanism or embedded asset pipeline
- Hardening other skills beyond the two identified in #51
- Automated compression-survival testing (out of scope for this change)

## Decisions

### D1: Decision table replaces opposing lists in forge-global

The "When to Forge / Don't forge" parallel lists will be replaced with a single decision table using columns for criteria, "forge" action, and "don't forge" action. Decision tables survive compression better because they are a single coherent structure rather than two lists that a compressor might flatten independently.

Format: markdown table with Signal / Forge / Skip columns.

### D2: TTL inlined into reservation step, not standalone bullet

Line 27 (`Set ttl_seconds to auto-release after timeout`) is a standalone bullet that compression drops. It will be inlined into step 1 of the protocol: `Workers MUST call comms_reserve(paths=[...], ttl_seconds=300) before editing`. The specific parameter becomes part of the action rather than a separate detail.

### D3: Protocol steps get explicit ordering language

The 5-step File Reservation Protocol will use explicit temporal markers ("FIRST", "THEN", "FINALLY") in addition to numbered steps. Ordered temporal language survives compression better than bare numbered lists because the ordering semantics are embedded in each item rather than implied by position.

### D4: hivemind_find moves to first position

"Check hivemind_find before solving problems from scratch" moves from last bullet (position 6, lowest survival) to first bullet (position 1, highest survival) in Tool Usage. First-position items have the highest compression survival rate.

### D5: Critical safety rules get dedicated section

"Never force push to main" is currently a mid-list bullet in Git Discipline, weighted equally with stylistic guidance. It moves to a new `## Critical Safety` section at the top of the file (after the title), giving it header-level protection. Compressors preserve section headers even when they drop individual bullets.

### D6: Lists reduced to 2-3 items via sub-headers

The current 5-item Code Quality list, 4-item Error Handling list, and 5-item Testing list will be reorganized under sub-headers to create groups of 2-3 items. Shorter lists lose fewer items under compression because the ratio of dropped items is lower. The 3-item Git Discipline section (which will have 2 items after D5 extracts the safety rule) is already within the threshold and does not need splitting.

## Risks / Trade-offs

### Risk: Structural changes misinterpreted as semantic changes

**Mitigation**: Each restructured section preserves the exact same rules. Review should verify no rules are added, removed, or weakened.

### Risk: Decision table format less scannable than bullet lists

**Accepted trade-off**: Decision tables require slightly more cognitive load to parse but survive compression significantly better. The table is small (3 rows) and the trade-off favors reliability over scanability.

### Risk: Longer file from sub-headers

**Accepted trade-off**: Adding sub-headers increases line count slightly but each list becomes shorter. The net effect on token count is minimal, and shorter lists have better compression survival.
