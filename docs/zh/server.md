# ONVIF 服务端

`server/` 是一个可嵌入也可独立运行的 ONVIF 服务端实现：带 Device、
Media、PTZ、Imaging 服务的虚拟摄像头模拟器。v2 传输层（`server/soap`）
面向真实设备嵌入重构——请求上下文、按动作鉴权、字节级可预测的 XML
输出。

## 传输层架构

```
POST /onvif/device_service
        │
        ▼
soap.Handler.ServeHTTP
        │  1. 提取动作（local name，规范 化：GetStreamURI → GetStreamUri）
        │  2. 解码信封（WS-Security 头 + 原始 inner-XML 请求体）
        │  3. requiresAuth(action)? → authenticate(header)
        │  4. 分发：ContextHandler(RequestContext, body []byte)
        ▼
响应写出器（golden 测试锁定的字节布局，RawXML 直通）
```

请求体以**原始 inner XML 字节**到达 handler——用 `soap.ParseRequest`
（或直接 `encoding/xml`）解码成具体请求类型。v1 传输层交给 handler 的
值是 `encoding/xml` 永远填不进去的空值，请求参数（如 `ProfileToken`）
静默丢失；v2 修好了这条管道。

## 请求上下文

以 `RegisterContextHandler` 注册的 handler 能感知请求来源——多网卡
地址通告的前提：

```go
handler.RegisterContextHandler("GetStreamUri", func(rc *soap.RequestContext, body []byte) (interface{}, error) {
    ip := rc.RemoteIP              // 客户端 IP（RemoteAddr 的 host 部分）
    ctx := rc.Context()            // 请求 context：取消、超时
    _ = rc.Request                 // 底层 *http.Request
    ...
})
```

`rc.Action` 携带 WSDL 规范动作名。旧签名
`RegisterHandler(action, func(body []byte) ...)` 作为薄包装继续可用。

### XAddr 回显

真实摄像头对每个客户端回答**从该客户端网络可达**的 URL。模拟器默认
采取同样行为：

- `Config.AdvertiseHost` 为空（默认）：把请求方源 IP 回显为所有通告
  URL 的 host——GetCapabilities/GetServices 的 XAddr、流/快照 URI。
- `Config.AdvertiseHost` 非空：该 host 处处生效（固定 DNS 名、反向
  代理等场景）。

## 按动作鉴权（#16）

配置了凭据时，默认策略只对**写类动作**要求认证——`Set*`、`Remove*`、
`Create*`、`Go*`，外加 `SystemReboot` 和 `Config.AuthProtectedActions`
里列出的名字。读操作对无凭据的发现客户端保持开放。未配置凭据时全部
开放（行为不变）。

WS-Security UsernameToken 两种口令形式都接受：

| 形式 | 校验 |
|---|---|
| PasswordDigest（默认） | `Base64(SHA1(nonce + created + password))`，常量时间比较 |
| PasswordText | 明文比较；`AuthPolicy{AllowPasswordText: false}` 可关闭 |

```go
handler := soap.NewHandlerWithOptions(soap.HandlerOptions{
    Username: "admin",
    Password: "secret",
    Auth: &soap.AuthPolicy{          // nil → DefaultAuthPolicy()
        Prefixes: []string{"Set", "Remove", "Create", "Go"},
        Actions: []string{"SystemReboot"},
        All:     false,              // true = 严格模式（v1 行为）
    },
})
```

## 字节级 XML 输出（#18）

响应写出器手工拼装信封；字节布局由 golden 测试锁定，做字节级匹配的
消费方因此保持稳定。

**规范大小写。** 响应元素使用 ONVIF WSDL 拼写——
`GetStreamUriResponse`/`GetSnapshotUriResponse`（不是 `...URI...`），
内层 `MediaUri`/`Uri`。进入的旧拼写（`GetStreamURI`）仍分发到规范
handler。

**默认形式**（与历史线上格式一致）：

```xml
<?xml version="1.0" encoding="UTF-8"?>
<Envelope xmlns="http://www.w3.org/2003/05/soap-envelope">
  <Body>
    <GetStreamUriResponse xmlns="http://www.onvif.org/ver10/media/wsdl">
      <MediaUri>
        <Uri>rtsp://198.51.100.7:8554/stream0</Uri>
```

**显式前缀**（`Config.ExplicitPrefixes` 或
`HandlerOptions.ExplicitPrefixes`）：

```xml
<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
  <s:Body>
    <trt:GetStreamUriResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
      <trt:MediaUri>
        <trt:Uri>rtsp://198.51.100.7:8554/stream0</trt:Uri>
```

前缀表：`tds:` `trt:` `tptz:` `timg:` `tev:` `tt:` `trc:` `tan:`。
无约定前缀的命名空间保持默认 `xmlns` 声明形式。

**Raw 通道。** 任意 handler 返回 `soap.RawXML` 即可把预构建字节原样
嵌入——完全绕过 `encoding/xml`，精确控制元素名、前缀、格式。Raw 输出
不会被前缀模式改写。

## 后续

M3（#23）把模拟器的内存状态改造为可插拔的 provider 接口（DeviceInfo /
StreamURI / Snapshot / Imaging / PTZ），让同一传输层可以直接前接真实
摄像头栈。见 [v2-architecture.md](v2-architecture.md)。
