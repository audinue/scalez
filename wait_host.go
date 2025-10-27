package main

import (
	"errors"
	"net"
	"time"
)

func waitHost(dialAddr string, timeout time.Duration) error {
	start := time.Now()
	for time.Since(start) < timeout {
		conn, err := net.DialTimeout("tcp", dialAddr, 500*time.Millisecond)
		if err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(time.Second)
	}
	return errors.New("timeout")
}
