---
name: chaitin-cli
description: "Use when running chaitin-cli commands to manage Chaitin security products: SafeLine WAF (site management, IP blocking, ACL, policy rules, attack logs), X-Ray vulnerability scanner (scan tasks, results, assets), CodeInsight (projects, repository configs, scan tasks, reports), CodeForce (projects, AI tasks, denoise, repositories), CloudWalker CWPP (events, vulnerabilities, assets), and T-Answer (semantic security operations for alarms, assets, policies, response actions, and Open API fallback)."
version: 1.0.0
author: chaitin
tags: [chaitin-cli, safeline, xray, codeinsight, codeforce, cloudwalker, tanswer, waf, security, chaitin, cli, ddr, veinmind]
---

# chaitin-cli Usage Guide

> Unified CLI for Chaitin security products. Manage SafeLine WAF, DDR, X-Ray scanner, CodeInsight, CodeForce, CloudWalker CWPP, VeinMind container security,  and T-Answer through a single tool.

## No-Argument Behavior

When `/chaitin-cli` is invoked without any arguments (empty `ARGUMENTS`):

1. Greet the user and introduce this skill in a well-formatted way, based on the SKILL.md content.
2. Run `command -v chaitin-cli` to check if `chaitin-cli` is already installed.
3. If found, report the installed path (e.g. `chaitin-cli is installed at /opt/homebrew/bin/chaitin-cli`).
4. If not found, install it per platform:
   - Windows: Tell the user to manually download the latest release from `https://github.com/chaitin/chaitin-cli/releases`, extract `chaitin-cli.exe`, and add it to PATH. Do not attempt automated installation on Windows.
   - macOS, Linux: Run `bash scripts/install-chaitin-cli.sh`. The script outputs the installed binary path on stdout (last line). Remember this path — subsequent commands must use the full path (e.g. `/home/user/.local/bin/chaitin-cli`) because each Bash invocation starts a new shell and the install directory may not yet be in PATH.
5. After the setup check, briefly tell the user what they can do next — for example: "You can now use chaitin-cli to manage SafeLine, DDR, X-Ray, CodeInsight, CodeForce, CloudWalker, VeinMind, or T-Answer. Tell me what you'd like to do, or run `chaitin-cli --help` to explore commands."

## Tool Resolution

When this skill needs `chaitin-cli`, do not run a preflight availability check before every command.

1. Run the requested `chaitin-cli ...` command directly.
2. Only if the shell reports `command not found`, `No such file or directory`, or exit code `127` because `chaitin-cli` is missing, install it per platform:
   - Windows: Tell the user to manually download the latest release from `https://github.com/chaitin/chaitin-cli/releases`, extract `chaitin-cli.exe`, and add it to PATH. Do not attempt automated installation on Windows.
   - macOS, Linux: Run `bash scripts/install-chaitin-cli.sh`. The script outputs the installed binary path on stdout (last line, e.g. `/home/user/.local/bin/chaitin-cli`). Remember this path for the rest of the session.
3. After installation, run subsequent `chaitin-cli` commands using the full installed path (e.g. `/home/user/.local/bin/chaitin-cli safeline site list`) instead of bare `chaitin-cli`, because each Bash invocation starts a new shell and the install directory may not yet be in PATH.
4. If `chaitin-cli` already exists, do not query GitHub Releases, do not reinstall, and do not do version checks unless the user explicitly asks.

The installer detects the current OS/architecture, downloads the latest matching `chaitin-cli` release archive from `https://github.com/chaitin/chaitin-cli/releases`, and installs it as a directly runnable `chaitin-cli`, preferring a user PATH directory and falling back to system-wide install when possible.

> **Windows**: There is no automated installer for Windows. Download the latest release from `https://github.com/chaitin/chaitin-cli/releases`, extract `chaitin-cli.exe`, and add it to PATH manually.

## Install & Run

```bash
# On-demand installer used only when `chaitin-cli` is missing
# macOS / Linux
bash scripts/install-chaitin-cli.sh

# The installer fetches the matching GitHub release package
# and installs the extracted binary as `chaitin-cli`.

# Windows: download the latest release manually from
# https://github.com/chaitin/chaitin-cli/releases
# Extract chaitin-cli.exe and add it to PATH.

# Or build from source
git clone https://github.com/chaitin/chaitin-cli.git
cd chaitin-cli
go build -o chaitin-cli .

# Run
chaitin-cli <product> <command> [flags]
```

## Prerequisites

Before running any `chaitin-cli` command:

1. **Network reachability** — the machine running `chaitin-cli` must be able to reach each product's console / API endpoint.
2. **API key** — generate one from each product's UI (SafeLine → System → API Token; X-Ray → System Settings → API Key; etc.) and supply it via `--api-key`, the product env var, or `config.yaml`.
3. **TLS with self-signed certs** — `chaitin-cli xray` takes `--insecure` (off by default). `chaitin-cli safeline` also exposes `--insecure`, but its default is `true` (already skipping verification); pass `--insecure=false` to re-enable verification. `chaitin-cli tanswer` exposes `--insecure` for environments that require skipping TLS certificate verification. Other products differ, so check product help when unsure.
4. **Build from source** — Go 1.25+ (see `go.mod`). Otherwise use the pre-built binary from GitHub Releases.

## Configuration

For a project-specific setup, create a recognized `./config.yaml` in the working directory (it must contain at least one chaitin-cli product name such as `tanswer`). Otherwise, place the same configuration in `~/.chaitin-cli/config.yaml`, which is the default configuration path:
For T-Answer, use `--url`, `--api-key`, `--timeout`, and `--insecure`; `TANSWER_URL`, `TANSWER_API_KEY`, `TANSWER_TIMEOUT`, and `TANSWER_INSECURE`; or `tanswer.url`, `tanswer.api_key`, `tanswer.timeout`, and `tanswer.insecure` in the shared configuration file.


```yaml
safeline:
  url: https://your-safeline-server
  api_key: YOUR_API_KEY

xray:
  url: https://your-xray-server/api/v2
  api_key: YOUR_API_KEY

cloudwalker:
  url: https://your-cloudwalker-server/rpc
  api_key: YOUR_API_KEY

ddr:
  url: https://your-ddr-server/qzh/api/v1
  api_key: YOUR_API_KEY

veinmind:
  url: https://your-veinmind-server
  api_key: YOUR_64_CHARACTER_API_TOKEN

tanswer:
  url: 'https://<全悉 Web 端 IP>'
  api_key: '<全悉 OpenAPI Token>'
  timeout: 30s
  insecure: false

codeinsight:
  url: https://your-codeinsight-server
  access_token: YOUR_ACCESS_TOKEN
```

Or use environment variables / `.env` file:

```bash
SAFELINE_URL=https://your-safeline-server
SAFELINE_API_KEY=YOUR_API_KEY
XRAY_URL=https://your-xray-server/api/v2
XRAY_API_KEY=YOUR_API_KEY
DDR_URL=https://your-ddr-server/qzh/api/v1
DDR_API_KEY=YOUR_API_KEY
CODEINSIGHT_URL=https://your-codeinsight-server
CODEINSIGHT_TOKEN=YOUR_ACCESS_TOKEN
TANSWER_URL='https://<全悉 Web 端 IP>'
TANSWER_API_KEY='<全悉 OpenAPI Token>'
TANSWER_TIMEOUT=30s
TANSWER_INSECURE=false
```

Priority: `flags > environment/.env > recognized ./config.yaml > ~/.chaitin-cli/config.yaml`. A local `./config.yaml` that does not contain a recognized chaitin-cli product is ignored, so it will not accidentally override another project's configuration.

Use `-c` to switch between config files (e.g., multiple environments):

```bash
chaitin-cli -c ./configs/prod.yaml safeline stats overview
chaitin-cli -c ./configs/staging.yaml safeline stats overview
```

### Global Flags

| Flag | Description |
|------|-------------|
| `-c, --config` | Explicit config file path. Without this flag, CLI tries recognized `./config.yaml` first, then `~/.chaitin-cli/config.yaml`. |
| `--dry-run` | Print the API request without executing when the product honors it. Applied by the root command to `xray`, `cloudwalker`, and `veinmind`. `safeline` registers its own `--dry-run` and forwards it to subcommands. `safeline-ce` inherits the root flag, but the current codebase stores the value without using it; `tanswer` ignores it. |

### Discovering Commands

`--help` is the authoritative source — this document does not enumerate every flag.

```bash
chaitin-cli <product> --help                # List subcommand groups for a product
chaitin-cli <product> <group> --help        # List commands in a group
chaitin-cli <product> <group> <cmd> --help  # List flags for a specific command
```

`chaitin-cli xray` commands are auto-generated from the X-Ray OpenAPI spec (hundreds of operations); `chaitin-cli xray <category> --help` is the only complete reference. `chaitin-cli cloudwalker` has 60+ command groups with similar depth. `chaitin-cli veinmind` is generated from the VeinMind OpenAPI spec; always confirm leaf flags with `chaitin-cli veinmind <group> <cmd> --help`.

### Operating Rules

For SafeLine, X-Ray, DDR, CodeInsight, CloudWalker, VeinMind, T-Answer, and SafeLine-CE tasks, treat `chaitin-cli` as the only supported operator interface.

- Prefer `chaitin-cli ... --help` and existing `chaitin-cli` subcommands over `curl`, ad-hoc HTTP requests, browser debugging, or guessed endpoints.
- If `chaitin-cli` does not expose the requested product operation, stop and say that the current CLI does not support it. Do not fall back to direct API calls just to "try it".
- Do not use `curl` or raw HTTP requests to perform state-changing or potentially dangerous product operations that are not implemented by `chaitin-cli`.
- Use source inspection to confirm command availability and behavior, not to bypass the CLI and reconstruct private API calls.
- When a supported command may change state and the product actually honors `--dry-run`, prefer checking that path first.

### Output Formats

Each product uses its own output convention — there is no unified `-f` / `--format` flag across `chaitin-cli`.

| Product | Default | Switch to JSON | Other |
|---------|---------|----------------|-------|
| `chaitin-cli safeline` | table | `--indent` | — |
| `chaitin-cli safeline-ce` | table | `-o json` (or `--output json`) | `--verbose` |
| `chaitin-cli ddr` | JSON | `-o json` is already the default | `-o table` for quick manual reading; `-v` for request debug |
| `chaitin-cli xray` | JSON (no alternative) | — | `--debug` for debug logs |
| `chaitin-cli cloudwalker` | text | `-f json` (or `--format json`) | `--no-trunc` to disable text truncation |
| `chaitin-cli veinmind` | table | `-o json` (or `--output json`) | `-v` for request debug; `--dry-run` prints request summary |
| `chaitin-cli tanswer` | JSON | default; no format switch required | `--insecure` for TLS certificate verification bypass |

When piping into `jq`, note that SafeLine uses `--indent` (not `-o`/`-f`), and T-Answer outputs JSON by default.

---

## Quick Lookup by Capability

Pick by task, not by product name. Items are listed most- to least-common.

| Task | Command path |
|------|--------------|
| Block/allow IP, rate-limit, manual ACL | `safeline acl` · `safeline ip-group` · `safeline-ce rule` |
| Add a custom rule on URL path / header / body | `safeline policy-rule` · `safeline-ce rule` |
| Manage protected sites / web services | `safeline site` · `safeline-ce site` |
| Query attack / access / rate-limit logs | `safeline log` · `safeline-ce log` |
| Enable detection modules (SQLi, XSS, …) | `safeline policy-group` · `safeline-ce module` · `safeline-ce skynet` |
| Launch / stop a vulnerability scan | `xray plan` |
| Query scan results, vulns, generate reports | `xray result` · `xray vulnerability` · `xray report` |
| Manage CodeInsight projects and scan tasks | `codeinsight project` · `codeinsight task` · `codeinsight repo-config` |
| Manage CodeForce projects, AI tasks, repositories, and denoise workflows | `codeforce project` · `codeforce audit` · `codeforce denoise` · `codeforce repository` · `codeforce git-auth` |
| Asset inventory (web / domain / IP) | `xray web_asset` · `xray domain_asset` · `xray ip_asset` |
| Baseline / compliance check | `xray baseline` · `cloudwalker baseline_v2` |
| Container security agent management | `veinmind agent` |
| Container / image / Kubernetes inventory | `veinmind img` · `veinmind container` · `veinmind cluster` · `veinmind host` |
| Container software / website / web framework inventory | `veinmind app` · `veinmind website` · `veinmind web-framework` |
| Container security baseline / compliance events | `veinmind baseline` |
| Container runtime and image risk events | `veinmind risk` |
| Host-level event response (webshell, reverse shell, brute force) | `cloudwalker webshell_event` · `cloudwalker revshell_event` · `cloudwalker brute_force` |
| Host asset inventory (process / port / container / user) | `cloudwalker process_asset` · `cloudwalker port_asset` · `cloudwalker docker_container` · `cloudwalker user_asset` |
| Ransomware protection, file quarantine, kill process | `cloudwalker anti_ransomware` · `cloudwalker file_disposal` · `cloudwalker process_kill` |
| Host firewall / network block | `cloudwalker firewall` · `cloudwalker network_reject` |
| DDR device inventory / file scan tasks | `ddr device list` · `ddr device filescantask` · `ddr device filescantask list` |
| DDR approvals / policy logs | `ddr disposal approvalinstance list` · `ddr policylog channel list` · `ddr policylog softwarenetwork list` |
| DDR behavior control policies | `ddr policy channel` · `ddr policy landing` · `ddr policy email` · `ddr policy codecontrol` · `ddr policy clipboard` · `ddr policy webpost-control` |
| DDR device uninstall | `ddr device status-action --operation uninstall` |
| Traffic threat detection and response operations | `tanswer alarm` · `tanswer file-alarm` · `tanswer asset` · `tanswer policy` · `tanswer response` |
| System info / license management | `safeline system` · `safeline-ce cert info/get` · `xray system_info` · `xray system_service PostSystemLicense` |

---

## SafeLine (雷池 WAF)

### Global Flags (SafeLine)

| Flag | Env Var | Description |
|------|---------|-------------|
| `--url` | `SAFELINE_URL` | SafeLine Skyview API address (required) |
| `--api-key` | `SAFELINE_API_KEY` | API token |
| `--indent` | — | Output JSON (pretty-printed) instead of the default table. SafeLine does not expose a separate `-o/--output` flag — this is the only way to switch format. |
| `--insecure` | — | Skip TLS certificate verification. **Default: `true`** — SafeLine already skips verification out of the box. Pass `--insecure=false` to re-enable verification. |

---

### Complete Workflow: Responding to an Attack

This walkthrough covers a typical incident response flow — from spotting an attack to blocking the attacker.

#### Step 1: Check the dashboard

```bash
# View last 24 hours stats
chaitin-cli safeline stats overview --duration h

# View last 30 days stats
chaitin-cli safeline stats overview --duration d
```

#### Step 2: List your protected sites

```bash
chaitin-cli safeline site list
```

#### Step 3: View recent attack logs

```bash
# List the latest 20 attack events
chaitin-cli safeline log detect list --count 20

# Get full details of a specific event
chaitin-cli safeline log detect get \
  --event-id "6edb4c7eb69042cd996045e3ee5526d9" \
  --timestamp "1774857841"
```

#### Step 4: Block the attacker's IP

**Option A — Create an IP group and block it with an ACL rule:**

```bash
# Create an IP group for malicious IPs
chaitin-cli safeline ip-group create \
  --name "Blocklist" \
  --ips "203.0.113.42,198.51.100.7" \
  --comment "Attackers from incident 2024-01"

# Create an ACL template that forbids the group
chaitin-cli safeline acl template create \
  --name "Block Malicious IPs" \
  --template-type manual \
  --target-type cidr \
  --action forbid \
  --ip-groups <group-id>
```

**Option B — Block specific IPs directly without a group:**

```bash
chaitin-cli safeline acl template create \
  --name "Emergency Block" \
  --template-type manual \
  --target-type cidr \
  --action forbid \
  --targets "203.0.113.42,198.51.100.7"
```

#### Step 5: Add a custom rule to block a malicious path

```bash
# Block requests to /admin/upload with high risk level
chaitin-cli safeline policy-rule create \
  --comment "Block malicious upload path" \
  --target urlpath \
  --cmp infix \
  --value "/admin/upload" \
  --action deny \
  --risk-level 3
```

#### Step 6: Verify detection modules are enabled

```bash
# Check the policy group
chaitin-cli safeline policy-group list
chaitin-cli safeline policy-group get <id>

# Enable SQL injection and XSS detection
chaitin-cli safeline policy-group update <id> \
  --module m_sqli,m_xss \
  --state enabled
```

#### Step 7: Monitor access logs

```bash
chaitin-cli safeline log access list --count 50
chaitin-cli safeline log access get \
  --event-id "1e1ef8e9b21d42cd996045e3ee5526d9" \
  --req-start-time "1775117700"
```

#### Step 8: Unblock a false positive

```bash
# List active ACL rules (blocked IPs)
chaitin-cli safeline acl rule list --template-id <template-id>

# Remove the block and add IP to whitelist
chaitin-cli safeline acl rule delete <rule-id> --add-to-whitelist

# Or clear all rules for a template
chaitin-cli safeline acl rule clear --template-id <template-id>
```

---

### SafeLine Command Reference

#### stats

```bash
chaitin-cli safeline stats overview --duration h   # 24h stats
chaitin-cli safeline stats overview --duration d   # 30d stats
```

#### site

```bash
chaitin-cli safeline site list                                    # List all sites
chaitin-cli safeline site get <id>                                # Get site details
chaitin-cli safeline site enable <id>                             # Enable a site
chaitin-cli safeline site disable <id>                            # Disable a site
chaitin-cli safeline site update <id> --policy-group <group-id>  # Attach a policy group to a site
chaitin-cli safeline site update <id> --policy-group 0            # Detach policy group from a site
```

#### ip-group (alias: ipgroup)

```bash
chaitin-cli safeline ip-group list                                              # List all IP groups
chaitin-cli safeline ip-group list --name "office" --count 50 --offset 0       # Filter by name with pagination
chaitin-cli safeline ip-group get <id>                                          # Get IP group details
chaitin-cli safeline ip-group create --name "DC" --ips "172.16.0.0/16" --comment "Data center"  # Create a new IP group
chaitin-cli safeline ip-group delete <id>                                       # Delete an IP group
chaitin-cli safeline ip-group delete 1 2 3                                      # Batch delete IP groups
chaitin-cli safeline ip-group add-ip <id> --ips "10.0.1.0/24"                  # Add IPs to an IP group
chaitin-cli safeline ip-group remove-ip <id> --ips "10.0.1.0/24"               # Remove IPs from an IP group
```

#### acl template

```bash
chaitin-cli safeline acl template list                # List all ACL templates
chaitin-cli safeline acl template list --name "limit" # Filter templates by name
chaitin-cli safeline acl template get <id>            # Get ACL template details
chaitin-cli safeline acl template enable <id>         # Enable an ACL template
chaitin-cli safeline acl template disable <id>        # Disable an ACL template
chaitin-cli safeline acl template delete <id>         # Delete an ACL template

# Create manual block rule (specific IPs)
chaitin-cli safeline acl template create \
  --name "Block IPs" --template-type manual \
  --target-type cidr --action forbid \
  --targets "192.168.1.100,10.0.0.50"

# Create auto rate-limit rule
chaitin-cli safeline acl template create \
  --name "Rate Limit" --template-type auto \
  --period 60 --limit 100 --action forbid

# Create throttle rule (allow but slow down)
chaitin-cli safeline acl template create \
  --name "Throttle" --template-type auto \
  --period 60 --limit 100 \
  --action limit_rate \
  --limit-rate-limit 10 --limit-rate-period 60
```

#### acl rule (blocked IP entries)

```bash
chaitin-cli safeline acl rule list --template-id <id>                    # List blocked IP entries for a template
chaitin-cli safeline acl rule delete <id>                                # Delete a blocked IP entry
chaitin-cli safeline acl rule delete <id> --add-to-whitelist             # Delete and move IP to whitelist
chaitin-cli safeline acl rule clear --template-id <id>                   # Clear all blocked IP entries for a template
chaitin-cli safeline acl rule clear --template-id <id> --add-to-whitelist # Clear all and move IPs to whitelist
```

#### policy-group

```bash
chaitin-cli safeline policy-group list                                              # List all policy groups
chaitin-cli safeline policy-group get <id>                                          # Get policy group details
chaitin-cli safeline policy-group update <id> --module m_sqli,m_xss --state enabled  # Enable detection modules
chaitin-cli safeline policy-group update <id> --module m_cmd_injection --state disabled # Disable a detection module
```

Available modules: `m_sqli` `m_xss` `m_cmd_injection` `m_file_include` `m_file_upload` `m_php_code_injection` `m_php_unserialize` `m_java` `m_java_unserialize` `m_ssrf` `m_ssti` `m_csrf` `m_scanner` `m_response` `m_rule`

#### policy-rule

```bash
chaitin-cli safeline policy-rule list                  # List all policy rules (global by default)
chaitin-cli safeline policy-rule list --global=false   # List site-specific rules only
chaitin-cli safeline policy-rule get <id>              # Get policy rule details
chaitin-cli safeline policy-rule enable <id>           # Enable a policy rule
chaitin-cli safeline policy-rule disable <id>          # Disable a policy rule
chaitin-cli safeline policy-rule delete <id>           # Delete a policy rule

# Create simple rule
chaitin-cli safeline policy-rule create \
  --comment "Block /admin" \
  --target urlpath --cmp infix --value "/admin" \
  --action deny --risk-level 3

# List available targets and operators
chaitin-cli safeline policy-rule targets
chaitin-cli safeline policy-rule targets --cmp urlpath

# Actions: deny | dry_run | allow
# Risk levels: 0=none 1=low 2=medium 3=high
```

#### log

```bash
# Attack logs
chaitin-cli safeline log detect list --count 50
chaitin-cli safeline log detect list --current-page 1 --target-page 2
chaitin-cli safeline log detect get --event-id "<id>" --timestamp "<ts>"

# Access logs
chaitin-cli safeline log access list --count 50
chaitin-cli safeline log access get --event-id "<id>" --req-start-time "<ts>"

# Rate-limit logs (alias: rl)
chaitin-cli safeline log rate-limit list --count 50 --offset 0
```

#### system

```bash
chaitin-cli safeline system license                      # Get license information
chaitin-cli safeline system machine-id                   # Get machine ID (for license activation)
chaitin-cli safeline system log list --count 50 --offset 0  # List system operation logs
```

#### network (hardware mode only)

```bash
chaitin-cli safeline network workgroup list          # alias: wg list
chaitin-cli safeline network workgroup get <name>    # alias: wg get
chaitin-cli safeline network interface list          # alias: if list
chaitin-cli safeline network interface ip <name>     # alias: if ip
chaitin-cli safeline network gateway get             # alias: gw get
chaitin-cli safeline network route list              # alias: sr list
```

---

## X-Ray (洞鉴 Vulnerability Scanner)

### Global Flags (X-Ray)

| Flag | Env Var | Description |
|------|---------|-------------|
| `--url` | `XRAY_URL` | X-Ray API address (required) |
| `--api-key` | `XRAY_API_KEY` | API token |
| `--debug` | — | Enable debug logging |
| `--insecure` | — | Skip TLS certificate verification |

### Basic Commands

```bash
# Quick scan (create and immediately execute a task)
chaitin-cli xray plan PostPlanCreateQuick \
  --targets=10.3.0.4,10.3.0.5 \
  --engines=<engine-id> \
  --project-id=1

# List scan tasks
chaitin-cli xray plan PostPlanFilter \
  --filterPlan.limit=10 \
  --filterPlan.offset=0

# Stop a scan task
chaitin-cli xray plan PostPlanStop --stopPlanBody.id=<id>

# Resume a scan task
chaitin-cli xray plan PostPlanExecute --executePlanBody.id=<id>

# Delete a scan task
chaitin-cli xray plan DeletePlanID --id=<id>
```

### Command Categories

| Command | Description |
|---------|-------------|
| `chaitin-cli xray asset_property` | Asset management |
| `chaitin-cli xray audit_log` | Audit log management |
| `chaitin-cli xray baseline` | Baseline check management |
| `chaitin-cli xray custom_poc` | Custom POC management |
| `chaitin-cli xray domain_asset` | Domain asset management |
| `chaitin-cli xray insight` | Data insight and analytics |
| `chaitin-cli xray ip_asset` | IP/host asset management |
| `chaitin-cli xray plan` | Scan task management |
| `chaitin-cli xray project` | Project management |
| `chaitin-cli xray report` | Report management |
| `chaitin-cli xray result` | Scan result management |
| `chaitin-cli xray role` | Role management |
| `chaitin-cli xray service_asset` | Service asset management |
| `chaitin-cli xray system_info` | System information |
| `chaitin-cli xray system_service` | System service management |
| `chaitin-cli xray task_config` | Task configuration management |
| `chaitin-cli xray template` | Policy template management |
| `chaitin-cli xray user` | User management |
| `chaitin-cli xray vulnerability` | Vulnerability management |
| `chaitin-cli xray web_asset` | Web asset management |
| `chaitin-cli xray xprocess` | XProcess task instance management |
| `chaitin-cli xray xprocess_lite` | XProcess lite management |

---

## CloudWalker (云溯 CWPP)

### Global Flags (CloudWalker)

| Flag | Env Var | Description |
|------|---------|-------------|
| `--url` | `CLOUDWALKER_URL` | CloudWalker RPC address (required) |
| `--api-key` | `CLOUDWALKER_API_KEY` | API key |

> Note: CloudWalker does not expose `--insecure`, but its HTTP client always sets `InsecureSkipVerify: true`, so self-signed certs just work — no CA install or HTTP fallback needed.

### Command Categories

Each category has subcommands — run `chaitin-cli cloudwalker <category> --help` to list them.

#### Security Events

| Command | Description |
|---------|-------------|
| `chaitin-cli cloudwalker abnormal_login_event` | Abnormal login events |
| `chaitin-cli cloudwalker brute_force` | Brute-force events |
| `chaitin-cli cloudwalker elevation_process_event` | Privilege escalation process events |
| `chaitin-cli cloudwalker event_stat` | Event management and statistics |
| `chaitin-cli cloudwalker full_command` | Full command execution records |
| `chaitin-cli cloudwalker honeypot` | Honeypot trap events |
| `chaitin-cli cloudwalker malware_event` | Malware events |
| `chaitin-cli cloudwalker memory_webshell_event` | In-memory webshell events |
| `chaitin-cli cloudwalker network_audit_event` | Network anomaly events |
| `chaitin-cli cloudwalker non_white_process` | Non-whitelisted process events |
| `chaitin-cli cloudwalker revshell_event` | Reverse shell events |
| `chaitin-cli cloudwalker suspicious_operation` | Suspicious operation events |
| `chaitin-cli cloudwalker webshell_event` | Webshell events |

#### Asset Inventory

| Command | Description |
|---------|-------------|
| `chaitin-cli cloudwalker application_asset` | Application assets |
| `chaitin-cli cloudwalker asset_cert` | Certificate assets |
| `chaitin-cli cloudwalker asset_config` | Asset collection configuration |
| `chaitin-cli cloudwalker asset_crontab` | Scheduled task assets |
| `chaitin-cli cloudwalker asset_env` | Environment variable assets |
| `chaitin-cli cloudwalker asset_registry` | Registry assets |
| `chaitin-cli cloudwalker asset_startup` | Startup item assets |
| `chaitin-cli cloudwalker docker_container` | Docker container assets |
| `chaitin-cli cloudwalker docker_image` | Docker image assets |
| `chaitin-cli cloudwalker docker_network` | Docker network assets |
| `chaitin-cli cloudwalker host_asset` | Host assets (includes agent management) |
| `chaitin-cli cloudwalker host_discovery` | Unknown host discovery |
| `chaitin-cli cloudwalker host_nic_asset` | Network interface card assets |
| `chaitin-cli cloudwalker host_partition_asset` | Partition assets |
| `chaitin-cli cloudwalker host_route_asset` | Route assets |
| `chaitin-cli cloudwalker port_asset` | Port assets |
| `chaitin-cli cloudwalker process_asset` | Process assets |
| `chaitin-cli cloudwalker user_asset` | User assets |
| `chaitin-cli cloudwalker website_asset` | Website assets |

#### Security Protection

| Command | Description |
|---------|-------------|
| `chaitin-cli cloudwalker anti_ransomware` | Anti-ransomware protection |
| `chaitin-cli cloudwalker baseline_v2` | Baseline check management |
| `chaitin-cli cloudwalker detection_rule` | Detection rule management |
| `chaitin-cli cloudwalker file_disposal` | File disposal (quarantine/delete) |
| `chaitin-cli cloudwalker firewall` | Firewall rule management |
| `chaitin-cli cloudwalker mimicry` | Mimicry defense |
| `chaitin-cli cloudwalker network_reject` | Network block management |
| `chaitin-cli cloudwalker port_scan` | Port scan protection |
| `chaitin-cli cloudwalker process_kill` | Process termination |
| `chaitin-cli cloudwalker security_check` | Security checks |
| `chaitin-cli cloudwalker sensitive_file` | Sensitive file management |
| `chaitin-cli cloudwalker sensitive_file_scan` | Sensitive file scanning |
| `chaitin-cli cloudwalker sensitive_port` | Sensitive port management |
| `chaitin-cli cloudwalker sensitive_user` | Sensitive user management |
| `chaitin-cli cloudwalker tamper_proof` | File tamper-proof protection |
| `chaitin-cli cloudwalker vuln` | Vulnerability management |
| `chaitin-cli cloudwalker weak_passwd` | Weak password detection |
| `chaitin-cli cloudwalker whitelist` | Whitelist rule management |

#### Platform Management

| Command | Description |
|---------|-------------|
| `chaitin-cli cloudwalker admin_agent` | Agent module update management |
| `chaitin-cli cloudwalker admin_monitor` | System monitoring management |
| `chaitin-cli cloudwalker admin_strategy` | Strategy management |
| `chaitin-cli cloudwalker agent` | Agent management |
| `chaitin-cli cloudwalker agent_detector` | Malicious file agent management |
| `chaitin-cli cloudwalker agent_module` | Agent module management |
| `chaitin-cli cloudwalker alert_config` | Alert configuration |
| `chaitin-cli cloudwalker audit_log` | Audit log |
| `chaitin-cli cloudwalker business_group` | Business group management |
| `chaitin-cli cloudwalker crontab` | Scheduled task management |
| `chaitin-cli cloudwalker emergency_vuln_v1` | Emergency vulnerability management |
| `chaitin-cli cloudwalker endpoint` | Agent connection configuration |
| `chaitin-cli cloudwalker log_collect` | Log collection |
| `chaitin-cli cloudwalker message_queue` | Message queue management |
| `chaitin-cli cloudwalker organization` | Organization management |
| `chaitin-cli cloudwalker package_service` | Update package service |
| `chaitin-cli cloudwalker patch_info` | Patch intelligence |
| `chaitin-cli cloudwalker patch_info_event` | Patch risk events |
| `chaitin-cli cloudwalker report` | Report management |
| `chaitin-cli cloudwalker scout_agent_api` | Event collection agent management |
| `chaitin-cli cloudwalker security_strategy` | Security dimension strategy management |
| `chaitin-cli cloudwalker security_tool` | Security tools |
| `chaitin-cli cloudwalker statistics` | Event statistics overview |
| `chaitin-cli cloudwalker threat_overview` | Threat overview |
| `chaitin-cli cloudwalker vuln_info` | Vulnerability intelligence |

---

## VeinMind (容器安全)

VeinMind commands are generated from the embedded OpenAPI spec. Use `chaitin-cli veinmind --help` to list all groups, then drill down with `chaitin-cli veinmind <group> --help` and `chaitin-cli veinmind <group> <cmd> --help`.

### Global Flags (VeinMind)

| Flag | Env Var | Description |
|------|---------|-------------|
| `--url` | `VEINMIND_URL` | VeinMind API address |
| `--api-key` | `VEINMIND_API_KEY` | Complete 64-character VeinMind API token |
| `-o, --output` | — | Output format: `table` (default) or `json` |
| `-v, --verbose` | — | Print request URL, headers, and body |
| `--dry-run` | — | Print request summary without sending the product request |

> Note: VeinMind authenticates by splitting the 64-character API token into secret/key parts and creating a session. Its HTTP client skips TLS verification, so self-signed product certificates do not require an extra flag.

### Container Security Command Groups

| Command | Primary use |
|---------|-------------|
| `chaitin-cli veinmind agent` | Probe / scanner management, groups, install commands, start/stop/restart/repair/delete |
| `chaitin-cli veinmind app` | Software asset inventory by software name/version and related images |
| `chaitin-cli veinmind host` | Non-cluster host inventory and host-related images |
| `chaitin-cli veinmind baseline` | Shift-left compliance baseline metadata and events |
| `chaitin-cli veinmind img` | Image inventory, details, layers, history, source info, repair suggestions |
| `chaitin-cli veinmind container` | Container inventory, container detail, networks, ports, processes, volumes |
| `chaitin-cli veinmind cluster` | Kubernetes clusters and resources: nodes, pods, services, roles, secrets, workloads, PV/PVC |
| `chaitin-cli veinmind website` | Web site assets and containers that expose web sites |
| `chaitin-cli veinmind web-framework` | Web framework assets and containers that expose frameworks |
| `chaitin-cli veinmind risk` | Runtime/image risk events, event detail, response actions, event status updates |

### Common Read-Only Queries

Prefer `-o json` when parsing or summarizing results. Most list commands use `--offset` and `--page-size`; exact filters differ by command, so check leaf help before adding filters.

```bash
# Probes / scanners
chaitin-cli veinmind agent scanner-list --offset 0 --page-size 20 -o json
chaitin-cli veinmind agent scanner-list --state 2 --host-ip <ip> -o json

# Image and container inventory
chaitin-cli veinmind img list --offset 0 --page-size 20 --risk 5 -o json
chaitin-cli veinmind img base-info --id <image-list-id> -o json
chaitin-cli veinmind container list --offset 0 --page-size 20 --state 3 --risk 4 -o json
chaitin-cli veinmind container get --id <container-list-id> -o json

# Kubernetes and host inventory
chaitin-cli veinmind cluster cluster-list --offset 0 --page-size 20 --status 1 -o json
chaitin-cli veinmind cluster pod-list --cluster-id <cluster-id> --offset 0 --page-size 20 -o json
chaitin-cli veinmind host list --offset 0 --page-size 20 --name <host-name> -o json

# Software, web site, and web framework assets
chaitin-cli veinmind app agg-list --offset 0 --page-size 20 --name nginx -o json
chaitin-cli veinmind website list --offset 0 --page-size 20 --container-name <container-name> -o json
chaitin-cli veinmind web-framework list --offset 0 --page-size 20 --name spring -o json

# Baseline metadata and events
chaitin-cli veinmind baseline meta -o json
chaitin-cli veinmind baseline events --set-id <set-id> --item-id <item-id> --offset 0 --page-size 20 -o json

# Risk events
chaitin-cli veinmind risk real-time-event-list -o json
chaitin-cli veinmind risk container-webshell-list --offset 0 --page-size 20 --risk 5 --manage-status 1 -o json
chaitin-cli veinmind risk container-malicious-file-list --offset 0 --page-size 20 --risk 5 -o json
chaitin-cli veinmind risk container-revshell-event-list --offset 0 --page-size 20 --risk 4 -o json
chaitin-cli veinmind risk event-info --event-type 8 --id <event-id> -o json
```

Useful enum hints from command help:

| Field | Values |
|-------|--------|
| `--risk` | `1` no risk, `2` low, `3` medium, `4` high, `5` critical |
| `--state` on `container list` | `1` creating, `2` created, `3` running, `4` stopped |
| `--status` on `cluster cluster-list` | `1` reachable, `2` unreachable |
| `--manage-status` on risk event lists | `1` risky, `2` confirming, `3` resolved, `4` false positive, `5` ignored |
| `--event-type` on `risk event-info` | `1` image sensitive file, `2` image malicious file, `3` image webshell, `5` container command audit, `6` reverse shell, `7` container malicious file, `8` container webshell, `9` brute force, `10` weak password, `11` in-memory trojan, `12` emergency vuln, `22` escape, `23` abnormal connection |

### Mutating / Response Commands

Treat `agent` start/stop/restart/repair/delete, `agent` group create/update/delete, `cluster delete`, `risk batch-tag-events`, `risk container-op`, `risk file-op`, and `risk resume-*` commands as mutating. Use `--dry-run -v` first when the command supports a request body. Prefer `--body-file` over inline `--body` for JSON payloads.

```bash
# Inspect the request shape before changing event status
chaitin-cli --dry-run veinmind risk batch-tag-events --body-file /tmp/veinmind-event-status.json -v -o json

# Inspect the request shape before container / Pod response actions
chaitin-cli --dry-run veinmind risk container-op --body-file /tmp/veinmind-container-op.json -v -o json
```

Do not invent VeinMind request bodies. Confirm the leaf command with `--help`, inspect the relevant OpenAPI/source mapping if needed, and only execute mutating commands after the target event, asset, or probe IDs are confirmed.

---

## T-Answer (全悉 Traffic Threat Detection)

For T-Answer tasks, use the following protocol:

1. **Connect and authenticate.** Configure `TANSWER_URL` and `TANSWER_API_KEY` (or the `tanswer` config section). Run `chaitin-cli tanswer auth status` to inspect local configuration and `chaitin-cli tanswer auth check` before accessing the product. If configuration or authorization is unavailable, ask the user; do not invent values.
2. **Discover before operating.** First run `chaitin-cli tanswer --help`, then domain and leaf-command `--help`. Discover unknown commands, flags, output fields, and protected-write requirements from help or `chaitin-cli tanswer manifest`; do not depend on repository product documents or guess commands.
3. **Read operations.** Prefer semantic commands over `api`, execute read-only operations after authentication, and summarize the returned JSON for the user.
4. **Protected semantic writes require user authorization.** First run the target command with `--preview`. Present the target, change summary, impact, and risk warnings to the user, then wait for explicit confirmation for that specific change. Only after the user confirms may you invoke the command with its exact documented `--confirm` token. A token discoverable from help or manifest is a mechanical CLI requirement, not user authorization; never use it to execute a write on your own.
5. **Open API fallback.** Use `chaitin-cli tanswer api` only when no semantic command covers the requested capability and the user has supplied a known, authorized endpoint, method, and request body. Do not guess RPC methods, paths, or request bodies. GET/HEAD requests may run directly. For every other HTTP method, first run the command without `--confirm` or with `--preview`; present its method, path, query, body, and risks; wait for explicit confirmation for that exact request; then use `--confirm CONFIRM_TANSWER_RAW_API_WRITE`. The root-level `--dry-run` flag does not apply to `tanswer`.

### Fast Agent Workflow (T-Answer)

Before autonomous T-Answer work, run `chaitin-cli tanswer auth check` and `chaitin-cli tanswer manifest` unless the current session already verified the environment. For create, update, delete, import, enable, disable, or response actions, run the command with `--preview` first and do not fill `--confirm` unless the user or upstream system explicitly provides the exact confirmation token.

For routine read-only tasks, this section is enough. For complex routing, write-operation planning, or ambiguous policy/response decisions, continue with the relevant domain/leaf `--help` and `chaitin-cli tanswer manifest`; do not guess. Use `products/tanswer/COMMAND_REFERENCE.md` when a specific command's full flags, output fields, long examples, or boundaries are needed. Onboarding, Token setup, permission requirements, smoke tests, and troubleshooting live in `products/tanswer/README.md`.

### Quick Routing (T-Answer)

| User intent | Prefer |
| --- | --- |
| Verify connection, local config, or Token availability | `tanswer auth status`, `tanswer auth check` |
| Discover supported commands, flags, risk levels, output fields | `tanswer manifest` |
| System version, License, node status, health summary | `tanswer system status` |
| Overall threat posture, alarm distribution, attacker/victim top lists | `tanswer alarm overview` |
| Critical/high successful or compromised alarms for duty handling | `tanswer alarm high-priority` |
| Original threat alarm rows or one alarm detail | `tanswer alarm list`, then `tanswer alarm detail --id <doc_id>` |
| Related alarms around one source alarm | `tanswer alarm related --id <doc_id>` |
| Malicious file, Webshell, or sandbox-result investigation | `tanswer file-alarm overview`, `malicious`, `webshell`, `sandbox`, `detail` |
| Configured asset inventory, group tree, import/export, asset maintenance | `tanswer asset ...` |
| Traffic metadata by protocol, advanced query, detail, or alarm context | `tanswer metadata ...` |
| Detection whitelist or custom IOC intelligence | `tanswer policy ...` |
| Bypass block policies, block records, response whitelist, linkage devices, automatic response | `tanswer response ...` |

Use `alarm detail`, `file-alarm detail`, `asset detail`, `metadata detail`, `metadata near-alarm`, and alarm-derived policy/response commands only after the required IDs are available from list/detail commands or the user.

### Safety Rules (T-Answer)

- Do not bypass `chaitin-cli tanswer` with `curl` for state-changing operations.
- Do not guess RPC methods or request bodies. Use semantic commands, `manifest`, command `--help`, or user-provided authorized Open API documentation.
- Do not use `file-alarm` to download samples, submit samples, or trigger new sandbox analysis.
- Treat `metadata near-alarm` as investigation context, not standalone proof of an attack.
- Do not treat T-Answer response records as proof that a third-party device completed enforcement without separate validation.

---

## SafeLine-CE (雷池社区版)

SafeLine-CE is the community edition of SafeLine WAF. Its command structure differs from the enterprise edition.

### Global Flags (SafeLine-CE)

| Flag | Description |
|------|-------------|
| `--url` | SafeLine-CE server address (e.g. `https://your-server:9443`) |
| `--api-key` | API key for authentication |
| `-o, --output` | Output format: `table` (default) or `json` |
| `--verbose` | Verbose output |

> Note: SafeLine-CE does not expose `--insecure`, but its HTTP client always sets `InsecureSkipVerify: true`.

### Configuration

```yaml
safeline-ce:
  url: https://your-safeline-ce-server:9443
  api_key: YOUR_API_KEY
```

Or use environment variables:

```bash
SAFELINE_CE_URL=https://your-safeline-ce-server:9443
SAFELINE_CE_API_KEY=YOUR_API_KEY
```

### Command Reference

#### stat

```bash
chaitin-cli safeline-ce stat overview        # Aggregated stats: QPS, access, intercept counts
```

#### site

```bash
chaitin-cli safeline-ce site list            # List all web services
chaitin-cli safeline-ce site create          # Create a web service
chaitin-cli safeline-ce site update          # Update a web service
chaitin-cli safeline-ce site delete          # Delete a web service
```

#### rule (custom policy rules)

```bash
chaitin-cli safeline-ce rule list            # List all custom rules
chaitin-cli safeline-ce rule create          # Create a custom rule
chaitin-cli safeline-ce rule update          # Update a custom rule
chaitin-cli safeline-ce rule delete          # Delete a custom rule
chaitin-cli safeline-ce rule switch          # Enable or disable a custom rule
```

#### ipgroup

```bash
chaitin-cli safeline-ce ipgroup list         # List all IP groups
chaitin-cli safeline-ce ipgroup get          # Get IP group details
chaitin-cli safeline-ce ipgroup create       # Create an IP group
chaitin-cli safeline-ce ipgroup update       # Update an IP group
chaitin-cli safeline-ce ipgroup delete       # Delete an IP group
chaitin-cli safeline-ce ipgroup append       # Add IPs to an IP group
```

#### log

```bash
chaitin-cli safeline-ce log attack list      # List attack logs
chaitin-cli safeline-ce log attack get       # Get attack log detail by ID
chaitin-cli safeline-ce log rule list        # List rule-triggered attack logs
chaitin-cli safeline-ce log rule get         # Get rule-triggered attack log detail
chaitin-cli safeline-ce log audit list       # Get audit logs
```

#### skynet (enhanced detection rules)

```bash
chaitin-cli safeline-ce skynet get           # Get enhanced rule configuration
chaitin-cli safeline-ce skynet update        # Update enhanced rule configuration
chaitin-cli safeline-ce skynet switch get    # Get global enable status of enhanced rules
chaitin-cli safeline-ce skynet switch set    # Enable or disable enhanced rules globally
```

#### module (global semantics)

```bash
chaitin-cli safeline-ce module get           # Get global semantics mode
chaitin-cli safeline-ce module update        # Update global semantics mode
```

#### cert (system / license)

```bash
chaitin-cli safeline-ce cert info            # Get system info
chaitin-cli safeline-ce cert get             # Get license info
chaitin-cli safeline-ce cert update          # Update management certificate
```

---

## DDR (数据安全运营)

### Global Flags (DDR)

| Flag | Env Var | Description |
|------|---------|-------------|
| `--url` | `DDR_URL` | DDR API address, usually ending with `/qzh/api/v1` |
| `--api-key` | `DDR_API_KEY` | API key or Serval token |
| `-o, --output` | — | Output format: `json` (default) or `table` |
| `-v, --verbose` | — | Print request URL, headers, and body |

DDR commands default to JSON output. Keep `-o json` in examples when the result will be parsed, and use `-o table` only when the user asks for a compact human-readable summary.

### 创建资产扫描任务流程

1. 根据设备名称查询设备 UID。Example for `PC-000006`:

```bash
chaitin-cli ddr device list --accept application/json --accept-language zh --content-type application/json --if-none-match '' --x-cs-header-crypt none --x-cs-header-debug false --x-cs-header-timezone Asia/Shanghai -o json --search PC-000006 --limit 20 --page 1
```

2. 从返回结果里取设备 `id`，并用该设备名称和 ID 修改扫描任务 body 里的 `name`、`include[0].name`、`include[0].id`、`binding_config.include.device_ids`。建议创建临时 JSON 文件，避免命令行转义问题:

```json
{
  "file_size_lower_measure": "MB",
  "file_size_lower": 0,
  "file_size_upper_measure": "MB",
  "file_size_upper": 100,
  "name": "2222-1",
  "description": "3222",
  "scope": "common",
  "include": [
    {
      "name": "PC-000341",
      "id": "62515059-0700-46a8-8441-bb75413e75de",
      "type": "device",
      "expand": {
        "ip": "192.168.96.247",
        "code": "PC-000341",
        "username": "orange@orangedeMacBook-Pro",
        "os_icon": ["macOS"],
        "status": "activated",
        "conn_status": "online",
        "agent_version": "3.9.100",
        "tags": []
      }
    }
  ],
  "frequency": "immediately",
  "worktime_point": [
    {"begin": "00:00", "end": "23:59", "week": 1},
    {"begin": "00:00", "end": "23:59", "week": 2},
    {"begin": "00:00", "end": "23:59", "week": 3},
    {"begin": "00:00", "end": "23:59", "week": 4},
    {"begin": "00:00", "end": "23:59", "week": 5},
    {"begin": "00:00", "end": "23:59", "week": 6},
    {"begin": "00:00", "end": "23:59", "week": 7}
  ],
  "scan_mode": "custom",
  "dirs_scan_options": {
    "Windows": [{"path": "C:\\Users"}, {"path": "D:\\"}, {"path": "E:\\"}],
    "macOS": [{"path": "/tmp/123"}],
    "Linux": [{"path": "/home"}, {"path": "/data"}],
    "Kylin": [{"path": "/home"}, {"path": "/data"}],
    "UOS": [{"path": "/home"}, {"path": "/data"}]
  },
  "scan_all_das": true,
  "expand": {"web": {"file_size_operator": "lt", "file_size_num": 100, "file_size_unit": "MB"}},
  "file_scan_mode": "speed",
  "advanced_option": false,
  "binding_config": {
    "source": "file_scan",
    "include": {"device_ids": ["62515059-0700-46a8-8441-bb75413e75de"]},
    "exclude": {}
  },
  "file_exclude_ids": [],
  "file_include_ids": [],
  "custom_whitelist": [],
  "notification_config": [],
  "governance_id": null
}
```

Run the task with the JSON body content:

```bash
chaitin-cli ddr device filescantask --body "$(cat /tmp/ddr-filescantask.json)"
```

### 查看扫描任务结果

Query the latest scan task and summarize the result for the user in a compact table when possible:

```bash
chaitin-cli ddr device filescantask list --body '{"page": 1, "limit": 1, "search": ""}' -o json
```

### 查看审批任务列表

```bash
chaitin-cli ddr disposal approvalinstance list --page 1 --limit 10 --search "" -o json
```

### 外发管控日志查询

```bash
chaitin-cli ddr policylog channel list --page 1 --limit 20 --search "" -o json
```

### 网络管理日志查询

```bash
chaitin-cli ddr policylog softwarenetwork list --page 1 --limit 20 --search "" -o json
```

### 卸载设备

This is a mutating operation. Confirm the target device UUID before execution and prefer `--dry-run` first when available.

```bash
./bin/chaitin-cli ddr device status-action --device-id <device uuid> --operation uninstall
```

### 终端扫描
```bash
chaitin-cli ddr device filescantask list  # 终端扫描任务列表
chaitin-cli ddr device filescantask get --task-id <task id> # 终端扫描任务结果详情
chaitin-cli ddr device filescantask instance-device-list --task-id=<task_id> --instance-id=<instance_id> # 拉取扫描设备列表
chaitin-cli ddr device filescantask instance-results-list --task-id=<task_id> --instance-id=<instance_id> # 拉取命中结果列表
chaitin-cli ddr device filescantask remove --task-id=<task_id> # 删除资产扫描任务
```

### 渠道管理
```bash
chaitin-cli ddr system channeldefgroup list  # 获取渠道列表
```

### 外发管控
```bash
chaitin-cli ddr policy list # 外发管控策略列表
chaitin-cli ddr policy channel get --policy-id <policy_id> # 外发管控策略详情


chaitin-cli ddr policy channel --body='{"mode":"quick","name":"<name>","policy_group_ids":["<policy_group_id>"],"description":"222","risk_level":3,"scope":"common","expression_fields":[{"field":"channel","operator":"contains","value":["<channel_id>"]},{"field":"file_size_limit_mb","operator":"gt","value":10}],"action":"approval","action_template_id":"<action_template_id>","binding_config":{"source":"channel","include":{"device_ids":["<device_id>"]},"exclude":{}}}' # 创建外发管控策略 body 内容建议创建临时文件，避免转移问题, <name> 替换为策略名称, <policy_group_id> 替换为策略组ID, <channel_id> 替换为渠道ID, <device_id> 替换为设备ID, <action_template_id> 替换为动作模板ID
```

#### 创建外发管控策略 body 内容样例
```json
{"mode":"quick","name":"<name>","policy_group_ids":["0f2ad6aa-d1cc-4a4e-a0b0-3bd749579868"],"description":"222","risk_level":3,"scope":"common","expression_fields":[{"field":"channel","operator":"contains","value":["6509b799-78e7-4ede-9af4-2c743897e6c6"]},{"field":"file_size_limit_mb","operator":"gt","value":10}],"action":"approval","action_template_id":"90144b18-e9d2-49ad-9c1d-cef0a4146ccc","binding_config":{"source":"channel","include":{"device_ids":["ebf62d82-f02b-4805-b6c3-31d1475b8061"]},"exclude":{}}}
```

### 软件管控
```bash
chaitin-cli ddr softwaremanager list --page 1 --limit 50 --search "" # 软件管控-软件列表
chaitin-cli ddr softwaremanager software-hash-list --software-hash=<software-hash> --limit=10 --page=1 --search= --body='{"is_pirated":false,"use_default_query":false,"search":"","queries":[]}' # 软件管控-安装详情 <software-hash> 取 软件管控-软件列表 里的 total_view_digest 字段
```

### 行为管控策略通用约定

创建或更新策略时优先把 JSON 写到临时文件，再用 `--body-file`，避免 shell 转义错误。启停动作已用 `chaitin-cli --dry-run ddr ... -v` 校验，最小 body 是 `{"operation":"activate"}`；停用通常把 `activate` 换成 `deactivate`。

OpenAPI 里的 `webpost_control` 在 CLI 中会归一化为 `webpost-control`，命令必须写连字符。

### 落盘管控
```bash
chaitin-cli ddr system channeldefgroup landing-list -o json # 落盘管控源列表
chaitin-cli ddr policy landing list # 落盘管控列表
chaitin-cli ddr policy landing action --uid <policy_id> --body '{"operation":"activate"}' # 启用落盘管控策略
chaitin-cli ddr policy landing action --uid <policy_id> --body '{"operation":"deactivate"}' # 停用落盘管控策略
chaitin-cli ddr policy landing get --uid <policy_id> # 落盘管控详情
chaitin-cli ddr policy landing remove --uid <policy_id> # 删除落盘管控策略
chaitin-cli ddr policy landing --body-file /tmp/ddr-policy-landing.json # 创建落盘管控策略
```

#### 创建 落盘管控 body 内容样例
```json
{"name":"test_langding","scope":"common","expression_fields":[{"field":"channel","operator":"contains","value":["6509b799-78e7-4ede-9af4-2c743897e6c6"]}],"action":"notify","category":"landing","policy_group_ids":[null],"binding_config":{"source":"landing","include":{"device_ids":["0578ae3e-6369-4cc4-bad6-ff48f1cf6ceb"]},"exclude":{}}}
```

### 邮件管控 - email
```bash
chaitin-cli ddr system channeldefgroup email-list -o json # 邮件管控源列表
chaitin-cli ddr policy email list -o json # 邮件管控策略列表
chaitin-cli ddr policy email get --uid <policy_id> -o json # 邮件管控策略详情
chaitin-cli ddr policy email action --uid <policy_id> --body '{"operation":"activate"}' # 启用邮件管控策略
chaitin-cli ddr policy email action --uid <policy_id> --body '{"operation":"deactivate"}' # 停用邮件管控策略
chaitin-cli ddr policy email remove --uid <policy_id> # 删除邮件管控策略
chaitin-cli ddr policy email --body-file /tmp/ddr-policy-email.json # 创建邮件管控策略
chaitin-cli ddr policy email entitywhitelist-list -o json # 邮件管控实体白名单列表
chaitin-cli ddr policylog email timelinelist --body '{"time_range":{"begin":"2023-11-14 10:55:38","end":"2023-11-14 11:55:38"},"search":"","queries":[]}' -o json # 邮件管控时间线
chaitin-cli ddr policylog email stafflist --body '{"time_range":{"begin":"2023-11-14 10:55:38","end":"2023-11-14 11:55:38"},"search":"","queries":[]}' -o json # 邮件管控员工列表
```

#### 创建 邮件管控 body 内容样例
```json
{"name":"邮件策略","description":"策略描述","scope":"global","binding_config":{"source":"email","include":{},"exclude":{}},"action":"notify","expression":"N/A","expression_relation":"all","expression_fields":[{"field":"channel","operator":"contains","value":["all_channels"]}]}
```

### 代码管控 - codecontrol
```bash
chaitin-cli ddr policy codecontrol list --page 1 --page-size 20 --search "" -o json # 代码管控策略列表
chaitin-cli ddr policy codecontrol get --policy-id <policy_id> -o json # 代码管控策略详情
chaitin-cli ddr policy codecontrol action --policy-id <policy_id> --body '{"operation":"activate"}' # 启用代码管控策略
chaitin-cli ddr policy codecontrol action --policy-id <policy_id> --body '{"operation":"deactivate"}' # 停用代码管控策略
chaitin-cli ddr policy codecontrol remove --policy-id <policy_id> # 删除代码管控策略
chaitin-cli ddr policy codecontrol create --body-file /tmp/ddr-policy-codecontrol.json # 创建代码管控策略
chaitin-cli ddr policy codecontrol update --policy-id <policy_id> --body-file /tmp/ddr-policy-codecontrol.json # 更新代码管控策略
chaitin-cli ddr policy codecontrol configwhitelist-list -o json # 代码管控配置白名单列表
chaitin-cli ddr policy codecontrol controlwhitelist-list -o json # 代码管控管控白名单列表
chaitin-cli ddr policy codecontrol entitywhitelist-list -o json # 代码管控实体白名单列表
chaitin-cli ddr policylog code timelinelist --body '{"time_range":{"begin":"2023-11-14 10:55:38","end":"2023-11-14 11:55:38"},"search":"","queries":[]}' -o json # 代码管控时间线
chaitin-cli ddr policylog code stafflist --body '{"time_range":{"begin":"2023-11-14 10:55:38","end":"2023-11-14 11:55:38"},"search":"","queries":[]}' -o json # 代码管控员工列表
chaitin-cli ddr policylog code download --body '{"time_range":{"begin":"2023-11-14 10:55:38","end":"2023-11-14 11:55:38"},"search":"","queries":[]}' # 代码管控导出
```

#### 创建 代码管控 body 内容样例
```json
{"name":"代码管控策略","description":"策略描述","scope":"global","period_category":"permanent","period_options":{"period_value":["2023-01-01 00:00:00","2023-01-01 00:00:00"]},"binding_config":{"include":{"device_ids":[],"device_tag_ids":[],"staff_ids":[],"staff_tag_ids":[],"dept_ids":[]},"exclude":{},"source":"code_control"},"control_action":{"category":"upload","expression_fields":{"git":{"left":"repository_url","operator":"contains","value":["github.com/example"]},"svn":{"left":"repository_url","operator":"contains","value":["svn.example.com"]}}},"action":"notify","action_template_id":"<action_template_id>"}
```

### 剪贴板管控 - clipboard
```bash
chaitin-cli ddr policy clipboard --body-file /tmp/ddr-policy-clipboard.json # 创建剪贴板管控策略
chaitin-cli ddr policylog clipboard list --page 1 --limit 20 --search "" --body '{"time_range":{"begin":"2024-09-23T06:36:00Z","end":"2024-11-22T06:36:59Z"},"search":"","queries":[]}' -o json # 剪贴板管控日志列表
chaitin-cli ddr policylog clipboard timelinelist --body '{"time_range":{"begin":"2024-09-23T06:36:00Z","end":"2024-11-22T06:36:59Z"},"search":"","queries":[]}' -o json # 剪贴板管控时间线
chaitin-cli ddr policylog clipboard stafflist --body '{"time_range":{"begin":"2024-09-23T06:36:00Z","end":"2024-11-22T06:36:59Z"},"search":"","queries":[]}' -o json # 剪贴板管控员工列表
chaitin-cli ddr clipboardbehavior timelinelist --body '{"time_range":{"begin":"2024-09-23T06:36:00Z","end":"2024-11-22T06:36:59Z"},"search":"","queries":[]}' -o json # 剪贴板行为时间线
chaitin-cli ddr clipboardbehavior stafflist --body '{"time_range":{"begin":"2024-09-23T06:36:00Z","end":"2024-11-22T06:36:59Z"},"search":"","queries":[]}' -o json # 剪贴板行为员工列表
```

#### 创建 剪贴板管控 body 内容样例
```json
{"name":"剪贴板策略","period_category":"permanent","scope":"common","family":"Windows","control_category":"copy","expression_fields":[{"field":"source_process","operator":"contains","value":["example.exe"]}],"action":"notify","binding_config":{"source":"clipboard","include":{"device_ids":["<device_id>"]},"exclude":{}}}
```

### 网页管控 - webpost_control
```bash
chaitin-cli ddr policy webpost-control list --page 1 --limit 20 --search "" -o json # 网页管控策略列表
chaitin-cli ddr policy webpost-control get --policy-id <policy_id> -o json # 网页管控策略详情
chaitin-cli ddr policy webpost-control action --policy-id <policy_id> --body '{"operation":"activate"}' # 启用网页管控策略
chaitin-cli ddr policy webpost-control action --policy-id <policy_id> --body '{"operation":"deactivate"}' # 停用网页管控策略
chaitin-cli ddr policy webpost-control remove --policy-id <policy_id> # 删除网页管控策略
chaitin-cli ddr policy webpost-control create --body-file /tmp/ddr-policy-webpost-control.json # 创建网页管控策略
chaitin-cli ddr policy webpost-control update --policy-id <policy_id> --body-file /tmp/ddr-policy-webpost-control.json # 更新网页管控策略
chaitin-cli ddr policy webpost-control whitelist-list -o json # 网页管控管控白名单列表
chaitin-cli ddr policy webpost-control entitywhitelist-list -o json # 网页管控实体白名单列表
```

#### 创建 网页管控 body 内容样例
```json
{"name":"网页管控策略","description":"策略描述","scope":"global","period_category":"permanent","period_options":{"period_hours_value":[{"begin":"00:00","end":"23:59","week":1},{"begin":"00:00","end":"23:59","week":2},{"begin":"00:00","end":"23:59","week":3},{"begin":"00:00","end":"23:59","week":4},{"begin":"00:00","end":"23:59","week":5},{"begin":"00:00","end":"23:59","week":6},{"begin":"00:00","end":"23:59","week":7}]},"binding_config":{"include":{"device_ids":[],"device_tag_ids":[],"staff_ids":[],"staff_tag_ids":[],"dept_ids":[]},"exclude":{},"source":"webpost_control"},"control_action":{"expression_fields":[{"left":"content","operator":"contains","value":["secret"]}]},"action":"notify","action_template_id":"<action_template_id>","checked_urls_options":{"left":"url","operator":"contains","value":["example.com"]},"detect_position":"form_internal"}
```
