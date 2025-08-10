# Weave Net：sleeve 模式封包格式与封装责任

本文概述 Weave Net 的两种数据平面：sleeve（用户态 L2-over-UDP）与 fastdp（内核 VXLAN），重点回答两个问题：
- sleeve 模式 overlay 的 UDP 封包是什么格式？是不是“UDP 里装二层网络”？
- 封包/解包由谁负责？是 wRouter 吗？

## sleeve 模式封包格式（L2-over-UDP）

结论：是的，sleeve 是“UDP 里承载完整以太网帧”的 L2 overlay，UDP 负载中包含 Weave 自定义头和内层以太网帧。

封包结构（自外向内）：
- Underlay：Outer Ethernet | Outer IP（v4/v6）| UDP（常见目的端口 6783/udp，用于 sleeve 数据）
- Weave sleeve 头：版本/类型、标志位（如分片）、对端标识/流标识等（用于重组、去重和寻址）
- 可选加密元数据：若启用密码，携带随机数/nonce，负载经对称加密并附鉴别标签
- Inner：完整 L2 以太网帧（Dst MAC | Src MAC | EtherType | Payload），可承载 ARP/IPv4/IPv6/广播/组播

端口与通道（默认约定，实际以部署为准）：
- 控制面：6783/tcp
- sleeve 数据：6783/udp
- fastdp(VXLAN) 数据：6784/udp（Weave 默认；标准 VXLAN 常见 4789/udp）

MTU 与开销：
- 额外头部 ≈ Outer IP(20/40) + UDP(8) + Weave 头(+ 可选加密)，Weave 会自动下调容器侧 MTU 以避免路径分片
- 大包可能按 sleeve 的分片/重组逻辑处理

广播/学习：
- 因为是 L2 overlay，ARP/ND、广播/组播帧会按覆盖网络策略转发；Weave 维护 MAC→peer 映射以优化转发

抓包建议：
- sleeve：udp and port 6783（启用加密时无法直接看到内层 L2 帧）
- fastdp：udp and port 6784（VXLAN 格式可被 Wireshark 解码）

## 封包/解包由谁负责？

### sleeve（用户态数据面）
- 负责者：Weave Router（weaver 进程中的 router 组件，俗称 wRouter）
- 行为：
	- 在本地主机 weave Linux bridge 上获取以太网帧（用户态），封装 Weave 头，按对端 peer 通过 UDP 发送
	- 接收对端 UDP 后解封还原以太网帧，注入到本地 weave bridge
	- 若开启密码，在用户态完成加/解密与完整性校验

### fastdp / VXLAN（内核数据面）
- 负责者：Linux 内核（VXLAN 设备，如 vxlan-6784）执行实际封包/解包
- weaver 职责：通过 netlink 编程内核，维护 FDB（MAC→VTEP）、邻居/对端信息；自身不逐包转发
- 特点：性能更好，资源占用更低

### 其他组件职责澄清
- weave Linux bridge：本地 L2 交换，不做隧道封装
- weave-npc：网络策略（iptables），不参与封包/解包数据面

## sleeve 与 fastdp（VXLAN）的对比要点

- 数据面位置：sleeve 在用户态；fastdp 在内核
- 封装格式：sleeve 为 Weave 自定义头；fastdp 为标准 VXLAN 头（含 VNI 等）
- 端口：sleeve 常 6783/udp；fastdp 常 6784/udp（或标准 4789/udp）
- 兼容/性能：sleeve 兼容性强（无需内核 VXLAN），但 CPU 开销更大；fastdp 依赖内核 VXLAN，性能更优

## 实战排查清单

- 端口连通：确认 6783/udp（sleeve）或 6784/udp（fastdp）在节点间可达
- MTU：检查 Pod MTU 是否被下调；抓包观察是否发生路径分片
- 加密：启用密码后抓包无法看到内层帧，需靠端到端连通性和计数器验证
- FDB/peer：fastdp 下检查 VXLAN FDB；sleeve 下关注 weaver 日志的 peer 学习与重传

## 小结

- sleeve：UDP 中承载完整二层帧，是 L2-over-UDP；封包/解包由 weaver（wRouter）在用户态完成
- fastdp：使用内核 VXLAN，由内核封包/解包；weaver 负责控制与编程
- 选择建议：优先 fastdp（性能），在环境不支持或需最高兼容性时使用 sleeve

