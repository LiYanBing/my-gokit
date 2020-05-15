package props

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
)

func GetConfigFromFile(filePath string) ([]byte, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return content, nil
}

func JSONConfigFromFile(filePath string, obj interface{}) error {
	conf, err := GetConfigFromFile(filePath)
	if err != nil {
		return err
	}

	return json.NewDecoder(bytes.NewBuffer(conf)).Decode(obj)
}
