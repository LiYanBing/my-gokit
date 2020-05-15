package gokit_tool

import (
	"bytes"
	"html/template"
	"testing"
)

var str = `
	这里是测试
{{- range .Methods}}
	{{.Name}}->{{.RequestName}}->{{.ResponseName}}
	{{FirstLower .Name}}
{{- end}}
`

func TestParseProjectImportPath(t *testing.T) {
	data := Data{
		PkgName:     "shux",
		ServiceName: "mt",
		ImportPath:  "github.com/shux",
		Methods: []*Method{
			{
				Doc:          "// 这里是文档",
				Name:         "HelloWorld",
				RequestName:  "HelloWorldRequest",
				ResponseName: "HelloWorldResponse",
			},
			{
				Doc:          "// 这里是文档",
				Name:         "Ping",
				RequestName:  "PingRequest",
				ResponseName: "PingResponse",
			},
		},
	}

	tpl, err := template.New("").Funcs(tplFunc).Parse(str)
	if err != nil {
		t.Error(err)
		return
	}

	buf := bytes.NewBuffer(nil)
	err = tpl.Execute(buf, data)
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(buf.String())
}
