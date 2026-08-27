# 媒体与码流

媒体服务门面（`client.Media()`）覆盖 profile、流/快照 URI、编码配置。本文讲
其中经过真机验证的语义部分：profile 选择、StreamSetup 参数化、响应解析。

## 选对 Profile

盲目取 `profiles[0]` 会在「低分辨率 profile 排第一」的设备上静默录成子码流
——一切正常，直到你发现分辨率不对。两个助手封装了真机验证过的启发式：

```go
profiles, _ := client.Media().GetProfiles(ctx)

mainToken := onvif.SelectMainProfile(profiles)
subToken  := onvif.SelectSubProfile(profiles, mainToken) // "" = 无独立子码流
```

**`SelectMainProfile`** —— 取像素数（W×H）最大的 profile。仅在完全平手时用
命名线索裁决（`main`/`primary`/`主流`/`主码流`/`channel1` 胜出；
`sub`/`secondary`/`辅流`/`辅码流`/`extra` 出局），因为 OEM 固件乱起名，
命名永远不能当首要依据。全部无分辨率信息时回退列表序（取第一个）。

**`SelectSubProfile`** —— 除主码流外像素数最大、且**严格小于**主码流分辨率
者。与主码流**同分辨率**的第二个 profile 不算子码流：部分硬件（Amcrest
IP4M 模式）同分辨率的两个 token 是同一路流的两个句柄。返回 `""` 表示没有
独立子码流。

## 指定传输方式的流 URI

`GetStreamURI` 请求 RTP-Unicast + RTSP。部分设备按请求的协议决定返回什么
——ESP32 固件只在被请求 RTSP 时返回带 G.711 音轨的 RTSP 地址，否则返回
纯视频的 HTTP 地址。`GetStreamURIWithOptions` 把选择权交给你：

```go
uri, err := client.Media().GetStreamURIWithOptions(ctx, profileToken,
    onvif.StreamSetup{
        Stream:    onvif.StreamRTPUnicast,   // 或 StreamRTPMulticast
        Transport: &onvif.Transport{Protocol: onvif.ProtocolRTSP}, // HTTP / UDP / TCP
    })
```

Transport 传 nil 或空协议默认 RTSP。`GetStreamURI(ctx, token)` 恒等于
`RTP-Unicast + RTSP`——行为不变。

## 响应解析保证

ONVIF 媒体响应比规范更多变：命名空间前缀（`trt:`/`tt:`/默认）、SOAP 1.1 与
1.2 信封、偶尔缺失的 `MediaUri` 包装层。解析是分层的：

1. 类型化结构体按 local name 匹配——任意前缀、两种 SOAP 版本；
2. 类型化路径拿不到 URI 时，对原始响应做 local-name 扫描提取第一个 `Uri`
   元素；
3. 仍然没有 URI 时，返回显式 `ErrEmptyMediaURI` 错误并附截断的响应体摘要
   ——绝不再是历史上的「空字符串 + nil error」；
4. 无论 Fault 搭配什么 HTTP 状态（包括 200 带 Fault）都会被检测并返回
   结构化的 `*FaultError`。

`GetSnapshotURI` 共享同一响应形态，享受同样保证。

## 编码与 OSD 配置

除拉流外，门面还覆盖视频/音频编码配置（`Get/SetVideoEncoderConfiguration`
系列）、OSD 管理、组播配置（`Start/StopMulticastStreaming`）与同步点。完整
操作列表见
[Go 参考文档](https://pkg.go.dev/github.com/mickeyzzc/onvif-go)。
