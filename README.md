# qc

`qc` 是一个企查命令行工具。目前只搭建 CLI 框架，不包含查询实现。

```console
qc search ents "百度"
qc search pers "李彦宏"
```

## 技术栈

- Go 1.25：生成单文件可执行程序，启动快，适合跨平台分发。
- Kong：通过 Go struct 和 tag 定义多级命令，直接把位置参数映射为有类型的字段。
- Go 标准库 `testing`：现阶段只验证命令树、参数约束和退出行为，不引入额外测试依赖。

当前不引入配置、HTTP、表格渲染和日志库；等相应需求出现后再选择，避免 CLI 层提前绑定具体实现。

命令树由 Kong 的嵌套 struct 明确定义：`search` 是一级命令，`ents` 和 `pers` 是它的两个子命令，`<query>` 才是叶子命令的位置参数。

## 文件结构

```text
.
├── cmd/qc/main.go                 # 可执行程序入口，只处理进程退出
├── internal/cli/
│   ├── cli.go                     # Kong 根命令模型
│   ├── run.go                     # 解析、执行和退出码处理
│   ├── run_test.go                # 命令树和 CLI 行为测试
│   └── search.go                  # search、ents、pers 命令模型
├── go.mod
└── README.md
```

以后实现查询时，CLI 命令只负责读取参数和展示结果；企查客户端、领域模型和输出格式分别放入 `internal/qichacha`、`internal/model` 和 `internal/output`。这些目录在有实际代码前不创建。

## 开发

```console
go run ./cmd/qc --help
go test ./...
```
