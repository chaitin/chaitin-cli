# workspace-cli

[![CI](https://img.shields.io/github/actions/workflow/status/chaitin/workspace-cli/ci.yml?branch=main&label=CI)](https://github.com/chaitin/workspace-cli/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/chaitin/workspace-cli?label=Release)](https://github.com/chaitin/workspace-cli/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/chaitin/workspace-cli?label=Go)](https://github.com/chaitin/workspace-cli/blob/main/go.mod)
[![License](https://img.shields.io/github/license/chaitin/workspace-cli?label=License)](https://github.com/chaitin/workspace-cli/blob/main/LICENSE)

长亭安全产品统一命令行工具

[English README](./README.en.md)

## 演示

### CloudWalker
[![asciicast](https://asciinema.org/a/894643.svg)](https://asciinema.org/a/894643)

### T-Answer

[![asciicast](https://asciinema.org/a/Gs8KPOIcEnRRnXWr.svg)](https://asciinema.org/a/Gs8KPOIcEnRRnXWr)

### SafeLine

[![asciicast](https://asciinema.org/a/x7I8JDzbtjsRjb5M.svg)](https://asciinema.org/a/x7I8JDzbtjsRjb5M)

### SafeLine-CE

[![asciicast](https://asciinema.org/a/dzJzibRTm8arWRmU.svg)](https://asciinema.org/a/dzJzibRTm8arWRmU)

### X-Ray

[![asciicast](https://asciinema.org/a/XH6Hk9pWK0yp4VIt.svg)](https://asciinema.org/a/XH6Hk9pWK0yp4VIt)

## 配置

将各产品的连接信息写入 `./config.yaml`：

```yaml
cloudwalker:
  url: https://cloudwalker.example.com/rpc
  api_key: YOUR_API_KEY

tanswer:
  url: https://tanswer.example.com
  api_key: YOUR_API_KEY

xray:
  url: https://xray.example.com/api/v2
  api_key: YOUR_API_KEY
```

也可以把同样的配置放到环境变量或本地 `.env` 文件中。变量命名规则为 `<PRODUCT>_<FIELD>`：

```text
cloudwalker.url      -> CLOUDWALKER_URL
cloudwalker.api_key  -> CLOUDWALKER_API_KEY
tanswer.url          -> TANSWER_URL
tanswer.api_key      -> TANSWER_API_KEY
xray.url             -> XRAY_URL
xray.api_key         -> XRAY_API_KEY
safeline-ce.url      -> SAFELINE_CE_URL
safeline-ce.api_key  -> SAFELINE_CE_API_KEY
safeline.url         -> SAFELINE_URL
safeline.api_key     -> SAFELINE_API_KEY
```

`.env` 示例：

```bash
SAFELINE_URL=https://safeline.example.com
SAFELINE_API_KEY=YOUR_API_KEY
XRAY_URL=https://xray.example.com/api/v2
XRAY_API_KEY=YOUR_API_KEY
```

优先级为 `flags > environment/.env > config.yaml`

可以通过根命令的 `-c` 或 `--config` 指定其他配置文件。这在切换多个产品实例时很有用，例如多个 SafeLine 环境：

```bash
cws -c ./configs/safeline-prod.yaml safeline stats overview
cws -c ./configs/safeline-staging.yaml safeline stats overview
```

支持 dry-run 的命令可以使用根级别的 `--dry-run`：

```bash
cws --dry-run xray plan PostPlanFilter --filterPlan.limit=10
```

## 项目结构

```text
main.go                # 主入口和 CLI 装配逻辑
products/<name>/       # 每个产品一个独立目录
Taskfile.yml           # 构建、运行、检查任务
```

## 添加新产品

在 `products` 目录下新增产品实现

新增产品检查清单：

- 在 `main.go` 中导入产品包
- 在 `newApp()` 中通过 `a.registerProductCommand(...)` 注册命令
- 如果 `NewCommand()` 返回 `(*cobra.Command, error)`，需要在注册前处理错误
- 如果产品依赖 `config.yaml` 或根级运行时参数，在产品包里实现 `ApplyRuntimeConfig(...)`，并从 `main.go` 的 `wrapProductCommand()` 中调用
- 产品配置应在产品包内部从 `config.Raw` 解码，不要把产品字段解析逻辑塞进根命令

## BusyBox 风格调用

同一个二进制可以通过软链接，或者直接重命名后，以子命令名直接调用：

```bash
task build
ln -s ./bin/cws ./chaitin
./chaitin
```

等价于：

```bash
./bin/cws chaitin
```

## 常用任务

```bash
task build
task run:chaitin
task fmt
task lint
task test
task package GOOS=linux GOARCH=amd64
```
