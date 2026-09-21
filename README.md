# chaitin-cli

[![CI](https://img.shields.io/github/actions/workflow/status/chaitin/chaitin-cli/ci.yml?branch=main&label=CI)](https://github.com/chaitin/chaitin-cli/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/chaitin/chaitin-cli?label=Release)](https://github.com/chaitin/chaitin-cli/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/chaitin/chaitin-cli?label=Go)](https://github.com/chaitin/chaitin-cli/blob/main/go.mod)
[![License](https://img.shields.io/github/license/chaitin/chaitin-cli?label=License)](https://github.com/chaitin/chaitin-cli/blob/main/LICENSE)

长亭安全产品统一命令行工具

[English README](./README.en.md)

## 项目简介

`chaitin-cli` 是面向长亭安全产品的统一命令行工具，目标是在一个二进制中提供多产品的常用运维、查询和自动化能力。它解决了不同产品 API、认证方式和输出格式分散的问题，让开发者、运维人员和 AI Agent 可以用一致的方式管理 SafeLine、X-Ray、Cloud Atlas、CloudWalker、Veinmind、T-Answer、DDR 等产品。

核心能力：

- 统一入口：通过 `chaitin-cli <product> <command>` 调用不同产品能力
- 配置复用：支持配置文件、环境变量、`.env` 和命令行参数
- 自动化友好：提供 dry-run、JSON 输出和 AI Skill，便于脚本和 Agent 集成

## 快速开始

macOS / Linux 可直接运行安装脚本：

```bash
curl -fsSL https://raw.githubusercontent.com/chaitin/chaitin-cli/main/skills/chaitin-cli/scripts/install-chaitin-cli.sh | bash
```

Windows 用户请从 [GitHub Releases](https://github.com/chaitin/chaitin-cli/releases) 下载对应版本，解压 `chaitin-cli.exe` 并加入 PATH。

## Skill

本项目提供了 skill，安装后 AI Agent（Claude Code、Cursor 等）可以直接调用 `chaitin-cli` 命令管理长亭安全产品。

如果你希望 AI Agent 自动调用 `chaitin-cli`，可以安装 Skill：

```bash
npx skills add chaitin/chaitin-cli
```

安装后，向 AI Agent 描述需求即可，例如：

- "帮我查看 SafeLine 最近的攻击日志"
- "在 X-Ray 中创建一个扫描任务"
- "列出 Cloud Atlas 中待处理的漏洞"
- "列出 CloudWalker 中的漏洞事件"
- "在 CodeForce 中创建降噪任务并查看结果"

## 演示

### CloudWalker
[![asciicast](https://asciinema.org/a/894643.svg)](https://asciinema.org/a/894643)

### Veinmind
[![asciicast](https://asciinema.org/a/cTKHufj2Fewwl95j.svg)](https://asciinema.org/a/cTKHufj2Fewwl95j)

### T-Answer

[![asciicast](https://asciinema.org/a/Gs8KPOIcEnRRnXWr.svg)](https://asciinema.org/a/Gs8KPOIcEnRRnXWr)

### SafeLine

[![asciicast](https://asciinema.org/a/x7I8JDzbtjsRjb5M.svg)](https://asciinema.org/a/x7I8JDzbtjsRjb5M)

### SafeLine-CE

[![asciicast](https://asciinema.org/a/dzJzibRTm8arWRmU.svg)](https://asciinema.org/a/dzJzibRTm8arWRmU)

### DDR

[![asciicast](https://asciinema.org/a/IHulIbJ5nsy924qd.svg)](https://asciinema.org/a/t0cFuOjkLkExjREx)

### X-Ray

[![asciicast](https://asciinema.org/a/923077.svg)](https://asciinema.org/a/923077)

## 功能模块

| 模块 | 说明 |
| --- | --- |
| `chaitin` | 示例和基础命令 |
| `safeline` | SafeLine WAF 站点、策略、ACL、攻击日志和系统信息管理 |
| `safeline-3` | SafeLine-3 保护对象、策略、ACL、日志、监控、系统和网络管理 |
| `safeline-ce` | SafeLine CE 站点、规则、日志、证书和增强防护管理 |
| `xray` | X-Ray 扫描任务、资产、漏洞、报告和系统配置管理 |
| `cloudAtlas` | Cloud Atlas 资产、暴露面、风险、情报、策略和任务管理 |
| `cloudwalker` | CloudWalker CWPP 事件、资产、漏洞、防护策略和系统管理 |
| `veinmind` | CloudWalker CNAPP 容器、镜像、逃逸防护管理 |
| `tanswer` | T-Answer 流量检测、白名单和阻断规则管理 |
| `ddr` | DDR API Token 和连接配置辅助能力 |
| `apisec` | APISec API 资产、站点、应用、访问者、数据安全和风险事件管理 |
| `dsensor` | D-Sensor 谛听安全监控、探针、蜜罐、告警和威胁日志管理 |
| `codeinsight` | CodeInsight 项目、代码托管配置、扫描任务和报告导出管理 |
| `codeforce` | CodeForce 项目、项目 AI 员工、AI 开发任务、原生审计、降噪、代码包、仓库和 Git 授权配置管理 |

根命令负责配置加载、产品命令注册和 BusyBox 风格调用分发；各产品目录负责自己的命令、参数、配置解析和 API 调用逻辑。

## 配置

将各产品的连接信息写入 `./config.yaml`：

```yaml
cloudAtlas:
  url: https://cloud-atlas.example.com
  token: YOUR_TOKEN
  space_id: YOUR_SPACE_ID

cloudwalker:
  url: https://cloudwalker.example.com/rpc
  api_key: YOUR_API_KEY

tanswer:
  url: https://tanswer.example.com
  api_key: YOUR_API_KEY

# chaitin-cli ddr get-api-token --url https://ddr.example.com:8443 --jwt-token "YOUR_JWT_TOKEN" 可以直接获取 url & api_key & company_id
ddr:
  url: "https://ddr.example.com:8443/qzh/api/v1"
  api_key: "YOUR_API_KEY"
  company_id: "YOUR_COMPANY_ID"

veinmind:
  url: "https://veinmind.example.com"
  api_key: "YOUR_64_CHARACTER_API_TOKEN"

xray:
  url: https://xray.example.com/api/v2
  api_key: YOUR_API_KEY

apisec:
  url: https://apisec.example.com
  api_token: YOUR_API_TOKEN

safeline-3:
  url: https://safeline3.example.com
  api_token: YOUR_API_TOKEN

dsensor:
  url: https://dsensor.example.com
  api_key: YOUR_API_KEY

codeinsight:
  url: https://codeinsight.example.com
  access_token: YOUR_ACCESS_TOKEN

codeforce:
  url: https://codeforce.example.com
  access_token: YOUR_ACCESS_TOKEN
  account_type: admin
```
也可以把同样的配置放到环境变量或本地 `.env` 文件中。变量命名规则为 `<PRODUCT>_<FIELD>`：

```text
cloudAtlas.url       -> CLOUD_ATLAS_URL
cloudAtlas.token     -> CLOUD_ATLAS_TOKEN
cloudAtlas.space_id  -> CLOUD_ATLAS_SPACE_ID
cloudwalker.url      -> CLOUDWALKER_URL
cloudwalker.api_key  -> CLOUDWALKER_API_KEY
tanswer.url          -> TANSWER_URL
tanswer.api_key      -> TANSWER_API_KEY
ddr.url              -> DDR_URL
ddr.api_key          -> DDR_API_KEY
ddr.company_id       -> DDR_COMPANY_ID
veinmind.url         -> VEINMIND_URL
veinmind.api_key     -> VEINMIND_API_KEY
xray.url             -> XRAY_URL
xray.api_key         -> XRAY_API_KEY
apisec.url           -> APISEC_URL
apisec.api_token     -> APISEC_API_TOKEN
dsensor.url          -> DSENSOR_URL
dsensor.api_key      -> DSENSOR_API_KEY
codeinsight.url      -> CODEINSIGHT_URL
codeinsight.access_token -> CODEINSIGHT_ACCESS_TOKEN 或 CODEINSIGHT_TOKEN
codeforce.url        -> CODEFORCE_URL
codeforce.access_token -> CODEFORCE_ACCESS_TOKEN 或 CODEFORCE_API_KEY
codeforce.account_type -> CODEFORCE_ACCOUNT_TYPE
safeline-ce.url      -> SAFELINE_CE_URL
safeline-ce.api_key  -> SAFELINE_CE_API_KEY
safeline-3.url       -> SAFELINE_3_URL
safeline-3.api_token -> SAFELINE_3_API_TOKEN
safeline.url         -> SAFELINE_URL
safeline.api_key     -> SAFELINE_API_KEY
```

### Cloud Atlas

Cloud Atlas 命令由内置 OpenAPI Schema 生成，认证使用 `Token` 请求头。配置 `cloudAtlas.space_id` 后，查询命令会自动把它作为默认空间 ID；也可以通过根命令参数 `--space-id` 指定默认空间，或在具体子命令中用 `--space` 覆盖。

```bash
export CLOUD_ATLAS_URL=https://cloud-atlas.example.com
export CLOUD_ATLAS_TOKEN=YOUR_TOKEN
export CLOUD_ATLAS_SPACE_ID=YOUR_SPACE_ID

chaitin-cli cloudAtlas seed enterprise list --page 1 --size 20 --output json
chaitin-cli cloudAtlas asset ip list --status valid --page 1 --size 20 --output json
chaitin-cli cloudAtlas exposure website list --status valid --page 1 --size 20 --output json
chaitin-cli cloudAtlas risk vulnerability list --page 1 --size 20 --output json
chaitin-cli cloudAtlas task asset schedule list --page 1 --size 20 --output json
```

常用命令分组包括 `seed`、`asset`、`exposure`、`risk`、`intelligence`、`task`、`space`、`system`、`user` 和 `strategy`。批量修改类命令（如 `set-status`、`batch-delete`）需要显式加 `--yes` 才会执行。完整参数以 `chaitin-cli cloudAtlas --help` 及对应子命令的 `--help` 为准。

APISec 常用查询不需要手动传内部 `scope`，优先使用语义化命令：

```bash
# 查询站点资产
chaitin-cli apisec asset site list --query count=100 --query offset=0 --output json

# 查询 API 资产
chaitin-cli apisec asset api list --query count=100 --query offset=0 --output json

# 查询风险事件
chaitin-cli apisec risk event list --query count=20 --query offset=0 --output json
```

`apisec raw` 保留为高级入口，用于调用生成出的底层 API 操作；日常查询建议先运行 `chaitin-cli apisec --help` 或对应语义命令的 `--help`。

### SafeLine-3

SafeLine-3 命令使用 OpenAPI Token，请在配置中填写 `safeline-3.api_token`，或设置 `SAFELINE_3_API_TOKEN`。

AI Agent 的 SafeLine-3 调用策略见 [`products/safeline3/agent-skill.md`](products/safeline3/agent-skill.md)。

```bash
chaitin-cli safeline-3 node-group list --output json
chaitin-cli safeline-3 site list --type reverse-proxy --page 1 --page-size 20 --output json
chaitin-cli safeline-3 policy-group list --page 1 --page-size 20 --output json
chaitin-cli safeline-3 log attack list --start -24h --page 1 --page-size 20 --output json
chaitin-cli safeline-3 raw request GET /api/v3/license
```

创建、更新、删除等复杂请求优先使用实体命令的语义参数；复杂嵌套结构可使用对应的 `--payload-file`、`--application-file` 等文件入口。`raw request` 是兜底入口，可调用未封装的 `/api/v3/...` 接口。

### SafeLine 企业版 AI 站点操作

SafeLine 企业版命令支持面向 AI/AISOC 调度的环境检查、证书查询/上传、站点创建预览、站点创建和回退删除。

首版站点创建支持的部署模式：

- `Software Reverse Proxy`
- `Software Cluster Reverse Proxy`

推荐调度流程：

```bash
chaitin-cli safeline inspect --indent
chaitin-cli safeline site create capabilities --indent
chaitin-cli safeline cert list --indent
chaitin-cli safeline site create --check --name app-a --domain app.example.com --port 443 --ssl --cert-id 12 --upstream http://10.0.0.1:8080 --policy-group 3 --indent
chaitin-cli safeline site create --yes --name app-a --domain app.example.com --port 443 --ssl --cert-id 12 --upstream http://10.0.0.1:8080 --policy-group 3 --indent
chaitin-cli safeline site delete 123 --yes --indent
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
chaitin-cli -c ./configs/safeline-prod.yaml safeline stats overview
chaitin-cli -c ./configs/safeline-staging.yaml safeline stats overview
```

支持 dry-run 的命令可以使用根级别的 `--dry-run`：

```bash
chaitin-cli --dry-run xray plan PostPlanFilter --filterPlan.limit=10
```

### CodeInsight 项目与任务

```bash
export CODEINSIGHT_URL=https://codeinsight.example.com
export CODEINSIGHT_TOKEN=YOUR_ACCESS_TOKEN

chaitin-cli codeinsight project create --name demo-java --language java
chaitin-cli codeinsight repo-config create --name git-prod --repo-type git --git-provider gitlab --server-host https://git.example.com/group/demo.git --auth-type access_token --access-token GIT_TOKEN
chaitin-cli codeinsight task create repo --project-name demo-java --task-name demo-repo --rule-set-name Corax-Java --repo-config-name git-prod --ref-type branch --ref-name main
chaitin-cli codeinsight task result --task-id 12345
chaitin-cli codeinsight task result download --task-id 12345 --out ./reports/12345.json
```

### CodeForce 项目与任务

```bash
export CODEFORCE_URL=https://codeforce.example.com
export CODEFORCE_ACCESS_TOKEN=YOUR_ACCESS_TOKEN
export CODEFORCE_ACCOUNT_TYPE=admin

chaitin-cli codeforce project create --name demo-app --repository-id repo-1
chaitin-cli codeforce project ai-employee model-options --project-id project-1
chaitin-cli codeforce project ai-employee create --project-id project-1 --type dev --name backend-dev-agent --enabled
chaitin-cli codeforce project ai-dev create --project-id project-1 --employee-id employee-1 --title "Add repository health dashboard" --issue-url https://github.com/example/demo/issues/12 --branch feature/repo-health
chaitin-cli codeforce audit native create git --repository-id repo-1 --source-ref branch:main --audit-rule-id 101 --task-name "main native audit"
chaitin-cli codeforce code-management create --name demo-code-drop --version-description "2026-06 release" --file ./artifacts/demo.zip
chaitin-cli codeforce repository create project --name demo-git --platform gitlab --repositories-url https://git.example.com/group/demo.git --token GIT_TOKEN
chaitin-cli codeforce git-auth create --name github-personal --platform github --token GIT_TOKEN
chaitin-cli codeforce denoise parse --type sast --report-file ./reports/sast.json
chaitin-cli codeforce denoise create --type sast --name repo-noise-check --engineer-id engineer-1 --source-type repository --repository-name demo-repo --branch-or-tag main --report-file ./reports/sast.json
chaitin-cli codeforce denoise result --task-id denoise-task-1
```

说明：

- `account_type=admin|user` 走管理接口，适用于项目创建、AI 员工、AI 开发任务、原生审计、代码管理、项目仓库和 git-auth。
- `account_type=openapi` 走 CodeForce 对外 OpenAPI，适用于 `openapi whoami`、`denoise parse`、OpenAPI 方式的 `denoise create` 和部分结果查询。
- 如果当前令牌实际上是 OpenAPI key，管理接口命令会明确提示切换凭证类型，而不会伪造成功。

## 项目结构

```text
main.go                         # 主入口、根命令、产品注册和 BusyBox 风格调用
config/                         # 配置加载、环境变量/.env 覆盖和配置写入逻辑
products/<name>/                # 每个产品一个独立目录，包含命令和 API 客户端
products/<name>/cmd/            # 手写产品命令分组
products/<name>/client/         # 生成或封装的产品 API 客户端
skills/chaitin-cli/             # AI Agent Skill 和自动安装脚本
cmd/gen-cli/                    # CLI 生成工具
Taskfile.yml                    # 构建、运行、检查和打包任务
.github/workflows/ci.yml        # CI、跨平台打包和 GitHub Release 发布流程
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
ln -s ./bin/chaitin-cli ./chaitin
./chaitin
```

等价于：

```bash
./bin/chaitin-cli chaitin
```

## 开发

仓库地址：

```bash
git clone https://github.com/chaitin/chaitin-cli.git
cd chaitin-cli
```

环境准备：

- Go 版本以 [`go.mod`](./go.mod) 为准
- 安装 [Task](https://taskfile.dev/) 后可使用 `task` 命令

本地运行：

```bash
go run . chaitin
go run . safeline --help
```

常用任务:

```bash
task build
task run:chaitin
task fmt
task lint
task test
task package GOOS=linux GOARCH=amd64
```

## 维护与反馈

- 主线分支：[`main`](https://github.com/chaitin/chaitin-cli/tree/main)
- 版本发布：通过 [`GitHub Releases`](https://github.com/chaitin/chaitin-cli/releases) 提供跨平台安装包
- 发布流程：推送 `v*` tag 后由 GitHub Actions 自动测试、打包并创建 Release
- 维护状态：当前持续维护，欢迎通过 PR 贡献产品命令、文档和 bugfix

需求建议和 Bug 反馈请通过 [GitHub Issues](https://github.com/chaitin/chaitin-cli/issues) 提交。提交前请先搜索已有 Issue，避免重复反馈。请勿在公开 Issue 中粘贴 API Key、Token、真实业务地址或其他敏感信息。

## FAQ

**安装后找不到 `chaitin-cli` 怎么办？**

确认安装目录是否在 `PATH` 中。macOS / Linux 安装脚本会优先安装到用户或系统 PATH 目录；如果当前 shell 未刷新，可以重新打开终端或使用脚本输出的完整路径运行。

**配置从哪里读取？**

优先级为 `flags > environment/.env > config.yaml`。根命令的 `-c` / `--config` 可以指定其他配置文件。

**自签名证书连接失败怎么办？**

部分产品命令提供 `--insecure` 参数，SafeLine 默认跳过 TLS 证书校验；不同产品行为可能不同，可先运行对应产品命令的 `--help` 查看支持的参数。

**如何发布新版本？**

在 `main` 上创建并推送 `v*` tag，例如 `v2605.0.0`，GitHub Actions 会自动构建跨平台包并发布到 Releases。
