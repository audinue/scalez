package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/moby/moby/client"
)

func getIdByHostname(apiClient *client.Client, hostname string) (string, error) {
	defer apiClient.Close()
	containers, err := apiClient.ContainerList(context.Background(), client.ContainerListOptions{All: true})
	if err != nil {
		return "", err
	}
	wg := sync.WaitGroup{}
	mutex := sync.Mutex{}
	ids := make(map[string]string)
	for _, container := range containers {
		wg.Go(func() {
			inspect, err := apiClient.ContainerInspect(context.Background(), container.ID)
			if err != nil {
				log.Println(err)
			}
			mutex.Lock()
			defer mutex.Unlock()
			for _, value := range inspect.NetworkSettings.Networks {
				for _, alias := range value.Aliases {
					ids[alias] = container.ID
				}
			}
		})
	}
	wg.Wait()
	id, ok := ids[hostname]
	if !ok {
		return "", fmt.Errorf("container not found: %s", hostname)
	}
	return id, nil
}

type Container struct {
	id        string
	apiClient *client.Client
}

func NewContainer(hostname string) (*Container, error) {
	if hostname == "" {
		return nil, errors.New("missing hostname")
	}
	apiClient, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithVersion("1.51"),
	)
	if err != nil {
		return nil, err
	}
	id, err := getIdByHostname(apiClient, hostname)
	if err != nil {
		return nil, err
	}
	return &Container{id, apiClient}, nil
}

func (c *Container) Start() error {
	return c.apiClient.ContainerStart(context.Background(), c.id, client.ContainerStartOptions{})
}

func (c *Container) Stop() error {
	return c.apiClient.ContainerStop(context.Background(), c.id, client.ContainerStopOptions{})
}
