# Cosmos / AISOC CLI

Cosmos（万象/AISOC）命令行工具，通过 JSON-RPC 2.0 调用万象后端 API，覆盖告警分析研判、日志查询、威胁情报、IP 封禁、资产管理、通知策略、运维监控、SOAR 和漏洞管理等能力。

命令树由 `products/cosmos/apis/*.json` 中的 API 定义生成。实际请求会发送到 `<url>/pedestal/rpc`，认证方式为 `Authorization: Bearer <api_key>`。

## 配置

推荐在 `~/.chaitin-cli/config.yaml` 中配置；当前目录下可识别的 `./config.yaml` 会被优先读取：

```yaml
cosmos:
  url: https://cosmos.example.com
  api_key: YOUR_JWT_TOKEN
```

也可以使用环境变量或本地 `.env`：

```bash
export COSMOS_URL=https://cosmos.example.com
export COSMOS_API_KEY=YOUR_JWT_TOKEN
```

产品级覆盖参数统一为 `--url` 和 `--api-key`：

```bash
chaitin-cli cosmos --url https://cosmos.example.com --api-key YOUR_JWT_TOKEN asset search-host-asset --count 20 --offset 0
```

优先级为 `flags > environment/.env > 识别后的 ./config.yaml > ~/.chaitin-cli/config.yaml`。

注意事项：

- `url` 填万象实例地址即可，不要追加 `/pedestal/rpc`，CLI 会自动追加。
- `api_key` 填 JWT token 本体，不要包含 `Bearer ` 前缀。
- 根级 `--dry-run` 会打印脱敏后的请求摘要，不会发送请求。
- 产品级 `--raw` 会输出未格式化的 JSON；默认会尽量格式化 JSON 响应。

## 命令结构

通用格式：

```bash
chaitin-cli cosmos <service> <method> [flags]
```

服务名和方法名按 API 名称自动转换。例如 `AssetService.SearchHostAsset` 会变为：

```bash
chaitin-cli cosmos asset search-host-asset
```

常用服务：

| 服务命令 | 后端 Service | 能力 |
| --- | --- | --- |
| `agent` | `AgentService` | 安全设备列表查询 |
| `alarm` | `AlarmService` | 告警列表、详情、标签、扩展信息、SOAR 触发 |
| `judge` | `JudgeService` | 批量研判告警 |
| `asset` | `AssetService` | IP / Web 资产查询和主机资产保存 |
| `intelligence` | `IntelligenceService` | IP 情报查询、详情、标签和新增 |
| `ip-block` | `IpBlockService` | 黑名单、白名单、封禁历史和封禁设备 |
| `log` | `LogService` | 安全日志检索、详情、聚合统计和总数 |
| `nonstandard-log` | `NonstandardLogApiService` | 非标日志查询 |
| `notice` | `NoticeService` | Webhook、通知策略、SMTP、短信和订阅人 |
| `ops` | `OpsService` | 集群监控、系统告警、执行器和运维操作 |
| `data-mgr` | `DataMgrService` | 数据监控总览、状态和趋势 |
| `extra-setting` | `ExtraSettingService` | 节点和组件健康信息 |
| `analysis` | `AnalysisService` | SOAR 剧本和执行记录查询 |
| `scan-vuln-ip` | `ScanVulnIpService` | IP / Web 漏洞列表、详情和新增更新 |

查看完整命令和参数：

```bash
chaitin-cli cosmos --help
chaitin-cli cosmos asset --help
chaitin-cli cosmos alarm get-alarm-list --help
```

## 参数规则

- 普通字段会生成为同名 flag，例如 `--count`、`--offset`、`--ip`。
- 嵌套对象会使用点号展开，例如 `--ip_ioc.ip`、`--ip_ioc.severity`。
- 字符串数组、整数数组等基础数组可用逗号分隔或重复传入，例如 `--keyword nginx,ssh`。
- 复杂查询字段通常传 JSON 数组，例如 `--ip '[{"oper":"=","target":"10.0.0.1"}]'`。
- `TimeQuery` 字段支持完整 JSON，也支持 `startMs-endMs` 简写，例如 `--created_at 1717200000000-1717286400000`。
- 写操作中的复杂对象通常传 JSON 字符串，例如 `--category_ids '[{"id":41,"name":"Linux"}]'`。

`alarm get-alarm-list` 为了避免误扫大表，默认要求提供 `--created_at` 或 `--updated_at`。如果通过工单筛选，可使用 `--workflow_id`；确需跳过时间条件时再使用 `--exclude_time_filter`。

## 示例

查询 IP 资产列表：

```bash
chaitin-cli cosmos asset search-host-asset --count 20 --offset 0 --raw
```

按 IP 精确查询资产：

```bash
chaitin-cli cosmos asset search-host-asset \
  --ip '[{"oper":"=","target":"10.0.0.1"}]' \
  --count 20 \
  --offset 0
```

预览保存主机资产请求。`--asset_ip_type` 当前取值为 `1` 实际 IP、`2` 虚拟 IP：

```bash
chaitin-cli --dry-run cosmos asset save-host-asset \
  --ip 10.0.0.1/32 \
  --name demo-host \
  --organization_id 1 \
  --asset_ip_type 1 \
  --category_ids '[{"id":41,"name":"Linux"}]' \
  --group_id 1 \
  --raw
```

查询告警列表：

```bash
chaitin-cli cosmos alarm get-alarm-list \
  --created_at 1717200000000-1717286400000 \
  --count 20 \
  --offset 0 \
  --raw
```

查询告警详情。先从告警列表返回值中获取真实 `id`；`created_at` 为可选分区键，使用列表返回的 `created_at` 转为毫秒整数后可加速查询：

```bash
chaitin-cli cosmos alarm get-alarm-info --id ALARM_ID --created_at CREATED_AT_MS --raw
```

预览写回告警 AI 研判与处置字段。专用后端接口为 `AlarmService.UpdateAiFieldsContent`，参数必须包装在 `ai_fields` 中；去掉根级 `--dry-run` 后才会真实写入：

```bash
chaitin-cli --dry-run cosmos alarm update-ai-fields-content \
  --ai_fields.alarm_id ALARM_ID \
  --ai_fields.analysis_report '# AI研判报告' \
  --ai_fields.confidence_score 85 \
  --ai_fields.ai_analysis 1 \
  --ai_fields.ai_judgment_result '误报' \
  --ai_fields.ai_judgment_level 1 \
  --ai_fields.ai_event_nature '业务测试' \
  --ai_fields.disposal_report '# AI处置报告' \
  --ai_fields.ai_disposal 1 \
  --raw
```

字段映射为：`analysis_report -> ai_analysis_result`、`confidence_score -> ai_confidence_score`、`disposal_report -> ai_disposal_result`；其余 AI 参数与告警返回字段同名。`ai_fields.created_at` 是可选 RFC3339 分区键，只能使用告警详情返回的真实创建时间转换得到；不确定时不要传。

查询安全日志：

```bash
chaitin-cli cosmos log search-log-list \
  --time_range_start 1717200000000 \
  --time_range_end 1717286400000 \
  --keyword nginx,error \
  --count 20 \
  --offset 0
```

预览新增 IP 情报请求。去掉根级 `--dry-run` 后会真实写入情报：

```bash
chaitin-cli --dry-run cosmos intelligence create-ip-ioc \
  --ip_ioc.ip 192.0.2.10 \
  --ip_ioc.intel_type ipv4 \
  --ip_ioc.threat_type ATTACK_IP \
  --ip_ioc.severity 3 \
  --ip_ioc.confidence 80 \
  --ip_ioc.confidence_level 4 \
  --threat_tags '[{"types":1,"tag":"Scanner","validity":1}]'
```

预览创建封禁 IP：

```bash
chaitin-cli --dry-run cosmos ip-block create-black-ip \
  --configs '[{"ips":["192.0.2.10"],"dev_ids":["DEVICE_ID"],"expire_at":253402271999000,"type":"black","remark":"manual block"}]' \
  --raw
```

查询 SOAR 剧本并预览触发通用剧本。SOAR 查询依赖目标环境的 SOAR 服务状态；触发剧本前先将 `SOAR_UUID` 替换为真实剧本 ID：

```bash
chaitin-cli cosmos analysis get-soar-playbook-list --types 1 --ai_call 1

chaitin-cli --dry-run cosmos alarm trigger-playbook-general \
  --soar_id SOAR_UUID \
  --params '{"title":"告警标题","msg":"通知内容"}'
```

查询集群节点 CPU 平均使用率。先查询节点信息，选择目标环境中真实存在的节点 IP；运维监控类接口中的 `start` / `end` 使用 Unix 秒级时间戳，并依赖目标环境有对应监控数据：

```bash
chaitin-cli cosmos extra-setting get-node-info --raw

chaitin-cli cosmos ops bigdata-cpu-avg \
  --ip NODE_IP \
  --start 1781515764 \
  --end 1781519364
```

## 维护 API 定义

新增或调整 Cosmos 接口时，优先修改 `products/cosmos/apis/*.json`：

- `name` 使用后端 JSON-RPC 方法名，例如 `AlarmService.GetAlarmList`。
- `comment` 会作为命令的 `Short` 文案和 help 说明。
- `args` 描述入参字段、类型、是否可选和查询类型。
- `paramStyle: "array"` 表示 JSON-RPC `params` 需要包装成单元素数组。

修改 API 定义后运行：

```bash
go test ./products/cosmos
```
