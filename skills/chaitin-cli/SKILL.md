---
name: chaitin-cli
description: "Use when running chaitin-cli commands to manage Chaitin security products: SafeLine WAF (site management, IP blocking, ACL, policy rules, attack logs), X-Ray vulnerability scanner (scan tasks, results, assets), CloudWalker CWPP (events, vulnerabilities, assets), and T-Answer (firewall rules, blocklists)."
version: 1.0.0
author: chaitin
tags: [chaitin-cli, safeline, xray, cloudwalker, tanswer, waf, security, chaitin, cli]
---

# chaitin-cli Usage Guide

> Unified CLI for Chaitin security products. Manage SafeLine WAF, X-Ray scanner, CloudWalker CWPP, and T-Answer through a single tool.

## No-Argument Behavior

When `/chaitin-cli` is invoked without any arguments (empty `ARGUMENTS`):

1. Greet the user and introduce this skill in a well-formatted way, based on the SKILL.md content.
2. Run `command -v chaitin-cli` to check if `chaitin-cli` is already installed.
3. If found, report the installed path (e.g. `chaitin-cli is installed at /opt/homebrew/bin/chaitin-cli`).
4. If not found, install it per platform:
   - Windows: Tell the user to manually download the latest release from `https://github.com/chaitin/chaitin-cli/releases`, extract `chaitin-cli.exe`, and add it to PATH. Do not attempt automated installation on Windows.
   - macOS, Linux: Run `bash scripts/install-chaitin-cli.sh`. The script outputs the installed binary path on stdout (last line). Remember this path — subsequent commands must use the full path (e.g. `/home/user/.local/bin/chaitin-cli`) because each Bash invocation starts a new shell and the install directory may not yet be in PATH.
5. After the setup check, briefly tell the user what they can do next — for example: "You can now use chaitin-cli to manage SafeLine, X-Ray, CloudWalker, or T-Answer. Tell me what you'd like to do, or run `chaitin-cli --help` to explore commands."

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
3. **TLS with self-signed certs** — `chaitin-cli xray` takes `--insecure` (off by default). `chaitin-cli safeline` also exposes `--insecure`, but its default is `true` (already skipping verification); pass `--insecure=false` to re-enable verification. `chaitin-cli safeline-ce`, `chaitin-cli cloudwalker`, and `chaitin-cli tanswer` don't expose the flag and always skip TLS verification in their HTTP clients.
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
chaitin-cli -c ./configs/prod.yaml safeline stats overview
chaitin-cli -c ./configs/staging.yaml safeline stats overview
```

### Global Flags

| Flag | Description |
|------|-------------|
| `-c, --config` | Config file path (default: `./config.yaml`) |
| `--dry-run` | Print the API request without executing. Applied by the root command to `xray` and `cloudwalker`. `safeline` registers its own `--dry-run` and forwards it to subcommands. `safeline-ce` inherits the root flag, but the current codebase stores the value without using it; `tanswer` ignores it. |

### Discovering Commands

`--help` is the authoritative source — this document does not enumerate every flag.

```bash
chaitin-cli <product> --help                # List subcommand groups for a product
chaitin-cli <product> <group> --help        # List commands in a group
chaitin-cli <product> <group> <cmd> --help  # List flags for a specific command
```

`chaitin-cli xray` commands are auto-generated from the X-Ray OpenAPI spec (hundreds of operations); `chaitin-cli xray <category> --help` is the only complete reference. `chaitin-cli cloudwalker` has 60+ command groups with similar depth.

### Operating Rules

For SafeLine, X-Ray, CloudWalker, T-Answer, and SafeLine-CE tasks, treat `chaitin-cli` as the only supported operator interface.

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
| `chaitin-cli xray` | JSON (no alternative) | — | `--debug` for debug logs |
| `chaitin-cli cloudwalker` | text | `-f json` (or `--format json`) | `--no-trunc` to disable text truncation |
| `chaitin-cli tanswer` | formatted text | `--raw` (bool) | — |

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
chaitin-cli tanswer firewall check-ip-is-white       # Check if IP is whitelisted
chaitin-cli tanswer firewall search-white-list       # Search whitelist entries
chaitin-cli tanswer firewall delete-white-list       # Remove from whitelist
chaitin-cli tanswer firewall update-white-list-status  # Enable/disable whitelist entry

# Block rules
chaitin-cli tanswer rules search-block-rules         # List block rules
chaitin-cli tanswer rules create-block-rules         # Create a block rule
chaitin-cli tanswer rules update-block-rules         # Update a block rule
chaitin-cli tanswer rules update-block-rules-status  # Enable/disable a block rule
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
