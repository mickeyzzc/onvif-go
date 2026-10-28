# 无硬件 ONVIF 开发闭环：虚拟相机 × CLI × NVR 联调

Go 生态被引用最多的 ONVIF 库
[use-go/onvif](https://github.com/use-go/onvif) 已停更（最后发布
2022-07，v0.0.9）。本教程是它的"当前路径"承接：把本仓库的全部零件——
虚拟相机 server、统一 CLI、诊断报告、客户端 API——串成一条
**端到端闭环，全程不需要任何真实摄像头**。每一步都给出可直接粘贴的
命令与真实输出。

## 0. 一个二进制开始

```bash
# 预编译包（linux/darwin/windows 全平台，零依赖静态二进制）
gh release download v2.7.0 -R mickeyzzc/onvif-go -p 'onvif_linux_amd64.tar.gz'
tar xzf onvif_linux_amd64.tar.gz && sha256sum -c SHA256SUMS

# 或按版本锚定安装 / 检出直跑
go install github.com/mickeyzzc/onvif-go/cmd/onvif@v2.7.0
go run ./cmd/onvif <子命令>
```

## 1. `onvif server`：立起一台虚拟多目相机

```bash
onvif server -password demo -manufacturer TestCam -model VC-1 -serial SIM-0001 -profiles 2
```

一台"多镜头 IP 相机"就跑起来了（服务树挂在
`http://<host>:<port>/onvif/device_service`，`SIGINT` 优雅停机）：

- **身份四件套** `-manufacturer/-model/-firmware/-serial`——
  `GetDeviceInformation` 如实回报，录像机的设备列表显示的就是它们；
- **`-profiles 1..10`**：十套轮换模板（1080p 主流、720p 广角、4K、
  鱼眼、热成像、60fps 车牌……），每套带 H264 编码器与快照配置——
  主/子流选择、多目 UI 布局全有真数据可测；
- **`-ptz`（默认开）**：PTZ 服务回答 `GetStatus`、continuous/absolute
  转动与预置位；**`-events`** 开启 pull-point 事件服务；
- **密码必填**（旗标或 `ONVIF_SERVER_PASSWORD`）——空凭证模式只适合
  嵌入式用法，不适合放在真实网络上的演示二进制。

更多旗标与嵌入自己二进制的方法见
[server 分册](onvif-go-cli-server.md)。

## 2. `onvif discover`：局域网发现它

```bash
onvif discover              # 主动 WS-Discovery Probe（默认全部接口）
onvif discover -interface eth0 -timeout 10s
```

同一网段里刚起的虚拟相机会带着你设置的身份出现在结果里。发现机制
三件套——主动探测、被动 Hello 监听、跨网段定向探测——以及
**ghost responder 过滤**（乱应答的假设备不进结果），见
[discover 分册](onvif-go-cli-discover.md) 与
[发现机制](onvif-go-discovery.md)。

## 3. `onvif quick`：交互体检

```bash
onvif quick
```

菜单驱动：`3` 连接 → `5` 取流地址 → `4` PTZ 演示。**关键提示：
"Camera IP:" 处输入裸 `host:port`**（工具自己拼 `http://` 前缀与
`/onvif/device_service` 后缀；输入完整 URL 会被双重拼接）：

```text
Choice: 3
Camera IP: 127.0.0.1:18081
Username [admin]: admin
Password: demo
✅ Connected!
📡 Stream: rtsp://127.0.0.1:8554/stream0
```

连上之后菜单里还有 PTZ 演示（对着虚拟云台走一圈 continuous/absolute/
preset），交互细节见 [quick 分册](onvif-go-cli-quick.md)。

## 4. `onvif diagnostics`：11 操作 JSON 报告

```bash
onvif diagnostics -endpoint http://127.0.0.1:18081/onvif/device_service \
  -username admin -password demo -output ./camera-logs
```

一次跑完 11 类操作（设备信息、能力、时间、profile 清单、逐 profile 的
视频编码器配置与流地址/快照地址、PTZ 状态与预置位），落盘一份带设备
身份与时间戳命名的 JSON 报告：

```text
✓ Diagnostic collection complete!
  Report saved to: ./camera-logs/TestCam_VC-1_1.0.0_20261029-064143.json
  Total errors: 2
```

报告即体检病历：`device_info`/`profiles` 看身份与能力面，
`video_encoders[i].error` 这类错误条目自带**哪个操作、哪个 profile、
原始 SOAP 信封**——出问题先看它。诊断分册
[cli-diagnostics](onvif-go-cli-diagnostics.md) 有 11 操作的逐节字段
解读与 `-capture-xml` 原始流量存档用法。

**仓规：在这份报告上提 issue。** 无论对真机还是对虚拟相机，报问题时
先跑 `diagnostics`、把 JSON（可选 `-capture-xml` 存档）附上——
可复现的证据比描述省一轮往返。

## 5. NVR 联调收口

证明"无硬件闭环"成立的最后一步：把虚拟相机喂给一台真录像机。

- 端点：`http://<host>:<port>/onvif/device_service`，凭证
  `admin` / 你设的密码；
- 任一支持 ONVIF 的录像机（NVR/VMS）按"手动添加"填入——发现、
  出图、录像、云台全走标准协议；
- `-profiles 10` 一进程十目，正好测 UI 的主/子流切换与多通道布局。

MiBeeNvr 以此为标准入网流程；其它厂商录像机同样适用——协议面就是
ONVIF Profile S + Media2。

## 6. CLI → 代码：同一件事的 API 形态

CLI 只是同一 client 的壳。换 Go API 做刚才的全部操作：

```go
import "github.com/mickeyzzc/onvif-go/v2"

client, _ := onvif.NewClient("127.0.0.1:18081",
    onvif.WithCredentials("admin", "demo"))
profiles, _ := client.Media().GetProfiles(ctx)
main := onvif.SelectMainProfile(profiles) // 主流 profile token
uri, _ := client.Media().GetStreamURI(ctx, main)
```

客户端全景（认证梯子、Media2/Profile M/imaging、事件订阅）见
[快速开始](onvif-go-quickstart.md)、
[鉴权](onvif-go-authentication.md) 与
[事件](onvif-go-events.md)；把虚拟相机塞进测试管线的姿势见
[testing](onvif-go-testing.md)。
