package gokit_tool

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

// 解析指定的 ***.pb.go 文件，只解析 serviceName+Server 服务端接口部分
func ParseProtoPBFile(fileName, serviceName, pkgName, importPath, projectPath string) (*Data, error) {
	var (
		methodList []*ast.Field
	)

	fileSet := token.NewFileSet()
	f, err := parser.ParseFile(fileSet, fileName, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	serviceServer := false
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.InterfaceType:
			methodList = x.Methods.List
		case *ast.TypeSpec:
			if x.Name.Name == fmt.Sprintf("%sServer", serviceName) {
				serviceServer = true
			}
		}
		return !serviceServer
	})
	return genData(pkgName, serviceName, importPath, projectPath, methodList), nil
}

func genData(pkgName, serviceName, importPath, projectPath string, methodList []*ast.Field) *Data {
	methods := make([]*Method, 0, len(methodList))
	for _, m := range methodList {
		curF := m.Type.(*ast.FuncType)

		doc := bytes.NewBuffer(nil)
		if m.Doc != nil {
			for _, v := range m.Doc.List {
				doc.WriteString(v.Text)
			}
			doc.WriteString("\n")
		}

		methods = append(methods, &Method{
			Doc:          doc.String(),
			Name:         m.Names[0].Name,
			RequestName:  curF.Params.List[1].Type.(*ast.StarExpr).X.(*ast.Ident).Name,
			ResponseName: curF.Results.List[0].Type.(*ast.StarExpr).X.(*ast.Ident).Name,
		})
	}

	return &Data{
		PkgName:     pkgName,
		ServiceName: serviceName,
		ImportPath:  importPath,
		Methods:     methods,
		Quote:       "`",
		ProjectPath: projectPath,
	}
}
