# qc

`qc` 是一个企查命令行工具。目前只搭建 CLI 框架，不包含查询实现。

```console
qc search ents "百度"
qc search pers "李彦宏"
qc search ents "百度" --source qcc
qc search ents "百度" --source qcc --source aiqicha
```

## 技术栈

- Go 1.25：生成单文件可执行程序，启动快，适合跨平台分发。
- Kong：通过 Go struct 和 tag 定义多级命令，直接把位置参数映射为有类型的字段。
- Go 标准库 `testing`：现阶段只验证命令树、参数约束和退出行为，不引入额外测试依赖。

`--source` 是可重复的可选参数，也接受逗号分隔的值；不传时默认使用 `qcc` 和 `aiqicha`。当前不引入配置、HTTP、表格渲染和日志库；等相应需求出现后再选择，避免 CLI 层提前绑定具体实现。

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
├── internal/qcc/api.go            # 企查查客户端（待接入）
├── internal/aiqicha/api.go        # 爱企查客户端（待接入）
├── internal/models/
│   ├── enterprise.go              # 统一的企业领域模型
│   └── person.go                  # 统一的人员领域模型
├── go.mod
└── README.md
```

`qcc.Client` 和 `aiqicha.Client` 各自负责对接自己的服务端。CLI 在 `search.go` 定义抽象的 `search` 接口并装配两个实现；`search_qcc.go` 和 `search_aiqicha.go` 分别实现该接口，再把调用转发给底层 Client。输出格式放入 `internal/output`，等实际需要时再创建。

## 开发

```console
go run ./cmd/qc --help
go test ./...
```
