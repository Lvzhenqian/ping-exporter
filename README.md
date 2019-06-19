## pinger

一个简单的多地址icmp包发送程序，支持导出到prometheus。
默认监听端口号：`http://ip:7777/metrics`

grafana 目录带有grafana 模板，导入就好。

编译：
`go build -mod=vendor src/main.go`