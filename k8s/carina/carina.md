# Carina CSI 插件学习笔记（存储）

> 本页聚焦 Carina 作为 Kubernetes CSI 存储插件的学习与实践。

## 概述与定位
- 目标：在每个节点上把本地磁盘/分区聚合为存储池，按需动态创建卷供 Pod 使用（偏 RWO、本地高性能）。
- 典型实现：基于 LVM（建议 thin-pool）或直盘；Controller 负责供给控制，Node 负责格式化/挂载与生命周期。
- 适用：需要简单、高性能、低延迟的本地持久卷；不追求跨节点副本（HA 需上层方案）。

## 前置知识与环境准备
- Kubernetes 存储基础：PV/PVC、StorageClass、AccessModes（RWO/ROX/RWX）、VolumeMode（Filesystem/Block）。
- CSI 基础：Controller/Node 插件；sidecar 组件（provisioner、attacher、resizer、snapshotter）。
- LVM 基础：PV/VG/LV、thin vs thick、thin-pool 元数据/数据空间监控；ext4/xfs 基操。
- 节点前提：
  - 安装 lvm2；创建用于 Carina 的 VG（例：carina-vg）。
  - 允许特权容器；挂载 /dev、/sys、/run/udev、/lib/modules（Node 插件需要）。
  - 建议启用 MountPropagation=Bidirectional；确认 SELinux/AppArmor 策略放行。
  - 如需快照/克隆：安装 CSI Snapshot CRDs 与 snapshot-controller。

## 架构与数据路径
- 组件：
  - Controller（Deployment）：csi-controller + external-provisioner/attacher/resizer（+ snapshotter 可选）。
  - Node（DaemonSet）：csi-node，执行 NodeStage/Publish、mkfs、mount/umount、LVM LV 创建/删除。
- 数据路径（Filesystem 卷）：PVC → provisioner 创建 LV → NodeStage 格式化挂载到全局路径 → NodePublish bind-mount 到容器目录。
- 调度：StorageClass 使用 volumeBindingMode: WaitForFirstConsumer，确保卷在目标节点创建。

## 能力与限制（常见）
- AccessModes：ReadWriteOnce（单节点写）。
- VolumeMode：Filesystem 与 Block。
- 动态供给/删除：支持；Retain/Recycle/Delete 由 reclaimPolicy 控制（本地盘常用 Delete）。
- 扩容：支持在线/离线扩容（文件系统需支持，如 xfs_growfs/ext4 resize2fs）。
- 快照/克隆：依赖 CSI Snapshot；底层可用 LVM snapshot/merge 或驱动自实现。
- 不提供：跨节点副本/远程复制（需 Longhorn/Ceph/备份体系等上层方案）。

## 安装与配置要点（示例占位）
- StorageClass（请按实际驱动名与参数调整）：
```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: carina-lvm
provisioner: <driver-name>           # 例：csi.carina.storage.io（以实际为准）
parameters:
  vgName: carina-vg                  # 目标 VG
  type: thin                         # thin 或 thick
  fstype: xfs                        # ext4/xfs
allowVolumeExpansion: true
reclaimPolicy: Delete
volumeBindingMode: WaitForFirstConsumer
```

- PVC（Filesystem）：
```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: demo-pvc
spec:
  accessModes: ["ReadWriteOnce"]
  storageClassName: carina-lvm
  resources:
    requests:
      storage: 20Gi
```

- Pod 示例：
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: demo
spec:
  containers:
  - name: app
    image: busybox
    command: ["sh","-c","sleep 3600"]
    volumeMounts:
    - name: data
      mountPath: /data
  volumes:
  - name: data
    persistentVolumeClaim:
      claimName: demo-pvc
```

## 常见运维与故障排查
- 容量管理：
  - 监控 thin-pool data/metadata 使用率；metadata 爆满会导致只读/失败。
  - 设置告警阈值与自动扩容策略（预留 VG 空间给元数据）。
- 扩容：
  - 扩 PVC → Controller 扩 LV → Node 扩文件系统；确认节点具备 xfs_growfs/resize2fs。
- 调度异常：
  - 无盘节点：使用节点标签与 nodeAffinity（或准入）确保 Pod 调度到有 carina-vg 的节点。
  - 绑定失败：确认 volumeBindingMode=WaitForFirstConsumer 与 provisioner 日志。
- 挂载失败：
  - 检查 Node 插件日志；核对 hostPath 挂载与权限、udev 事件是否到位。

## 监控与可观测
- 指标：
  - VG/ThinPool 容量（data/metadata）；LV 数量与空间；失败率/耗时（provision/attach/mount/resize）。
  - 节点磁盘 IOPS/吞吐/延迟（node-exporter + SMART/blkstats）。
- 日志：Controller/Node 插件与 sidecar（provisioner/attacher/resizer/snapshotter）。

## 与同类的取舍
- vs Longhorn：Longhorn 提供跨节点副本与 UI，网络数据路径开销更高；Carina 本地路径，性能更优但无 HA。
- vs OpenEBS LVM-LocalPV：功能定位相近；OpenEBS 引擎多、生态大；Carina 更聚焦简洁一致。
- vs 静态本地 PV：静态运维成本高；Carina 支持动态供给与扩容，体验更友好。

## 验证清单（最小实验）
- PVC → Pod 跑通；删除回收；扩容 20Gi→40Gi；
- Filesystem 与 Block 各验证一次；
- 节点重启/Pod 重建能正常重挂载；
- fio：4k randread/write 与 128k 顺序，记录 IOPS/吞吐与 P95/P99 延迟。

> 注：请以实际 Carina CSI 发行版的驱动名与参数为准（如 <driver-name>、vgName、type、fstype 等）。确认后可补充官方安装 YAML 与一键实验脚本。
