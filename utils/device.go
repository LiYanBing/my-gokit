package utils

import (
	"errors"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func GetLANHost() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		panic(err)
	}

	for _, addr := range addrs {
		a := addr.String()
		if strings.HasPrefix(a, "10.") || strings.HasPrefix(a, "192.168") || strings.HasPrefix(a, "172.") {
			if i := strings.IndexByte(a, '/'); i != -1 {
				return a[:i]
			} else {
				return a
			}
		}
	}
	name, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	return name
}

func GetAppName() string {
	return filepath.Base(os.Args[0])
}

func GetPort(addr string) (int, error) {
	arr := strings.Split(addr, ":")
	if len(arr) != 2 {
		return 0, errors.New("invalid addr")
	}
	ret, err := strconv.ParseInt(arr[1], 10, 64)
	return int(ret), err
}
