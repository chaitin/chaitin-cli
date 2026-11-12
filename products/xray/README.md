# 洞鉴(XRay) CLI

命令行工具，用于通过 OpenAPI v2 控制洞鉴(XRay)扫描平台。

## 命令列表

| 命令 | 说明 |
|------|------|
| `chaitin-cli xray asset_property` | 资产管理 |
| `chaitin-cli xray audit_log` | 审计日志管理 |
| `chaitin-cli xray baseline` | 基线检查管理 |
| `chaitin-cli xray custom_poc` | 自定义POC管理 |
| `chaitin-cli xray domain_asset` | 域名资产管理 |
| `chaitin-cli xray insight` | 数据洞察 |
| `chaitin-cli xray ip_asset` | 主机资产管理 |
| `chaitin-cli xray plan` | 任务计划管理 |
| `chaitin-cli xray project` | 工作区管理 |
| `chaitin-cli xray report` | 报表管理 |
| `chaitin-cli xray result` | 任务结果管理 |
| `chaitin-cli xray role` | 角色管理 |
| `chaitin-cli xray service_asset` | 服务资产管理 |
| `chaitin-cli xray system_info` | 系统信息管理 |
| `chaitin-cli xray system_service` | 系统服务管理 |
| `chaitin-cli xray task_config` | 任务配置管理 |
| `chaitin-cli xray template` | 策略模板管理 |
| `chaitin-cli xray user` | 用户管理 |
| `chaitin-cli xray vulnerability` | 漏洞资产管理 |
| `chaitin-cli xray web_asset` | Web资产管理 |
| `chaitin-cli xray xprocess` | XProcess任务实例管理 |
| `chaitin-cli xray xprocess_lite` | XProcess精简版管理 |

## 配置

在当前工作目录的 `config.yaml` 中配置：

```yaml
xray:
  url: https://x-ray-demo.chaitin.cn/api/v2
  api_key: YOUR_TOKEN
```

## 产品参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `--url` | API 地址 | 从 `config.yaml` 读取 |
| `--api-key` | 认证令牌 | 从 `config.yaml` 读取 |
| `--debug` | 输出调试日志 | false |

`--dry-run` 由主命令 `chaitin-cli` 提供，例如 `chaitin-cli --dry-run xray ...`。

## 任务管理

### 创建任务 (PostPlanCreateQuick)

快速创建立即、单次定时或周期扫描任务。不指定 `--plan-type` 时保持原有行为：创建后立即执行。

```bash
chaitin-cli xray plan PostPlanCreateQuick \
  --targets=10.3.0.4,10.3.0.5 \
  --engines=00000000000000000000000000000001 \
  --project-id=1
```

**参数说明：**

| 参数 | 必填 | 说明 |
|------|------|------|
| `--targets` | 是 | 扫描目标（逗号分隔） |
| `--engines` | 是 | 引擎 ID（逗号分隔） |
| `--project-id` | 是 | 项目 ID |
| `--template-name` | 否 | 模板名称搜索关键字（默认"基础服务漏洞扫描"） |
| `--name` | 否 | 任务名称（默认 quick-scan） |
| `--plan-type` | 否 | NOW/CLOCKED/DAY/WEEK/MONTH/MONTH_WEEK（默认 NOW） |
| `--exec-at` | 按类型 | CLOCKED 使用带时区 RFC3339；周期任务使用 HH:MM 或 HH:MM:SS |
| `--weekday` | 按类型 | WEEK/MONTH_WEEK 使用，1=周一，...，7=周日 |
| `--day-of-month` | 按类型 | MONTH 使用，范围 1-31 |
| `--week-of-month` | 按类型 | MONTH_WEEK 使用，范围 1-4 |

`--exec-at` 表示任务触发时间，不是扫描结束时间。周期计划不会在创建时额外触发扫描。

```bash
# 每天 10:30 执行
chaitin-cli xray plan PostPlanCreateQuick \
  --targets=10.3.0.4 \
  --engines=00000000000000000000000000000001 \
  --project-id=1 \
  --template-name=主机资产监控 \
  --plan-type=DAY \
  --exec-at=10:30

# 每月第 2 周的星期五 10:30 执行
chaitin-cli xray plan PostPlanCreateQuick \
  --targets=10.3.0.4 \
  --engines=00000000000000000000000000000001 \
  --project-id=1 \
  --plan-type=MONTH_WEEK \
  --week-of-month=2 \
  --weekday=5 \
  --exec-at=10:30
```

### 更新任务排期 (PostPlanUpdateQuick)

只修改已有任务的排期，保留名称、目标、引擎、扫描配置和执行窗口，并且不会立即触发扫描。

```bash
# 修改为每周五 11:00 执行
chaitin-cli xray plan PostPlanUpdateQuick \
  --id=783 \
  --plan-type=WEEK \
  --weekday=5 \
  --exec-at=11:00

# 保持当前计划类型，只修改触发时间
chaitin-cli xray plan PostPlanUpdateQuick \
  --id=783 \
  --exec-at=12:00
```

### 任务列表 (PostPlanFilter)

查看扫描任务列表。

```bash
chaitin-cli xray plan PostPlanFilter \
  --filterPlan.limit=10 \
  --filterPlan.offset=0 \
  --filterPlan.project-id=1
```

### 暂停任务 (PostPlanStop)

暂停正在运行的扫描任务。

```bash
chaitin-cli xray plan PostPlanStop \
  --stopPlanBody.id=783
```

### 恢复任务 (PostPlanExecute)

恢复已暂停的扫描任务。

```bash
chaitin-cli xray plan PostPlanExecute \
  --executePlanBody.id=783
```

### 删除任务 (DeletePlanID)

删除指定的扫描任务。

```bash
chaitin-cli xray plan DeletePlanID \
  --id=783
```
