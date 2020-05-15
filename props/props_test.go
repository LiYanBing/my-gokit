package props

import (
	"testing"
)

func TestReplacePlaceHolder(t *testing.T) {
	content, err := GetConfigFromFile("./app.conf")
	if err != nil {
		t.Error(err)
		return
	}

	ret := ReplacePlaceHolder(string(content), func(input []string) map[string]string {
		ret := make(map[string]string)
		for _, v := range input {
			ret[v] = v
		}
		return ret
	})

	expected := `{
    "taobao_api":{
        "app_key":"27755801",
        "app_secret":"order-taobao_secret"
    },
    "config":{
        "tests":[
            "name",
            "password"
        ]
    }
}`
	if ret != expected {
		t.Errorf("expected %v but got %v", expected, ret)
	}
}

func TestGetConfigFromConsul(t *testing.T) {
	conf, err := GetConfigFromConsul(Consul{
		Address: "127.0.0.1:8500",
	}, "/myapp")
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(string(conf))
}

type Config struct {
	Address string `json:"address"`
}

func TestDecodeConfigFromConsul(t *testing.T) {
	var conf Config
	err := JSONConfigFromConsul(Consul{
		Address: "127.0.0.1:8500",
	}, "/myapp", &conf)
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("%#v", conf)
}
