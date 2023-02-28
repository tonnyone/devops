# influx db start

## download
```shell
https://portal.influxdata.com/downloads/


https://dl.influxdata.com/influxdb/releases/influxdb2-2.6.1-darwin-amd64.tar.gz
tar zxvf influxdb2-2.6.1-darwin-amd64.tar.gz

## telegraf
https://dl.influxdata.com/telegraf/releases/telegraf-1.25.2_darwin_amd64.dmg
```

## get start

```shell
https://docs.influxdata.com/influxdb/v2.6/get-started/setup/
```

```shell
./influxd  admin/admin123!@#
```

### UI界面

### Cli 命令行

#### 写入数据

Line protocol elements
```shell
https://docs.influxdata.com/influxdb/v2.6/get-started/write/?t=influx+CLI
```

#### 查询数据

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



