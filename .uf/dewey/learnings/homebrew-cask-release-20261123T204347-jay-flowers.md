---
tag: homebrew-cask-release
author: jay-flowers
category: gotcha
created_at: 2026-11-23T20:43:47Z
identity: homebrew-cask-release-20261123T204347-jay-flowers
tier: draft
as_of: "replicator@6835a938448577a200ac14f629a7085fd8caf9fa"
---

Homebrew 7's cask postflight_steps DSL requires the literal placeholder {{staged_path}}; a Ruby interpolation "#{staged_path}" makes the whole cask unreadable and `brew` rejects it. When emitting that literal from a GoReleaser custom_block (which is itself a Go text/template), escape it with a backtick raw-string action: "{{ `{{staged_path}}` }}". A text-shape regression test (grep for the literal and reject the Ruby form) can guard the rendered output, but only a real `brew` load verifies parseability, so confirm the published cask loads before announcing a release. Also note that a cask template fix is forward-only: the tap is only republished by a release run, so an already-broken published cask needs a patch release to recover existing users.
