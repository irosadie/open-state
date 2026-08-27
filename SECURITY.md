# Security Policy

OpenState takes security seriously. We appreciate your efforts to responsibly
disclose vulnerabilities.

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| latest  | ✅                 |

## Reporting a Vulnerability

**Please do NOT open a public issue for security vulnerabilities.**

Instead, report privately. If a private security advisory channel is enabled on
this repository, use it. Otherwise, open a **private** issue or contact the
maintainers directly.

Please include:

- The affected version / commit
- A description of the vulnerability
- Steps to reproduce (if possible)
- Potential impact

We will acknowledge receipt, investigate, and respond as soon as possible.

## Security Best Practices for Operators

- Never store secrets (API keys, MCP credentials, tokens) in workflow
  definitions — use `credential_reference` and a secret manager (PRD §61).
- Run with TLS and encrypt data at rest and in transit (PRD §92).
- Apply least-privilege RBAC and enforce tenant isolation (PRD §80-81).
- Configure data retention and PII controls (PRD §89-90).
