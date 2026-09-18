---
name: core-search
description: Search QCC enterprise and person records with supported filters and JSON output
---

## Usage

Searches require a configured User-Agent and imported Cookies. Use `qc status` first when the session is uncertain.

### Enterprises

Search by enterprise name:

```console
qc search ents "百度" --provider qcc
```

Apply QCC page filters:

```console
qc search ents "建筑" --provider qcc \
  --match scope \
  --area 北京市 \
  --status active
```

| Flag | Accepted values |
| --- | --- |
| `--match` | `name`, `scope`, `introduction`, `address`, `brand`, `legal-representative`, `patent`, `trademark`, `shareholder`, `key-personnel` |
| `--area` | QCC province label or code, such as `北京市` or `BJ` |
| `--status` | `active`, `moved`, `establishing`, `cancelled`, `revoked` |

Omitting `--match` searches enterprise names. The status values mean 存续/在业, 迁出, 设立, 注销, and 吊销 respectively. Enterprise filter flags accept repeated or comma-separated values:

```console
qc search ents "科技" --provider qcc \
  --area 北京市 --area 上海市 \
  --status active,cancelled
```

### People

Search by person name:

```console
qc search pers "李彦宏" --provider qcc
```

Person area filters use the labels displayed by QCC. Join hierarchy levels with spaces and quote values containing spaces:

```console
qc search pers "李彦宏" --provider qcc \
  --area "广东省 深圳市"
```

### Consume the result

Both commands write a JSON array to standard output. Enterprise objects can include:

```text
id, detailUrl, name, registrationNumber, creditCode, legalRepresentative, status,
establishedDate, address, registeredCapital, phone, email, logoUrl
```

Person objects can include:

```text
id, detailUrl, name, mainCompanyId, mainCompanyName, role, relatedCompanyCount,
partnerCount, introduction, avatarUrl
```

Optional fields are omitted when QCC does not return them. Preserve the JSON output for downstream processing rather than scraping formatted terminal text.

## Key Points

- Pass `--provider qcc` explicitly until Aiqicha search is implemented.
- Place global flags before the command: `qc --profile work search ...`.
- Filter values are current CLI enums or QCC labels; use `qc search ents --help` or `qc search pers --help` before inventing a value.
- On authentication or login-page responses, stop retrying and follow the session diagnosis workflow.

<!--
Source references:
- https://raw.githubusercontent.com/edram/qc/main/README.md
- https://raw.githubusercontent.com/edram/qc/main/internal/cli/search.go
- https://raw.githubusercontent.com/edram/qc/main/internal/cli/search_qcc.go
- https://raw.githubusercontent.com/edram/qc/main/internal/models/enterprise.go
- https://raw.githubusercontent.com/edram/qc/main/internal/models/person.go
-->
