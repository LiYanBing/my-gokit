package gokit_tool

import (
	"bytes"
	"fmt"
	"go/format"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

// proto
func GenProto(filePath string, data *Data) error {
	return genFileWithTemplate(filePath, protoTemplate, false, false, data)
}

// api
func GenAPI(filePath string, data *Data) error {
	return genFileWithTemplate(filePath, apiTemplate, true, true, data)
}

// endpoints
func GenEndpoints(filePath string, data *Data) error {
	return genFileWithTemplate(filePath, endpointsTemplate, true, true, data)
}

// transport
func GenTransport(filePath string, data *Data) error {
	return genFileWithTemplate(filePath, transportTemplate, true, true, data)
}

// client
func GenClient(filePath string, data *Data) error {
	return genFileWithTemplate(filePath, clientTemplate, true, true, data)
}

// service
func GenService(filePath string, data *Data) error {
	return genFileWithTemplate(filePath, serviceTemplate, false, true, data)
}

// health
func GenHealth(filePath string, data *Data) error {
	return genFileWithTemplate(filePath, healthTemplate, false, true, data)
}

// conf
func GenConfig(filePath string, data *Data) error {
	return genFileWithTemplate(filePath, configTemplate, false, false, data)
}

// constant
func GenConstant(filePath string, data *Data) error {
	return genFileWithTemplate(filePath, constantTemplate, false, true, data)
	//return createFile(filepath.Join(path, "grpc", "constant.go"), fmt.Sprintf(constantTemplate, pkgName, serviceName), true, 0666)
}

// **.pb.go
func CompileProto(path, serviceName, pkgName string) error {
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("cd %s && protoc --proto_path=./grpc --go_out=plugins=grpc:./grpc ./grpc/*.proto", pkgName))
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

// main.go
func GenMain(filePath string, data *Data) error {
	return genFileWithTemplate(filePath, mainTemplate, false, true, data)
}

// build.sh
func GenBuild(filePath string, data *Data) error {
	return genFileWithTemplate(filePath, buildTemplate, false, false, data)
}

// Dockerfile
func GenDockerfile(filePath string, data *Data) error {
	return genFileWithTemplate(filePath, dockerFileTemplate, false, false, data)
}

// Makefile
func GenMakefile(filePath string, data *Data) error {
	return genFileWithTemplate(filePath, makefileTemplate, false, false, data)
}

// deployment.yaml
func GenK8sDeployment(filePath string, data *Data) error {
	return genFileWithTemplate(filePath, deploymentTemplate, false, false, data)
}

// configmap.yaml
func GenK8sConfigMap(filePath string, data *Data) error {
	return genFileWithTemplate(filePath, configMapTemplate, false, false, data)
}

func GenK8sService(filePath string, data *Data) error {
	return genFileWithTemplate(filePath, k8sServiceTemplate, false, false, data)
}

// server
func GenServer(filePath string, data *Data) error {
	return genFileWithTemplate(filePath, serverTemplate, false, true, data)
}

// client
func GenServerClient(filePath string, data *Data) error {
	return genFileWithTemplate(filePath, serverClientTemplate, false, true, data)
}

// cmd
func GenCmd(filePath string, data *Data) error {
	return genFileWithTemplate(filePath, cmdTemplate, false, true, data)
}

var (
	tplFunc = template.FuncMap{
		"FirstLower":       FirstLower,
		"FirstUpper":       FirstUpper,
		"ServicePath":      ServicePath,
		"ServerPath":       ServerPath,
		"CmdPath":          CMDPath,
		"ServerClientPath": ServerClientPath,
	}
	formatter = formatGo
)

type Data struct {
	PkgName      string
	ServiceName  string
	ImportPath   string
	Methods      []*Method
	Quote        string
	ProjectPath  string
	Port         int
	Namespace    string
	Registry     string
	ImportPrefix string
	MetricPort   int
}

type Method struct {
	Doc          string
	Name         string
	RequestName  string
	ResponseName string
}

func formatGo(src string) (string, error) {
	source, err := format.Source([]byte(src))
	if err != nil {
		return "", err
	}
	return string(source), nil
}

func genFileWithTemplate(filePath, content string, override, format bool, data *Data) error {
	contentBuf := bytes.NewBuffer(nil)
	tpl, err := template.New("").Funcs(tplFunc).Parse(content)
	if err != nil {
		return err
	}

	err = tpl.Execute(contentBuf, data)
	if err != nil {
		return err
	}

	if format && formatter != nil {
		content, _ := formatter(contentBuf.String())
		contentBuf.Reset()
		contentBuf.WriteString(content)
	}

	return createFile(filePath, contentBuf.String(), override, 0666)
}

func FirstLower(input string) string {
	if len(input) <= 0 {
		return input
	}

	return strings.ToLower(input[:1]) + input[1:]
}

func FirstUpper(input string) string {
	if len(input) <= 0 {
		return input
	}

	return strings.ToUpper(input[:1]) + input[1:]
}

func ServicePath(input string) string {
	dir := filepath.Dir(input)
	return filepath.Join(dir, "service")
}

func ServerPath(input string) string {
	dir := filepath.Dir(input)
	return filepath.Join(dir, "server")
}

func ServerClientPath(input string) string {
	dir := filepath.Dir(input)
	return filepath.Join(dir, "client")
}

func CMDPath(input string) string {
	dir := filepath.Dir(input)
	return filepath.Join(dir, "cmd")
}
