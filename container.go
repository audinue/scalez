package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
)

type Container struct {
	id     string
	client http.Client
}

func NewContainer(id string) (*Container, error) {
	if id == "" {
		return nil, errors.New("missing id")
	}
	client := http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", "/var/run/docker.sock")
			},
		},
	}
	return &Container{id, client}, nil
}

func (c *Container) Id() string {
	return c.id
}

func (c *Container) run(action string) error {
	url := fmt.Sprintf("http://localhost/containers/%s/%s", c.id, action)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}
	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	res.Body.Close()
	if res.StatusCode != http.StatusNoContent && res.StatusCode != http.StatusNotModified {
		return fmt.Errorf("action failed %s %s: %s", action, c.id, res.Status)
	}
	return nil
}

func (c *Container) Start() error {
	return c.run("start")
}

func (c *Container) Stop() error {
	return c.run("stop")
}
