# D-Sensor

D-Sensor (谛听) security monitoring platform commands for `chaitin-cli`.

The command tree is generated from the embedded D-Sensor OpenAPI schema and is available under:

```bash
chaitin-cli dsensor --help
```

## Configuration

Use `~/.chaitin-cli/config.yaml`, or a recognized local `./config.yaml`:

```yaml
dsensor:
  url: https://dsensor.example.com
  api_key: YOUR_API_KEY
```

Or environment variables:

```bash
export DSENSOR_URL=https://dsensor.example.com
export DSENSOR_API_KEY=YOUR_API_KEY
```

Flags override config and environment values:

```bash
chaitin-cli dsensor --url https://dsensor.example.com --api-key YOUR_API_KEY agent list
```

## Examples

```bash
chaitin-cli dsensor --help
chaitin-cli dsensor agent --help
chaitin-cli dsensor agent list --limit 10
chaitin-cli dsensor agent detail --sn abc123
chaitin-cli dsensor alarm smtp set --body '{"server":"smtp.example.com","port":25}'
chaitin-cli dsensor honeypot list --honeynet-id xxx
```

## Request Bodies

For API operations with request bodies, pass either field flags or raw JSON:

```bash
chaitin-cli dsensor agent detail --sn abc123 --node-sn node1
chaitin-cli dsensor alarm smtp set --body '{"server":"smtp.example.com","port":25}'
chaitin-cli dsensor alarm smtp set --body-file ./smtp-config.json
```

`--body` / `--body-file` and field-level flags are mutually exclusive.

## Cache

The D-Sensor command stores a small server-version cache in `~/.dsensor/cache/`.

```bash
chaitin-cli dsensor --refresh-cache agent list
```
