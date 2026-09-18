---
name: qc
description: Searches Chinese enterprise records and enterprise-related people with the qc CLI, returning structured QCC data and managing its browser-backed session. Use when asked to query 企查查 or QCC; look up 企业信息、公司信息或工商信息; find a legal representative, founder, entrepreneur, executive, shareholder, or related company; or configure qc profiles, User-Agent, cookies, status, and authentication.
metadata:
  author: edram
  version: 2026.09.16
---

Task-oriented reference for the `qc` command-line client. It searches authenticated QCC data and emits provider-independent JSON for scripts and downstream analysis.

- Use `qc` instead of recreating QCC HTTP requests, signatures, or Cookie headers
- Pass `--provider qcc` explicitly; the Aiqicha provider is not implemented
- Keep `--profile` consistent across configuration, Cookie import, status, and search
- Use the exact User-Agent from the browser session that supplied the Cookies; never guess or substitute a generic value
- Do not print, inspect, or relay Cookie values; `qc status` exposes safe metadata
- Check `qc <command> --help` when the installed CLI differs from this reference

```text
qc
├── config set user-agent
├── auth import
├── status
└── search
    ├── ents
    └── pers
```

## Core

| Topic | Description | Reference |
|-------|-------------|-----------|
| Session | Install checks, User-Agent, Cookie import, profiles, status, and recovery | [core-session](references/core-session.md) |
| Search | Enterprise and person searches, filters, and JSON output | [core-search](references/core-search.md) |
