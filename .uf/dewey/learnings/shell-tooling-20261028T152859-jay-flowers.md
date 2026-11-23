---
tag: shell-tooling
author: jay-flowers
category: gotcha
created_at: 2026-10-28T15:28:59Z
identity: shell-tooling-20261028T152859-jay-flowers
tier: draft
---

In the opencode orchestration shell layer, BSD sed in-place edit with an empty backup suffix (`sed -i '' "Nd" file`) gets mangled: sed reports "can't read Nd: No such file or directory" and the edit silently never happens (the empty-string argument is not passed adjacent to -i). For in-place single-line surgery use head/tail composition (head -n N-1 file > tmp; tail -n +N+1 file >> tmp; mv) or `sed -i.bak` with a non-empty suffix. Verified while running the pin-test mutation verification (task 5.2) on unbound-force.
