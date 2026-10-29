# 真机兼容矩阵（逐型号证据）

README 的兼容性主张来自真实部署。本页把每一条主张落到**可核验的
型号级证据**上：固件版本、验证过的操作面、实际工作的认证模式、怪癖
与对应处理（链到代码 / 文档 / fixture / issue）。检索长尾（"DS-2CD2143G2-I
onvif golang" 这类查询）会直接命中本页。

**口径**：只登记有第一手证据的行——诊断报告、issue 里的真机字段向量、
或生产部署日志；"应该兼容"不登记。最后一列 = 证据采集时点与库版本。

| 厂商 | 型号 / 固件 | 验证过的操作面 | 实际工作的认证 | 怪癖 → 处理 | 最后验证 |
|---|---|---|---|---|---|
| Hikvision | DS-2CD2143G2-I @ 5.7.1 | 诊断 11 操作全通过（设备信息/能力/时间/Profile/编码器/流地址/快照/PTZ） | WSSE digest | 无——行为基准机；诊断示例产物即来自它（[diagnostics 手册](onvif-go-cli-diagnostics.md)示例命名） | 2026-09-20 · v2.2 期 |
| Hikvision | DS-IPC-T12HV3-IA | MotionAlarm 事件字段 | — | 事件字段向量进 fixture（[#126](https://github.com/mickeyzzc/onvif-go/issues/126)） | 2026-10 · v2.5 期 |
| Hikvision | （型号待补） | 摘要认证 | WSSE digest | "sender not authorized" 实为**时钟偏移**——摘要时间戳超窗被误报为授权失败；`WithAutoClockSkew` 自动校时（[authentication](onvif-go-authentication.md)） | 2026-09 · v2.1 期 |
| TP-Link | TL-IPC44AW-COLOR | Motion 事件（CellMotionDetector） | — | `IsMotion` 方言：该家族用 CellMotionDetector 而非 VideoSource/MotionAlarm（#126 字段向量） | 2026-10 · v2.5 期 |
| Amcrest | IP4M 系列 | Profile/流地址/子码流选择 | WSSE digest | 同分辨率**双 token** 的子码流选择启发式——按分辨率区分主流/子流（[media](onvif-go-media.md)） | 2026-10 · v2.3 期 |
| iDRPENG（中维世纪方案） | PSH_H_V1.0 @ V2.2.1144，ONVIF 2.4 | 连接/设备信息/Profile/流地址 | digest（401 挑战后） | **Digest → 200 + 空体**：旧栈对合法应答回空 SOAP 体，认证梯子的空体修复即为此机（[#140](https://github.com/mickeyzzc/onvif-go/issues/140)/#141） | 2026-10-14 · v2.6.0 |
| Dahua | （型号待补） | 目录/能力/Profile | WSSE digest | 严格信封命名空间解析——不声明默认命名空间的信封曾被跳过（v2.1 线格式修复） | 2026-09 · v2.1 期 |
| Uniview / TP-Link / 自研固件 | — | wsnt:Subscribe 推送订阅 + Notify 接收 | — | push 订阅信封真机验证（v2.4 批次） | 2026-10 · v2.4 期 |
| 海思 OEM / ESP32 / "以 SOAP Fault 回答一切"的怪固件 | — | 连接梯子存活 | 梯子各级 | Fault-only 固件走 `DiagnoseAuth` 定位到可用档位（[authentication](onvif-go-authentication.md)） | 持续 · 生产 |

## 怪癖 → 处理的对应索引

- **时钟偏移误报授权失败** → `onvif.WithAutoClockSkew`（客户端自动校时）
- **Digest 200 空体** → 认证梯子空体容错（#141，v2.6.0）
- **同分辨率双 token** → `SelectSubProfile` 分辨率启发式（media 包）
- **CellMotionDetector 方言** → 事件解析双主题兼容（#126）
- **严格命名空间信封** → 解码器逐前缀绑定（v2.1 修复）
- **Fault-only 固件** → `DiagnoseAuth`（认证诊断）

## 维护规则

1. 每次真机联调后**顺手补行**：型号/固件/操作面/认证档/怪癖，附诊断
   报告文件名或 issue 号；
2. 行内链接必须指向仓内位置（docs 页 / fixture / issue）——外部链接
   会腐烂；
3. 型号待补的行**保留**（如实呈现证据缺口），补齐后移除待注记。

对应中文检索词：海康 onvif golang、大华 onvif go、TP-Link ONVIF 事件、
Amcrest 子码流、中维世纪 ONVIF。
