---
name: workspace-cli-usage
description: "Use when running cws commands to manage Chaitin security products: SafeLine WAF (site management, IP blocking, ACL, policy rules, attack logs), X-Ray vulnerability scanner (scan tasks, results, assets), CloudWalker CWPP (events, vulnerabilities, assets), and T-Answer (firewall rules, blocklists)."
version: 1.0.0
author: chaitin
tags: [cws, safeline, xray, cloudwalker, tanswer, waf, security, chaitin, cli]
---

# workspace-cli (cws) Usage Guide

> Unified CLI for Chaitin security products. Manage SafeLine WAF, X-Ray scanner, CloudWalker CWPP, and T-Answer through a single tool.

## Install & Run

```bash
# Download pre-built binary from GitHub Releases
# https://github.com/chaitin/workspace-cli/releases

# Or build from source
git clone https://github.com/chaitin/workspace-cli.git
cd workspace-cli
go build -o cws .

# Run
cws <product> <command> [flags]
```

## Prerequisites

Before running any `cws` command:

1. **Network reachability** — the machine running `cws` must be able to reach each product's console / API endpoint.
2. **API key** — generate one from each product's UI (SafeLine → System → API Token; X-Ray → System Settings → API Key; etc.) and supply it via `--api-key`, the product env var, or `config.yaml`.
3. **TLS with self-signed certs** — `cws xray` takes `--insecure` (off by default). `cws safeline` also exposes `--insecure`, but its default is `true` (already skipping verification); pass `--insecure=false` to re-enable verification. `cws safeline-ce`, `cws cloudwalker`, and `cws tanswer` don't expose the flag and always skip TLS verification in their HTTP clients.
4. **Build from source** — Go 1.25+ (see `go.mod`). Otherwise use the pre-built binary from GitHub Releases.

## Configuration

Create `config.yaml` in the working directory:

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

tanswer:
  url: https://your-tanswer-server
  api_key: YOUR_API_KEY
```

Or use environment variables / `.env` file:

```bash
SAFELINE_URL=https://your-safeline-server
SAFELINE_API_KEY=YOUR_API_KEY
XRAY_URL=https://your-xray-server/api/v2
XRAY_API_KEY=YOUR_API_KEY
```

Priority: `flags > environment/.env > config.yaml`

Use `-c` to switch between config files (e.g., multiple environments):

```bash
cws -c ./configs/prod.yaml safeline stats overview
cws -c ./configs/staging.yaml safeline stats overview
```

### Global Flags

| Flag | Description |
|------|-------------|
| `-c, --config` | Config file path (default: `./config.yaml`) |
| `--dry-run` | Print the API request without executing. Applied by the root command to `xray` and `cloudwalker`. `safeline` registers its own `--dry-run` and forwards it to subcommands. `safeline-ce` inherits the root flag, but the current codebase stores the value without using it; `tanswer` ignores it. |

### Discovering Commands

`--help` is the authoritative source — this document does not enumerate every flag.

```bash
cws <product> --help                # List subcommand groups for a product
cws <product> <group> --help        # List commands in a group
cws <product> <group> <cmd> --help  # List flags for a specific command
```

`cws xray` commands are auto-generated from the X-Ray OpenAPI spec (hundreds of operations); `cws xray <category> --help` is the only complete reference. `cws cloudwalker` has 60+ command groups with similar depth.

### Operating Rules

For SafeLine, X-Ray, CloudWalker, T-Answer, and SafeLine-CE tasks, treat `cws` as the only supported operator interface.

- Prefer `cws ... --help` and existing `cws` subcommands over `curl`, ad-hoc HTTP requests, browser debugging, or guessed endpoints.
- If `cws` does not expose the requested product operation, stop and say that the current CLI does not support it. Do not fall back to direct API calls just to "try it".
- Do not use `curl` or raw HTTP requests to perform state-changing or potentially dangerous product operations that are not implemented by `cws`.
- Use source inspection to confirm command availability and behavior, not to bypass the CLI and reconstruct private API calls.
- When a supported command may change state and the product actually honors `--dry-run`, prefer checking that path first.

### Output Formats

Each product uses its own output convention — there is no unified `-f` / `--format` flag across `cws`.

| Product | Default | Switch to JSON | Other |
|---------|---------|----------------|-------|
| `cws safeline` | table | `--indent` | — |
| `cws safeline-ce` | table | `-o json` (or `--output json`) | `--verbose` |
| `cws xray` | JSON (no alternative) | — | `--debug` for debug logs |
| `cws cloudwalker` | text | `-f json` (or `--format json`) | `--no-trunc` to disable text truncation |
| `cws tanswer` | formatted text | `--raw` (bool) | — |

When piping into `jq`, note that SafeLine uses `--indent` (not `-o`/`-f`), and T-Answer uses `--raw`.

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
| Asset inventory (web / domain / IP) | `xray web_asset` · `xray domain_asset` · `xray ip_asset` |
| Baseline / compliance check | `xray baseline` · `cloudwalker baseline_v2` |
| Host-level event response (webshell, reverse shell, brute force) | `cloudwalker webshell_event` · `cloudwalker revshell_event` · `cloudwalker brute_force` |
| Host asset inventory (process / port / container / user) | `cloudwalker process_asset` · `cloudwalker port_asset` · `cloudwalker docker_container` · `cloudwalker user_asset` |
| Ransomware protection, file quarantine, kill process | `cloudwalker anti_ransomware` · `cloudwalker file_disposal` · `cloudwalker process_kill` |
| Host firewall / network block | `cloudwalker firewall` · `cloudwalker network_reject` |
| Traffic-level threat detection firewall (whitelist / block rules) | `tanswer firewall` · `tanswer rules` |
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
cws safeline stats overview --duration h

# View last 30 days stats
cws safeline stats overview --duration d
```

#### Step 2: List your protected sites

```bash
cws safeline site list
```

#### Step 3: View recent attack logs

```bash
# List the latest 20 attack events
cws safeline log detect list --count 20

# Get full details of a specific event
cws safeline log detect get \
  --event-id "6edb4c7eb69042cd996045e3ee5526d9" \
  --timestamp "1774857841"
```

#### Step 4: Block the attacker's IP

**Option A — Create an IP group and block it with an ACL rule:**

```bash
# Create an IP group for malicious IPs
cws safeline ip-group create \
  --name "Blocklist" \
  --ips "203.0.113.42,198.51.100.7" \
  --comment "Attackers from incident 2024-01"

# Create an ACL template that forbids the group
cws safeline acl template create \
  --name "Block Malicious IPs" \
  --template-type manual \
  --target-type cidr \
  --action forbid \
  --ip-groups <group-id>
```

**Option B — Block specific IPs directly without a group:**

```bash
cws safeline acl template create \
  --name "Emergency Block" \
  --template-type manual \
  --target-type cidr \
  --action forbid \
  --targets "203.0.113.42,198.51.100.7"
```

#### Step 5: Add a custom rule to block a malicious path

```bash
# Block requests to /admin/upload with high risk level
cws safeline policy-rule create \
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
cws safeline policy-group list
cws safeline policy-group get <id>

# Enable SQL injection and XSS detection
cws safeline policy-group update <id> \
  --module m_sqli,m_xss \
  --state enabled
```

#### Step 7: Monitor access logs

```bash
cws safeline log access list --count 50
cws safeline log access get \
  --event-id "1e1ef8e9b21d42cd996045e3ee5526d9" \
  --req-start-time "1775117700"
```

#### Step 8: Unblock a false positive

```bash
# List active ACL rules (blocked IPs)
cws safeline acl rule list --template-id <template-id>

# Remove the block and add IP to whitelist
cws safeline acl rule delete <rule-id> --add-to-whitelist

# Or clear all rules for a template
cws safeline acl rule clear --template-id <template-id>
```

---

### SafeLine Command Reference

#### stats

```bash
cws safeline stats overview --duration h   # 24h stats
cws safeline stats overview --duration d   # 30d stats
```

#### site

```bash
cws safeline site list                                    # List all sites
cws safeline site get <id>                                # Get site details
cws safeline site enable <id>                             # Enable a site
cws safeline site disable <id>                            # Disable a site
cws safeline site update <id> --policy-group <group-id>  # Attach a policy group to a site
cws safeline site update <id> --policy-group 0            # Detach policy group from a site
```

#### ip-group (alias: ipgroup)

```bash
cws safeline ip-group list                                              # List all IP groups
cws safeline ip-group list --name "office" --count 50 --offset 0       # Filter by name with pagination
cws safeline ip-group get <id>                                          # Get IP group details
cws safeline ip-group create --name "DC" --ips "172.16.0.0/16" --comment "Data center"  # Create a new IP group
cws safeline ip-group delete <id>                                       # Delete an IP group
cws safeline ip-group delete 1 2 3                                      # Batch delete IP groups
cws safeline ip-group add-ip <id> --ips "10.0.1.0/24"                  # Add IPs to an IP group
cws safeline ip-group remove-ip <id> --ips "10.0.1.0/24"               # Remove IPs from an IP group
```

#### acl template

```bash
cws safeline acl template list                # List all ACL templates
cws safeline acl template list --name "limit" # Filter templates by name
cws safeline acl template get <id>            # Get ACL template details
cws safeline acl template enable <id>         # Enable an ACL template
cws safeline acl template disable <id>        # Disable an ACL template
cws safeline acl template delete <id>         # Delete an ACL template

# Create manual block rule (specific IPs)
cws safeline acl template create \
  --name "Block IPs" --template-type manual \
  --target-type cidr --action forbid \
  --targets "192.168.1.100,10.0.0.50"

# Create auto rate-limit rule
cws safeline acl template create \
  --name "Rate Limit" --template-type auto \
  --period 60 --limit 100 --action forbid

# Create throttle rule (allow but slow down)
cws safeline acl template create \
  --name "Throttle" --template-type auto \
  --period 60 --limit 100 \
  --action limit_rate \
  --limit-rate-limit 10 --limit-rate-period 60
```

#### acl rule (blocked IP entries)

```bash
cws safeline acl rule list --template-id <id>                    # List blocked IP entries for a template
cws safeline acl rule delete <id>                                # Delete a blocked IP entry
cws safeline acl rule delete <id> --add-to-whitelist             # Delete and move IP to whitelist
cws safeline acl rule clear --template-id <id>                   # Clear all blocked IP entries for a template
cws safeline acl rule clear --template-id <id> --add-to-whitelist # Clear all and move IPs to whitelist
```

#### policy-group

```bash
cws safeline policy-group list                                              # List all policy groups
cws safeline policy-group get <id>                                          # Get policy group details
cws safeline policy-group update <id> --module m_sqli,m_xss --state enabled  # Enable detection modules
cws safeline policy-group update <id> --module m_cmd_injection --state disabled # Disable a detection module
```

Available modules: `m_sqli` `m_xss` `m_cmd_injection` `m_file_include` `m_file_upload` `m_php_code_injection` `m_php_unserialize` `m_java` `m_java_unserialize` `m_ssrf` `m_ssti` `m_csrf` `m_scanner` `m_response` `m_rule`

#### policy-rule

```bash
cws safeline policy-rule list                  # List all policy rules (global by default)
cws safeline policy-rule list --global=false   # List site-specific rules only
cws safeline policy-rule get <id>              # Get policy rule details
cws safeline policy-rule enable <id>           # Enable a policy rule
cws safeline policy-rule disable <id>          # Disable a policy rule
cws safeline policy-rule delete <id>           # Delete a policy rule

# Create simple rule
cws safeline policy-rule create \
  --comment "Block /admin" \
  --target urlpath --cmp infix --value "/admin" \
  --action deny --risk-level 3

# List available targets and operators
cws safeline policy-rule targets
cws safeline policy-rule targets --cmp urlpath

# Actions: deny | dry_run | allow
# Risk levels: 0=none 1=low 2=medium 3=high
```

#### log

```bash
# Attack logs
cws safeline log detect list --count 50
cws safeline log detect list --current-page 1 --target-page 2
cws safeline log detect get --event-id "<id>" --timestamp "<ts>"

# Access logs
cws safeline log access list --count 50
cws safeline log access get --event-id "<id>" --req-start-time "<ts>"

# Rate-limit logs (alias: rl)
cws safeline log rate-limit list --count 50 --offset 0
```

#### system

```bash
cws safeline system license                      # Get license information
cws safeline system machine-id                   # Get machine ID (for license activation)
cws safeline system log list --count 50 --offset 0  # List system operation logs
```

#### network (hardware mode only)

```bash
cws safeline network workgroup list          # alias: wg list
cws safeline network workgroup get <name>    # alias: wg get
cws safeline network interface list          # alias: if list
cws safeline network interface ip <name>     # alias: if ip
cws safeline network gateway get             # alias: gw get
cws safeline network route list              # alias: sr list
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
cws xray plan PostPlanCreateQuick \
  --targets=10.3.0.4,10.3.0.5 \
  --engines=<engine-id> \
  --project-id=1

# List scan tasks
cws xray plan PostPlanFilter \
  --filterPlan.limit=10 \
  --filterPlan.offset=0

# Stop a scan task
cws xray plan PostPlanStop --stopPlanBody.id=<id>

# Resume a scan task
cws xray plan PostPlanExecute --executePlanBody.id=<id>

# Delete a scan task
cws xray plan DeletePlanID --id=<id>
```

### Command Categories

| Command | Description |
|---------|-------------|
| `cws xray asset_property` | Asset management |
| `cws xray audit_log` | Audit log management |
| `cws xray baseline` | Baseline check management |
| `cws xray custom_poc` | Custom POC management |
| `cws xray domain_asset` | Domain asset management |
| `cws xray insight` | Data insight and analytics |
| `cws xray ip_asset` | IP/host asset management |
| `cws xray plan` | Scan task management |
| `cws xray project` | Project/workspace management |
| `cws xray report` | Report management |
| `cws xray result` | Scan result management |
| `cws xray role` | Role management |
| `cws xray service_asset` | Service asset management |
| `cws xray system_info` | System information |
| `cws xray system_service` | System service management |
| `cws xray task_config` | Task configuration management |
| `cws xray template` | Policy template management |
| `cws xray user` | User management |
| `cws xray vulnerability` | Vulnerability management |
| `cws xray web_asset` | Web asset management |
| `cws xray xprocess` | XProcess task instance management |
| `cws xray xprocess_lite` | XProcess lite management |

---

## CloudWalker (云溯 CWPP)

### Global Flags (CloudWalker)

| Flag | Env Var | Description |
|------|---------|-------------|
| `--url` | `CLOUDWALKER_URL` | CloudWalker RPC address (required) |
| `--api-key` | `CLOUDWALKER_API_KEY` | API key |

> Note: CloudWalker does not expose `--insecure`, but its HTTP client always sets `InsecureSkipVerify: true`, so self-signed certs just work — no CA install or HTTP fallback needed.

### Command Categories

Each category has subcommands — run `cws cloudwalker <category> --help` to list them.

#### Security Events

| Command | Description |
|---------|-------------|
| `cws cloudwalker abnormal_login_event` | Abnormal login events |
| `cws cloudwalker brute_force` | Brute-force events |
| `cws cloudwalker elevation_process_event` | Privilege escalation process events |
| `cws cloudwalker event_stat` | Event management and statistics |
| `cws cloudwalker full_command` | Full command execution records |
| `cws cloudwalker honeypot` | Honeypot trap events |
| `cws cloudwalker malware_event` | Malware events |
| `cws cloudwalker memory_webshell_event` | In-memory webshell events |
| `cws cloudwalker network_audit_event` | Network anomaly events |
| `cws cloudwalker non_white_process` | Non-whitelisted process events |
| `cws cloudwalker revshell_event` | Reverse shell events |
| `cws cloudwalker suspicious_operation` | Suspicious operation events |
| `cws cloudwalker webshell_event` | Webshell events |

#### Asset Inventory

| Command | Description |
|---------|-------------|
| `cws cloudwalker application_asset` | Application assets |
| `cws cloudwalker asset_cert` | Certificate assets |
| `cws cloudwalker asset_config` | Asset collection configuration |
| `cws cloudwalker asset_crontab` | Scheduled task assets |
| `cws cloudwalker asset_env` | Environment variable assets |
| `cws cloudwalker asset_registry` | Registry assets |
| `cws cloudwalker asset_startup` | Startup item assets |
| `cws cloudwalker docker_container` | Docker container assets |
| `cws cloudwalker docker_image` | Docker image assets |
| `cws cloudwalker docker_network` | Docker network assets |
| `cws cloudwalker host_asset` | Host assets (includes agent management) |
| `cws cloudwalker host_discovery` | Unknown host discovery |
| `cws cloudwalker host_nic_asset` | Network interface card assets |
| `cws cloudwalker host_partition_asset` | Partition assets |
| `cws cloudwalker host_route_asset` | Route assets |
| `cws cloudwalker port_asset` | Port assets |
| `cws cloudwalker process_asset` | Process assets |
| `cws cloudwalker user_asset` | User assets |
| `cws cloudwalker website_asset` | Website assets |

#### Security Protection

| Command | Description |
|---------|-------------|
| `cws cloudwalker anti_ransomware` | Anti-ransomware protection |
| `cws cloudwalker baseline_v2` | Baseline check management |
| `cws cloudwalker detection_rule` | Detection rule management |
| `cws cloudwalker file_disposal` | File disposal (quarantine/delete) |
| `cws cloudwalker firewall` | Firewall rule management |
| `cws cloudwalker mimicry` | Mimicry defense |
| `cws cloudwalker network_reject` | Network block management |
| `cws cloudwalker port_scan` | Port scan protection |
| `cws cloudwalker process_kill` | Process termination |
| `cws cloudwalker security_check` | Security checks |
| `cws cloudwalker sensitive_file` | Sensitive file management |
| `cws cloudwalker sensitive_file_scan` | Sensitive file scanning |
| `cws cloudwalker sensitive_port` | Sensitive port management |
| `cws cloudwalker sensitive_user` | Sensitive user management |
| `cws cloudwalker tamper_proof` | File tamper-proof protection |
| `cws cloudwalker vuln` | Vulnerability management |
| `cws cloudwalker weak_passwd` | Weak password detection |
| `cws cloudwalker whitelist` | Whitelist rule management |

#### Platform Management

| Command | Description |
|---------|-------------|
| `cws cloudwalker admin_agent` | Agent module update management |
| `cws cloudwalker admin_monitor` | System monitoring management |
| `cws cloudwalker admin_strategy` | Strategy management |
| `cws cloudwalker agent` | Agent management |
| `cws cloudwalker agent_detector` | Malicious file agent management |
| `cws cloudwalker agent_module` | Agent module management |
| `cws cloudwalker alert_config` | Alert configuration |
| `cws cloudwalker audit_log` | Audit log |
| `cws cloudwalker business_group` | Business group management |
| `cws cloudwalker crontab` | Scheduled task management |
| `cws cloudwalker emergency_vuln_v1` | Emergency vulnerability management |
| `cws cloudwalker endpoint` | Agent connection configuration |
| `cws cloudwalker log_collect` | Log collection |
| `cws cloudwalker message_queue` | Message queue management |
| `cws cloudwalker organization` | Organization management |
| `cws cloudwalker package_service` | Update package service |
| `cws cloudwalker patch_info` | Patch intelligence |
| `cws cloudwalker patch_info_event` | Patch risk events |
| `cws cloudwalker report` | Report management |
| `cws cloudwalker scout_agent_api` | Event collection agent management |
| `cws cloudwalker security_strategy` | Security dimension strategy management |
| `cws cloudwalker security_tool` | Security tools |
| `cws cloudwalker statistics` | Event statistics overview |
| `cws cloudwalker threat_overview` | Threat overview |
| `cws cloudwalker vuln_info` | Vulnerability intelligence |

---

## T-Answer (全悉 Traffic Threat Detection)

### Global Flags (T-Answer)

| Flag | Env Var | Description |
|------|---------|-------------|
| `--url` | `TANSWER_URL` | T-Answer server address (required) |
| `--api-key` | `TANSWER_API_KEY` | API token |

> Note: T-Answer does not expose `--insecure`, but its HTTP client always sets `InsecureSkipVerify: true`, so self-signed certs just work — no CA install or HTTP fallback needed.

### Commands

```bash
# Firewall whitelist
cws tanswer firewall check-ip-is-white       # Check if IP is whitelisted
cws tanswer firewall search-white-list       # Search whitelist entries
cws tanswer firewall delete-white-list       # Remove from whitelist
cws tanswer firewall update-white-list-status  # Enable/disable whitelist entry

# Block rules
cws tanswer rules search-block-rules         # List block rules
cws tanswer rules create-block-rules         # Create a block rule
cws tanswer rules update-block-rules         # Update a block rule
cws tanswer rules update-block-rules-status  # Enable/disable a block rule
```

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
cws safeline-ce stat overview        # Aggregated stats: QPS, access, intercept counts
```

#### site

```bash
cws safeline-ce site list            # List all web services
cws safeline-ce site create          # Create a web service
cws safeline-ce site update          # Update a web service
cws safeline-ce site delete          # Delete a web service
```

#### rule (custom policy rules)

```bash
cws safeline-ce rule list            # List all custom rules
cws safeline-ce rule create          # Create a custom rule
cws safeline-ce rule update          # Update a custom rule
cws safeline-ce rule delete          # Delete a custom rule
cws safeline-ce rule switch          # Enable or disable a custom rule
```

#### ipgroup

```bash
cws safeline-ce ipgroup list         # List all IP groups
cws safeline-ce ipgroup get          # Get IP group details
cws safeline-ce ipgroup create       # Create an IP group
cws safeline-ce ipgroup update       # Update an IP group
cws safeline-ce ipgroup delete       # Delete an IP group
cws safeline-ce ipgroup append       # Add IPs to an IP group
```

#### log

```bash
cws safeline-ce log attack list      # List attack logs
cws safeline-ce log attack get       # Get attack log detail by ID
cws safeline-ce log rule list        # List rule-triggered attack logs
cws safeline-ce log rule get         # Get rule-triggered attack log detail
cws safeline-ce log audit list       # Get audit logs
```

#### skynet (enhanced detection rules)

```bash
cws safeline-ce skynet get           # Get enhanced rule configuration
cws safeline-ce skynet update        # Update enhanced rule configuration
cws safeline-ce skynet switch get    # Get global enable status of enhanced rules
cws safeline-ce skynet switch set    # Enable or disable enhanced rules globally
```

#### module (global semantics)

```bash
cws safeline-ce module get           # Get global semantics mode
cws safeline-ce module update        # Update global semantics mode
```

#### cert (system / license)

```bash
cws safeline-ce cert info            # Get system info
cws safeline-ce cert get             # Get license info
cws safeline-ce cert update          # Update management certificate
```
