# 生产部署与加固

把 onvif-go server（或嵌入它的宿主）部署到真实网络。贯穿全文的主题：
库自带安全默认值，但相机后端是贴着互联网的基础设施——下面每个开关
都来自真实部署的需要。

## TLS 部署

SOAP server 说纯 HTTP（ONVIF 的常见形态）。在入口层终结 TLS，库留在
其后：

```nginx
# nginx：HTTPS 对外，纯 HTTP 到 :8080 的 onvif-go server
location /onvif/ {
    proxy_pass http://127.0.0.1:8080/onvif/;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
}
```

终结 TLS 之后有两件事要跟：

1. **宣告正确的主机地址。** Discovery `XAddrs` 与 `GetStreamUri` 应答
   必须带*客户端*可达的地址——设置 `Config.AdvertiseHost`（地址会变的
   场景用 `WithAdvertiseHostProvider`，例如 DHCP 续租）。Responder 绝
   不回显请求方地址（那个 bug 会把 NVR 注册成相机——见 discovery 指
   南）。
2. **认证锁定用的客户端 IP。** 防爆破按源 IP 记账。代理后面要转发真
   实客户端 IP（上面片段已做），并确认你的 HTTP 层把它呈现为
   `RemoteAddr` 或嵌入宿主的 `RequestContext`。

多网卡主机（相机常见：有线 + 无线）：显式设置 `XAddrs`，不要依赖按
对端推导的地址。

## 认证策略选择

`Config.Username`/`Config.Password` 启用 WS-UsernameToken 校验；按动作
的策略决定哪些动作真正需要它（细节见[认证指南](authentication.md)）：

| 部署形态 | 策略 |
|---|---|
| 封闭监控 LAN、只有 NVR 客户端 | 默认：写类动作认证（`Set*`、`Remove*`、`Create*`、`Go*`、`SystemReboot`），读开放——只拉流/拉 profile 的 NVR 无需凭证。 |
| 交换机以外可达的任何环境 | `AuthPolicy{All: true}`——所有动作认证。配合 digest token（见下）。 |
| 实验台 / 抓包回放 | `AllowAnonymous` 文档化模式——绝不上生产网。 |

优先使用 `PasswordDigest` 客户端：密码不过线，且重放窗口（nonce +
Created）有效。`AllowPasswordText` 为互操作默认开启；两端都可控时关
掉（`AuthPolicy.AllowPasswordText = false`）。

## 限速与包体上限

- `HandlerOptions.MaxBodyBytes`（默认 1 MiB）：超限 SOAP 请求在解析
  前即以 413 拒绝。保持默认；只有真的需要超大厂商扩展时才调高。
- `AuthFailureLimit`（默认 5）/ `AuthLockout`（默认 60s）：按源 IP 的
  防爆破后端。锁定源收到 401 且不做任何凭证运算。对锁定计数告警
  （见下）——那是有人在敲门。
- WS-Discovery：responder 在组播组上应答 Probe。设计上无认证；暴露
  的只有你配置的 scopes。若 discovery 会到不可信网段，`Scopes` 里别
  放站点名。

## 监控指标接入

`metrics` 包是观测缝（本模块零 Prometheus 依赖）：

```go
bridge := myprom.NewOnvifBridge()           // 实现 metrics.Hooks
srv, _ := server.New(cfg, server.WithMetrics(bridge))
responder := discoveryserver.NewResponder(discoveryserver.Config{
    // ...
    Metrics: bridge,
})
```

值得上面板的五个计数：按 action 的 `SoapRequest`（流量构成）、按
action 的 `SoapFault`（故障率分子——只计 handler 错误，不含认证拒
绝）、`AuthFail`（凭证猜测）、`AuthLockout`（生效中的锁定）、
`DiscoveryProbeAnswered`（扫描噪音基线）。可运行的桥接示例在
`examples/metrics-bridge`。

## NVR 对接清单

字节稳定是契约——NVR 侧解析器普遍按裸 SOAP 局部名匹配：

- [ ] `GetStreamUriResponse → MediaUri/Uri` 元素顺序不变（golden 已
      固化；任何变更都是破坏性发版）。
- [ ] ProbeMatches scopes 宣告 `onvif://www.onvif.org/name/…`、
      `hardware/…`、`location/…`——填上，NVR 会显示。
- [ ] 流 URI 带宣告的主机地址，不是 loopback 或 0.0.0.0。
- [ ] Digest 互操作：nonce + Created 齐备，`PasswordDigest` 类型 URI
      精确；NVR 时钟偏移须落在你的重放窗口内。
- [ ] 受保护动作缺凭证时回 401（不是 403）——ONVIF 客户端把 401 当
      挑战信号。
- [ ] 快照端点（默认 `/onvif/snapshot`）可达（若 NVR 轮询它）；在
      `All` 策略下它与写动作同等认证。

把 NVR 指过去之前，先用 `cmd/onvif-diagnostics` 对已部署端点跑一
遍——它按 NVR 的方式过一遍 discovery、认证、profile 与流 URI。
