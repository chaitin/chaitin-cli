# Xray CLI

命令行工具，用于通过 OpenAPI v2 控制 Xray 扫描平台。

## 命令列表

| 命令 | 说明 |
|------|------|
| `asset_property` | 资产属性管理 |
| `audit_log` | 审计日志 |
| `baseline` | 基线检查 |
| `custom_poc` | 自定义 PoC |
| `domain_asset` | 域名资产管理 |
| `insight` | 洞察分析 |
| `ip_asset` | IP 资产管理 |
| `plan` | 任务管理 |
| `project` | 项目管理 |
| `report` | 报告管理 |
| `result` | 结果查询 |
| `role` | 角色管理 |
| `service_asset` | 服务资产管理 |
| `system_info` | 系统信息 |
| `system_service` | 系统服务 |
| `task_config` | 任务配置 |
| `template` | 策略模板 |
| `user` | 用户管理 |
| `vulnerability` | 漏洞管理 |
| `web_asset` | Web 资产管理 |
| `xprocess` | 扫描进程 |
| `xprocess_lite` | 轻量级扫描进程 |

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

快速创建扫描任务，自动使用"基础服务漏洞扫描"模板。

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
| `--template-id` | 否 | 指定模板 ID（默认自动查找"基础服务漏洞扫描"） |
| `--template-name` | 否 | 模板名称搜索关键字 |
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
