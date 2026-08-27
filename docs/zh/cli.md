# CLI 工具

`cmd/` 目录附带辅助工具。为你的平台构建：

```bash
make build          # go build ./...（最快的语法/编译检查）
make cross          # CGO_ENABLED=0 构建 linux/arm64 到 build/
go build -o bin/ ./cmd/...
```

所有工具都是零依赖单文件二进制；没有发布流水线——从源码构建。

## discover

带接口选择的组播 WS-Discovery 探测（多网卡主机有用）：

```bash
go run ./cmd/discover -timeout 5s
go run ./cmd/discover -interface eth0
```

## onvif-quick

对一台相机的一次性摘要：设备信息、profile、流 URI。用来回答「这个库到底
能不能和我的相机说上话」。

```bash
go run ./cmd/onvif-quick
```

## onvif-diagnostics

针对单台相机的深度诊断收集器——提 issue 前该跑的工具。跑遍所有主要操作、
逐项打印结果，还能抓取原始 SOAP 交换用作回归 fixture：

```bash
go run ./cmd/onvif-diagnostics \
    -endpoint http://192.168.1.100/onvif/device_service \
    -username admin -password '***' \
    -verbose

# 抓取原始 XML 请求/响应对（附到 issue 前先抹掉凭据；可直接成为
# testdata/captures fixture）：
go run ./cmd/onvif-diagnostics -endpoint ... -username ... -password *** \
    -capture-xml -output diag.json
```

参数：`-endpoint`、`-username`、`-password`、`-timeout`（秒）、`-verbose`、
`-capture-xml`、`-capture-all`、`-output`。

## onvif-server

虚拟 ONVIF 相机模拟器，可独立运行——不用硬件就能测录像软件的发现/接入流程：

```bash
go run ./cmd/onvif-server -port 8000 -manufacturer TestCam -model V1 \
    -serial SIM-0001
```

功能开关：`-info`、`-ptz`、`-imaging`、`-events`、`-version`。

## generate-tests

开发者工具：把抓取的 SOAP 交换（来自 `onvif-diagnostics`）转成 Go 测试脚手架，
并维护抓包注册表。fixture 工作流见[测试](testing.md)。
