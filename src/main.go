package main

import (
	"ping-exporter/src/config"
	"ping-exporter/src/ping"
	"ping-exporter/src/state"
	"sync"
	"time"
)



func worker(ip string, c int, wg *sync.WaitGroup) {
	wk, err :=ping.NewPinger(ip,c)
	if err != nil {
		panic(err)
	}
	wk.Interval = time.Second * 1
	wk.Timeout = time.Second *	1
	defer wk.Close()
	wk.Run()
	wg.Done()
}


func main() {
	cfg, e := config.Read("./conf/ping.toml")
	if e != nil {
		panic(e)
	}
	var wg sync.WaitGroup
	for _,v := range cfg["ips"]{
		wg.Add(1)
		go worker(v,0,&wg)
	}
	go state.Start()
	wg.Wait()
}