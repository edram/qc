---
name: qc
description: Searches Chinese administrative area and industry catalogs locally, queries enterprise records with QCC or Aiqicha and enterprise-related people with QCC through the qc CLI, returns structured JSON, and manages browser-backed sessions. Use when asked to find a Chinese area or industry code; query 企查查, QCC, 爱企查, or Aiqicha; look up 企业信息、公司信息或工商信息; find a legal representative, founder, entrepreneur, executive, shareholder, or related company; or configure qc profiles, User-Agent, cookies, status, and authentication.
metadata:
  author: edram
  version: 2026.10.30
---

Task-oriented reference for the `qc` command-line client. It searches authenticated QCC and Aiqicha data and emits provider-independent JSON for scripts and downstream analysis.

- Use `qc` instead of recreating provider HTTP requests, signatures, or Cookie headers
- Select `--provider qcc` or `--provider aiqicha` explicitly; Aiqicha supports enterprise keyword searches on the first page without filters and mainland enterprise details
- Keep `--profile` consistent across configuration, Cookie import, status, and search
- For QCC, use the exact User-Agent from the browser session that supplied the Cookies; never guess or substitute a generic value. User-Agent configuration is optional for Aiqicha
- Do not print, inspect, or relay Cookie values; `qc status` exposes safe metadata
- Check `qc <command> --help` when the installed CLI differs from this reference

```text
qc
├── config set user-agent
├── auth import
├── status
├── update
├── area list
├── industry list
├── ent
│   └── view
└── search
    ├── ents
    └── pers
```

## Core

| Topic | Description | Reference |
|-------|-------------|-----------|
| Session | Install checks, User-Agent, Cookie import, profiles, status, and recovery | [core-session](references/core-session.md) |
| Search | Enterprise and person searches, filters, and JSON output | [core-search](references/core-search.md) |
