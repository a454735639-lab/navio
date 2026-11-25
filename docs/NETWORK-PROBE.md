# 网络位置探针（Network Location Probe）

> 解决的问题：面板用 **https 域名**访问时，浏览器禁止页面探测 `http://192.168.x` / `http://100.x`，
> 导致"内外网自动切换"失效。探针提供两个 **https** 端点，让前端在 https 页面下也能判断
> 当前是「内网 / Tailscale / 外网」，从而按 192 → 100.x → 域名 的顺序打开卡片。

## 架构

```
sun-panel 容器
  ├─ :3002  面板（http）
  ├─ :8443  TLS 探针（GET /probe → 204，证书热加载）
  └─ 内置 ACME（provider=acme 时）：DNS-01 签发/续期，Token 加密存库

数据流（前端）：
  fetch(https://lan-probe.example.com:8443/) 通 → 在家 → 用卡片内网地址
  fetch(https://ts-probe.example.com:8443/)  通 → tailnet → 用 100.x 地址
  都不通 → 外网域名
```

## 前置条件

1. 一个托管在支持 DNS API 的域名（当前实现：Cloudflare）
2. 两条 **灰云（DNS only，不开代理）** 的 A 记录：

| 记录 | 内容 | 用途 |
|---|---|---|
| `lan-probe.example.com` | `192.168.3.20`（面板所在内网 IP） | 判断"在家" |
| `ts-probe.example.com` | `100.66.1.2`（面板的 Tailscale/组网 IP） | 判断"在 tailnet" |

> 私网 IP 做 A 记录是合法的；**必须灰云**，开了代理探测会打到 Cloudflare，判定永远"可达"。

## 部署

```yaml
# docker-compose.yaml 片段
services:
  sun-panel:
    image: ghcr.io/xxx/sun-panel:latest
    ports:
      - "3002:3002"
      - "8443:8443"          # 探针端口
    volumes:
      - ./conf:/app/conf     # 数据 + 证书（provider=acme 时面板要可写 conf/probe）
```

启动后在「外观设置 → 网络位置探针」配置：

- **path**：已有证书文件（`certPath/keyPath`），适合已经有反代证书的用户
- **upload**：把 pem 上传到 `conf/probe/` 后填路径
- **acme**：填邮箱、两个探针域名、DNS 提供商、API Token（保存后加密存储，界面不再显示）；
  面板会自动签发并在剩余 <30 天时自动续期，无需外部容器

另需在前端配置两个探测地址（同一设置页）：

```
LAN 探针地址       https://lan-probe.example.com:8443/
Tailscale 探针地址 https://ts-probe.example.com:8443/
```

## 续期

- 12 小时巡检一次，证书剩余 <30 天自动续期
- 新证书原子替换旧文件，**面板按 mtime 热加载，无需重启**
- 记录在 `docker logs sun-panel`（过滤 `[probe]`）

## 故障排查

| 现象 | 检查 |
|---|---|
| 首页角标一直是 WAN | 两条 A 记录是否灰云；探针 URL 是否填了 https；从设备 `curl -v https://lan-probe…:8443/` |
| 8443 不通 | `docker logs sun-panel | grep probe`；证书文件是否存在、权限 600 |
| acme 签发失败 | Token 权限需 `Zone.DNS:Edit`；看日志中 `cloudflare:` / `propagation` 行 |
| DNS 传播慢 | 已默认在 TXT 写入后等 25s；仍失败可稍后重试（12h 巡检会自动重试） |
| 证书换新后不生效 | 证书路径不变时自动热加载；改路径在设置里保存即可，无需重启 |
| 面板重启后探针没起来 | 8443 被占用会跳过并打日志，换端口即可 |

## 安全说明

- DNS API Token 用 AES-GCM 加密存数据库，密钥文件 `conf/probe/.secret`（600，仅本机有效）
- API 永不回传 Token（`GET` 只返回 `hasToken`）
- 证书私钥文件权限 600；挂载目录建议不要暴露给其他容器
- 探针端点只响应 `GET /probe`，其余 404；建议只在内网/组网 DNS 里解析，不要公网可达
