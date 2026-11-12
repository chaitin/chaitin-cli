# 全悉 CLI（`tanswer`）

`chaitin-cli tanswer` 是全悉面向人类操作者和 AI Agent 的命令行入口。它把高频安全运营动作封装为语义命令，并提供 `api <METHOD> <PATH>` 通用调用入口，用于访问用户已知且已授权的全悉 Open API。

## 文档定位

本文档提供人工可读的入门、配置、权限、故障排查和常见操作说明。它不替代运行时命令契约：安装后的 `chaitin-cli tanswer --help`、领域/具体命令的 `--help` 与 `chaitin-cli tanswer manifest` 才是当前版本的命令、参数、输出和确认要求事实来源。

AI Agent 必须先用运行时 help 发现命令，需要机器可读完整契约时读取 manifest；不得根据本文档猜测命令、RPC 方法、路径或请求体。统一 AI 使用规则见 [`skills/chaitin-cli/SKILL.md`](../../skills/chaitin-cli/SKILL.md)。

## 为什么使用全悉 CLI？

- Agent 原生：命令按安全运营任务命名，输出稳定 JSON。
- 读写边界清晰：查询命令直接执行，写操作必须 preview 和 confirm。
- 覆盖核心安全运营场景：系统状态、威胁告警、文件告警、流量元数据、资产配置、安全策略和响应处置。
- 行为可预期：命令、参数、输出字段和变更确认要求均可通过 manifest 查询。

## 快速开始

先按仓库根目录 [README](../../README.md) 安装 `chaitin-cli`，或从 [GitHub Releases](https://github.com/chaitin/chaitin-cli/releases) 下载对应平台的二进制。

```bash
export TANSWER_URL='https://<全悉 Web 端 IP>'
export TANSWER_API_KEY='<全悉 OpenAPI Token>'

chaitin-cli tanswer auth status
chaitin-cli tanswer auth check
chaitin-cli tanswer system status
chaitin-cli tanswer alarm overview --time today
```

命令默认输出 JSON。若目标环境使用自签或非标准证书，可设置：

```bash
chaitin-cli tanswer --insecure auth check
```

## 最小可用性验证

首次接入或发布新版本后，建议按下面顺序做 smoke test，确认二进制版本、连接配置、Token 权限和核心只读能力可用：

```bash
chaitin-cli tanswer manifest
chaitin-cli tanswer auth check
chaitin-cli tanswer system status
chaitin-cli tanswer alarm overview --time today
chaitin-cli tanswer asset list --page-size 5
chaitin-cli tanswer response block-policies --page-size 5
```

如果 `manifest` 不存在，通常表示当前安装的 `chaitin-cli` 版本不包含全悉语义 CLI。请先升级到包含 `tanswer` 语义命令的 Release。

## 快速开始：AI Agent

Agent 应先读取命令清单，再按用户意图选择语义命令：

```bash
chaitin-cli tanswer manifest
chaitin-cli tanswer alarm high-priority --time 24h
chaitin-cli tanswer asset list --page-size 10
chaitin-cli tanswer response block-policies --page-size 10
```

当语义命令已覆盖目标任务时，不要优先使用 `tanswer api`。只有目标能力已开放 Open API、且当前版本没有对应语义命令时，才使用通用调用入口：

```bash
chaitin-cli tanswer api POST /rpc --body '{"jsonrpc":"2.0","method":"OpsService.GetBaseInfo","params":{},"id":"1"}'
chaitin-cli tanswer api GET /api/example --query '{"count":10,"offset":0}'
chaitin-cli tanswer api POST /rpc --body @./request.json
```

`api <METHOD> <PATH>` 输出 `status_code` 和 `raw` 原始响应；它不提供专属参数解释或业务摘要。

## 配置和认证

连接信息可通过 `--url`、`--api-key`、`--timeout`、`--insecure`；环境变量 `TANSWER_URL`、`TANSWER_API_KEY`、`TANSWER_TIMEOUT`、`TANSWER_INSECURE`；或配置文件中的 `tanswer.url`、`tanswer.api_key`、`tanswer.timeout`、`tanswer.insecure` 提供。优先级为 flags > environment/.env > 识别到的 `./config.yaml` > `~/.chaitin-cli/config.yaml`。


`chaitin-cli tanswer` 支持通过配置文件、环境变量和全局 flag 配置连接信息。

### 获取 OpenAPI Token

在全悉 Web 控制台中生成 OpenAPI Token：

1. 登录全悉 Web 控制台。
2. 进入 `系统管理` -> `Open API`。
3. 点击 `新增 API Token`，选择用于 Open API 调用的 Token 类型。
4. 保存生成出的 Token，并配置到 `TANSWER_API_KEY` 或 `tanswer.api_key`。

API Token 生成后不会持续明文展示，请妥善保存。不要把真实 Token 写入脚本、日志、共享文档、commit message 或 MR 描述。

| 配置 | 说明 |
| --- | --- |
| `tanswer.url` / `TANSWER_URL` | 全悉控制台地址，例如 `https://<全悉 Web 端 IP>`。 |
| `tanswer.api_key` / `TANSWER_API_KEY` | OpenAPI Token。不要在脚本、日志或共享文档中暴露真实值。 |
| `--timeout` | 请求超时时间，默认 `30s`。 |
| `--insecure` | 是否跳过 TLS 证书校验，默认 `false`。 |

常用 flag：

```bash
chaitin-cli tanswer --url 'https://<全悉 Web 端 IP>' --api-key '<全悉 OpenAPI Token>' auth check
chaitin-cli tanswer --insecure auth check
```

### 权限要求

CLI 调用继承全悉现有 OpenAPI Token 的权限、有效期、频率限制和 IP 访问策略。建议按实际自动化范围分配最小权限：

| 使用范围 | 需要的能力 |
| --- | --- |
| 只读巡检 | 系统状态、威胁告警、文件告警、资产、元数据、策略和响应记录查询。 |
| 资产维护 | 资产创建、更新、删除、导入、导出、批量维护和资产组维护。 |
| 策略维护 | 检测白名单、自定义 IOC 情报的创建、更新、启停、删除、导入和导出。 |
| 响应处置 | 阻断策略、响应白名单、告警生成处置对象、联动设备配置和自动响应记录查询。 |

如果查询命令可用但写操作失败，优先检查 Token 绑定角色是否包含对应写权限；如果所有命令都失败，先运行 `chaitin-cli tanswer auth check` 排查地址、Token、证书、有效期、IP 策略和频率限制。

### 版本和兼容性

客户端需要使用包含全悉语义命令的 `chaitin-cli` Release。服务端需要开放全悉 OpenAPI TokenAuth 和对应 RPC/Open API 能力。不同全悉版本开放的后端接口可能不同；以 `chaitin-cli tanswer manifest`、本文件和 [COMMAND_REFERENCE.md](./COMMAND_REFERENCE.md) 记录的命令为准。

## 命令层级

| 层级 | 命令形态 | 用途 |
| --- | --- | --- |
| Foundation | `chaitin-cli tanswer auth status` | 查看当前目标环境、Token 是否已配置、请求超时和 TLS 跳过状态。 |
| Foundation | `chaitin-cli tanswer auth check` | 校验 OpenAPI Token 能否访问当前全悉环境。 |
| Manifest | `chaitin-cli tanswer manifest` | 输出 AI 可读命令清单、风险等级、字段和确认要求。 |
| Semantic shortcut | `chaitin-cli tanswer alarm overview` | 用稳定业务字段回答安全运营问题。 |
| Protected write | `chaitin-cli tanswer asset create --preview` | 先返回写操作预览，不直接修改产品状态。 |
| Open API 通用调用 | `chaitin-cli tanswer api <METHOD> <PATH>` | 访问用户已知且已授权的全悉 Open API。 |

语义命令按领域组织；Agent 应先通过运行时 `--help` 发现命令，并通过 `chaitin-cli tanswer manifest` 区分语义命令、风险等级和 Open API 通用调用入口。根级 `--dry-run` 不适用于 `tanswer`。

## 常用场景

```bash
chaitin-cli tanswer system status
chaitin-cli tanswer alarm overview --time 24h --severity critical,high
chaitin-cli tanswer alarm detail --id '<doc_id>'
chaitin-cli tanswer file-alarm malicious --time 7d --page-size 10
```

```bash
chaitin-cli tanswer asset list --page-size 10
chaitin-cli tanswer asset group-tree --depth 2
chaitin-cli tanswer metadata protocol --protocol http --time today
chaitin-cli tanswer policy detection-whitelist --page-size 10
```

```bash
chaitin-cli tanswer response block-policies --page-size 10
chaitin-cli tanswer response block-records --time 24h --page-size 10
chaitin-cli tanswer asset create --name core-db --ip 192.0.2.10 --preview
chaitin-cli tanswer asset create --name core-db --ip 192.0.2.10 --confirm CONFIRM_ASSET_CREATE
```

完整命令表、参数和长示例见 [COMMAND_REFERENCE.md](./COMMAND_REFERENCE.md)。

## JSON 输出契约

成功输出：

```json
{
  "success": true,
  "task": "查看威胁告警概览",
  "command": "chaitin-cli tanswer alarm overview",
  "query": {},
  "data": {},
  "warnings": []
}
```

失败输出：

```json
{
  "success": false,
  "task": "查看威胁告警概览",
  "command": "chaitin-cli tanswer alarm overview",
  "error": {
    "code": "ALARM_OVERVIEW_FAILED",
    "message": "error detail",
    "retryable": true
  }
}
```

文件类命令返回文件路径、大小、格式和导入/导出摘要，不在日志中输出 Token。

## 受保护写操作

高影响写命令不能静默执行。首次调用应使用 `--preview`，输出必须包含 `requires_confirmation=true`、`confirmed=false`、`operation_type`、`target`、`change_summary`、`impact`、`risk_warnings` 和命令专属 `confirmation_token`。

确认执行时必须重新运行命令并传入完全一致的 `--confirm` 值，例如：

```bash
chaitin-cli tanswer response block-policy-create --name block-bad-ip --object 192.0.2.30 --duration 3600 --preview
chaitin-cli tanswer response block-policy-create --name block-bad-ip --object 192.0.2.30 --duration 3600 --confirm CONFIRM_RESPONSE_BLOCK_POLICY_CREATE
```

确认大小写敏感。执行结果必须包含 `confirmed=true`、`result`、`object` 和 `audit` 字段。

## 安全边界

- 运行时 help 与 manifest 是命令事实来源；本文档和命令索引不替代它们。
- 语义写操作必须先 `--preview`，展示目标、影响和风险，并在人工明确确认后才使用精确 `--confirm` token。对 AI Agent，还必须等待用户对本次具体变更的明确确认。
- `tanswer api` 的 GET/HEAD 可直接执行；其他 HTTP 方法必须先 preview，并在明确确认本次 method、path、query 和 body 后才使用 `--confirm CONFIRM_TANSWER_RAW_API_WRITE`。

- 不在脚本、日志、共享文档、commit message 或 MR 描述中暴露真实 Token、真实环境地址或真实内网地址。
- 查询命令可以直接执行；资产、安全策略、元数据配置和响应处置写命令必须先预览变更，再使用确认令牌执行，并返回审计信息。
- 使用 `chaitin-cli tanswer manifest` 查看当前版本支持的命令、参数、输出字段和确认要求。
- `file-alarm` 命令只读取文件告警，不下载样本，也不触发新的沙箱分析。
- `metadata near-alarm` 只提供上下文，不能单独作为攻击证据。
- `response` 产品记录不能替代第三方联动设备执行证明。

## 常见问题

| 现象 | 处理方式 |
| --- | --- |
| `missing Quanxi address` | 设置 `TANSWER_URL`，或在命令中传入 `--url`。 |
| `missing OpenAPI token` | 设置 `TANSWER_API_KEY`，或在命令中传入 `--api-key`。 |
| `TOKEN_CHECK_FAILED` | 检查全悉地址、Token 是否正确、Token 是否过期/禁用、IP 访问策略和频率限制。 |
| TLS 证书校验失败 | 在确需跳过证书校验的环境中使用 `chaitin-cli tanswer --insecure auth check`。 |
| 命令不存在 | 升级到包含全悉语义 CLI 的 `chaitin-cli` Release，并重新运行 `chaitin-cli tanswer --help`。 |
| 读命令成功但写命令失败 | 检查 Token 绑定角色是否具备资产维护、策略维护或响应处置写权限。 |

## 文档索引

- [COMMAND_REFERENCE.md](./COMMAND_REFERENCE.md)：完整命令参考和长示例。
- [统一 chaitin-cli Skill](../../skills/chaitin-cli/SKILL.md)：AI Agent 使用规则。
- [manifest.go](./manifest.go)：`chaitin-cli tanswer manifest` 使用的 AI 可读命令清单。
