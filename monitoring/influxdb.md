# influx db start

## InfluxDB 2 vs 3 对比

| 对比项 | InfluxDB 2 | InfluxDB 3 |
|---|---|---|
| 存储引擎 | TSM + WAL，TSI 索引可选（行式/LSM 风格） | IOx（Rust）基于 Apache Arrow/Parquet（列式），面向高基数与长期留存 |
| 查询语言 | Flux（主推），InfluxQL 通过兼容端点 | SQL 为主（FlightSQL/Arrow/DataFusion），不再支持 Flux；提供一定 InfluxQL 兼容 |
| 写入协议 | Line Protocol（/api/v2/write） | Line Protocol 继续支持；查询走 SQL/FlightSQL |
| 架构/存储介质 | 本地盘为主，OSS 单机；集群能力在企业/云 | 设计适配对象存储（S3 等）与本地，冷热分层，更易横向扩展（完整分布式多在云/商用版） |
| 高基数数据 | 可用但索引/内存压力较大 | 核心目标之一，列存 + 压缩 + 向量化扫描更友好 |
| 功能特性 | 内置 UI、Dashboards、Tasks、Notebooks | 更聚焦引擎/SQL 与对象存储；任务/告警更偏外部工具或云侧能力 |
| 生态集成 | Flux 生态、InfluxQL、Telegraf | SQL/Arrow 生态、FlightSQL、DataFrames，Telegraf 仍可用 |
| 伸缩与 HA | OSS 单机；HA/集群需企业或云 | 面向水平扩展的设计；本地版功能有限，云/商用版提供完整扩展性 |
| 适用场景 | 低延迟写入、短保留、本地部署、Flux 分析 | 高基数、长期留存、对象存储、SQL 分析/数据仓库式工作负载 |
| 开源状态 | OSS 单机开源；企业/云为商业 | IOx/核心引擎开源；社区版开源可用，云/商用版提供更多分布式/运维特性 |

选择建议：
- 需要 Flux/内置任务与看板/本地单机：选 InfluxDB 2。
- 需要 SQL、Arrow/Parquet、对象存储、高基数与长留存：选 InfluxDB 3。




## InfluxData 组件速览

| 组件 | 类型 | 一句话作用 | 主要场景/备注 |
|---|---|---|---|
| Telegraf | 采集器/Agent | 插件化收集系统/中间件等指标与日志，写入 InfluxDB（行协议）或其他后端 | 支持大量输入/处理/输出插件；适配 1.x/2.x/3 |
| InfluxDB CLI | 命令行工具 | 管理组织/桶/Token，写入数据，查询（Flux/InfluxQL 视版本） | 2.x 原生；3.x 以 SQL/FlightSQL 客户端为主 |
| InfluxDB OSS | 数据库（开源） | 自托管的开源单机 InfluxDB | 1.x/2.x 明确；3.x 引擎开源，完整分布式多在云/商用 |
| Chronograf | 可视化/管理 UI | 仪表盘、探索数据、与 Kapacitor 集成配置 | 主要用于 1.x；2.x 内置 UI 基本取代之 |
| Kapacitor | 流处理/告警 | 订阅 InfluxDB 流数据，执行 TICKscript 做 ETL/告警/降采样 | 1.x 常用；2.x 用 Tasks/通知；3.x 多用外部任务/云能力 |

## install influxdb
```
sudo apt update
sudo apt install gnupg

curl -fsSL https://repos.influxdata.com/influxdata-archive_compat.key \
  | gpg --dearmor \
  | sudo tee /etc/apt/trusted.gpg.d/influxdata.gpg > /dev/null

wget -q https://repos.influxdata.com/influxdata-archive.key
gpg --show-keys --with-fingerprint --with-colons ./influxdata-archive.key 2>&1 | grep -q '^fpr:\+24C975CBA61A024EE1B631787C3D57159FC2F927:$' && cat influxdata-archive.key | gpg --dearmor | sudo tee /etc/apt/trusted.gpg.d/influxdata-archive.gpg > /dev/null
echo 'deb [signed-by=/etc/apt/trusted.gpg.d/influxdata-archive.gpg] https://repos.influxdata.com/debian stable main' | sudo tee /etc/apt/sources.list.d/influxdata.list

sudo apt-get update && sudo apt-get install influxdb2
```

### cli install
```
wget -q https://repos.influxdata.com/influxdata-archive.key
gpg --show-keys --with-fingerprint --with-colons ./influxdata-archive.key 2>&1 | grep -q '^fpr:\+24C975CBA61A024EE1B631787C3D57159FC2F927:$' && cat influxdata-archive.key | gpg --dearmor | sudo tee /etc/apt/trusted.gpg.d/influxdata-archive.gpg > /dev/null
echo 'deb [signed-by=/etc/apt/trusted.gpg.d/influxdata-archive.gpg] https://repos.influxdata.com/debian stable main' | sudo tee /etc/apt/sources.list.d/influxdata.list

sudo apt-get update && sudo apt-get install influxdb2-cli
```

## start

```
influxd --http-bind-address=:8086
```

## 查询数据

UI 访问地址

```
http://localhost:8086
```


command
```
influx setup

QJhVYwdZPDwITK9C1mVSaRZq1SsWSqdJLOq9ZexkTA-EI7C0w7Hrt_5okbbnBZOu6Cghpz4BbfEuPJQZin9rVA=="
```

```shell
Flux: A functional scripting language designed to query and process data from InfluxDB and other data sources.
InfluxQL: A SQL-like query language designed to query time series data from InfluxDB.
```

##### influxQl

```shell
export INFLUX_HOST='http://127.0.0.1:8086'
export INFLUX_ORG='admin'
export INFLUX_TOKEN='aONaRPaDfqq5snMLA

./influx v1 shell
```


```shell
SELECT co,hum,temp,room FROM "get-started".autogen.home WHERE time >= '2022-01-01T08:00:00Z' AND time <= '2022-01-01T20:00:00Z'
```

#### Processing data

```shell

```




