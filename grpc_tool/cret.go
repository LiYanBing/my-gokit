package grpc_tool

import (
	"crypto/tls"
	"errors"
)

var (
	ErrEmptyConf = errors.New("empty conf")
)

type Cert struct {
	File bool   `json:"file"` // 是否是文件的形式
	Cert string `json:"cert"` // 证书
	Key  string `json:"key"`  // 密钥
}

func LoadCertificates(cnf *Cert) (*tls.Certificate, error) {
	if cnf == nil {
		return nil, ErrEmptyConf
	}

	var cert tls.Certificate
	var err error

	if cnf.File {
		cert, err = tls.LoadX509KeyPair(cnf.Cert, cnf.Key)
	} else {
		cert, err = tls.X509KeyPair([]byte(cnf.Cert), []byte(cnf.Key))
	}
	if err != nil {
		return nil, err
	}

	return &cert, nil
}
