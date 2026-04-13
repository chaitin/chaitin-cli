# CWS Commands Reference

> Comprehensive command reference for all Chaitin products. Use `cws <product> --help` to discover dynamically.

## Quick Lookup by Intent

### CloudWalker (牧云)

| Intent | Commands |
|--------|----------|
| **资产清点** | |
| 主机资产 | `cws cloudwalker host-asset list` |
| 端口资产 | `cws cloudwalker port-asset list` |
| 网站资产 | `cws cloudwalker website-asset list` |
| 应用资产 | `cws cloudwalker application-asset list` |
| Docker资产 | `cws cloudwalker docker-container/image/network list` |
| 证书/启动项/注册表 | `cws cloudwalker asset-cert/startup/registry list` |
| **漏洞管理** | |
| 漏洞列表 | `cws cloudwalker vuln list` |
| 漏洞详情 | `cws cloudwalker vuln-info get --id=<id>` |
| 应急漏洞 | `cws cloudwalker emergency-vuln-v1 list` |
| 补丁情报 | `cws cloudwalker patch-info list` |
| **威胁检测** | |
| 恶意软件事件 | `cws cloudwalker malware-event list` |
| Webshell事件 | `cws cloudwalker webshell-event list` |
| 网络异常事件 | `cws cloudwalker network-audit-event list` |
| 暴力破解 | `cws cloudwalker brute-force list` |
| 弱口令 | `cws cloudwalker weak-passwd list` |
| 内存马/反弹Shell | `cws cloudwalker memory-webshell-event/revshell-event list` |
| 提权进程 | `cws cloudwalker elevation-process-event list` |
| **入侵检测** | |
| 异常登录 | `cws cloudwalker abnormal-login-event list` |
| 可疑操作 | `cws cloudwalker suspicious-operation list` |
| 蜜罐诱捕 | `cws cloudwalker honeypot list` |
| 防篡改 | `cws cloudwalker tamper-proof list` |
| **安全基线** | |
| 基线检查 | `cws cloudwalker baseline-v2 check` |
| 基线结果 | `cws cloudwalker baseline-v2 get` |
| 安全检查 | `cws cloudwalker security-check list` |
| 检测规则 | `cws cloudwalker detection-rule list` |
| **探针管理** | |
| 探针列表 | `cws cloudwalker agent list` |
| 探针分组 | `cws cloudwalker agent get-agent-group-tree` |
| 恶意文件探针 | `cws cloudwalker agent-detector list` |
| 探针模块 | `cws cloudwalker agent-module get-agent-list` |
| **策略管理** | |
| 安全策略 | `cws cloudwalker security-strategy get-system-strategy` |
| 登录控制 | `cws cloudwalker security-strategy get-login-control` |
| LDAP/Radius | `cws cloudwalker security-strategy get-ldap/get-radius` |
| **统计概览** | |
| 威胁概览 | `cws cloudwalker threat-overview list` |
| 事件统计 | `cws cloudwalker event-stat list` |
| 审计日志 | `cws cloudwalker audit-log list` |

### SafeLine (雷池)

| Intent | Commands |
|--------|----------|
| **统计** | |
| 防护概览 | `cws safeline stats overview` |
| 攻击类型 | `cws safeline stats attack-types` |
| **日志** | |
| 拦截事件 | `cws safeline logs event` |
| 攻击日志 | `cws safeline logs attack` |
| **规则** | |
| 规则列表 | `cws safeline rule list` |
| 启用/禁用规则 | `cws safeline rule enable/disable --id=<id>` |
| 策略规则 | `cws safeline policy-rule list` |
| **站点** | |
| 站点列表 | `cws safeline site list` |
| 策略组 | `cws safeline policy-group list` |
| **访问控制** | |
| ACL列表 | `cws safeline acl list` |
| IP组 | `cws safeline ip-group list` |
| **配置** | |
| 系统配置 | `cws safeline config show` |
| 防护状态 | `cws safeline protection status` |

### SafeLine-CE (雷池社区版)

| Intent | Commands |
|--------|----------|
| 统计概览 | `cws safeline-ce stat overview` |
| 站点管理 | `cws safeline-ce site list/add/update/delete` |
| 攻击日志 | `cws safeline-ce log attack list` |
| 规则管理 | `cws safeline-ce rule list/add/update/delete` |
| IP组 | `cws safeline-ce ipgroup list/add/update` |
| 证书 | `cws safeline-ce cert list/add` |
| 语义模块 | `cws safeline-ce module list/update` |
| SkyNet规则 | `cws safeline-ce skynet list/add` |

### T-Answer (全悉)

| Intent | Commands |
|--------|----------|
| 防火墙规则 | `cws tanswer firewall list/create/update` |
| 规则管理 | `cws tanswer rules list/create/update` |

### X-Ray (洞鉴)

| Intent | Commands |
|--------|----------|
| **漏洞扫描** | |
| 快速扫描 | `cws xray plan PostPlanCreateQuick --targets=<target> --engines=<engine>` |
| 扫描任务列表 | `cws xray plan PostPlanFilter` |
| 任务详情 | `cws xray plan GetPlanID --id=<id>` |
| 启动/恢复扫描 | `cws xray plan PostPlanExecute --executePlanBody.id=<id>` |
| 停止扫描 | `cws xray plan PostPlanStop --stopPlanBody.id=<id>` |
| 删除任务 | `cws xray plan DeletePlanID --id=<id>` |
| **资产管理** | |
| 资产属性 | `cws xray asset_property list` |
| 域名资产 | `cws xray domain_asset list` |
| 主机资产 | `cws xray ip_asset list` |
| Web资产 | `cws xray web_asset list` |
| 漏洞资产 | `cws xray vulnerability list` |
| **基线检查** | |
| 基线列表 | `cws xray baseline list` |
| 执行检查 | `cws xray baseline check --id=<id>` |
| **POC管理** | |
| POC列表 | `cws xray custom_poc list` |
| 创建POC | `cws xray custom_poc create` |
| **报表** | |
| 报告列表 | `cws xray report list` |
| 生成报告 | `cws xray report generate --plan-id=<id>` |
| **系统管理** | |
| 用户管理 | `cws xray user list` |
| 工作区 | `cws xray project list` |
| 审计日志 | `cws xray audit_log list` |

---

## CloudWalker (牧云) - 80+ Commands

### 资产清点 Asset Management
```bash
cws cloudwalker host-asset list                              # 主机资产列表
cws cloudwalker host-asset get --id=<id>                    # 主机详情
cws cloudwalker host-asset get-agent-source                 # 探针来源
cws cloudwalker host-asset get-agent-group-tree            # 主机分组

cws cloudwalker port-asset list                             # 端口资产列表
cws cloudwalker port-asset list --host=<ip>                # 按主机筛选

cws cloudwalker website-asset list                          # 网站资产
cws cloudwalker application-asset list                      # 应用资产
cws cloudwalker application-asset get --id=<id>           # 应用详情

cws cloudwalker user-asset list                             # 用户资产
cws cloudwalker process-asset list                          # 进程资产
cws cloudwalker docker-container list                       # Docker容器
cws cloudwalker docker-image list                           # Docker镜像
cws cloudwalker docker-network list                         # Docker网络

cws cloudwalker asset-cert list                             # 证书资产
cws cloudwalker asset-startup list                          # 启动项资产
cws cloudwalker asset-env list                              # 环境变量资产
cws cloudwalker asset-registry list                         # 注册表资产
cws cloudwalker asset-crontab list                          # 计划任务资产

cws cloudwalker host-nic-asset list                         # 网卡资产
cws cloudwalker host-partition-asset list                   # 分区资产
cws cloudwalker host-route-asset list                       # 路由资产
cws cloudwalker sensitive-port list                          # 敏感端口
cws cloudwalker sensitive-file list                          # 敏感文件
cws cloudwalker sensitive-user list                          # 敏感用户
```

### 漏洞管理 Vulnerability
```bash
cws cloudwalker vuln list                                   # 漏洞列表
cws cloudwalker vuln list --severity=high                  # 按等级筛选
cws cloudwalker vuln-info get --id=<vuln_id>               # 漏洞详情
cws cloudwalker vuln-info get-detail --id=<vuln_id>       # 漏洞详细信息

cws cloudwalker emergency-vuln-v1 list                      # 应急漏洞
cws cloudwalker patch-info list                             # 补丁情报
cws cloudwalker patch-info-event list                       # 补丁风险事件
```

### 威胁检测 Threat Detection
```bash
cws cloudwalker malware-event list                          # 恶意软件事件
cws cloudwalker malware-event list --state=open             # 未处理事件
cws cloudwalker malware-event mark-as-read --id=<id>       # 标记已读

cws cloudwalker webshell-event list                         # Webshell事件
cws cloudwalker webshell-event list --state=open

cws cloudwalker network-audit-event list                    # 网络异常事件
cws cloudwalker memory-webshell-event list                 # 内存马事件
cws cloudwalker revshell-event list                         # 反弹Shell事件
cws cloudwalker elevation-process-event list               # 提权进程事件

cws cloudwalker brute-force list                             # 暴力破解
cws cloudwalker brute-force list --state=open

cws cloudwalker suspicious-operation list                   # 可疑操作
cws cloudwalker tamper-proof list                           # 防篡改
cws cloudwalker honeypot list                              # 蜜罐诱捕
cws cloudwalker mimicry list                               # 拟态防护

cws cloudwalker non-white-process list                      # 非白名单进程
```

### 入侵检测 Intrusion Detection
```bash
cws cloudwalker abnormal-login-event list                 # 异常登录事件
cws cloudwalker abnormal-login-event list --state=open
cws cloudwalker abnormal-login-event create-whitelist      # 创建白名单
cws cloudwalker abnormal-login-event generate-firewall-rule # 生成防火墙规则

cws cloudwalker file-disposal list                         # 文件处置
cws cloudwalker process-kill list                          # 进程阻断
cws cloudwalker network-reject list                         # 网络阻断
```

### 探针管理 Agent Management
```bash
cws cloudwalker agent list                                 # 探针列表
cws cloudwalker agent get-agent-source                    # 探针来源
cws cloudwalker agent get-agent-group-tree                # 分组树
cws cloudwalker agent list-agent-ip-cfg                   # IP配置列表
cws cloudwalker agent create-agent-ip-cfg                 # 创建IP配置
cws cloudwalker agent update-agent-ip-cfg --id=<id>       # 更新IP配置
cws cloudwalker agent delete-agent-ip-cfg --id=<id>       # 删除IP配置
cws cloudwalker agent enable-agent-ip-cfg --id=<id>       # 启用IP配置

cws cloudwalker agent-detector list                        # 恶意文件探针
cws cloudwalker agent-detector get-agent-detector-cfg    # 探针配置
cws cloudwalker agent-detector install-detector-agent    # 安装探针
cws cloudwalker agent-detector upgrade-detector-agent    # 升级探针
cws cloudwalker agent-detector uninstall-detector-agent  # 卸载探针
cws cloudwalker agent-detector stop-detector-agent       # 停止探针
cws cloudwalker agent-detector reboot-detector-agent     # 重启探针

cws cloudwalker agent-module get-agent-list               # 探针模块列表
cws cloudwalker agent-module operate-module               # 操作模块
cws cloudwalker agent-module set-overload                # 设置过载配置
cws cloudwalker agent-module set-resource-limit           # 设置资源限制
cws cloudwalker agent-module set-running-mode-config     # 运行模式
cws cloudwalker agent-module set-auto-downgrade-config   # 自动降级配置
cws cloudwalker agent-module set-log-config              # 日志配置
cws cloudwalker agent-module set-net-conn-collect        # 网络连接采集
cws cloudwalker agent-module operate-veinmind-agent     # Veinmind探针操作
cws cloudwalker agent-module delete-module --id=<id>     # 删除模块

cws cloudwalker admin-agent apply-package                 # 应用更新包
cws cloudwalker admin-agent get-module-list               # 模块列表
cws cloudwalker admin-agent get-module-detail --id=<id>  # 模块详情
cws cloudwalker admin-agent delete-module --id=<id>      # 删除模块
```

### 安全策略 Security Strategy
```bash
cws cloudwalker security-strategy get-system-strategy    # 获取系统策略
cws cloudwalker security-strategy set-system-strategy      # 设置系统策略
cws cloudwalker security-strategy get-login-control       # 登录控制策略
cws cloudwalker security-strategy set-login-control        # 设置登录控制
cws cloudwalker security-strategy get-ldap               # LDAP策略
cws cloudwalker security-strategy set-ldap               # 设置LDAP
cws cloudwalker security-strategy get-radius             # Radius策略
cws cloudwalker security-strategy set-radius             # 设置Radius
cws cloudwalker security-strategy restore-system-strategy  # 恢复默认策略

cws cloudwalker baseline-v2 check                         # 基线检查
cws cloudwalker baseline-v2 get                           # 基线结果

cws cloudwalker security-check list                       # 安全检查
cws cloudwalker security-tool list                        # 安全工具
cws cloudwalker detection-rule list                       # 检测规则
```

### 告警与日志 Alert & Log
```bash
cws cloudwalker alert-config list                        # 告警配置
cws cloudwalker alert-config update --id=<id>            # 更新告警配置

cws cloudwalker audit-log list                           # 审计日志
cws cloudwalker log-collect list                         # 日志采集
cws cloudwalker full-command list                        # 全量命令记录

cws cloudwalker statistics overview                       # 统计概览
cws cloudwalker event-stat list                          # 事件统计
cws cloudwalker threat-overview list                      # 威胁概览
```

---

## SafeLine (雷池) WAF

### 统计 Statistics
```bash
cws safeline stats overview                              # 防护概览
cws safeline stats attack-types                         # 攻击类型统计
cws safeline stats top-events                           # TOP事件
cws safeline stats trend                                 # 趋势统计
```

### 日志 Logs
```bash
cws safeline logs event                                  # 拦截事件
cws safeline logs event --limit=100                     # 限制条数
cws safeline logs event --start-time=<timestamp>        # 时间筛选
cws safeline logs attack                                # 攻击日志
```

### 规则 Rules
```bash
cws safeline rule list                                   # 规则列表
cws safeline rule list --type=<rule_type>               # 按类型筛选
cws safeline rule get --id=<rule_id>                    # 规则详情
cws safeline rule update --id=<rule_id>                 # 更新规则
cws safeline rule enable --id=<rule_id>                 # 启用规则
cws safeline rule disable --id=<rule_id>                # 禁用规则
cws safeline rule create                                 # 创建规则

cws safeline policy-rule list                           # 策略规则
cws safeline policy-rule create                         # 创建策略规则
cws safeline policy-rule update --id=<id>              # 更新策略规则
```

### 站点与策略 Site & Policy
```bash
cws safeline site list                                  # 站点列表
cws safeline site get --id=<site_id>                   # 站点详情
cws safeline site create                                # 创建站点
cws safeline site update --id=<site_id>                # 更新站点

cws safeline policy-group list                          # 策略组
cws safeline policy-group create                        # 创建策略组
```

### ACL & IP
```bash
cws safeline acl list                                    # ACL列表
cws safeline acl create                                 # 创建ACL
cws safeline acl update --id=<id>                      # 更新ACL

cws safeline ip-group list                              # IP组
cws safeline ip-group create                            # 创建IP组
cws safeline ip-group update --id=<id>                 # 更新IP组
```

### 配置 Config
```bash
cws safeline config show                                 # 显示配置
cws safeline config update                               # 更新配置
cws safeline protection status                          # 防护状态
cws safeline system info                                 # 系统信息
cws safeline network list                               # 网络配置
```

---

## SafeLine-CE (社区版)

```bash
# 统计
cws safeline-ce stat overview                           # 统计概览

# 站点管理
cws safeline-ce site list                              # 站点列表
cws safeline-ce site get --id=<id>                    # 站点详情
cws safeline-ce site add                               # 添加站点
cws safeline-ce site update --id=<id>                 # 更新站点
cws safeline-ce site delete --id=<id>                 # 删除站点

# 日志
cws safeline-ce log attack list                        # 攻击日志
cws safeline-ce log event list                        # 事件日志

# 规则
cws safeline-ce rule list                              # 规则列表
cws safeline-ce rule add                               # 添加规则
cws safeline-ce rule update --id=<id>                 # 更新规则
cws safeline-ce rule delete --id=<id>                 # 删除规则

# IP管理
cws safeline-ce ipgroup list                           # IP组列表
cws safeline-ce ipgroup add                            # 添加IP组
cws safeline-ce ipgroup update --id=<id>              # 更新IP组

# 其他
cws safeline-ce cert list                              # 证书列表
cws safeline-ce module list                            # 语义分析模块
cws safeline-ce skynet list                           # SkyNet规则
```

---

## T-Answer (全悉)

```bash
cws tanswer firewall list                              # 防火墙规则
cws tanswer firewall create                             # 创建规则
cws tanswer firewall update --id=<id>                  # 更新规则

cws tanswer rules list                                 # 规则列表
cws tanswer rules create                               # 创建规则
cws tanswer rules update --id=<id>                     # 更新规则
```

---

## X-Ray (洞鉴)

### 扫描任务 Scan Plans
```bash
# 创建快速扫描
cws xray plan PostPlanCreateQuick \
  --targets=<target> \
  --engines=00000000000000000000000000000001 \
  --project-id=1 \
  --name=<scan_name>

# 任务管理
cws xray plan PostPlanFilter                           # 列出所有任务
cws xray plan PostPlanFilter --limit=50               # 限制条数
cws xray plan GetPlanID --id=<plan_id>                # 任务详情
cws xray plan PostPlanExecute --executePlanBody.id=<id> # 执行/恢复
cws xray plan PostPlanStop --stopPlanBody.id=<id>      # 停止
cws xray plan DeletePlanID --id=<plan_id>             # 删除任务
```

### 资产管理 Assets
```bash
cws xray asset_property list                           # 资产属性
cws xray domain_asset list                             # 域名资产
cws xray ip_asset list                                 # 主机资产
cws xray service_asset list                            # 服务资产
cws xray web_asset list                               # Web资产
cws xray vulnerability list                             # 漏洞资产
cws xray vulnerability get --id=<vuln_id>            # 漏洞详情
```

### 基线检查 Baseline
```bash
cws xray baseline list                                 # 基线列表
cws xray baseline get --id=<id>                        # 基线详情
cws xray baseline check --id=<id>                      # 执行检查
```

### POC管理 Custom POC
```bash
cws xray custom_poc list                               # POC列表
cws xray custom_poc create                             # 创建POC
cws xray custom_poc update --id=<id>                   # 更新POC
cws xray custom_poc delete --id=<id>                   # 删除POC
```

### 报表与结果 Report & Result
```bash
cws xray report list                                   # 报告列表
cws xray report get --id=<report_id>                  # 报告详情
cws xray report generate --plan-id=<id>              # 生成报告

cws xray result list                                   # 结果列表
cws xray result get --id=<result_id>                  # 结果详情
cws xray result export --id=<result_id>                # 导出结果
```

### 系统管理 System
```bash
cws xray system_info get                              # 系统信息
cws xray system_service list                          # 系统服务
cws xray user list                                     # 用户列表
cws xray role list                                     # 角色列表
cws xray project list                                 # 工作区列表
cws xray template list                                # 策略模板
cws xray audit_log list                               # 审计日志
```

---

## X-Ray 引擎 ID

| 引擎 | ID |
|-----|-----|
| 基础扫描 | 00000000000000000000000000000001 |
| SQL注入 | 00000000000000000000000000000002 |
| XSS | 00000000000000000000000000000004 |
| 命令注入 | 00000000000000000000000000000008 |
| 目录穿越 | 00000000000000000000000000000010 |
| 组合多个引擎用 `\|`，如 `00000000000000000000000000000001\|00000000000000000000000000000002` |

---

## 全局参数 Global Flags

| 参数 | 说明 |
|-----|------|
| `-c, --config <path>` | 配置文件路径（默认 config.yaml） |
| `--dry-run` | 仅验证，不发送请求 |
| `-f, --format <format>` | 输出格式：json, text |
| `--no-trunc` | 不截断输出 |
| `--insecure` | 跳过TLS证书验证 |
| `--limit <n>` | 限制返回条数 |
| `--offset <n>` | 分页偏移 |
