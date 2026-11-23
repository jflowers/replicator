---
tag: cli-testing
author: jay-flowers
category: gotcha
created_at: 2026-10-27T13:36:36Z
identity: cli-testing-20261027T133636-jay-flowers
tier: draft
---

cobra's root.SetOut(&buf) captures only cobra-layer output — flag-parse deprecation notices (pflag writes them to the FlagSet output, which cobra buffers in flagErrorBuf and re-prints via Command.Print to OutOrStderr, hence SetOut-capturable), help, and usage errors. Output a command's RunE produces via fmt.Println or a styles writer bound to os.Stdout bypasses SetOut entirely and goes to the real stdout. A test suite can therefore contain output-capturing tests while having zero coverage of the command's rendering contract: in replicator's init smart-sync PR (2026-10-26), both cobra-level tests captured output via SetOut but could assert only the --force deprecation notice, while the new "updated .opencode/<path>" rendering required by the spec ("CLI output MUST distinguish updated") was untestable because runInit renders through ui.NewStyles(os.Stdout). Review heuristic: for every acceptance-scenario clause of the form "the CLI renders/prints X", map it to a named test asserting X in captured output; if the only captured-output tests assert framework-emitted messages (notices, help), the rendering clauses are uncovered. Fix pattern: inject an io.Writer into the command's options struct (go pack AP-003), as cmd/replicator already does for writeDocs(&buf, reg) in docs_test.go.
