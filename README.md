IP Service
====
基于 https://github.com/teambition/gear  Go Web 框架和 http://www.ipip.net/ IP 数据库实现的 IP 查询服务。

## 运行

### 从源码运行

```bash
go get github.com/xusss/ipservice
cd path_to_ipservice

# set env
export CONFL_VAULT_TOKEN=06900225-b34b-69de-7872-21a2c8b52306
export VAULT_ADDR='http://127.0.0.1:8200'
export CONFL_CONF_PATH=/confl/test2
export CONFL_VAULT_ADDRESS=http://localhost:8200
export CONFL_VAULT_AUTH_TYPE=token
export CONFL_ETCD_CLUSTERS=http://localhost:2379

go run app.go
```

### Docker (15.01 MB)

Build docker image with https://github.com/hesion3d/slimage:
```sh
cp docker.sh path-to-slimage/ipservice.sh
cd path-to-slimage
./run.sh -f ipservice.sh -l extra -n ipservice
```

Please edit docker.sh yourself.

Run image:
```sh
docker images
docker run --rm -p 8080:8080 ipservice
curl 127.0.0.1:8080/json/8.8.8.8
```

## API

### GET /json/:ip

```bash
curl 127.0.0.1:8080/json/8.8.8.8
# 返回 JSON 数据
{"IP":"8.8.8.8","Status":200,"Message":"","Data":{"Country":"GOOGLE","Region":"GOOGLE","City":"N/A","Isp":"N/A"}}
```

### GET /json/:ip?callback=xxx

```bash
# callback=xxxx 返回 JSONP 数据
curl 127.0.0.1:8080/json/8.8.8.8?callback=readIP
# 返回 JSONP 数据
/**/ typeof readIP === "function" && readIP({"IP":"8.8.8.8","Status":200,"Message":"","Data":{"Country":"GOOGLE","Region":"GOOGLE","City":"N/A","Isp":"N/A"}});
```

## Bench

Environment: MacBook Pro, 2.4 GHz Intel Core i5, 8 GB 1600 MHz DDR3

Start service: `./ipservice --data=./data/17monipdb.dat`

Result: **41132.68 req/sec**
```bash
wrk 'http://localhost:8080/json/8.8.8.8' -d 60 -c 100 -t 4
Running 1m test @ http://localhost:8080/json/8.8.8.8
  4 threads and 100 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     3.51ms    6.45ms 199.21ms   95.37%
    Req/Sec    10.34k     1.14k   14.33k    86.12%
  2470564 requests in 1.00m, 558.40MB read
Requests/sec:  41132.68
Transfer/sec:      9.30MB
```
