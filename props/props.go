package props

import (
	"errors"
	"flag"
	"fmt"
	"regexp"
	"strings"
)

var (
	ConfFilePath     string
	ConsulAddress    string
	ConsulSchema     string
	ConsulDataCenter string
	ConsulWaitTime   int64
	ConsulToken      string
	ConfNode         string
)

func InitFlag() {
	flag.StringVar(&ConfFilePath, "config-path", "", "config file path")
	flag.StringVar(&ConsulAddress, "consul-addr", "", "consul address")
	flag.StringVar(&ConsulSchema, "consul-schema", "", "consul schema")
	flag.StringVar(&ConsulDataCenter, "consul-data-center", "", "consul data center")
	flag.Int64Var(&ConsulWaitTime, "consul-wait-time", 0, "consul wait time")
	flag.StringVar(&ConsulToken, "consul-token", "", "consul token")
	flag.StringVar(&ConfNode, "config-node", "", "config node in consul")
}

func LoadConfig(obj interface{}) error {
	if !flag.Parsed() {
		flag.Parse()
	}

	if ConfFilePath != "" {
		return JSONConfigFromFile(ConfFilePath, obj)
	}

	if ConsulAddress == "" || ConfNode == "" {
		return errors.New("consul address and config node can not be empty")
	}

	return JSONConfigFromConsul(Consul{
		Address:    ConsulAddress,
		Scheme:     ConsulSchema,
		DataCenter: ConsulDataCenter,
		WaitTime:   ConsulWaitTime,
		Token:      ConsulToken,
	}, ConfNode, obj)
}

var placeHolder = regexp.MustCompile(`\[\[\.(.*)+?\]\]`)

type placeHolderReplaceFunc func([]string) map[string]string

func ReplacePlaceHolder(content string, replacerFn placeHolderReplaceFunc) string {
	keys := placeHolder.FindAllStringSubmatch(content, -1)

	cache := make(map[string]struct{})
	oldKeys := make([]string, 0, len(keys))
	for _, v := range keys {
		if len(v) < 2 {
			continue
		}

		if _, ok := cache[v[1]]; ok {
			continue
		}

		oldKeys = append(oldKeys, v[1])
		cache[v[1]] = struct{}{}
	}

	newKeys := replacerFn(oldKeys)
	oldNewParis := make([]string, 0, 2*len(oldKeys))
	for _, old := range oldKeys {
		oldNewParis = append(oldNewParis, fmt.Sprintf("[[.%v]]", old), newKeys[old])
	}

	replacer := strings.NewReplacer(oldNewParis...)
	return replacer.Replace(content)
}
