# Insight CLI

Exposure & Risk Management (暴露面与风险管理) API 命令行工具。

## 简介

Insight CLI 提供了与 Insight 暴露面与风险管理平台进行交互的自动化能力。它支持任务调度、资产快照获取、漏洞信息查询以及工单流水管理，非常适合安全运营工程师（SecOps）用来编写定时脚本、与 CI/CD 管道对接，或供 AI Agent 调用以实现自动化闭环。

## 安装

在仓库根目录执行构建：

```bash
task build
# 编译后的文件位于 bin/chaitin-cli
```

## 配置

在当前工作目录的 `config.yaml` 中配置：

```yaml
insight:
  url: "https://your-insight-platform.com"
  api_key: "YOUR_JWT_TOKEN"
```

同时，也支持通过命令行参数或环境变量进行配置：

| 参数 | 环境变量 | 说明 |
|------|----------|------|
| `--url` | `INSIGHT_URL` | Insight 平台地址（必填） |
| `--api-key` | `INSIGHT_API_KEY` | API Key（必填，请在后台生成） |
| `--output` / `-o`| - | 输出格式，支持 `table` 或 `json` (默认 json) |
| `--verbose` / `-v`| - | 打印完整的 HTTP 请求与响应信息 |
| `--dry-run` | - | 模拟执行，仅打印组装好的请求，不向服务器发送真实报文 |

**API Key 格式提醒：**
Insight 的 API Key 是一串 JWT 格式的字符串。配置后，CLI 会自动将其同时注入至请求的 `Authorization` 和 `Cookie` 中，以满足后端的鉴权需求。

## 命令概览

```bash
chaitin-cli insight [全局参数] <命令> [子命令] [参数]
```

### 全局参数示例

```bash
# JSON 格式输出，跳过 TLS 校验
chaitin-cli insight task list --insecure --output json
```

---

## 核心命令详情

### 1. 任务流管理 (Task & Execution)

这是平台运转的心脏，最适合通过 CLI 触发扫描任务或查看任务执行记录。

#### 1.1 列出所有任务 (`task list`)

查看系统里配置了哪些巡检和扫描任务。

```bash
chaitin-cli insight task list [flags]
```
**可用参数：**
* `--count int`: 每页返回数量 (默认 20)
* `--offset uint`: 跳过数量 (默认 0)

#### 1.2 重新执行/触发扫描 (`task start`)

对某个已配置好的任务立即发起一次排查执行。
```bash
chaitin-cli insight task start --id <task_id>
```

#### 1.3 停止运行中的任务 (`task stop`)

遇到误报或影响业务时紧急掐断任务执行。
```bash
chaitin-cli insight task stop --id <task_id>
```

#### 1.4 查询执行记录详情 (`task execution`)

查询某一次排查任务执行记录的详情。Insight 后端接口为 `GET /exposure/api/task/execution?id=<execution_id>`，其中 `id` 是执行记录 ID。
```bash
chaitin-cli insight task execution --id <execution_id>
```

---

### 2. 漏洞与排查结果感知 (Vuln & Result)

拉取风险点对应的具体漏洞详情。由于使用了 JSON-RPC，高级查询可以按需通过 JSON Body 传入，CLI 默认支持分页拉取。

#### 2.1 查询 IP 漏洞 (`vuln ip`)

拉取所有与 IP 资产相关的漏洞风险列表。
```bash
chaitin-cli insight vuln ip [flags]
```
**可用参数：**
* `--count int`: 每页返回数量 (默认 20)
* `--offset uint`: 跳过数量 (默认 0)

*后端完整支持的过滤字段（API 层）：*
`id`, `port`, `service`, `protocol`, `vuln_ip`, `vuln_location`, `rel_asset_name`, `vuln_status`, `name` 等。

#### 2.2 查询 Web 漏洞 (`vuln web`)

拉取所有与 Web 站点、域名等相关的漏洞风险列表。
```bash
chaitin-cli insight vuln web [flags]
```
**可用参数：**
* `--count int`: 每页返回数量 (默认 20)
* `--offset uint`: 跳过数量 (默认 0)

#### 2.3 获取特定排查任务的风险点 (`result list`)

获取某次任务最新排查出的风险列表。CLI 使用业务语义参数 `--task-id`，请求 Insight REST 接口时会映射为 `id=<task_id>`。
```bash
chaitin-cli insight result list --task-id <task_id>
chaitin-cli insight result list --task-id <task_id> --execution-id <execution_id>
```

#### 2.4 获取排查风险增量对比 (`result diff`)

获取本次排查和上次排查的增量/消除情况对比。
```bash
chaitin-cli insight result diff --exec-id <execution_id>
```

---

### 3. 资产快照与台账 (Asset & Snapshot)

提取或同步 CMDB 和资产管理系统的数据。资产列表查询基于后端的 RPC 接口。

#### 3.1 提取 IP 资产台账 (`asset ip`)

查询所有内部或外部暴露的 IP 维度资产。
```bash
chaitin-cli insight asset ip [flags]
```
**可用参数：**
* `--count int`: 每页返回数量 (默认 20)
* `--offset uint`: 跳过数量 (默认 0)

*后端完整支持的过滤字段（API 层）：*
`id`, `name`, `asset_state`, `external` (是否互联网暴露), `business_id`, `scope_id`, `online_status` 等。

#### 3.2 提取 Web 资产台账 (`asset web`)

查询所有 Web 业务、域名及网站维度的资产台账。
```bash
chaitin-cli insight asset web [flags]
```
**可用参数：**
* `--count int`: 每页返回数量 (默认 20)
* `--offset uint`: 跳过数量 (默认 0)

#### 3.3 提取主机软件资产台账 (`asset software`)

查询所有从主机提取出的软件组件与中间件版本等资产。
```bash
chaitin-cli insight asset software [flags]
```
**可用参数：**
* `--count int`: 每页返回数量 (默认 20)
* `--offset uint`: 跳过数量 (默认 0)

*后端完整支持的过滤字段（API 层）：*
`id`, `name`, `ip`, `agent_id`, `owner_id` 等。

#### 3.4 提取资产标签列表 (`asset tag`)

查询系统中定义的所有资产标签列表，这通常用于其他脚本自动化给资产打标之前获取合法的标签数据。
```bash
chaitin-cli insight asset tag [flags]
```
**可用参数：**
* `--count int`: 每页返回数量 (默认 20)
* `--offset uint`: 跳过数量 (默认 0)

*后端完整支持的过滤字段（API 层）：*
`id`, `tag_name`, `description`, `auto_mark_enable`, `updated_by_id` 等。

#### 3.5 提取业务系统列表 (`asset business`)

查询系统中定义的业务系统台账列表，方便将漏洞或资产根据所属业务系统进行进一步聚合。
```bash
chaitin-cli insight asset business [flags]
```
**可用参数：**
* `--count int`: 每页返回数量 (默认 20)
* `--offset uint`: 跳过数量 (默认 0)

*后端完整支持的过滤字段（API 层）：*
`id`, `name`, `full_name`, `importance`, `organization_id`, `maintainer_id` 等。

#### 3.6 平台资产全量快照 (`snapshot asset`)

查看某个排查任务的资产快照数据。CLI 使用业务语义参数 `--task-id`，请求 Insight REST 接口时会映射为 `id=<task_id>`。
```bash
chaitin-cli insight snapshot asset --task-id <task_id>
chaitin-cli insight snapshot asset --task-id <task_id> --execution-id <execution_id>
```

---

### 4. 工单流转管理 (Order / Workflow)

将 Insight 生成的待处理工单同步至外部 ITSM、Jira 或企微机器人。

#### 4.1 查询全部工单列表 (`order list`)

获取系统内的审批、漏洞修复或资产认领等工单记录。
```bash
chaitin-cli insight order list [flags]
```
**可用参数：**
* `--page int`: 页码 (默认 1)
* `--size int`: 每页返回数量 (默认 20)
* `--name string`: 按工单名称进行模糊搜索
* `--status int`: 按工单状态值过滤 (1:进行中, 2:已完结等)
* `--is-timeout bool`: 过滤出已超时的工单 (true/false)

---

### 5. 系统与授权管理 (System)

用于系统上线初始化、授权监控的刚需低频指令。

#### 5.1 查看 License 信息
在续费、更新授权或者系统监控自动化脚本里使用。
```bash
chaitin-cli insight system license
```

#### 5.2 获取 Machine ID
用于离线申请授权凭证时获取机器码。
```bash
chaitin-cli insight system machine-id
```

---

## 示例

### 场景一：通过 CI/CD 或 Agent 触发扫描

```bash
# 1. 查找指定资产相关的任务
chaitin-cli insight task list --count 100

# 2. 拿到 Task ID 后，触发重新执行
chaitin-cli insight task start --id "1234"

# 3. 拿到 execution_id 后，查看本次执行记录详情
chaitin-cli insight task execution --id "456"
```

### 场景二：每日定时拉取超时工单

```bash
# 编写定时脚本拉取超时工单，并通过 JQ 等工具解析处理发给告警群
chaitin-cli insight order list --is-timeout true --size 50
```

### 场景三：漏洞大盘数据同步

```bash
# 第一页拉取前 100 个 Web 漏洞
chaitin-cli insight vuln web --count 100

# 翻页拉取后面的 100 个
chaitin-cli insight vuln web --count 100 --offset 100
```
