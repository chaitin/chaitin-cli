# 全悉 CLI 命令索引

本文档是全悉 CLI 的人工可读快速参考，供产品用户、开发者和仓库维护者快速发现常见任务与命令。

在实际操作或让 AI Agent 操作前，必须使用当前安装版本的运行时契约：

```bash
chaitin-cli tanswer --help
chaitin-cli tanswer <领域或命令> --help
chaitin-cli tanswer manifest
```

首次使用请先按 [全悉 CLI 入门的配置说明](./README.md#快速使用) 设置连接信息，并运行 `chaitin-cli tanswer auth check`。可使用 `--url`、`--api-key`、`--timeout`、`--insecure`；环境变量 `TANSWER_URL`、`TANSWER_API_KEY`、`TANSWER_TIMEOUT`、`TANSWER_INSECURE`；或 `config.yaml` 的 `tanswer.url`、`tanswer.api_key`、`tanswer.timeout`、`tanswer.insecure`。

本文档提供稳定的操作入口和安全示例；当前版本的准确参数、枚举、输出字段、风险等级和 `--confirm` token，以运行时 help 与 manifest 为准。

## 常用任务快捷示例

以下示例可作为人工操作者的起点。前五类为只读查询；后三类均带有 `--preview`，不会提交产品变更。示例中的文档保留地址仅用于演示，执行前请替换成受控环境中的实际值。

```bash
# 1. 首次配置后，检查连接和 Token 权限。
chaitin-cli tanswer auth check

# 2. 确认系统、License 和节点状态。
chaitin-cli tanswer system status

# 3. 查看最近 24 小时的高优先级威胁告警。
chaitin-cli tanswer alarm high-priority --time 24h --page-size 10

# 4. 查看最近 7 天的恶意文件告警。
chaitin-cli tanswer file-alarm malicious --time 7d --page-size 10

# 5. 查询资产清单。
chaitin-cli tanswer asset list --page-size 10

# 6. 预览创建资产，不会写入后端。
chaitin-cli tanswer asset create --name core-db --ip 192.0.2.10 --preview

# 7. 预览创建检测白名单，不会写入后端。必须提供过期时间或有效时长。
chaitin-cli tanswer policy detection-whitelist-create --name trusted-source --src-ip 192.0.2.20 --valid-time 3600 --preview

# 8. 预览创建响应封禁策略，不会写入后端。
chaitin-cli tanswer response block-policy-create --name block-malicious-ip --object 192.0.2.30 --preview
```

如需实际执行第 6 至 8 类操作，先阅读该命令的 `--help` 或 manifest 中的确认要求，检查 preview 输出并确认本次变更后，再使用指定的精确 `--confirm` token。对 AI Agent，还必须先取得用户的明确确认；token 本身不等同于用户授权。

## Foundation

```text
tanswer auth status
tanswer auth check
tanswer system status
tanswer manifest
tanswer api <METHOD> <PATH>
```

`api` 仅用于已知且已授权、又没有语义命令覆盖的 Open API；不要据此猜测 RPC 方法、路径或请求体。GET/HEAD 可直接执行；其他 HTTP 方法默认只输出请求 preview、不会发送请求。人工审阅方法、路径、query 和 body 后，AI Agent 还必须取得用户对该次请求的明确确认，才能使用 `--confirm CONFIRM_TANSWER_RAW_API_WRITE` 执行。根级 `--dry-run` 不适用于 `tanswer`。

## Threat alarm

```text
tanswer alarm overview
tanswer alarm list
tanswer alarm timeline
tanswer alarm high-priority
tanswer alarm detail
tanswer alarm by-attacker
tanswer alarm by-victim
tanswer alarm by-threat
tanswer alarm important-assets
tanswer alarm attacker-rank
tanswer alarm victim-rank
tanswer alarm phase-distribution
tanswer alarm related
```

## File alarm

```text
tanswer file-alarm overview
tanswer file-alarm malicious
tanswer file-alarm webshell
tanswer file-alarm sandbox
tanswer file-alarm detail
```

## Asset

```text
tanswer asset list
tanswer asset detail
tanswer asset group-tree
tanswer asset download-template
tanswer asset export
tanswer asset create
tanswer asset update
tanswer asset delete
tanswer asset batch-maintain
tanswer asset batch-tag
tanswer asset group-create
tanswer asset group-rename
tanswer asset group-delete
tanswer asset tree-move
tanswer asset import
```

`create`、`update`、`delete`、批量操作、资产组操作和导入属于受保护写操作。

## Metadata

```text
tanswer metadata protocol
tanswer metadata search
tanswer metadata detail
tanswer metadata near-alarm
tanswer metadata config
tanswer metadata config-update
```

`config-update` 属于受保护写操作。

## Security policy

```text
tanswer policy detection-whitelist
tanswer policy detection-whitelist-create
tanswer policy detection-whitelist-update
tanswer policy detection-whitelist-enable
tanswer policy detection-whitelist-disable
tanswer policy detection-whitelist-delete
tanswer policy detection-whitelist-from-alarm
tanswer policy detection-whitelist-export
tanswer policy detection-whitelist-import

tanswer policy custom-intelligence
tanswer policy custom-intelligence-create
tanswer policy custom-intelligence-update
tanswer policy custom-intelligence-enable
tanswer policy custom-intelligence-disable
tanswer policy custom-intelligence-delete
tanswer policy custom-intelligence-export
tanswer policy custom-intelligence-import
```

除查询和导出外，策略命令通常会改变产品状态，必须先查看对应命令的 help 或 manifest。

## Response

```text
tanswer response block-policies
tanswer response block-policy-create
tanswer response block-policy-update
tanswer response block-policy-enable
tanswer response block-policy-disable
tanswer response block-policy-delete
tanswer response block-records
tanswer response whitelist
tanswer response whitelist-create
tanswer response whitelist-update
tanswer response whitelist-enable
tanswer response whitelist-disable
tanswer response whitelist-delete
tanswer response block-policy-from-alarm
tanswer response whitelist-from-alarm
tanswer response devices
tanswer response device-records
tanswer response auto-policies
tanswer response auto-list
```

创建、更新、启停、删除和从告警创建策略/白名单属于受保护写操作。

## 受保护写操作

统一流程如下（适用于各领域的受保护语义写操作；raw API 非 GET/HEAD 请求按上文的专用确认流程执行）：

```bash
# 1. 用实际命令及参数查看影响；该步骤不执行产品变更。
chaitin-cli tanswer <write-command> ... --preview

# 2. 读取 preview 后，使用该命令 help 或 manifest 给出的精确 token 执行。
chaitin-cli tanswer <write-command> ... --confirm CONFIRM_<OPERATION>
```

不要复用或猜测 `CONFIRM_<OPERATION>`；每个操作需要的精确 token 以运行时返回内容为准。
