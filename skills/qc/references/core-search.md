---
name: core-search
description: Search QCC enterprise and person records with supported filters and JSON output
---

## Usage

Enterprise and person searches require a configured User-Agent and imported Cookies. Use `qc status` first when the session is uncertain. The local area and industry catalogs do not require authentication.

### Local catalogs

Find a supported area name and provider code:

```console
qc area list --search "深圳" --provider qcc
```

Find a supported industry name and provider code:

```console
qc industry list --search "软件" --provider qcc
```

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
  --industry 建筑业 \
  --status active
```

| Flag | Accepted values |
| --- | --- |
| `--match` | `name`, `scope`, `introduction`, `address`, `brand`, `legal-representative`, `patent`, `trademark`, `shareholder`, `key-personnel` |
| `--area` | Area name, full path, or QCC code; use `qc area list --search <query>` to find values |
| `--industry` | Industry name, such as `建筑业`; use `qc industry list --search <query>` to find names |
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

Person searches accept area and industry names:

```console
qc search pers "李彦宏" --provider qcc \
  --area "广东省 深圳市" \
  --industry "软件和信息技术服务业"
```

### Consume the result

Enterprise searches write an object containing the full match count, the current page, and supported aggregations:

```json
{
  "total": 12451,
  "enterprises": [
    {
      "id": "...",
      "name": "百度在线网络技术（北京）有限公司",
      "tags": ["被执行人", "港股VIE", "美股VIE", "高新技术企业", "企业技术中心"]
    }
  ],
  "aggregations": {
    "provinces": [{"code": "JS", "name": "江苏省", "count": 909}],
    "industries": [{"code": "F", "name": "批发和零售业", "count": 4754}]
  }
}
```

Enterprise objects can include:

```text
id, detailUrl, name, registrationNumber, creditCode, legalRepresentative, status,
establishedDate, address, registeredCapital, phone, email, logoUrl, tags, risk
```

When present, `risk` contains counts only:

```json
{
  "risk": {
    "direct": {"count": 5},
    "associated": {"count": 27}
  }
}
```

`direct` counts risks belonging directly to the enterprise. `associated` counts risks belonging to associated entities. A returned `count` of `0` means the provider confirmed that there are no risks in that scope. If the provider does not return risk statistics, `risk` is omitted rather than treated as zero. Risk details require a separate query.

Person searches continue to write a JSON array. Person objects can include:

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
- https://raw.githubusercontent.com/edram/qc/main/internal/cli/area.go
- https://raw.githubusercontent.com/edram/qc/main/internal/cli/industry.go
- https://raw.githubusercontent.com/edram/qc/main/internal/cli/search.go
- https://raw.githubusercontent.com/edram/qc/main/internal/cli/search_qcc.go
- https://raw.githubusercontent.com/edram/qc/main/internal/models/enterprise.go
- https://raw.githubusercontent.com/edram/qc/main/internal/models/person.go
-->
