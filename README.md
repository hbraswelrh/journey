# Gemara User Journey

Gemara User Journey is a role-based tutorial guide for the
[Gemara](https://github.com/gemaraproj/gemara) governance,
risk, and compliance (GRC) schema project. It helps users
identify their role and activities, discover which Gemara
tutorials are relevant, walk through them with tailored
context, and then hand off to
[OpenCode](https://opencode.ai) with the
[gemara-mcp](https://github.com/gemaraproj/gemara-mcp)
server for artifact authoring. Gemara User Journey is the guide;
the MCP server is the authoring tool.

![Gemara User Journey Web UI](docs/images/web-ui-preview.png)

## User Journey

Gemara User Journey answers three questions for every user — **why**
Gemara matters for their role, **how** they will use it in
their daily work, and **what** each tutorial covers.

### 1. Discover

Tell Gemara User Journey your role and describe your daily activities.
It extracts domain keywords, maps them to Gemara's
seven-layer model, and recommends the artifact types you
will produce — with the specific MCP wizard or
collaborative authoring approach for each.

### 2. Learn

Walk through Gemara tutorials section by section, tailored
to your role. Sections matching your activity keywords are
highlighted as focus areas. Each section explains why it
matters for your work, how you will apply it, and what you
will learn.

### 3. Author

After completing a tutorial, Gemara User Journey presents a handoff
summary directing you to OpenCode with the gemara-mcp
server. The summary names the exact MCP wizard prompt
(e.g., `threat_assessment`), available tools
(`validate_gemara_artifact`), resources
(`gemara://lexicon`, `gemara://schema/definitions`), and
a preparation checklist — so you arrive ready to author.

## Prerequisites

- Linux or macOS (Windows: use WSL)
- [Go](https://go.dev/dl/) 1.21 or later
- [CUE](https://cuelang.org/docs/introduction/installation/)
  v0.15.1 or later
- [Git](https://git-scm.com/downloads) (with DCO sign-off
  and commit signing configured)
- [OpenCode](https://opencode.ai) — the AI coding agent
  that serves as the tutorial and authoring interface
- [gemara-mcp](https://github.com/gemaraproj/gemara-mcp)
  server (recommended — build from source)

## Getting Started

### Step 1: Clone and Build

```bash
git clone https://github.com/hbraswelrh/gemara-user-journey.git
cd gemara-user-journey
make build
```

### Step 2: Verify Your Environment

```bash
./gemara-user-journey --doctor
```

This checks your Go version, CUE installation, gemara-mcp
server availability, and `opencode.json` configuration.

### Step 3: Launch OpenCode

```bash
opencode
```

OpenCode reads the project's `AGENTS.md` for context and
starts the gemara-mcp server automatically if configured
in `opencode.json`.

### Step 4: Tell OpenCode Your Role

In your OpenCode session, describe your role and goals:

> "I'm a Security Engineer working on CI/CD pipeline
> security. Help me get started with Gemara."

Gemara User Journey will identify your relevant Gemara layers,
recommend tutorials, and guide you through them.

## Upstream Projects

| Project | Repository | Role |
|:--------|:-----------|:-----|
| Gemara | [gemaraproj/gemara](https://github.com/gemaraproj/gemara) | GRC schema definitions, tutorials, lexicon |
| Gemara MCP Server | [gemaraproj/gemara-mcp](https://github.com/gemaraproj/gemara-mcp) | Validation, lexicon, and schema documentation tools |

## Learn More

- [Gemara Layer Reference](docs/layer-reference.md) —
  the seven-layer model that Gemara User Journey uses for routing
- [Project Structure](docs/project-structure.md) —
  directory layout and architecture decision records
- [Keeping gemara-mcp Up to Date](docs/mcp-update-guide.md) —
  syncing and rebuilding the MCP server
- [Contributing](CONTRIBUTING.md) —
  branching, commits, review, and code standards

## License

Apache License 2.0 — see [LICENSE](LICENSE) for details.
