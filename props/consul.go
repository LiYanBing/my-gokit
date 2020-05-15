package props

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/hashicorp/consul/api"
)

type Consul struct {
	Address    string `json:"address"`
	Scheme     string `json:"scheme"`
	DataCenter string `json:"data_center"`
	WaitTime   int64  `json:"wait_time"`
	Token      string `json:"token"`
}

var EmptyPathError = errors.New("empty config_node")

// get project config from consul
func GetConfigFromConsul(consul Consul, configNode string) ([]byte, error) {
	if configNode == "" {
		return nil, EmptyPathError
	}

	client, err := api.NewClient(&api.Config{
		Address:    consul.Address,
		Datacenter: consul.DataCenter,
	})
	if err != nil {
		return nil, err
	}

	pair, _, err := client.KV().Get(configNode, nil)
	if err != nil {
		return nil, err
	}

	return pair.Value, nil
}

func JSONConfigFromConsul(consul Consul, configNode string, obj interface{}) error {
	conf, err := GetConfigFromConsul(consul, configNode)
	if err != nil {
		return err
	}

	return json.NewDecoder(bytes.NewBuffer(conf)).Decode(obj)
}
