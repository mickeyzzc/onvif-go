# onvif-go

[![CI](https://github.com/mickeyzzc/onvif-go/actions/workflows/ci.yml/badge.svg)](https://github.com/mickeyzzc/onvif-go/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/mickeyzzc/onvif-go/v2.svg)](https://pkg.go.dev/github.com/mickeyzzc/onvif-go/v2)
[![Go Report Card](https://goreportcard.com/badge/github.com/mickeyzzc/onvif-go/v2)](https://goreportcard.com/report/github.com/mickeyzzc/onvif-go/v2)
[![License](https://img.shields.io/github/license/mickeyzzc/onvif-go)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.26-00ADD8.svg)](https://go.dev/)

> 面向 Go 的 ONVIF 客户端 + 虚拟相机服务器——服务门面 API、主动与被动双模发现、
> 能对付各种怪固件的鉴权梯队、PTZ / 媒体 / 事件 / 成像。零第三方依赖。
> 由 [@mickeyzzc](https://github.com/mickeyzzc) 维护。

> [**English**](README.md)

一个经过生产锤炼的 Go 库，用于与 ONVIF 兼容的网络摄像头、NVR 和监控设备通信。
设备兼容性来自真实部署：海康、安讯士（Axis）、大华、博世（Bosch）、Amcrest、
海思 OEM 硬件、极简嵌入式实现（ESP32），以及那种「不管问什么都回 SOAP Fault」
的固件。它是 [MiBeeNvr](https://github.com/Mi-Bee-Studio/MiBeeNvr) 的 ONVIF
底层——后者是跑在 ARM64 上 7×24 小时的 NVR。

## 功能特性

**客户端** —— 每个 ONVIF 操作挂在其所属的服务对象上，与 ONVIF 服务模型一一对应：

- `client.Device()` —— 设备信息、能力（带缓存、single-flight、弱设备可降级到
  最小能力集）、网络/系统/DNS/NTP 配置、存储、WiFi、证书、用户管理
- `client.Media()` —— profile 与主/子码流选择助手、流 URI
  （`GetStreamURIWithOptions`：RTSP/HTTP/UDP × 单播/组播，命名空间宽容解析）、
  快照、编码/音频/OSD 配置
- `client.PTZ()` —— 转动、状态、预置位
- `client.Imaging()` —— 曝光、聚焦、成像设置
- `client.Events()` —— 托管 PullPoint 订阅（后台轮询、自动续订、
  `ErrEventsNotSupported` 哨兵）+ 原始原语
- `client.DeviceIO()`、`client.Security()` —— 继电器、I/O、用户管理

**为真实固件打造的鉴权** —— `WithAuthMode` 选择 digest / 明文 token /
HTTP Basic / 不鉴权；`WithAuthFallback` 提供自动回退梯队，并记住设备接受的
第一个模式。`errors.Is(err, onvif.ErrUnauthorized)` 可靠分类鉴权失败
（HTTP 401/403、NotAuthorized Fault、200 带 Fault）。`WithAutoClockSkew`
自动测量并校正设备时钟偏差（海康 "sender not authorized" 陷阱），
`DiagnoseAuth` 能区分时钟偏差、密码错误和拒绝 ONVIF 鉴权的设备。

**三种发现方式** —— `discovery.Discover`（组播探测）、`discovery.Listener`
（被动：听相机上电 Hello 和别人探测的应答，与主动发现共存）、
`discovery.ProbeEndpoint` / `ProbeSerial`（纯 HTTP 定向探测，跨子网触达组播
够不着的设备）。后处理：`FilterONVIFDevices`（滤掉 Synology/Windows/打印机
幽灵应答者）、`ParseScopes`、并行 `EnrichDevices`（限并发身份补全）。

**健壮性** —— 每次调用都做 Fault 检测（200 带 Fault 绝不会被判成成功）、
结构化 `FaultError`/`HTTPStatusError`、空 URI 响应变成带响应体摘要的显式
错误、能力 XAddr 修复（相机漫游后的陈旧宣传地址）、以及有文档有测试的并发
契约：一个 `Client`、多个 goroutine、无需外部加锁。

**虚拟相机服务器** —— `server/` 模拟 ONVIF 相机，不用硬件就能测你的录像软件。

## 安装

```bash
go get github.com/mickeyzzc/onvif-go/v2
```

要求 Go **1.26+**。库模块零第三方依赖。

## 快速上手

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/mickeyzzc/onvif-go/v2"
)

func main() {
    client, err := onvif.NewClient("http://192.168.1.100/onvif/device_service",
        onvif.WithCredentials("admin", "pass"),
        onvif.WithAutoClockSkew()) // 应对时钟偏差的相机
    if err != nil {
        log.Fatal(err)
    }
    if err := client.Initialize(context.Background()); err != nil {
        log.Fatal(err)
    }

    info, err := client.Device().GetDeviceInformation(context.Background())
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("%s %s（固件 %s，序列号 %s）\n",
        info.Manufacturer, info.Model, info.FirmwareVersion, info.SerialNumber)

    profiles, err := client.Media().GetProfiles(context.Background())
    if err != nil {
        log.Fatal(err)
    }
    mainToken := onvif.SelectMainProfile(profiles) // 分辨率最高者，而非 profiles[0]
    uri, err := client.Media().GetStreamURI(context.Background(), mainToken)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("主码流:", uri.URI)
}
```

固件拒绝 digest 鉴权？加一条梯队：

```go
client, err := onvif.NewClient(endpoint,
    onvif.WithCredentials("admin", "pass"),
    onvif.WithAuthFallback(onvif.AuthPasswordText, onvif.AuthHTTPBasic, onvif.AuthNone),
)
```

发现——主动、被动或定向：

```go
devices, _ := discovery.Discover(ctx, 5*time.Second)          // 组播探测
usable := discovery.FilterONVIFDevices(devices)               // 滤掉幽灵应答者
discovery.EnrichDevices(ctx, usable)                          // 并行补全身份

listener, _ := discovery.NewListener("", func(d *discovery.Device) { // 被动
    fmt.Println("相机上线:", d.GetDeviceEndpoint())
})
go listener.Start(ctx)
defer listener.Stop()

dev := discovery.ProbeEndpoint(ctx, "192.168.2.50", 80, 1200*time.Millisecond) // 跨子网
serial, ok := discovery.ProbeSerial(ctx, "192.168.2.50", nil)                  // 身份锚点
```

事件——托管、自动续订：

```go
sub, err := client.Events().SubscribeEvents(ctx, func(msg onvif.NotificationMessage) {
    fmt.Println(msg.Topic, msg.Message.Data)
}, nil)
if errors.Is(err, onvif.ErrEventsNotSupported) {
    // 设备没有事件服务；缓存阴性结果
}
defer sub.Unsubscribe(context.Background())
```

鉴权怎么都不通时，做个分诊：

```go
diag, _ := client.DiagnoseAuth(ctx)
// diag.Status: "ok" | "clock-skew"（修相机 NTP）| "bad-credentials"
```

## 文档

主题指南，中英双语：

| 主题 | 中文 | English |
|---|---|---|
| 架构 | [架构](docs/zh/architecture.md) | [architecture.md](docs/en/architecture.md) |
| 鉴权与安全 | [鉴权与安全](docs/zh/authentication.md) | [authentication.md](docs/en/authentication.md) |
| 设备发现 | [设备发现](docs/zh/discovery.md) | [discovery.md](docs/en/discovery.md) |
| 媒体与码流 | [媒体与码流](docs/zh/media.md) | [media.md](docs/en/media.md) |
| 事件 | [事件](docs/zh/events.md) | [events.md](docs/en/events.md) |
| 服务端 | [服务端](docs/zh/server.md) | [server.md](docs/en/server.md) |
| 并发模型 | [并发模型](docs/zh/concurrency.md) | [concurrency.md](docs/en/concurrency.md) |
| 测试 | [测试](docs/zh/testing.md) | [testing.md](docs/en/testing.md) |
| CLI 工具 | [CLI 工具](docs/zh/cli.md) | [cli.md](docs/en/cli.md) |

API 参考：[pkg.go.dev/github.com/mickeyzzc/onvif-go](https://pkg.go.dev/github.com/mickeyzzc/onvif-go/v2)。

## 项目结构

| 路径 | 用途 |
|---|---|
| `client.go` 等（根目录） | `Client`（鉴权梯队、时钟偏差、下载）+ v1 兼容 alias |
| `device/` `security/` `deviceio/` `media/` `ptz/` `imaging/` `events/` | 域服务包（v2 拆分，issue #20） |
| `types/` | 共享数据模型 leaf |
| `discovery/` | WS-Discovery：主动探测、被动监听、定向 HTTP 探测、后处理 |
| `internal/soap/` | SOAP 传输 + WS-Security（digest/明文模式、Fault 检测） |
| `server/` | 虚拟 ONVIF 相机服务器（测试用模拟器） |
| `testing/` | 测试助手：mock server、抓包回放、golden 文件 |
| `testdata/captures/` | 真机 SOAP 抓包回归 fixture |
| `docs/{en,zh}/` | 主题文档（架构、鉴权、发现、媒体、事件、并发、测试、CLI） |
| `cmd/` | 辅助 CLI：`discover`、`onvif-quick`、`onvif-diagnostics`、`onvif-server` |
| `examples/` | 按功能划分的可运行示例 |

## 开发

```bash
make build    # go build ./...
make test     # go test -race ./...
make lint     # golangci-lint run（安装：make lint-install）
make fmt      # 通过 golangci-lint fmt 执行 gofumpt + goimports
```

CI（[ci.yml](.github/workflows/ci.yml)）在每次推送到 `main` 时运行 lint +
格式检查 + race 测试 + 构建；分支要求三个任务全绿，且只接受 PR 合并。

## 血统与许可

本项目是 [0x524a/onvif-go](https://github.com/0x524a/onvif-go) 的持续维护版
（其本身源自更早的 ONVIF Go 尝试——见 git 历史）。模块路径在 **v1.2.0** 起变更为
`github.com/mickeyzzc/onvif-go`；此前以 `github.com/0x524a/onvif-go` 发布
（tag v1.0.0–v1.1.7 保留在本仓库作历史存档，永不删除）。v1.2.0 之后的延续性
变更记录在[变更日志](CHANGELOG.md)。

基于 [MIT 许可证](LICENSE) 发布。衷心感谢原作者的贡献，它们仍受同一许可保护。
