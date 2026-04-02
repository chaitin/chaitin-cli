# T-Answer CLI

全悉（T-Answer）命令行工具，通过 OpenAPI 控制全悉流量威胁检测系统。

## 配置

在当前工作目录的 `config.yaml` 中配置：

```yaml
tanswer:
  url: https://<全悉 Web 端 IP>
  api_key: <全悉 OpenAPI Token>
```

## 命令列表

| 命令 | 说明 |
| ---- | ---- |
| cws tanswer firewall check-ip-is-white         | CheckIpIsWhite 检查 IP 是否在白名单中 |
| cws tanswer firewall delete-white-list         | DeleteWhiteList 响应处置 / 响应白名单：删除响应白名单 |
| cws tanswer firewall search-white-list         | SearchWhiteList 响应处置 / 响应白名单：搜索响应白名单 |
| cws tanswer firewall update-white-list-status  | UpdateWhiteListStatus 响应处置 / 响应白名单：启用或禁用响应白名单 |
| cws tanswer rules    create-block-rules        | CreateBlockRules 响应处置 / 旁路阻断策略：创建旁路阻断策略 |
| cws tanswer rules    search-block-rules        | SearchBlockRules 响应处置 / 旁路阻断策略：搜索旁路阻断策略 |
| cws tanswer rules    update-block-rules        | UpdateBlockRules 响应处置 / 旁路阻断策略：编辑旁路阻断策略 |
| cws tanswer rules    update-block-rules-status | UpdateBlockRulesStatus 响应处置 / 旁路阻断策略：启用或禁用旁路阻断策略 |

产品级覆盖参数统一为 `--url` 和 `--api-key`。
