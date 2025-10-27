package main

import (
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/audinue/scalez/namedargs"
)

func main() {
	log.SetFlags(log.Ltime | log.Llongfile)
	args, err := namedargs.ParseArgs()
	if err != nil {
		log.Fatal(err)
	}
	target, ok := args["target"]
	if !ok {
		log.Fatal("missing target")
	}
	host, port, err := net.SplitHostPort(target)
	if err != nil {
		log.Fatal(err)
	}
	if host == "" {
		log.Fatal("missing host")
	}
	listenAddr := fmt.Sprintf(":%s", port)
	dialAddr := fmt.Sprintf("%s:%s", host, port)
	timeout, err := time.ParseDuration(args["timeout"])
	if err != nil {
		log.Fatal(err)
	}
	container, err := NewContainer(host)
	if err != nil {
		log.Fatal(err)
	}
	mutex := sync.Mutex{}
	nilTime := time.Now()
	started := atomic.Bool{}
	last := atomic.Value{}
	last.Store(nilTime)
	proxy := Proxy{
		ListenAddr: listenAddr,
		DialAddr:   dialAddr,
		OnServe: func() error {
			if !started.Load() {
				mutex.Lock()
				if started.Load() {
					return nil
				}
				defer mutex.Unlock()
				err := container.Start()
				if err != nil {
					return err
				}
				err = waitHost(dialAddr, time.Minute)
				if err != nil {
					return err
				}
				started.Store(true)
			}
			return nil
		},
		OnMessage: func() {
			if timeout != 0 {
				last.Store(time.Now())
			}
		},
	}
	err = proxy.Start()
	if err != nil {
		log.Fatal(err)
	}
	for {
		if timeout != 0 && started.Load() {
			l := last.Load().(time.Time)
			if l != nilTime && time.Since(l) > timeout {
				mutex.Lock()
				err := container.Stop()
				started.Store(false)
				last.Store(nilTime)
				mutex.Unlock()
				if err != nil {
					log.Println(err)
					continue
				}
			}
		}
		time.Sleep(time.Second)
	}
}
