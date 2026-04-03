# 洞鉴(XRay) CLI

命令行工具，用于通过 OpenAPI v2 控制洞鉴(XRay)扫描平台。

## 命令列表

| 命令 | 说明 |
|------|------|
| `cws xray asset_property` | 资产管理 |
| `cws xray audit_log` | 审计日志管理 |
| `cws xray baseline` | 基线检查管理 |
| `cws xray custom_poc` | 自定义POC管理 |
| `cws xray domain_asset` | 域名资产管理 |
| `cws xray insight` | 数据洞察 |
| `cws xray ip_asset` | 主机资产管理 |
| `cws xray plan` | 任务计划管理 |
| `cws xray project` | 工作区管理 |
| `cws xray report` | 报表管理 |
| `cws xray result` | 任务结果管理 |
| `cws xray role` | 角色管理 |
| `cws xray service_asset` | 服务资产管理 |
| `cws xray system_info` | 系统信息管理 |
| `cws xray system_service` | 系统服务管理 |
| `cws xray task_config` | 任务配置管理 |
| `cws xray template` | 策略模板管理 |
| `cws xray user` | 用户管理 |
| `cws xray vulnerability` | 漏洞资产管理 |
| `cws xray web_asset` | Web资产管理 |
| `cws xray xprocess` | XProcess任务实例管理 |
| `cws xray xprocess_lite` | XProcess精简版管理 |

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

`--dry-run` 由主命令 `cws` 提供，例如 `cws --dry-run xray ...`。

## 任务管理

### 创建任务 (PostPlanCreateQuick)

快速创建扫描任务，立即执行（马上扫一次）。

```bash
cws xray plan PostPlanCreateQuick \
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

### 任务列表 (PostPlanFilter)

查看扫描任务列表。

```bash
cws xray plan PostPlanFilter \
  --filterPlan.limit=10 \
  --filterPlan.offset=0 \
  --filterPlan.project-id=1
```

### 暂停任务 (PostPlanStop)

暂停正在运行的扫描任务。

```bash
cws xray plan PostPlanStop \
  --stopPlanBody.id=783
```

### 恢复任务 (PostPlanExecute)

恢复已暂停的扫描任务。

```bash
cws xray plan PostPlanExecute \
  --executePlanBody.id=783
```

### 删除任务 (DeletePlanID)

删除指定的扫描任务。

```bash
cws xray plan DeletePlanID \
  --id=783
```
