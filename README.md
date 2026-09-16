# qc

`qc` 是一个企查命令行工具。目前支持通过企查查搜索企业和人员，并将结果转换为统一模型后输出 JSON。

```console
qc auth import --browser chrome
qc --profile work auth import --browser chrome --browser-profile "Profile 1"
qc status
qc --profile work status
qc search ents "百度"
qc search pers "李彦宏"
qc --profile work search ents "百度" --source qcc
qc search ents "百度" --source qcc
qc search pers "李彦宏" --source qcc
qc search ents "百度" --source qcc --source aiqicha
```

企查查企业和人员搜索已可用；爱企查数据源仍待接入。

`--profile` 选择认证 profile，默认值为 `default`，也可通过 `QC_PROFILE` 设置。profile 决定 Cookie 缓存文件，例如 `qcc.default.json` 或 `qcc.work.json`；status、搜索和后续请求共享同一 profile。

`qc auth import` 从浏览器同步企查查 Cookie 到当前认证 profile。`--browser` 选择浏览器，默认是 Chrome；`--browser-profile` 选择浏览器内部的用户目录。它和全局 `--profile` 含义不同，例如 `qc --profile work auth import --browser chrome --browser-profile "Profile 1"` 会从 Chrome 的 `Profile 1` 读取，并写入 qc 的 `work` 缓存。

`qc status` 显示当前 profile、Cookie 清单和各数据源的当前账号。Cookie 区块列出请求使用的安全元数据；企查查账号区块通过身份接口验证当前用户，爱企查账号暂时标记为尚未接入。Cookie 值不会输出，手机号和邮箱会脱敏。

## 技术栈

- Go 1.25：生成单文件可执行程序，启动快，适合跨平台分发。
- Kong：通过 Go struct 和 tag 定义多级命令，直接把位置参数映射为有类型的字段。
- Go 标准库 `testing`：验证命令行为及服务端请求、响应映射，不引入额外测试依赖。

`--source` 是可重复的可选参数，也接受逗号分隔的值；不传时默认使用 `qcc` 和 `aiqicha`。企查查数据源使用标准库 HTTP 客户端；当前不引入表格渲染和日志库。

命令树由 Kong 的嵌套 struct 明确定义：`search` 是一级命令，`ents` 和 `pers` 是它的两个子命令，`<query>` 才是叶子命令的位置参数。

## 文件结构

```text
.
├── cmd/qc/main.go                 # 可执行程序入口，只处理进程退出
├── internal/cli/
│   ├── cli.go                     # Kong 根命令模型
│   ├── run.go                     # 解析、执行和退出码处理
│   ├── run_test.go                # 命令树和 CLI 行为测试
│   ├── auth.go                    # 从浏览器同步 Cookie
│   ├── status.go                  # 聚合并显示各数据源状态
│   ├── status_cookies.go          # 各数据源 Cookie 元数据
│   ├── status_qcc.go              # 企查查当前账号
│   ├── status_aiqicha.go          # 爱企查当前账号（待接入）
│   ├── search.go                  # search 命令和数据源接口
│   ├── search_qcc.go              # 接入企查查客户端
│   └── search_aiqicha.go          # 接入爱企查客户端
├── internal/qcc/
│   ├── api.go                     # 企查查 HTTP 客户端
│   └── auth.go                    # 安全 Cookie 元数据和账号身份
├── internal/aiqicha/api.go        # 爱企查客户端（待接入）
├── internal/models/
│   ├── enterprise.go              # 统一的企业领域模型
│   └── person.go                  # 统一的人员领域模型
├── go.mod
└── README.md
```

`qcc.Client` 和 `aiqicha.Client` 各自负责对接自己的服务端。CLI 在 `search.go` 定义抽象的 `search` 接口并装配两个实现；`search_qcc.go` 将企查查响应转换为 `models.Enterprise` 或 `models.Person`，搜索命令将模型数组输出为 JSON。

## 开发

```console
go run ./cmd/qc --help
go test ./...
```

## 发布

推送语义化版本标签后，GitHub Actions 会构建 macOS、Linux 和 Windows 的 amd64、arm64 版本，并将压缩包和校验文件发布到 GitHub Release：

```console
git tag v0.1.0
git push origin v0.1.0
```

手动运行 `Release` workflow 只执行 snapshot 构建，不会创建 GitHub Release，可用于验证发布配置。
