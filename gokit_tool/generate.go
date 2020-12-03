package gokit_tool

import (
	"bytes"
	"go/format"
	"path/filepath"
	"strings"
	"text/template"
)

// api
func GenAPI(filePath string, data *Data) error {
	return genFileWithTemplate(filePath, apiTemplate, true, data)
}

// endpoints
func GenEndpoints(filePath string, data *Data) error {
	return genFileWithTemplate(filePath, endpointsTemplate, true, data)
}

// transport
func GenTransport(filePath string, data *Data) error {
	return genFileWithTemplate(filePath, transportTemplate, true, data)
}

// client
func GenClient(filePath string, data *Data) error {
	return genFileWithTemplate(filePath, clientTemplate, true, data)
}

// service
func GenService(filePath string, data *Data) error {
	return genFileWithTemplate(filePath, serviceTemplate, false, data)
}

// server
func GenServer(filePath string, data *Data) error {
	return genFileWithTemplate(filePath, serverTemplate, false, data)
}

// client
func GenServerClient(filePath string, data *Data) error {
	return genFileWithTemplate(filePath, serverClientTemplate, false, data)
}

// cmd
func GenCmd(filePath string, data *Data) error {
	return genFileWithTemplate(filePath, cmdTemplate, false, data)
}

// main.go
func GenMain(filePath string, data *Data) error {
	return genFileWithTemplate(filePath, mainTemplate, false, data)
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
	PkgName     string
	ServiceName string
	ImportPath  string
	Methods     []*Method
	Quote       string
	ProjectPath string
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

func genFileWithTemplate(filePath, content string, override bool, data *Data) error {
	contentBuf := bytes.NewBuffer(nil)
	tpl, err := template.New("").Funcs(tplFunc).Parse(content)
	if err != nil {
		return err
	}

	err = tpl.Execute(contentBuf, data)
	if err != nil {
		return err
	}

	if formatter != nil {
		content, err := formatter(contentBuf.String())
		if err != nil {
			return err
		}
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
