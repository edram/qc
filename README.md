# qc

`qc` 是一个企查命令行工具。目前支持通过企查查搜索企业和人员，并将结果转换为统一模型后输出 JSON。

```console
qc search ents "百度"
qc search pers "李彦宏"
qc search ents "百度" --source qcc
qc search pers "李彦宏" --source qcc
qc search ents "百度" --source qcc --source aiqicha
```

企查查企业和人员搜索已可用；爱企查数据源仍待接入。

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
│   ├── search.go                  # search 命令和数据源接口
│   ├── search_qcc.go              # 接入企查查客户端
│   └── search_aiqicha.go          # 接入爱企查客户端
├── internal/qcc/api.go            # 企查查 HTTP 客户端
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
