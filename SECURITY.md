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

## Secrets Management (PRD §139)

All secrets are injected at runtime via environment variables and are never
baked into container images or committed to source. Required secrets:

- `DATABASE_URL` — PostgreSQL connection string.
- `JWT_SECRET` — JWT signing key, **at least 32 characters** (the API fails to
  start if it is missing or too short).
- `SSO_*_CLIENT_SECRET` — OIDC provider client secrets (when SSO is enabled).
- Capability provider credentials — stored as a `credential_reference` (a
  key/vault path), never as a plaintext value in the registry or API responses.

`.env*` files are ignored by git; only `.env.example` placeholders are tracked.
CI runs a secret scan (`trufflehog`) on every pull request and push to `main`.

### Rotating secrets

1. **JWT_SECRET**: set the new value in the environment, then restart the API.
   All previously issued tokens become invalid (users re-authenticate). Rotate
   on any suspected compromise.
2. **DATABASE_URL**: update the environment to point at the new credential,
   then restart the API and the worker. No code change is required.
3. **SSO client secrets**: rotate in the provider console (Google/GitHub/Entra),
   update `SSO_*_CLIENT_SECRET`, then restart the API.
4. **Capability provider credentials**: rotate the secret in the backing vault,
   update the `credential_reference` target, and re-create the binding if needed.

Secrets are read only at process start; after any rotation, restart the affected
service so the new value is picked up.
