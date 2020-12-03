package props

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type Config struct {
	TaoBaoAPI struct {
		APPKey    string `json:"app_key"`
		AppSecret string `json:"app_secret"`
	} `json:"taobao_api"`
	Config struct {
		Tests []string `json:"tests"`
	} `json:"config"`
}

func TestConfigFromFile(t *testing.T) {
	var cfg Config
	err := ConfigFromFile(JsonDecoder(&cfg), Always, FilePath("./app.conf"))
	assert.Nil(t, err)
	assert.Equal(t, "27755801", cfg.TaoBaoAPI.APPKey)
	assert.Equal(t, "[[.order-taobao_secret]]", cfg.TaoBaoAPI.AppSecret)
	assert.Equal(t, 2, len(cfg.Config.Tests))
	assert.Equal(t, "[[.name]]", cfg.Config.Tests[0])
	assert.Equal(t, "[[.password]]", cfg.Config.Tests[1])
}
