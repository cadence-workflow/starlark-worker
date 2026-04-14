<!-- 1-2 line summary of WHAT changed technically:
- Always link the relevant GitHub issue, unless it is a minor bugfix
- Good: "Added `hashlib` plugin with SHA-256 and MD5 support for Starlark scripts #42"
- Bad: "added new plugin" -->
**What changed?**


<!-- Your goal is to provide all the required context for a future maintainer
to understand the reasons for making this change (see https://cbea.ms/git-commit/#why-not-how).
How did this work previously (and what was wrong with it)? What has changed, and why did you solve it
this way?
- Good: "Starlark scripts that need to hash data currently have no way to do so without
  importing unsafe native Go code. This prevents multi-tenant environments from offering
  common cryptographic utilities safely. This change adds a sandboxed `hashlib` plugin
  that exposes SHA-256 and MD5 through the standard plugin registry, keeping the Starlark
  sandbox intact while covering the most common hashing use cases."
- Bad: "Adds hashing support" -->
**Why?**


<!-- Include specific test commands and setup. Please include the exact commands such that
another maintainer or contributor can reproduce the test steps taken.
- e.g. Unit test commands with exact invocation
  `go test -v ./plugin/hashlib/... -run TestHashlib`
- For integration tests against Cadence, include setup steps
  Example: "Started local Cadence via Docker Compose, registered the default domain,
  then ran `go test -v ./test/... -run TestCadenceIntegration`"
- For Starlark script changes, include which testdata scripts were exercised
  Example: "Ran `go test -v ./star/... -run TestExecFile` which exercises test/testdata/ping.star"
- Good: Full commands that reviewers can copy-paste to verify
- Bad: "Tested locally" or "Added tests" -->
**How did you test it?**


<!-- If there are risks that the release engineer or downstream consumers should know about,
document them here. For example:
- Has a plugin API changed? Will existing Starlark scripts break?
- Has the workflow or activity registration signature changed? Is it backwards compatible?
- Has a Cadence SDK dependency been bumped? Are there known breaking changes?
- Could this affect Starlark sandbox safety or thread isolation?
- Is there a performance concern for long-running or concurrent workflows?
- If truly N/A, you can mark it as such -->
**Potential risks**


<!-- If this PR completes a user-facing feature or changes functionality, add release notes here.
Your release notes should allow a Go module consumer or Starlark script author to understand
the changes with little context. Always link the relevant GitHub issue when applicable.
- For new plugins: describe the Starlark-facing API
- For SDK version bumps: note the old and new versions
- Skip for: incremental work, internal refactors, partial implementations -->
**Release notes**


<!-- Consider whether this change requires documentation updates:
- Does the README.md need updating (new setup steps, changed CLI flags)?
- Does CONTRIBUTING.md need updating (new dev workflow, changed test commands)?
- Are new exported Go APIs documented via godoc comments for pkg.go.dev?
- Do Starlark-facing changes need examples in testdata or doc comments?
- Only mark N/A if you're certain no docs are affected -->
**Documentation Changes**

