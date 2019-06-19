package state

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

var (
	PingerSend = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "pinger",
		Subsystem: "Packets",
		Name: "Send",
		Help: "total send packets to remote address",
	},
	[]string{"ip"})

	PingerRecv = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "pinger",
		Subsystem: "Packets",
		Name: "Recv",
		Help: "total recv packets from remote address",
	},
		[]string{"ip"})
	PingerLost = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "pinger",
		Subsystem: "Packets",
		Name: "Lost",
		Help: "total lost packets from remote address",
	},
		[]string{"ip"})
	PingerTime = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "pinger",
		Subsystem: "Packets",
		Name: "Delay",
		Help: "Delay statistics",
	},
		[]string{"ip"})
)

func init() {
	prometheus.MustRegister(PingerTime)
	prometheus.MustRegister(PingerLost)
	prometheus.MustRegister(PingerRecv)
	prometheus.MustRegister(PingerSend)
}

func Start() error {
	http.Handle("/metrics",promhttp.Handler())
	return http.ListenAndServe(":7777",nil)
}