# ADR-0002: MCP Server Installation Security and Update Strategy

**Status**: Accepted
**Date**: 2026-03-13
**Deciders**: Project maintainers
**Constitution Version**: 1.3.0

## Context

Gemara User Journey's first-launch setup offers to install the Gemara MCP
server (`gemara-mcp`) from source. The original implementation
had three issues:

1. **SSH/HTTPS prompt**: Users were asked to choose between SSH
   and HTTPS cloning, but most users cannot determine whether
   their SSH keys are configured for GitHub. Choosing SSH when
   keys are not configured results in a confusing git clone
   failure.

2. **Prerelease handling**: The installer fetched the latest
   *stable* release via GitHub's `/releases/latest` endpoint.
   When the upstream repository only has prereleases (as is
   currently the case with `gemaraproj/gemara-mcp` at `v0.0.0`),
   this returned a 404 error with raw GitHub API JSON, leaving
   users unable to install.

3. **No update checking**: Once installed, Gemara User Journey never checks
   for newer versions of `gemara-mcp`. Users must manually
   discover and install updates.

4. **Supply chain security**: The installer must ensure that the
   exact code fetched from upstream matches what was published
   in the release. Git tags are mutable (they can be force-pushed
   to point at different commits), so pinning to tags alone is
   insufficient.

## Decision

### SSH Auto-Detection

Remove the SSH/HTTPS prompt. Instead, auto-detect SSH access by
probing `ssh -T git@github.com` with a 5-second timeout:

- If the probe succeeds (GitHub responds with "successfully
  authenticated"), use SSH cloning.
- Otherwise, default to HTTPS cloning.

Users are never asked a question they cannot answer. The
detection result is displayed to the user for transparency.

### Prerelease Fallback

When no stable release exists (GitHub `/releases/latest` returns
404), fall back to the releases list (`/releases?per_page=1`)
which includes prereleases. The `ReleaseInfo` struct carries a
`Prerelease` flag so the UI can indicate when a non-stable
release is being used.

Once the upstream project publishes a stable release, the
`/releases/latest` endpoint returns it directly and the
prerelease fallback is never reached. No code changes are needed
when that transition happens.

### SHA-Pinned Installation

All installations pin to the release's `target_commitish` SHA
digest, not the tag name. The installation sequence is:

1. `git clone <url> <dest>`
2. `git checkout <SHA>`
3. `make build`

Tags can be moved to point at different commits (via
`git tag -f`). SHA digests are cryptographic hashes of the
commit content and cannot be forged without breaking the hash.
This ensures that even if a tag is maliciously reassigned, the
installed code matches exactly what was published in the GitHub
release.

### Programmatic Update Checking

On session start, when an MCP server binary is already installed
from a prior build-from-source installation, Gemara User Journey checks for
a newer release:

1. Read the installed release metadata (tag and SHA) from a
   local file (`~/.local/share/gemara-user-journey/installed-release.json`).
2. Fetch the latest release from the GitHub API (stable or
   prerelease, using the same fallback logic).
3. Compare the installed SHA against the latest release SHA.
4. If different, display both versions with their SHAs and ask
   the user for explicit confirmation before updating.
5. If the user confirms, clone at the new pinned SHA, build, and
   update the local metadata file.

The user always sees the full SHA comparison and must explicitly
approve updates. No automatic updates occur.

## Consequences

### Benefits

- Users are never asked whether they have SSH keys configured —
  the system detects it automatically and defaults to the safe
  choice (HTTPS).
- Prereleases are usable immediately, enabling early testing of
  `gemara-mcp` before stable releases are cut.
- SHA pinning prevents tag substitution attacks. The installed
  code is verified by its cryptographic hash, not a mutable
  label.
- Update checking keeps the MCP server current without requiring
  users to monitor upstream releases manually. Users retain full
  control via explicit confirmation.

### Risks and Trade-offs

- **SSH detection latency**: The `ssh -T` probe adds up to 5
  seconds to the first-launch flow when SSH is not configured
  (timeout). This is acceptable because it only runs once during
  source build installation.
- **Prerelease instability**: Prereleases may contain bugs or
  incomplete features. The `[prerelease]` label in the UI
  mitigates this by setting clear expectations.
- **Network dependency**: Update checking requires network access
  to the GitHub API. When offline, the check is skipped silently
  and the existing installation continues to work.

### Follow-up Actions

- Implement `InstalledRelease` metadata file read/write in
  `internal/mcp/install.go`.
- Add `CheckForUpdate` function that compares installed vs
  latest release.
- Wire update check into the session start flow in
  `internal/cli/setup.go`.
- Add update flow to the CLI with SHA comparison display and
  explicit user confirmation.
