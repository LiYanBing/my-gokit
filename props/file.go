package props

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"reflect"
	"time"

	"gopkg.in/yaml.v2"
)

type Options struct {
	FilePath string
	Retry    int
}

func newOptions() *Options {
	return &Options{
		FilePath: "./app",
		Retry:    1,
	}
}

type Option func(opt *Options)

func FilePath(filepath string) Option {
	return func(opt *Options) {
		opt.FilePath = filepath
	}
}

func Retry(retry int) Option {
	return func(opt *Options) {
		opt.Retry = retry
	}
}

type Decoder func(content []byte) error

func JsonDecoder(obj interface{}) Decoder {
	return func(content []byte) error {
		return json.NewDecoder(bytes.NewReader(content)).Decode(obj)
	}
}

func XMLDecoder(obj interface{}) Decoder {
	return func(content []byte) error {
		return xml.NewDecoder(bytes.NewReader(content)).Decode(obj)
	}
}

func YAMLDecoder(obj interface{}) Decoder {
	return func(content []byte) error {
		return yaml.NewDecoder(bytes.NewReader(content)).Decode(obj)
	}
}

func check(obj interface{}) error {
	if reflect.TypeOf(obj).Kind() != reflect.Ptr {
		return errors.New("obj must be a ptr")
	}
	return nil
}

func GetConfigFromFile(filePath string) ([]byte, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return content, nil
}

func Do(attempts int) time.Duration {
	if attempts > 13 {
		return 2 * time.Minute
	}
	return time.Duration(math.Pow(float64(attempts), math.E)) * time.Millisecond * 100
}

type Callback func(n int, err error) bool

func Always(n int, err error) bool {
	return true
}

func ConfigFromFile(decoder Decoder, call Callback, opts ...Option) error {
	o := newOptions()
	for _, opt := range opts {
		opt(o)
	}

	var (
		content []byte
		err     error
	)

	for i := 1; i <= o.Retry; i++ {
		content, err = GetConfigFromFile(o.FilePath)
		if err != nil {
			retry := call(i, err)
			if !retry {
				return err
			}
			time.Sleep(Do(i))
			continue
		}
		return decoder(content)
	}
	return fmt.Errorf("retry more than max")
}
