package ping

import (
	"errors"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"ping-exporter/src/state"
	"net"
	"os"
	"time"
)

var Data = []byte("hello")

type ping struct {
	Addr 		string
	Conn 		net.Conn
	Data 		[]byte
	Timeout 	time.Duration
	Interval 	time.Duration
	Count  		int
	PacketsSent int32
	PacketsRecv	int32
	OnRecv 		func(r Reply)
	OnTimeOut  	func(r Reply)
	seq			int
	lost		int
}

type Reply struct {
	Addr 	string
	Seq		int
	Time  	time.Duration
	TTL   	uint8
	Error 	error
}

type Statistics struct {
	Addr  			string
	SendPackets		int32
	RecvPackets		int32
	LostPercent		float64
}

func NewPinger(addr string,c int) (*ping,error) {
	p := new(ping)
	ipaddr, err := net.ResolveIPAddr("ip", addr)
	if err != nil {
		return nil, err
	}
	wb, err := MarshalMsg(8, Data)
	if err != nil {
		return nil, err
	}
	p = &ping{Data: wb, Addr: ipaddr.String(),Timeout: 5,Count: c}
	return p, nil
}

func MarshalMsg(req int, data []byte) ([]byte, error) {
	xid, xseq := os.Getpid()&0xffff, req
	wm := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID: xid, Seq: xseq,
			Data: data,
		},
	}
	return wm.Marshal(nil)
}

func (p *ping) Close() error {
	return p.Conn.Close()
}

func (p *ping) GetConn() {
	var ConnErr error
	p.Conn, ConnErr = net.Dial("ip4:icmp", p.Addr)
	if ConnErr != nil {
		panic(ConnErr)
	}
	settimouterr := p.Conn.SetDeadline(time.Now().Add(p.Timeout))
	if settimouterr != nil {
		panic(settimouterr)
	}
}

func (p *ping) Run() {
	psend := state.PingerSend.With(prometheus.Labels{"ip":p.Addr})
	precv := state.PingerRecv.With(prometheus.Labels{"ip":p.Addr})
	plost := state.PingerLost.With(prometheus.Labels{"ip":p.Addr})
	c := 0
	for {
		if p.Count > 0 && c >= p.Count {
			return
		}
		p.GetConn()
		psend.Inc()
		r := p.sendPingMsg()
		if r.Error != nil {
			if opt, ok := r.Error.(*net.OpError); ok && opt.Timeout() {
				r.Addr = p.Addr
				r.Seq = p.seq
				if p.OnTimeOut != nil {
					p.OnTimeOut(r)
				}
				p.lost++
				plost.Inc()
			}
		} else {
			if p.OnRecv != nil{
				p.OnRecv(r)
			}
			p.PacketsRecv++
			precv.Inc()
		}
		time.Sleep(p.Interval)
		c++
	}
}

func (p *ping) sendPingMsg() (reply Reply) {
	start := time.Now()
	p.seq++
	if _, reply.Error = p.Conn.Write(p.Data); reply.Error != nil {
		return
	}
	p.PacketsSent++
	rb := make([]byte, 500)
	var n int
	n, reply.Error = p.Conn.Read(rb)
	if reply.Error != nil {
		return
	}

	duration := time.Since(start)
	ptime := state.PingerTime.With(prometheus.Labels{"ip":p.Addr})
	ptime.Observe(float64(duration) / float64(time.Millisecond))
	ttl := uint8(rb[8])
	rb = func(b []byte) []byte {
		if len(b) < 20 {
			return b
		}
		hdrlen := int(b[0]&0x0f) << 2
		return b[hdrlen:]
	}(rb)
	var rm *icmp.Message
	rm, reply.Error = icmp.ParseMessage(1, rb[:n])
	if reply.Error != nil {
		return
	}

	switch rm.Type {
	case ipv4.ICMPTypeEchoReply:
		reply = Reply{p.Addr,p.seq,duration, ttl, nil}
	case ipv4.ICMPTypeDestinationUnreachable:
		reply.Error = errors.New("Destination Unreachable")
	default:
		reply.Error = fmt.Errorf("Not ICMPTypeEchoReply %v", rm)
	}
	return
}

func (p *ping) Getstatistics() Statistics {
	loss := float64(p.lost) / float64(p.PacketsSent) * 100
	return Statistics{p.Addr,p.PacketsSent,p.PacketsRecv,loss}
}
