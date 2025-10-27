package main

import (
	"errors"
	"io"
	"log"
	"net"
)

type Proxy struct {
	ListenAddr string
	DialAddr   string
	OnServe    func() error
	OnMessage  func()
}

func (p *Proxy) copy(source net.Conn, target net.Conn) {
	buffer := make([]byte, 32*1024)
	for {
		nr, er := source.Read(buffer)
		if nr > 0 {
			nw, ew := target.Write(buffer[:nr])
			if ew != nil {
				if !errors.Is(ew, net.ErrClosed) {
					log.Println(ew)
				}
				break
			}
			if nw != nr {
				log.Println(io.ErrShortWrite)
				break
			}
			if p.OnMessage != nil {
				p.OnMessage()
			}
		}
		if er != nil {
			if !errors.Is(er, io.EOF) && !errors.Is(er, net.ErrClosed) {
				log.Println(er)
			}
			break
		}
	}
	source.Close()
	target.Close()
}

func (p *Proxy) serve(client net.Conn) {
	if p.OnServe != nil {
		err := p.OnServe()
		if err != nil {
			client.Close()
			log.Println(err)
			return
		}
	}
	server, err := net.Dial("tcp", p.DialAddr)
	if err != nil {
		client.Close()
		log.Println(err)
		return
	}
	go p.copy(client, server)
	go p.copy(server, client)
}

func (p *Proxy) accept(listener net.Listener) {
	for {
		client, err := listener.Accept()
		if err != nil {
			log.Println(err)
			break
		}
		go p.serve(client)
	}
}

func (p *Proxy) listen() error {
	listener, err := net.Listen("tcp", p.ListenAddr)
	if err != nil {
		return err
	}
	log.Println(listener.Addr())
	go p.accept(listener)
	return nil
}

func (p *Proxy) Start() error {
	if p.ListenAddr == "" {
		return errors.New("missing ListenAddr")
	}
	if p.DialAddr == "" {
		return errors.New("missing DialAddr")
	}
	return p.listen()
}
