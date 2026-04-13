---
name: chaitin-cws
description: "Use when working with Chaitin security products via cws CLI: CloudWalker (牧云), SafeLine (雷池), SafeLine-CE (雷池社区版), T-Answer (全悉), and X-Ray (洞鉴). Supports dynamic command discovery via --help."
version: 1.1.0
author: chaitin
tags: [chaitin, cws, cli, cloudwalker, safeline, safeLine, tanswer, xray, security, vulnerability-scan, waf, asset-management]
---

# Chaitin Security Products CLI (CWS)

> Unified CLI tool for Chaitin security products. One binary, multiple products.

## Install & Run

**自动安装**：

1. 检测系统平台（OS + ARCH）
2. 从 GitHub Releases 获取最新版本
3. 下载对应平台的压缩包并解压到任意 PATH 目录
4. 添加执行权限

```bash
# 示例：macOS ARM64 安装到 ~/.local/bin/
curl -fsSL https://github.com/chaitin/workspace-cli/releases/latest/download/cws_v{version}_darwin_arm64.tar.gz | tar -xzf - -C ~/.local/bin/
chmod +x ~/.local/bin/cws
```

验证：`cws --help`

完整安装命令由 Agent 根据上述步骤自动生成执行。

## Products

| Product | Command | Description |
|---------|---------|-------------|
| CloudWalker | `cws cloudwalker` | 牧云, 云工作负载保护平台 |
| SafeLine | `cws safeline` | 雷池, 下一代WAF |
| SafeLine-CE | `cws safeline-ce` | 雷池, 社区版 |
| T-Answer | `cws tanswer` | 全悉, 威胁感知与响应 |
| X-Ray | `cws xray` | 洞鉴, 漏洞扫描与风险评估 |

## Command Discovery

Always use `--help` to discover commands dynamically:

```bash
cws <product> --help                  # List subcommands
cws <product> <subcommand> --help    # List sub-subcommands
cws <product> <sub> <action> --help  # Full parameters
```

## Quick Commands Reference

See [commands.md](./commands.md) for detailed command reference by product.

### CloudWalker
```bash
cws cloudwalker host-asset list                    # 主机资产
cws cloudwalker vuln list                         # 漏洞列表
cws cloudwalker malware-event list                 # 恶意软件事件
cws cloudwalker baseline-v2 check                  # 基线检查
cws cloudwalker agent list                         # 探针列表
```

### SafeLine
```bash
cws safeline stats overview                        # 防护概览
cws safeline logs event                            # 拦截事件
cws safeline rule list                            # 规则列表
```

### X-Ray
```bash
cws xray plan PostPlanCreateQuick --targets=<target> --engines=00000000000000000000000000000001  # 快速扫描
cws xray plan PostPlanFilter                       # 任务列表
```

## Configuration

```yaml
# config.yaml
cloudwalker:
  url: https://your-cloudwalker.example.com/rpc
  api_key: YOUR_API_KEY
safeline:
  url: https://your-safeline.example.com
  api_key: YOUR_API_KEY
xray:
  url: https://your-xray.example.com/api/v2
  api_key: YOUR_API_KEY
tanswer:
  url: https://your-tanswer.example.com
  api_key: YOUR_API_KEY
safeline-ce:
  url: https://your-safeline-ce.example.com:9443
  api_key: YOUR_API_KEY
```

Or use environment variables: `CLOUDWALKER_URL`, `CLOUDWALKER_API_KEY`, etc.

**Priority**: Flags > Env > config.yaml

## Output Formats

```bash
cws cloudwalker host-asset list          # Text (default)
cws cloudwalker host-asset list -f json # JSON
cws cloudwalker host-asset list --no-trunc  # Full content
```

## Global Flags

| Flag | Description |
|------|-------------|
| `-c, --config <path>` | Config file path |
| `--dry-run` | Validate without making changes |
| `-f, --format <format>` | Output format (json/text) |
| `--insecure` | Skip TLS verification |

## See Also

- [commands.md](./commands.md) - Full command reference by product with Quick Lookup tables
