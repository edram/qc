# qc

[简体中文](README.zh-CN.md)

`qc` is a command-line and automation tool for querying Chinese enterprise information. It imports login cookies from a local browser, searches Qichacha (企查查) for enterprises or people, and returns stable JSON.

Current status: enterprise and people searches through Qichacha are available. The Aiqicha data source is not connected yet.

## Features

- Search enterprise information by company name
- Search people and their associated enterprises by name
- Convert responses from different data sources into one JSON model
- Import Qichacha cookies from Chrome, Brave, Edge, Firefox, or Safari
- Isolate work and personal login sessions with profiles
- View cookie metadata and the current Qichacha account without displaying cookie values

## Installation

Download the archive for your operating system and architecture from [GitHub Releases](https://github.com/edram/qc/releases), extract it, and put `qc` in your `PATH`.

Building from source requires Go 1.25 or later:

```console
git clone https://github.com/edram/qc.git
cd qc
go install ./cmd/qc
```

## Quick start

Qichacha validates the browser User-Agent associated with your login cookies. If they do not match, Qichacha may update `QCCSESSID` and log the browser out. Save the real User-Agent from the browser that is currently logged in before your first import.

1. Open developer tools in a browser that is logged in to Qichacha and run this in the Console:

   ```javascript
   navigator.userAgent
   ```

2. Save the complete User-Agent output:

   ```console
   qc config set user-agent "Mozilla/5.0 ..."
   ```

3. Import Qichacha cookies from the same browser:

   ```console
   qc auth import --browser chrome
   ```

   If the account is in another Chrome user directory, specify the browser profile:

   ```console
   qc auth import --browser chrome --browser-profile "Profile 1"
   ```

4. Check the current session:

   ```console
   qc status
   ```

5. Search for an enterprise or person:

   ```console
   qc search ents "百度" --provider qcc
   qc search pers "李彦宏" --provider qcc
   ```

Example enterprise search JSON:

```json
[
  {
    "id": "3f603703d59a04cb",
    "detailUrl": "https://www.qcc.com/firm/3f603703d59a04cb.html",
    "name": "百度在线网络技术（北京）有限公司"
  }
]
```

## Agent skill

This repository includes an agent skill at [`skills/qc/SKILL.md`](skills/qc/SKILL.md). It teaches AI coding assistants when and how to use `qc` for QCC searches, local area and industry catalogs, profile setup, authentication, and structured JSON output.

### Install the skill

Install it from GitHub with the skills CLI:

```console
npx skills@latest add edram/qc --skill qc
```

Before using the skill, install `qc` and complete the [quick start](#quick-start) to save the browser User-Agent and import Cookies.

### Use the skill

Start a new AI assistant session after installation, then describe the task in plain language. For example:

```text
Use the qc skill to find the QCC industry code for software in Shenzhen.
```

```text
Use qc to search QCC enterprises named 百度 in the active status, then return the JSON fields name, status, risk, and province aggregation.
```

## Commands

| Command | Description |
| --- | --- |
| `qc auth import` | Import Qichacha cookies from a browser |
| `qc config set user-agent <value>` | Save the browser User-Agent for the current profile |
| `qc status` | Show the current profile, cookie metadata, and account status |
| `qc search ents <query>` | Search enterprises |
| `qc search pers <query>` | Search people |
| `qc area list` | Search the local area catalog |
| `qc industry list` | Search the local industry catalog |
| `qc --help` | Show complete command help |

`--provider` can be repeated and also accepts comma-separated values:

```console
qc search ents "百度" --provider qcc
qc search ents "百度" --provider qcc --provider aiqicha
```

Aiqicha search is not connected yet, so use `--provider qcc` explicitly for now.

Area catalogs search by name and return the name and provider code:

```console
qc area list --search "深圳" --provider qcc
```

Industry catalogs search by name and return the name and provider code:

```console
qc industry list --search "软件" --provider qcc
```

### Search filters

Enterprise searches support filters for match scope, area, industry, and registration status. For example, find active construction enterprises in Beijing whose business scope contains “建筑”:

```console
qc search ents "建筑" --provider qcc \
  --match scope \
  --area 北京市 \
  --industry 建筑业 \
  --status active
```

Filter parameters can be repeated or supplied as comma-separated values:

| Parameter | Accepted values |
| --- | --- |
| `--match` | `name`, `scope`, `introduction`, `address`, `brand`, `legal-representative`, `patent`, `trademark`, `shareholder`, `key-personnel` |
| `--area` | An area name, full path, or Qichacha code such as `深圳市`, `广东省 深圳市`, or `440300`; use `qc area list --search <keyword>` to find one |
| `--industry` | An industry name such as `建筑业`; use `qc industry list --search <keyword>` to find one |
| `--status` | `active`, `moved`, `establishing`, `cancelled`, `revoked` |

Without `--match`, the search matches enterprise names only. The status values correspond to active, moved, establishing, cancelled, and revoked enterprises.

Enterprise searches return the total number of matches, the enterprises on the current page, and province and national-standard-industry aggregations. An enterprise's `tags` can include labels returned by Qichacha, such as small and micro enterprise, high-tech enterprise, and specialized and innovative small and medium-sized enterprise:

```json
{
  "total": 12451,
  "enterprises": [
    {
      "id": "...",
      "name": "百度在线网络技术（北京）有限公司",
      "tags": ["被执行人", "港股VIE", "美股VIE", "高新技术企业", "企业技术中心"],
      "risk": {
        "direct": {"count": 5},
        "associated": {"count": 27}
      }
    }
  ],
  "aggregations": {
    "provinces": [{"code": "JS", "name": "江苏省", "count": 909}],
    "industries": [{"code": "F", "name": "批发和零售业", "count": 4754}]
  }
}
```

An enterprise's `risk` is a summary of risk counts. `direct` covers risks belonging directly to the enterprise, while `associated` covers risks from related entities. `count: 0` means that the statistic was returned and confirmed to be zero. If the data source does not return a risk statistic, the `risk` field is omitted and must not be treated as zero.

People searches support area and industry name filters:

```console
qc search pers "李彦宏" --provider qcc \
  --area "广东省 深圳市" \
  --industry "软件和信息技术服务业"
```

## Profiles

The global `--profile` flag selects the qc configuration and cookie namespace. The default is `default`. You can also set the default profile with `QC_PROFILE`.

```console
qc --profile work config set user-agent "Mozilla/5.0 ..."
qc --profile work auth import --browser chrome --browser-profile "Profile 1"
qc --profile work status
qc --profile work search ents "百度" --provider qcc
```

These three options serve different purposes:

| Option | Purpose |
| --- | --- |
| `--profile work` | Select qc's `work` configuration and cookies |
| `--browser chrome` | Select the browser that provides the cookies |
| `--browser-profile "Profile 1"` | Select the browser's internal user directory |

## Configuration and precedence

User-Agent precedence is:

```text
--user-agent > config.<profile>.json > config.default.json
```

Non-default profiles inherit `config.default.json`, then override it with their own non-empty fields. `config set` changes only the current profile; it does not copy inherited values.

For example, `work` can inherit the default User-Agent or override it:

```console
qc config set user-agent "Mozilla/5.0 ..."
qc --profile work config set user-agent "Mozilla/5.0 ..."
```

You can also override the User-Agent for one command without writing it to the configuration:

```console
qc --user-agent "Mozilla/5.0 ..." status
```

The configuration directory follows the operating system's user configuration directory:

| System | Directory |
| --- | --- |
| macOS | `~/Library/Application Support/qc` |
| Linux | `~/.config/qc` |
| Windows | `%AppData%\qc` |

Typical files look like this:

```text
qc/
├── config.default.json
├── config.work.json
└── cookies/
    ├── qcc.default.json
    └── qcc.work.json
```

Configuration is separated by profile and shared by all data sources. Cookies are isolated by both data source and profile. Configuration files and cookie caches may contain sensitive information; do not commit or share them publicly.

## Session diagnostics

Run the following command to inspect the cookies and account for the current profile:

```console
qc status
```

The output includes:

- The current qc profile
- Cookie names, domains, paths, expiration times, and security attributes for each data source
- The current Qichacha account status and a masked phone number and email address
- The status of data sources that are not connected yet

`qc status` does not display cookie values. If Qichacha asks you to log in again, make sure the User-Agent comes from the same browser and browser profile used to import the cookies, then log in again and rerun `qc auth import`.

## FAQ

### Why are some Chromium cookies skipped on Windows?

Chromium v20 App-Bound Encryption on Windows binds some cookie values to the browser process. When Sweet Cookie cannot decrypt one of those values, it skips that cookie, returns a warning, and keeps other readable cookies. See [Chrome's App-Bound Encryption announcement](https://security.googleblog.com/2024/07/improving-security-of-chrome-cookies-on.html) for background.

If the cookies required by Qichacha are skipped, either manually configure those cookies in the current profile's cookie cache or import them from another supported browser that can be read successfully, such as Firefox. Keep the browser User-Agent matched to the cookies, and protect any manually configured cookie values.

## Development

```console
go run ./cmd/qc --help
go test -skip '^TestManual' ./...
go vet ./...
```

`TestManualSearchMulti` makes real requests to Qichacha with the local configuration and cookies. Run it separately only when you need to manually verify the interface:

```console
go test -run '^TestManualSearchMulti$' -v ./internal/qcc
```

## Releases

After you push a semantic version tag, GitHub Actions builds amd64 and arm64 archives for macOS, Linux, and Windows, then publishes the archives and checksum files:

```console
git tag v0.1.0
git push origin v0.1.0
```

Running the `Release` workflow manually performs a snapshot build only and does not create a GitHub Release.

## Disclaimer

All content in this repository is provided for learning and reference only and must not be used for commercial purposes. No person or organization may use the content for illegal activities or to infringe on the lawful rights and interests of others. When using this project, comply with applicable laws and third-party service terms.

## License

This project is distributed under the [MIT License](LICENSE).
