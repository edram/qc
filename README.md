# qc

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
   qc search ents "百度" --source qcc
   qc search pers "李彦宏" --source qcc
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

## 命令

| 命令 | 说明 |
| --- | --- |
| `qc auth import` | 从浏览器同步企查查 Cookie |
| `qc config set user-agent <value>` | 保存当前 profile 的浏览器 User-Agent |
| `qc status` | 显示当前 profile、Cookie 元数据和账号状态 |
| `qc search ents <query>` | 搜索企业 |
| `qc search pers <query>` | 搜索人员 |
| `qc --help` | 查看完整命令帮助 |

`--source` 可以重复使用，也接受逗号分隔的值：

```console
qc search ents "百度" --source qcc
qc search ents "百度" --source qcc --source aiqicha
```

爱企查尚未接入，因此当前应显式使用 `--source qcc`。

### 搜索筛选

企业搜索支持按查找范围、省份、国标行业和登记状态筛选。例如，查找北京建筑业中经营范围包含“建筑”的存续企业：

```console
qc search ents "建筑" --source qcc \
  --match scope \
  --area 北京市 \
  --industry 建筑业 \
  --status active
```

企业筛选参数可以重复使用，也可以传入逗号分隔的多个值：

| 参数 | 可用值 |
| --- | --- |
| `--match` | `name`、`scope`、`introduction`、`address`、`brand`、`legal-representative`、`patent`、`trademark`、`shareholder`、`key-personnel` |
| `--area` | 企查查页面显示的省份名称或省份代码，例如 `北京市`、`BJ` |
| `--industry` | 国标行业门类名称或代码，例如 `建筑业`、`E` |
| `--status` | `active`、`moved`、`establishing`、`cancelled`、`revoked` |

不传 `--match` 时默认只匹配企业名。状态值依次对应存续/在业、迁出、设立、注销和吊销。

人员搜索支持页面中的省份地区和国标行业筛选。层级名称之间使用空格连接：

```console
qc search pers "李彦宏" --source qcc \
  --area "广东省 深圳市" \
  --industry "信息传输、软件和信息技术服务业"
```

## Profile

全局 `--profile` 用于选择 qc 的配置和 Cookie 命名空间，默认值为 `default`。也可以通过 `QC_PROFILE` 设置默认 profile。

```console
qc --profile work config set user-agent "Mozilla/5.0 ..."
qc --profile work auth import --browser chrome --browser-profile "Profile 1"
qc --profile work status
qc --profile work search ents "百度" --source qcc
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

## License

本项目使用 [MIT License](LICENSE)。
