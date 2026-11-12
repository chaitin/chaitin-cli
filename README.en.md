# chaitin-cli

[![CI](https://img.shields.io/github/actions/workflow/status/chaitin/chaitin-cli/ci.yml?branch=main&label=CI)](https://github.com/chaitin/chaitin-cli/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/chaitin/chaitin-cli?label=Release)](https://github.com/chaitin/chaitin-cli/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/chaitin/chaitin-cli?label=Go)](https://github.com/chaitin/chaitin-cli/blob/main/go.mod)
[![License](https://img.shields.io/github/license/chaitin/chaitin-cli?label=License)](https://github.com/chaitin/chaitin-cli/blob/main/LICENSE)

Unified CLI for Chaitin security products

## Overview

`chaitin-cli` is a unified command-line tool for Chaitin security products. It provides common operations, queries, and automation capabilities for multiple products from one binary. The goal is to reduce the friction caused by separate APIs, authentication patterns, and output formats, so developers, operators, and AI agents can manage SafeLine, X-Ray, CloudWalker, Veinmind, T-Answer, DDR, and related products consistently.

Core capabilities:

- Unified entrypoint: run product commands through `chaitin-cli <product> <command>`
- Reusable configuration: supports config files, environment variables, `.env`, and command-line flags
- Automation friendly: supports dry-run, JSON output where available, and an AI Skill for agent integration

## Quick Start

Install directly on macOS / Linux:

```bash
curl -fsSL https://raw.githubusercontent.com/chaitin/chaitin-cli/main/skills/chaitin-cli/scripts/install-chaitin-cli.sh | bash
```

Windows users should download the matching package from [GitHub Releases](https://github.com/chaitin/chaitin-cli/releases), extract `chaitin-cli.exe`, and add it to PATH.

Check the installed version:

```bash
chaitin-cli version
chaitin-cli --version
```

## Skill

This project provides an AI Agent skill. Once installed, AI agents (Claude Code, Cursor, etc.) can invoke `chaitin-cli` commands to manage Chaitin security products directly.

If you want an AI agent to call `chaitin-cli` automatically, install the Skill:

```bash
npx skills add chaitin/chaitin-cli
```

After installation, simply describe your needs to the AI agent, for example:

- "Show me recent attack logs in SafeLine"
- "Create a scan task in X-Ray"
- "List vulnerability events in CloudWalker"
- "Create a denoise task in CodeForce and check its result"

## Demo

### CloudWalker

[![asciicast](https://asciinema.org/a/894643.svg)](https://asciinema.org/a/894643)

### Veinmind

[![asciicast](https://asciinema.org/a/cTKHufj2Fewwl95j.svg)](https://asciinema.org/a/cTKHufj2Fewwl95j)

### T-Answer

[![asciicast](https://asciinema.org/a/Gs8KPOIcEnRRnXWr.svg)](https://asciinema.org/a/Gs8KPOIcEnRRnXWr)

### SafeLine

[![asciicast](https://asciinema.org/a/x7I8JDzbtjsRjb5M.svg)](https://asciinema.org/a/x7I8JDzbtjsRjb5M)

### SafeLine-CE

[![asciicast](https://asciinema.org/a/dzJzibRTm8arWRmU.svg)](https://asciinema.org/a/dzJzibRTm8arWRmU)

### DDR

[![asciicast](https://asciinema.org/a/IHulIbJ5nsy924qd.svg)](https://asciinema.org/a/t0cFuOjkLkExjREx)

### X-Ray

[![asciicast](https://asciinema.org/a/923077.svg)](https://asciinema.org/a/923077)

## Modules

| Module | Description |
| --- | --- |
| `chaitin` | Demo and basic commands |
| `safeline` | SafeLine WAF site, policy, ACL, attack log, and system information management |
| `safeline-3` | SafeLine-3 protected object, policy, ACL, log, monitor, system, and network management |
| `safeline-ce` | SafeLine CE site, rule, log, certificate, and enhanced protection management |
| `xray` | X-Ray scan task, asset, vulnerability, report, and system configuration management |
| `cloudwalker` | CloudWalker CWPP event, asset, vulnerability, protection policy, and system management |
| `veinmind` | CloudWalker CNAPP container, image, and escape protection management |
| `tanswer` | T-Answer semantic security operations for alarms, file alarms, assets, policies, response actions, and Open API fallback |
| `ddr` | DDR API token and connection configuration helpers |
| `dsensor` | D-Sensor security monitoring, agent, honeypot, alarm, and threat log management |
| `codeinsight` | CodeInsight project, repository configuration, scan task, and report export management |
| `codeforce` | CodeForce project, project AI employee, AI dev task, native audit, denoise, code package, repository, and Git auth management |
| `cosmos` | Cosmos / AISOC alarm, log, intelligence, IP block, asset, notice, ops, SOAR, and vulnerability management |

The root command handles configuration loading, product command registration, and BusyBox-style dispatch. Each product directory owns its commands, flags, configuration decoding, and API calls.

## Configuration

Prefer putting product connection settings in `~/.chaitin-cli/config.yaml`. If `./config.yaml` exists in the current directory and has a chaitin-cli product name at the top level, the CLI reads that local config first to keep existing workflows compatible:

```yaml
cloudwalker:
  url: https://cloudwalker.example.com/rpc
  api_key: YOUR_API_KEY

tanswer:
  url: 'https://<全悉 Web 端 IP>'
  api_key: '<全悉 OpenAPI Token>'

# chaitin-cli ddr get-api-token --url https://ddr.example.com:8443 --jwt-token "YOUR_JWT_TOKEN" can directly get url & api_key & company_id
ddr:
  url: "https://ddr.example.com:8443/qzh/api/v1"
  api_key: "YOUR_API_KEY"
  company_id: "YOUR_COMPANY_ID"

veinmind:
  url: "https://veinmind.example.com"
  api_key: "YOUR_64_CHARACTER_API_TOKEN"

xray:
  url: https://xray.example.com/api/v2
  api_key: YOUR_API_KEY

safeline-3:
  url: https://safeline3.example.com
  api_token: YOUR_API_TOKEN

dsensor:
  url: https://dsensor.example.com
  api_key: YOUR_API_KEY

codeinsight:
  url: https://codeinsight.example.com
  access_token: YOUR_ACCESS_TOKEN

codeforce:
  url: https://codeforce.example.com
  access_token: YOUR_ACCESS_TOKEN
  account_type: admin

cosmos:
  url: https://cosmos.example.com
  api_key: YOUR_JWT_BEARER_TOKEN

agent-compose:
  url: https://agent-compose.example.com
  api_token: YOUR_API_TOKEN
  default_project: YOUR_PROJECT
```
You can also put the same keys into environment variables or a local `.env` file. Variable names follow `<PRODUCT>_<FIELD>`:

```text
cloudwalker.url      -> CLOUDWALKER_URL
cloudwalker.api_key  -> CLOUDWALKER_API_KEY
tanswer.url          -> TANSWER_URL
tanswer.api_key      -> TANSWER_API_KEY
ddr.url              -> DDR_URL
ddr.api_key          -> DDR_API_KEY
ddr.company_id       -> DDR_COMPANY_ID
veinmind.url         -> VEINMIND_URL
veinmind.api_key     -> VEINMIND_API_KEY
xray.url             -> XRAY_URL
xray.api_key         -> XRAY_API_KEY
dsensor.url          -> DSENSOR_URL
dsensor.api_key      -> DSENSOR_API_KEY
codeinsight.url      -> CODEINSIGHT_URL
codeinsight.access_token -> CODEINSIGHT_ACCESS_TOKEN or CODEINSIGHT_TOKEN
codeforce.url        -> CODEFORCE_URL
codeforce.access_token -> CODEFORCE_ACCESS_TOKEN or CODEFORCE_API_KEY
codeforce.account_type -> CODEFORCE_ACCOUNT_TYPE
cosmos.url           -> COSMOS_URL
cosmos.api_key       -> COSMOS_API_KEY
safeline-ce.url      -> SAFELINE_CE_URL
safeline-ce.api_key  -> SAFELINE_CE_API_KEY
safeline-3.url       -> SAFELINE_3_URL
safeline-3.api_token -> SAFELINE_3_API_TOKEN
safeline.url         -> SAFELINE_URL
safeline.api_key     -> SAFELINE_API_KEY
agent-compose.url    -> AGENT_COMPOSE_URL
agent-compose.api_token -> AGENT_COMPOSE_API_TOKEN
agent-compose.api_key -> AGENT_COMPOSE_API_KEY (alias of api_token)
agent-compose.default_project -> AGENT_COMPOSE_DEFAULT_PROJECT
```

### T-Answer Quick Start

T-Answer commands use an OpenAPI Token. Create it in the T-Answer web console under `System Management` -> `Open API`, then configure `tanswer.api_key` or `TANSWER_API_KEY`.

Configure T-Answer with `--url`, `--api-key`, `--timeout`, and `--insecure`; environment variables `TANSWER_URL`, `TANSWER_API_KEY`, `TANSWER_TIMEOUT`, and `TANSWER_INSECURE`; or `tanswer.url`, `tanswer.api_key`, `tanswer.timeout`, and `tanswer.insecure` in `config.yaml`. Never commit a real token.

```bash
chaitin-cli tanswer auth check
chaitin-cli tanswer --help
chaitin-cli tanswer manifest
chaitin-cli tanswer system status
chaitin-cli tanswer alarm overview --time today
chaitin-cli tanswer asset list --page-size 10
```

For human operators using this repository, see [the T-Answer guide (Chinese)](./products/tanswer/README.md) and [command reference (Chinese)](./products/tanswer/COMMAND_REFERENCE.md). Users who only have the installed binary should start with `chaitin-cli tanswer --help`, then use domain or leaf-command help and `manifest` for the current contract. The root-level `--dry-run` flag does not apply to `tanswer`.

AI agents must use `chaitin-cli tanswer`, discover commands with help, and use `manifest` for machine-readable parameters, output fields, risks, and confirmation requirements. Prefer semantic commands. Use `tanswer api` only for a known, authorized endpoint, method, and request body not covered by a semantic command; never guess RPC methods, paths, or request bodies. For every semantic protected write, run `--preview`, present the target, impact, and risk to the user, and wait for explicit confirmation of that specific change before using the exact `--confirm` token. `tanswer api` GET/HEAD requests may run directly; every other method returns a preview by default and must receive explicit confirmation for that exact method, path, query, and body before `--confirm CONFIRM_TANSWER_RAW_API_WRITE` sends it. A confirmation token is a mechanical CLI requirement, not user authorization.

### SafeLine-3

SafeLine-3 commands use an OpenAPI token. Configure `safeline-3.api_token` or set `SAFELINE_3_API_TOKEN`.

SafeLine-3 calling guidance for AI agents lives in [`products/safeline3/agent-skill.md`](products/safeline3/agent-skill.md).

```bash
chaitin-cli safeline-3 node-group list --output json
chaitin-cli safeline-3 site list --type reverse-proxy --page 1 --page-size 20 --output json
chaitin-cli safeline-3 policy-group list --page 1 --page-size 20 --output json
chaitin-cli safeline-3 log attack list --start -24h --page 1 --page-size 20 --output json
chaitin-cli safeline-3 raw request GET /api/v3/license
```

For complex create, update, and delete requests, prefer semantic entity commands. Use file inputs such as `--payload-file` or `--application-file` for nested payloads. `raw request` is a fallback for unwrapped `/api/v3/...` endpoints.

Example `.env`:

```bash
SAFELINE_URL=https://safeline.example.com
SAFELINE_API_KEY=YOUR_API_KEY
XRAY_URL=https://xray.example.com/api/v2
XRAY_API_KEY=YOUR_API_KEY
```

Priority is `flags > environment/.env > recognized ./config.yaml > ~/.chaitin-cli/config.yaml`. If `./config.yaml` does not contain a chaitin-cli product name, it is treated as another project's config and ignored. When a local config is recognized, global config is not merged.

Use root-level `-c` or `--config` to load a different config file. This is useful when you switch between multiple product instances or temporarily use a project-local config file, for example multiple SafeLine environments:

```bash
chaitin-cli -c ./configs/safeline-prod.yaml safeline stats overview
chaitin-cli -c ./configs/safeline-staging.yaml safeline stats overview
```

Commands that create or update config use the same path selection by default: if a recognized `./config.yaml` exists in the current directory, they write to it first; otherwise they write to `~/.chaitin-cli/config.yaml`. When `-c` / `--config` is set explicitly, they write only to the specified file.

Use root-level `--dry-run` for commands that support dry-run:

```bash
chaitin-cli --dry-run xray plan PostPlanFilter --filterPlan.limit=10
chaitin-cli --dry-run cosmos asset save-host-asset --ip 10.0.0.1/32 --name demo-host --organization_id 1 --asset_ip_type 1 --category_ids '[{"id":41,"name":"Linux"}]' --group_id 1
```

### Cosmos / AISOC

Full module documentation lives in [`products/cosmos/README.md`](products/cosmos/README.md).

Cosmos generic JSON-RPC commands cover alarms, logs, intelligence, IP blocks, assets, notices, ops, SOAR, and vulnerabilities. Root-level `--dry-run` prints a redacted request summary without sending the request. Saving a host asset usually needs asset type, category, and group parameters on current deployments. `--asset_ip_type` currently uses `1` for a real IP and `2` for a virtual IP:

```bash
chaitin-cli cosmos asset search-host-asset --count 20 --offset 0 --raw

chaitin-cli --dry-run cosmos asset save-host-asset \
  --ip 10.0.0.1/32 \
  --name demo-host \
  --organization_id 1 \
  --asset_ip_type 1 \
  --category_ids '[{"id":41,"name":"Linux"}]' \
  --group_id 1 \
  --raw
```

### CodeInsight Projects And Tasks

```bash
export CODEINSIGHT_URL=https://codeinsight.example.com
export CODEINSIGHT_TOKEN=YOUR_ACCESS_TOKEN

chaitin-cli codeinsight project create --name demo-java --language java
chaitin-cli codeinsight repo-config create --name git-prod --repo-type git --git-provider gitlab --server-host https://git.example.com/group/demo.git --auth-type access_token --access-token GIT_TOKEN
chaitin-cli codeinsight task create repo --project-name demo-java --task-name demo-repo --rule-set-name Corax-Java --repo-config-name git-prod --ref-type branch --ref-name main
chaitin-cli codeinsight task result --task-id 12345
chaitin-cli codeinsight task result download --task-id 12345 --out ./reports/12345.json
```

### CodeForce Projects And Tasks

```bash
export CODEFORCE_URL=https://codeforce.example.com
export CODEFORCE_ACCESS_TOKEN=YOUR_ACCESS_TOKEN
export CODEFORCE_ACCOUNT_TYPE=admin

chaitin-cli codeforce project create --name demo-app --repository-id repo-1
chaitin-cli codeforce project ai-employee model-options --project-id project-1
chaitin-cli codeforce project ai-employee create --project-id project-1 --type dev --name backend-dev-agent --enabled
chaitin-cli codeforce project ai-dev create --project-id project-1 --employee-id employee-1 --title "Add repository health dashboard" --issue-url https://github.com/example/demo/issues/12 --branch feature/repo-health
chaitin-cli codeforce audit native create git --repository-id repo-1 --source-ref branch:main --audit-rule-id 101 --task-name "main native audit"
chaitin-cli codeforce code-management create --name demo-code-drop --version-description "2026-06 release" --file ./artifacts/demo.zip
chaitin-cli codeforce repository create project --name demo-git --platform gitlab --repositories-url https://git.example.com/group/demo.git --token GIT_TOKEN
chaitin-cli codeforce git-auth create --name github-personal --platform github --token GIT_TOKEN
chaitin-cli codeforce denoise parse --type sast --report-file ./reports/sast.json
chaitin-cli codeforce denoise create --type sast --name repo-noise-check --engineer-id engineer-1 --source-type repository --repository-name demo-repo --branch-or-tag main --report-file ./reports/sast.json
chaitin-cli codeforce denoise result --task-id denoise-task-1
```

Notes:

- `account_type=admin|user` targets the management APIs for project creation, AI employees, AI dev tasks, native audit, code management, project repositories, and git-auth.
- `account_type=openapi` targets the public CodeForce OpenAPI and is appropriate for `openapi whoami`, `denoise parse`, OpenAPI-backed `denoise create`, and some result lookups.
- If the current credential is really an OpenAPI key, management commands fail with an explicit message instead of pretending they succeeded.

## Project Structure

```text
main.go                         # Main entry point, root command, product registration, and BusyBox-style dispatch
config/                         # Config loading, environment/.env overrides, and config writing
products/<name>/                # One dedicated directory per product, with commands and API clients
products/<name>/cmd/            # Hand-written product command groups
products/<name>/client/         # Generated or wrapped product API clients
skills/chaitin-cli/             # AI Agent Skill and installer script
cmd/gen-cli/                    # CLI generation tooling
Taskfile.yml                    # Build, run, check, and package tasks
.github/workflows/ci.yml        # CI, cross-platform packaging, and GitHub Release publishing
```

## More Products

Add to `products` directory

Checklist for a new product:

- Add the product package import in `main.go`.
- Register the command in `newApp()` with `a.registerProductCommand(...)`.
- If `NewCommand()` returns `(*cobra.Command, error)`, handle the error before registration.
- If the product needs `config.yaml` or root-level runtime flags, implement `ApplyRuntimeConfig(...)` in the product package and call it from `wrapProductCommand()` in `main.go`.
- Decode product-specific config inside the product package from `config.Raw`; do not add config field parsing to the root command.

## BusyBox-Style Invocation

The same binary can be invoked directly by subcommand name through a symlink or by renaming the executable:

```bash
task build
ln -s ./bin/chaitin-cli ./chaitin
./chaitin
```

This is equivalent to:

```bash
./bin/chaitin-cli chaitin
```

## Development

Repository:

```bash
git clone https://github.com/chaitin/chaitin-cli.git
cd chaitin-cli
```

Environment:

- Use the Go version declared in [`go.mod`](./go.mod)
- Install [Task](https://taskfile.dev/) to use the `task` commands

Run locally:

```bash
go run . chaitin
go run . safeline --help
```

tasks:

```bash
task build
task run:chaitin
task fmt
task lint
task test
task package GOOS=linux GOARCH=amd64
```

Versioning:

- `task build` / `task package` inject the version with `-ldflags "-X main.version=..."`; by default it comes from `git describe --tags --always --dirty`.
- Pass an explicit version when needed, for example `task build VERSION=v1.2.3`.
- When no version is injected (for example `go run .` or a plain `go build`), the command falls back to Go build info: an untagged build that can be traced to a commit shows `dev+<commit>`, a `go install <module>@<version>` of a released tag shows the module version, and `dev` is shown when neither is available.
- Verify with `task build VERSION=v0.0.0-test && ./bin/chaitin-cli version`.

## Maintenance and Feedback

- Main branch: [`main`](https://github.com/chaitin/chaitin-cli/tree/main)
- Releases: cross-platform packages are published through [`GitHub Releases`](https://github.com/chaitin/chaitin-cli/releases)
- Release flow: pushing a `v*` tag triggers GitHub Actions to test, package, and publish a release
- Maintenance status: actively maintained; PRs for product commands, docs, and bug fixes are welcome

Please submit feature requests and bug reports through [GitHub Issues](https://github.com/chaitin/chaitin-cli/issues). Search existing issues before opening a new one. Do not post API keys, tokens, real service URLs, or other sensitive information in public issues.

## FAQ

**What should I do if `chaitin-cli` is not found after installation?**

Check whether the install directory is in `PATH`. The macOS / Linux installer prefers user or system PATH directories. If the current shell has not refreshed, reopen the terminal or run the binary with the full path printed by the installer.

**Where is configuration loaded from?**

By default, configuration is loaded first from a current-directory `./config.yaml` recognized as chaitin-cli config, otherwise from `~/.chaitin-cli/config.yaml`. Priority is `flags > environment/.env > recognized ./config.yaml > ~/.chaitin-cli/config.yaml`. Use root-level `-c` or `--config` to load a different config file.

**What should I do if a self-signed certificate causes connection failures?**

Some product commands expose `--insecure`. SafeLine skips TLS verification by default. Behavior differs by product, so run the relevant product command with `--help` to check supported flags.

**How do I publish a new version?**

Create and push a `v*` tag from `main`, for example `v2605.0.0`. GitHub Actions will build cross-platform packages and publish them to Releases.
