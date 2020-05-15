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
    // say hello world
    rpc HelloWorld (HelloWorldRequest) returns (HelloWorldResponse) {}
	// Ping
	rpc Ping (PingRequest) returns (PingResponse) {}
}

message HelloWorldRequest {
    string input = 1;
}

message HelloWorldResponse {
    string output = 1;
}

message PingRequest {
	string service = 1;
}

message PingResponse {
	string status = 1;
}
`
	compileTemplate = `#!/bin/sh
protoc *.proto --go_out=plugins=grpc:.`

	constantTemplate = `package %s 

var (
	ServiceName = _%s_serviceDesc.ServiceName
)`
)

// create  project/grpc/pkgname.proto、 compile.sh and auto gen pkgname.pb.go
func CreateProtoAndCompile(path, serviceName, pkgName string) error {
	// create proto file
	protoFilePath := filepath.Join(path, fmt.Sprintf("%v.proto", pkgName))
	err := createFile(protoFilePath, fmt.Sprintf(protoTemplate, pkgName, serviceName), 0666)
	if err != nil {
		return err
	}

	// create compile file
	compileFilePath := filepath.Join(path, "compile.sh")
	err = createFile(compileFilePath, compileTemplate, 0777)
	if err != nil {
		return err
	}

	// execute compile file generate ***.pb.go file
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("cd %v && ./compile.sh", filepath.Dir(compileFilePath)))
	if err := cmd.Run(); err != nil {
		return errors.New(fmt.Sprintf("please enter %s then execute compile.sh", filepath.Dir(compileFilePath)))
	}

	return createFile(filepath.Join(path, "constant.go"), fmt.Sprintf(constantTemplate, pkgName, FirstUpper(serviceName)), 0666)
}
