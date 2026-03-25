# Keeping gemara-mcp Up to Date

The gemara-mcp server is built from source against the
upstream [gemaraproj/gemara-mcp](https://github.com/gemaraproj/gemara-mcp)
repository. To ensure your build reflects the latest schema
support and bug fixes, sync with upstream regularly.

## If You Cloned Directly from gemaraproj (No Fork)

```bash
cd gemara-mcp
git fetch origin
git checkout main
git pull origin main
make build
```

## If You Cloned from a Personal Fork

```bash
cd gemara-mcp

# Add upstream remote (one-time setup)
git remote add upstream \
  https://github.com/gemaraproj/gemara-mcp.git

# Fetch and merge upstream changes
git fetch upstream
git checkout main
git merge upstream/main

# Rebuild
make build
```

## Verify the Build

```bash
# Check the binary runs
./bin/gemara-mcp --version

# Or use Gemara User Journey's doctor command
cd /path/to/gemara-user-journey
./gemara-user-journey --doctor
```

## When to Sync

- Before starting a new tutorial or authoring session
- When `./gemara-user-journey --doctor` reports a version mismatch
  between the MCP server and your selected Gemara schema
  version
- When new Gemara schema releases are published at
  [gemaraproj/gemara](https://github.com/gemaraproj/gemara/releases)

The gemara-mcp server's schema version must be compatible
with the Gemara schema version used in your Gemara User Journey
session. If there is a version mismatch, Gemara User Journey will
warn you in the handoff summary after completing a
tutorial.

---

[Back to README](../README.md)
