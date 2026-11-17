# Security Policy

## Reporting a Vulnerability

**Please do not open a public issue for security vulnerabilities.**

Report vulnerabilities privately via [GitHub Security Advisories](https://github.com/Guliveer/twitch-miner-go/security/advisories/new). This lets us discuss, patch, and coordinate disclosure before details become public.

Include:
- A description of the vulnerability and its potential impact
- Steps to reproduce or a proof-of-concept
- The version(s) affected
- Any suggested fix, if you have one

You can expect an initial response within **72 hours** and a patch within **14 days** for confirmed vulnerabilities.

## Scope

This project is a local automation tool that runs under your own Twitch account. The attack surface is limited to:

- **Config files and environment variables** — parsed locally; no remote config loading
- **Twitch API communication** — HTTPS only, no custom certificate handling
- **Notification webhooks** — outbound only, URLs are user-supplied
- **Analytics HTTP server** — local only by default (port 8080); optionally protected with HTTP basic auth via `DASHBOARD_USER` / `DASHBOARD_PASSWORD_SHA256`

Out of scope: issues that require physical access to the machine, social engineering, or rely on a compromised Twitch account.

## Supported Versions

Only the [latest release](https://github.com/Guliveer/twitch-miner-go/releases/latest) receives security fixes.
