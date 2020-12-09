package grpc_tool

import (
	"errors"

	"github.com/go-kit/kit/sd/consul"
	"github.com/hashicorp/consul/api"
	"github.com/liyanbing/my-gokit/utils"
)

type ConsulServer struct {
	Address    string `json:"address"`
	Scheme     string `json:"scheme"`
	DataCenter string `json:"data_center"`
	WaitTime   int64  `json:"wait_time"`
	Token      string `json:"token"`
}

type ConsulRegistration struct {
	ID                string                 `json:"id,omitempty"`
	Name              string                 `json:"name,omitempty"`
	Tags              []string               `json:"tags,omitempty"`
	Port              int                    `json:"port,omitempty"`
	Address           string                 `json:"address,omitempty"`
	EnableTagOverride bool                   `json:"enable_tag_override,omitempty"`
	Check             *api.AgentServiceCheck `json:"-"`
	Checks            api.AgentServiceChecks `json:"-"`
}

type ConsulConf struct {
	Server       ConsulServer       `json:"server"`
	Registration ConsulRegistration `json:"registration"`
}

type Deregister func() error

func RegisterServiceWithConsul(conf *ConsulConf) (Deregister, error) {
	if conf == nil {
		return nil, errors.New("empty consul config")
	}

	if conf.Registration.ID == "" {
		return nil, errors.New("empty consul registration id")
	}

	if conf.Server.Address == "" {
		return nil, errors.New("empty consul server address")
	}

	if conf.Registration.Address == "" {
		conf.Registration.Address = utils.GetLANHost()
	}

	consulClient, err := api.NewClient(&api.Config{
		Address:    conf.Server.Address,
		Datacenter: conf.Server.DataCenter,
	})
	if err != nil {
		return nil, err
	}

	sdClient := consul.NewClient(consulClient)
	if sdClient == nil {
		return nil, errors.New("create consulsd client failed")
	}

	sdRegConfig := &api.AgentServiceRegistration{
		ID:      conf.Registration.ID,
		Name:    conf.Registration.Name,
		Tags:    conf.Registration.Tags,
		Port:    conf.Registration.Port,
		Address: conf.Registration.Address,
		Check:   conf.Registration.Check,
		Checks:  conf.Registration.Checks,
	}

	err = sdClient.Register(sdRegConfig)
	if err != nil {
		return nil, err
	}

	return func() error {
		return sdClient.Deregister(sdRegConfig)
	}, nil
}
