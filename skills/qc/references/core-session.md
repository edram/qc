---
name: core-session
description: Prepare and diagnose the profile-scoped browser identity used by qc
---

## Usage

Verify the installed command before using it:

```console
qc --version
qc --help
```

If `qc` is not installed, download the appropriate binary from [GitHub Releases](https://github.com/edram/qc/releases) and place it in `PATH`.

### First-time setup

QCC binds authenticated Cookies to the browser User-Agent. Obtain the exact value from the same logged-in browser by evaluating this in its developer console:

```javascript
navigator.userAgent
```

Persist that value, then import Cookies from the same browser:

```console
qc config set user-agent "Mozilla/5.0 ..."
qc auth import --browser chrome
qc status
```

Supported browser values are `chrome`, `brave`, `edge`, `firefox`, and `safari`. Chrome is the default. Select a non-default browser profile explicitly:

```console
qc auth import --browser chrome --browser-profile "Profile 1"
```

Do not infer the User-Agent from the operating system, `curl`, or an HTTP library. A mismatch can rotate `QCCSESSID` and sign the browser out.

### qc profiles

The global `--profile` flag selects both configuration and cached Cookies. Put it before the command and use the same value throughout a workflow:

```console
qc --profile work config set user-agent "Mozilla/5.0 ..."
qc --profile work auth import --browser chrome --browser-profile "Profile 1"
qc --profile work status
qc --profile work search ents "百度" --source qcc
```

`--profile work` selects qc's namespace; `--browser-profile "Profile 1"` selects the browser's user directory. They are independent. `QC_PROFILE` can provide the default qc profile.

A one-command User-Agent override does not modify configuration:

```console
qc --user-agent "Mozilla/5.0 ..." status
```

User-Agent precedence is:

```text
--user-agent > config.<profile>.json > config.default.json
```

Non-default profiles inherit non-empty settings from `default` and then apply their own values.

### Diagnose a session

Run:

```console
qc --profile work status
```

Status reports the active profile, Cookie names and attributes, and the authenticated QCC account. It masks account identifiers and never displays Cookie values.

| Symptom | Action |
| --- | --- |
| User-Agent required | Save the exact User-Agent from the logged-in browser, or pass `--user-agent` once |
| No Cookies | Confirm the browser is logged in, then repeat `auth import` with the correct browser and browser profile |
| Not authenticated | Re-login in the browser, confirm its User-Agent, import again, and rerun `status` |
| Browser becomes logged out after a request | Treat it as a User-Agent mismatch; re-login before capturing the User-Agent and importing fresh Cookies |
| Wrong QCC account | Check both the qc `--profile` and browser `--browser-profile` selections |

Cookie import replaces the cache only after the browser read succeeds. A failed browser read leaves the existing cache intact.

## Key Points

- Configuration is profile-scoped and shared by data sources; Cookie caches are scoped by both data source and profile.
- Do not expose files under the qc configuration directory: Cookie caches contain sensitive values even though `qc status` does not show them.
- Diagnose authentication with `qc status` before retrying a failed search or falling back to raw HTTP.

<!--
Source references:
- https://raw.githubusercontent.com/edram/qc/main/README.md
- https://raw.githubusercontent.com/edram/qc/main/internal/cli/auth.go
- https://raw.githubusercontent.com/edram/qc/main/internal/cli/config.go
- https://raw.githubusercontent.com/edram/qc/main/internal/cli/status.go
- https://raw.githubusercontent.com/edram/qc/main/internal/config/config.go
- https://raw.githubusercontent.com/edram/qc/main/internal/qcc/api.go
-->
