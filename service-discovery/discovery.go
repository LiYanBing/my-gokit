package service_discovery

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"

	"github.com/go-kit/kit/sd/consul"
	"github.com/golang/glog"
	"github.com/hashicorp/consul/api"
	"sobe-kit/deviceinfo"
)

var (
	serDisConsulConf *string
	sereDisConsul    Consul
)

type Consul struct {
	Address    string `json:"address"`
	Scheme     string `json:"scheme"`
	DataCenter string `json:"data_center"`
	WaitTime   int64  `json:"wait_time"`
	Token      string `json:"token"`
}

func InitConsulFlag() {
	serDisConsulConf = flag.String("sc", "", `config consul: 
	{
		"address": string,
		"scheme": string,
		"data_center": string,
		"wait_time": int64,
		"token": string,
	}`)
}

func LoadConsul() (bool, error) {
	if !flag.Parsed() {
		flag.Parse()
	}

	if *serDisConsulConf == "" {
		return false, nil
	}

	err := json.NewDecoder(bytes.NewBufferString(*serDisConsulConf)).Decode(&sereDisConsul)
	if err != nil {
		return true, err
	}

	return true, nil
}

func RegisterWithConsul(reg *api.AgentServiceRegistration, event <-chan struct{}) error {
	if reg == nil {
		return errors.New("empty reg")
	}

	if reg.ID == "" {
		return errors.New("empty consul registration id")
	}

	if reg.Address == "" {
		reg.Address = deviceinfo.GetLANHost()
	}

	consulConfig, err := api.NewClient(&api.Config{
		Address:    sereDisConsul.Address,
		Datacenter: sereDisConsul.DataCenter,
	})
	if err != nil {
		return err
	}

	consulClient := consul.NewClient(consulConfig)
	if consulClient == nil {
		return errors.New("create consul client failed")
	}

	sdRegConfig := &api.AgentServiceRegistration{
		ID:      reg.ID,
		Name:    reg.Name,
		Tags:    reg.Tags,
		Port:    reg.Port,
		Address: reg.Address,
		Check:   reg.Check,
	}

	err = consulClient.Register(sdRegConfig)
	if err != nil {
		return err
	}

	go func() {
		select {
		case <-event:
			err = consulClient.Deregister(reg)
			if err != nil {
				glog.Errorf("deregister service with err:%v", err)
			}
		}
	}()

	return nil
}
