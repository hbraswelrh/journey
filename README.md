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

**[Try it live](https://hbraswelrh.github.io/pacman/)** | [View source](https://github.com/hbraswelrh/pacman)

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
- [Node.js](https://nodejs.org/) 18 or later (for the
  web interface)
- [Go](https://go.dev/dl/) 1.21 or later (for data
  generation)
- [Git](https://git-scm.com/downloads)
- [OpenCode](https://opencode.ai) — the AI coding agent
  used for tutorials and artifact authoring
- [gemara-mcp](https://github.com/gemaraproj/gemara-mcp)
  server (recommended — build from source)

## Getting Started

### Step 1: Clone and Install

```bash
git clone https://github.com/hbraswelrh/gemara-user-journey.git
cd gemara-user-journey
cd web && npm install && cd ..
```

### Step 2: Launch the Web Interface

```bash
make web-dev
```

Open **http://localhost:5173/** in your browser.

### Step 3: Discover Your Role

Select a predefined role or type your own, describe your
daily activities, and the app maps you to the relevant
Gemara layers, tutorials, and artifact types.

### Step 4: Author with OpenCode

After completing the guided journey, launch OpenCode in
this project directory:

```bash
opencode
```

OpenCode reads `AGENTS.md` for context and starts the
gemara-mcp server automatically if configured in
`opencode.json`. Tell it your role and goals:

> "I'm a Security Engineer working on CI/CD pipeline
> security. Help me get started with Gemara."

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
