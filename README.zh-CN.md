# qc

[English](README.md)

`qc` 是一个面向命令行和自动化脚本的企业信息查询工具。它从本地浏览器同步登录 Cookie，通过企查查搜索企业或人员，并输出结构稳定的 JSON。

当前状态：企查查企业搜索和人员搜索可用；爱企查数据源尚未接入。

## 功能

- 按企业名称搜索企业信息
- 按姓名搜索人员及关联企业信息
- 将不同数据源的响应转换为统一 JSON 模型
- 从 Chrome、Brave、Edge、Firefox 或 Safari 同步企查查 Cookie
- 使用 profile 隔离工作、个人等不同登录会话
- 查看 Cookie 元数据和当前企查查账号，不输出 Cookie 值

## 安装

从 [GitHub Releases](https://github.com/edram/qc/releases) 下载适合当前系统和架构的压缩包，解压后将 `qc` 放入 `PATH`。

从源码构建需要 Go 1.25 或更高版本：

```console
git clone https://github.com/edram/qc.git
cd qc
go install ./cmd/qc
```

## 快速开始

企查查会校验登录 Cookie 对应的浏览器 User-Agent。两者不一致可能导致 `QCCSESSID` 更新，并使浏览器退出登录。因此，首次使用时必须保存当前登录浏览器的真实 User-Agent。

1. 在已登录企查查的浏览器中打开开发者工具，在 Console 执行：

   ```javascript
   navigator.userAgent
   ```

2. 保存输出的完整 User-Agent：

   ```console
   qc config set user-agent "Mozilla/5.0 ..."
   ```

3. 从同一个浏览器同步企查查 Cookie：

   ```console
   qc auth import --browser chrome
   ```

   如果登录账号位于 Chrome 的其他用户目录，请指定浏览器 profile：

   ```console
   qc auth import --browser chrome --browser-profile "Profile 1"
   ```

4. 检查当前会话：

   ```console
   qc status
   ```

5. 搜索企业或人员：

   ```console
   qc search ents "百度" --provider qcc
   qc search pers "李彦宏" --provider qcc
   ```

企业搜索的 JSON 结构示例：

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

本仓库包含一个供 AI 编程助手使用的 skill：[skills/qc/SKILL.md](skills/qc/SKILL.md)。它说明 AI 何时以及如何使用 `qc` 查询企查查、搜索本地地区和行业目录、配置 profile、处理登录状态，并消费结构化 JSON 输出。

### 安装 skill

使用 skills CLI 从 GitHub 安装：

```console
npx skills@latest add edram/qc --skill qc
```

使用这个 skill 前，请先安装 `qc`，再按照[快速开始](#快速开始)保存浏览器 User-Agent 并导入 Cookie。

### 使用 skill

安装完成后，新建一个 AI 助手会话，然后用自然语言描述任务。例如：

```text
使用 qc skill 查找深圳“软件”行业对应的企查查行业编码。
```

```text
使用 qc 搜索名称为百度且状态为存续的企查查企业，然后返回 name、status、risk 和省份聚合字段的 JSON。
```

## 命令

| 命令 | 说明 |
| --- | --- |
| `qc auth import` | 从浏览器同步企查查 Cookie |
| `qc config set user-agent <value>` | 保存当前 profile 的浏览器 User-Agent |
| `qc status` | 显示当前 profile、Cookie 元数据和账号状态 |
| `qc search ents <query>` | 搜索企业 |
| `qc search pers <query>` | 搜索人员 |
| `qc area list` | 搜索本地地区目录 |
| `qc industry list` | 搜索本地行业目录 |
| `qc --help` | 查看完整命令帮助 |

`--provider` 可以重复使用，也接受逗号分隔的值：

```console
qc search ents "百度" --provider qcc
qc search ents "百度" --provider qcc --provider aiqicha
```

爱企查搜索尚未接入，因此当前应显式使用 `--provider qcc`。

地区目录按名称搜索，并返回名称和对应 provider 编码：

```console
qc area list --search "深圳" --provider qcc
```

行业目录按名称搜索，并返回名称和对应 provider 编码：

```console
qc industry list --search "软件" --provider qcc
```

### 搜索筛选

企业搜索支持按查找范围、地区、行业和登记状态筛选。例如，查找北京建筑业中经营范围包含“建筑”的存续企业：

```console
qc search ents "建筑" --provider qcc \
  --match scope \
  --area 北京市 \
  --industry 建筑业 \
  --status active
```

企业筛选参数可以重复使用，也可以传入逗号分隔的多个值：

| 参数 | 可用值 |
| --- | --- |
| `--match` | `name`、`scope`、`introduction`、`address`、`brand`、`legal-representative`、`patent`、`trademark`、`shareholder`、`key-personnel` |
| `--area` | 地区名称、完整路径或企查查编码，例如 `深圳市`、`广东省 深圳市`、`440300`；可以先用 `qc area list --search <关键词>` 查找 |
| `--industry` | 行业名称，例如 `建筑业`；可以先用 `qc industry list --search <关键词>` 查找 |
| `--status` | `active`、`moved`、`establishing`、`cancelled`、`revoked` |

不传 `--match` 时默认只匹配企业名。状态值依次对应存续/在业、迁出、设立、注销和吊销。

企业搜索返回命中总量、当前页企业和省份/国标行业聚合。企业的 `tags` 包含企查查返回的小微企业、高新技术企业、专精特新中小企业等标签：

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

企业的 `risk` 是风险数量摘要：`direct` 表示直接归属于企业自身的风险，`associated` 表示关联主体产生的风险。`count: 0` 表示已获取统计且确认没有风险；如果数据源没有返回风险统计，`risk` 字段会被省略，不应按 0 处理。

人员搜索支持地区和行业名称筛选：

```console
qc search pers "李彦宏" --provider qcc \
  --area "广东省 深圳市" \
  --industry "软件和信息技术服务业"
```

## Profile

全局 `--profile` 用于选择 qc 的配置和 Cookie 命名空间，默认值为 `default`。也可以通过 `QC_PROFILE` 设置默认 profile。

```console
qc --profile work config set user-agent "Mozilla/5.0 ..."
qc --profile work auth import --browser chrome --browser-profile "Profile 1"
qc --profile work status
qc --profile work search ents "百度" --provider qcc
```

以下三个参数用途不同：

| 参数 | 作用 |
| --- | --- |
| `--profile work` | 选择 qc 的 `work` 配置和 Cookie |
| `--browser chrome` | 选择 Cookie 来源浏览器 |
| `--browser-profile "Profile 1"` | 选择浏览器内部的用户目录 |

## 配置与优先级

User-Agent 的优先级为：

```text
--user-agent > config.<profile>.json > config.default.json
```

非 default profile 会继承 `config.default.json`，再用自己的非空字段覆盖默认配置。`config set` 只修改当前 profile，不会复制继承值。

例如，`work` 可以继承默认 User-Agent，也可以单独覆盖：

```console
qc config set user-agent "Mozilla/5.0 ..."
qc --profile work config set user-agent "Mozilla/5.0 ..."
```

也可以只为单次命令覆盖 User-Agent，不写入配置：

```console
qc --user-agent "Mozilla/5.0 ..." status
```

配置目录遵循操作系统的用户配置目录：

| 系统 | 目录 |
| --- | --- |
| macOS | `~/Library/Application Support/qc` |
| Linux | `~/.config/qc` |
| Windows | `%AppData%\qc` |

典型文件如下：

```text
qc/
├── config.default.json
├── config.work.json
└── cookies/
    ├── qcc.default.json
    └── qcc.work.json
```

配置只按 profile 区分，并由所有数据源共享；Cookie 同时按数据源和 profile 隔离。配置文件和 Cookie 缓存可能包含敏感信息，不应提交到版本控制或公开分享。

## 会话诊断

运行以下命令检查当前 profile 使用的 Cookie 和账号：

```console
qc status
```

输出包含：

- 当前 qc profile
- 各数据源的 Cookie 名称、域名、路径、过期时间和安全属性
- 当前企查查账号状态及脱敏后的手机号、邮箱
- 尚未接入的数据源状态

`qc status` 不会显示 Cookie 值。如果企查查提示重新登录，请确认 User-Agent 来自导入 Cookie 的同一浏览器和同一浏览器 profile，然后重新登录并再次运行 `qc auth import`。

## 常见问题

### 为什么 Windows 上部分 Chromium Cookie 会被跳过？

Windows 上的 Chromium v20 App-Bound Encryption 会将部分 Cookie 值绑定到浏览器进程。当 Sweet Cookie 无法解密其中某个值时，会跳过该 Cookie，返回警告，并保留其他可以读取的 Cookie。背景说明请参阅 [Chrome 的 App-Bound Encryption 公告](https://security.googleblog.com/2024/07/improving-security-of-chrome-cookies-on.html)。

如果企查查所需的 Cookie 被跳过，可以手动将这些 Cookie 配置到当前 profile 的 Cookie 缓存中，或者改用其他能够成功读取 Cookie 的支持浏览器，例如 Firefox。请保持浏览器 User-Agent 与 Cookie 匹配，并妥善保护手动配置的 Cookie 值。

## 开发

```console
go run ./cmd/qc --help
go test -skip '^TestManual' ./...
go vet ./...
```

`TestManualSearchMulti` 会使用本机配置和 Cookie 真实请求企查查，只应在需要手动验证接口时单独运行：

```console
go test -run '^TestManualSearchMulti$' -v ./internal/qcc
```

## 发布

推送语义化版本标签后，GitHub Actions 会构建 macOS、Linux 和 Windows 的 amd64、arm64 版本，并发布压缩包和校验文件：

```console
git tag v0.1.0
git push origin v0.1.0
```

手动运行 `Release` workflow 只执行 snapshot 构建，不创建 GitHub Release。

## 免责声明

本仓库的所有内容仅供学习和参考之用，禁止用于商业用途。任何人或组织不得将本仓库的内容用于非法用途或侵犯他人合法权益。使用本项目时，请遵守适用法律法规和第三方服务条款。

## License

本项目使用 [MIT License](LICENSE)。
