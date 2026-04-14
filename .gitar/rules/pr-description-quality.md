---
title: PR Description Quality Standards
description: Ensures PR descriptions meet starlark-worker quality criteria using guidance from PR template
when: PR description is created or updated
actions: Read PR template for guidance then report requirement status
---

# PR Description Quality Standards

When evaluating a pull request description:

1. **Read the PR template guidance** at `.github/pull_request_guidance.md` to understand the expected guidance for each section
2. Apply that guidance to evaluate the current PR description
3. Provide recommendations for how to improve the description.

## Core Principle: Why Not How

From https://cbea.ms/git-commit/#why-not-how:
- **"A diff shows WHAT changed, but only the description can explain WHY"**
- Focus on: the problem being solved, the reasoning behind the solution, context
- The code itself documents HOW - the PR description documents WHY

## Evaluation Criteria

### Required Sections (must exist with substantive content per PR template guidance)

1. **What changed?**
   - 1-2 line summary of WHAT changed technically
   - Focus on key modification, not implementation details
   - Template has good/bad examples

2. **Why?**
   - Full context and motivation
   - What problem does this solve? What's the use case?
   - What's the impact if we don't make this change?
   - **CRITICAL**: Technical rationale required for ALL changes (not just large ones)
   - Must explain WHY this implementation approach was chosen

3. **How did you test it?**
   - Concrete, copyable commands with exact invocations
   - ✅ GOOD: `go test -v ./plugin/hashlib/... -run TestHashlib`
   - ❌ BAD: "Tested locally" or "See tests/foo_test.go"
   - For integration tests: setup steps + commands
   - For integration tests: include Cadence setup steps + commands

4. **Potential risks**
   - Plugin API compatibility concerns?
   - Starlark sandbox or thread safety impact?
   - Cadence SDK compatibility?
   - Performance impact on concurrent or long-running workflows?

5. **Release notes**
   - If this completes a user-facing feature, describe it for Go module consumers and Starlark script authors
   - Skip for: incremental work, internal refactors, partial implementations

6. **Documentation Changes**
   - README.md or CONTRIBUTING.md updates needed?
   - New exported Go APIs documented via godoc?
   - Starlark-facing changes need examples in testdata?
   - Only mark N/A if certain no docs affected

### Quality Checks

- **Skip obvious things** - Don't flag items clear from folder structure
- **Skip trivial refactors** - Minor formatting/style changes don't need deep rationale
- **Don't check automated items** - Issue links, CI, linting are automated

## FORBIDDEN - Never Include

- ❌ "Issues Found", "Testing Evidence Quality", "Documentation Reasoning", "Summary" sections
- ❌ "Note:" paragraphs or explanatory text outside recommendations
- ❌ Grouping recommendations by type

## Section Names (Use EXACT Brackets)

- **[What changed?]**
- **[Why?]**
- **[How did you test it?]**
- **[Potential risks]**
- **[Release notes]**
- **[Documentation Changes]**
