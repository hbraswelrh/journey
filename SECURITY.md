# Security Policy

## Supported Versions

| Version | Supported |
|:--------|:----------|
| Latest on `main` | Yes |
| Feature branches | Development only |

## Reporting a Vulnerability

If you discover a security vulnerability in Gemara User Journey, please
report it responsibly. **Do not open a public issue.**

### How to Report

1. Email the maintainers with a description of the
   vulnerability, steps to reproduce, and any relevant logs
   or screenshots.
2. If the vulnerability involves a dependency, also report it
   to the upstream project.

### What to Expect

- **Acknowledgment**: We will acknowledge receipt of your
  report within 48 hours.
- **Assessment**: We will assess the severity and impact
  within 5 business days.
- **Resolution**: We will work on a fix and coordinate
  disclosure with you before making any public announcement.

### Scope

This security policy covers the Gemara User Journey codebase and its
direct dependencies. Vulnerabilities in the Gemara schema
project or the Gemara MCP server should be reported to their
respective maintainers.

## Security Practices

### Secret Scanning

All contributors MUST have
[Gitleaks](https://github.com/gitleaks/gitleaks) installed
and configured as a pre-commit hook. Gitleaks runs
`gitleaks dir --staged` against staged changes before each
commit, blocking commits that contain secrets.

Install Gitleaks:

```sh
brew install gitleaks
```

### Commit Signing

All commits MUST be cryptographically signed using a GPG,
SSH, or S/MIME key registered with the contributor's GitHub
account. This ensures the integrity and authenticity of all
contributions.

### Dependency Management

- Dependencies are tracked in `go.mod` and `go.sum`.
- The Gemara MCP server installation pins to SHA256 commit
  digests (not mutable tags) to prevent tag substitution
  attacks and ensure reproducible builds.
- Dependencies SHOULD be reviewed for known vulnerabilities
  before upgrading.

### Build Integrity

The `Makefile` is the single entry point for all build
operations. All builds MUST pass `make build`, `make test`,
and `make lint` before merge.
