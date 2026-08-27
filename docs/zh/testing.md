# 测试

测试套件默认完全离线：所有行为测试都跑在 `httptest` mock 设备或原始 XML
fixture 上。真机集成测试存在，但由环境变量门控，从不在 CI 运行。

## 运行

```bash
make test        # go test -race ./...
make lint        # golangci-lint run（要求零 findings）
make fmt         # 通过 golangci-lint fmt 执行 gofumpt + goimports
make check       # lint + test
```

CI（`.github/workflows/ci.yml`）在每次推送到 `main` 时运行 lint、格式检查、
race 测试与构建；三个任务都是必需状态检查。

少数发现测试需要主机能加入 WS-Discovery 组播组。没有可用组播路由的主机
（开着 VPN 的笔记本、部分 CI runner）会自动 skip——套件保持绿色。

## 测试分层

| 层 | 位置 | 覆盖 |
|---|---|---|
| mock 设备测试 | 代码旁的 `*_test.go` | 对 `httptest` SOAP 设备的完整 client 行为：鉴权梯队迁移、Fault 处理、响应形态变体、托管订阅生命周期 |
| 原始 fixture | `testdata/captures/*.xml` | 把真实形态的信封回放进 client（如 issue #3 背后的 GetStreamUri 命名空间变体） |
| 解析器单测 | 如 `TestParseScopes`、`TestLooseExtractURI`、`TestSelectMainProfile` | 纯函数、无 I/O |
| 并发矩阵 | `concurrency_test.go` | 单个共享 client 上的混合操作 + 配置变更，`-race` 下有实际意义 |
| 抓包回放助手 | `testing/` 包 | mock server、抓包注册表、golden 文件——大型套件在用 |

## 真机集成测试

部分测试文件（`device_real_camera_test.go`、`media_real_camera_test.go`）
在提供连接信息时（且仅在当时）对真实硬件运行：

```bash
export ONVIF_TEST_ENDPOINT="http://192.168.x.x/onvif/device_service"
export ONVIF_TEST_USERNAME="..."
export ONVIF_TEST_PASSWORD="..."
go test -v -run RealCamera ./...
```

文档和 issue 里一律使用占位符凭据——绝不放真实凭据。设备行为异常时，抓下
它的原始 SOAP 响应，作为 fixture 加进 `testdata/captures/`，让回归在相机
离场后依然存活。`cmd/onvif-diagnostics`（见 [CLI 工具](cli.md)）就是为产出
这些抓包而生的。

## 新测试的约定

- 优先用 `httptest` server 而非手写 mock transport；按请求体内容分类请求，
  不按调用顺序——传输细节变化时测试仍然有效。
- 对设备会「看请求内容行事」的部分断言**报文形态**（如 `GetStreamUri`
  请求的 StreamSetup 字段）。
- 时间敏感的循环（托管订阅、发现监听器）应暴露确定性退出信号（`Done()`
  channel），不依赖 sleep。
