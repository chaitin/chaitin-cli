1.从 config.yaml 加载 openapi 配置
2.从 openapi.json 读取嵌入的 swagger.json -> 遍历 paths -> 为每个 API 动态实例化 cobra.Command -> 遍历 parameters 动态添加 Flags (支持 string, int, bool)
3.实现通用的 HTTP Client，拦截 CLI 的输入指令，自动拼装 URL 参数和 JSON Body。
4.自动在 Request Header 中注入对应产品的 Token (X-SLCE-API-TOKEN)。

2.1 站点与证书管理 (Site & Cert)
● 1. 获取防护站点列表
    ○ ct-cli safeline site list --page 1 --size 20
    ○ (映射: GET /api/open/sites)
● 2. 创建新防护站点
    ○ ct-cli safeline site create --domain api.example.com --port 443 --upstream 10.0.0.1:8080 --ssl true
    ○ (映射: POST /api/open/sites)
● 3. 删除指定站点
    ○ ct-cli safeline site delete --id 1024
    ○ (映射: DELETE /api/open/sites/{id})
● 4. 上传 SSL 证书
    ○ ct-cli safeline cert upload --cert-file ./server.crt --key-file ./server.key
    ○ (映射: POST /api/open/certs - 自动处理文件类型的 payload)
● 5. 查看证书列表与过期时间
    ○ ct-cli safeline cert list -o json
2.2 防护规则与黑白名单 (Rules & ACL)
● 6. 获取当前自定义规则列表
    ○ ct-cli safeline rule list
● 7. 创建自定义拦截规则 (如拦截特定 UA)
    ○ ct-cli safeline rule create --name "Block-Bad-Bot" --condition "User-Agent contains python-requests" --action block
    ○ (映射: POST /api/open/rules)
● 8. 添加 IP 黑名单 (高频应急场景)
    ○ ct-cli safeline blacklist add --ip 1.2.3.4 --reason "Malicious scanning" --expire 3600
● 9. 添加 IP 白名单 (业务放行场景)
    ○ ct-cli safeline whitelist add --ip 192.168.1.0/24
    ○ (映射: POST /api/open/allowlist)
2.3 日志与数据统计 (Logs & Stats)
● 10. 查看全局拦截统计大盘数据
    ○ ct-cli safeline stat overview
    ○ (映射: GET /api/open/stat)
● 11. 导出攻击日志 (用于对接 SIEM/SOC)
    ○ ct-cli safeline log attack --start-time "2026-03-30T00:00:00Z" --limit 500 -o json
    ○ (映射: GET /api/open/logs/attack)
● 12. 获取 WAF 节点运行状态与负载
    ○ ct-cli safeline node status
    ○ (映射: GET /api/open/nodes/status)
