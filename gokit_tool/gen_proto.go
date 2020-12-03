package gokit_tool

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
)

const (
	protoTemplate = `syntax = "proto3";
package %s;

service %s {
	// Ping
	rpc Ping (PingRequest) returns (PingResponse) {}
}

message PingRequest {
	string service = 1;
}

message PingResponse {
	string status = 1;
}
`
	compileTemplate = `#!/bin/sh
protoc  --go_out=plugins=grpc:./.. ./*.proto`

	constantTemplate = `package %s 

var (
	ServiceName = _%s_serviceDesc.ServiceName
)`
)

// create  project/grpc/pkgname.proto、 compile.sh and auto gen pkgname.pb.go
func CreateProtoAndCompile(path, serviceName, pkgName string) error {
	// create proto file
	protoFilePath := filepath.Join(path, "protos", fmt.Sprintf("%v.proto", pkgName))
	err := createFile(protoFilePath, fmt.Sprintf(protoTemplate, pkgName, serviceName), false, 0666)
	if err != nil {
		return err
	}

	// create compile file
	compileFilePath := filepath.Join(path, "protos", "compile.sh")
	err = createFile(compileFilePath, compileTemplate, false, 0777)
	if err != nil {
		return err
	}

	if err := CompileProto(path, serviceName, pkgName); err != nil {
		return err
	}
	return nil
}

func CompileProto(path, serviceName, pkgName string) error {
	protoPath := filepath.Join(path, "protos")
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("cd %v && ./compile.sh", protoPath))
	if err := cmd.Run(); err != nil {
		return errors.New(fmt.Sprintf("please enter %s then execute compile.sh", protoPath))
	}
	return createFile(filepath.Join(path, "constant.go"), fmt.Sprintf(constantTemplate, pkgName, FirstUpper(serviceName)), true, 0666)
}
