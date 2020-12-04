package cmd

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"sobe-kit/gokit_tool"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	create      bool
	port        int
	reGen       bool
	projectPath string
	serviceName string
)

var rootCmd = &cobra.Command{
	Use:   "gokit",
	Short: "gokit gen micr service",
	Run: func(cmd *cobra.Command, args []string) {
		work()
	},
}

func init() {
	rootCmd.Flags().BoolVarP(&create, "create", "c", false, "是否是生成项目")
	rootCmd.Flags().IntVarP(&port, "port", "P", 2048, "服务端口，默认2048")
	rootCmd.Flags().StringVarP(&projectPath, "project-path", "p", "", "生成项目路径，需要包含项目名称,例如 ./app")
	rootCmd.Flags().StringVarP(&serviceName, "service-name", "s", "", "proto 服务名称；默认跟项目名称相同")
	rootCmd.Flags().BoolVarP(&reGen, "re-generate", "r", false, "是否重新生成 api grpc/client、endpoints、transport")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func FirstUpper(input string) string {
	if len(input) <= 0 {
		return input
	}

	return strings.ToUpper(input[:1]) + input[1:]
}

func work() {
	if serviceName == "" {
		serviceName = filepath.Base(projectPath)
	}

	if len(strings.TrimSpace(projectPath)) == 0 {
		log.Fatal("please input -p flags")
	}

	serviceName = FirstUpper(serviceName)
	err := genGRPCServerAndClient(projectPath)
	if err != nil {
		log.Fatal(err)
	}
}

func genGRPCServerAndClient(projectPath string) error {
	data := &gokit_tool.Data{
		PkgName:     filepath.Base(projectPath),
		ServiceName: serviceName,
		ImportPath:  filepath.Join(gokit_tool.ParseProjectImportPath(projectPath), "grpc"),
		Quote:       "`",
		ProjectPath: projectPath,
		Port:        port,
	}
	grpcPath := filepath.Join(projectPath, "grpc")
	protoFilePath := filepath.Join(grpcPath, fmt.Sprintf("%v.pb.go", data.PkgName))

	if create {
		// create proto
		err := gokit_tool.GenProto(filepath.Join(projectPath, "grpc", "protos", fmt.Sprintf("%v.proto", data.PkgName)), data)
		if err != nil {
			log.Fatal(err)
		}

		// create /conf/xx.conf
		err = gokit_tool.GenConfig(filepath.Join(projectPath, "conf", fmt.Sprintf("%v.conf", data.PkgName)), data)
		if err != nil {
			return err
		}

		// create project/service/service.go
		err = gokit_tool.GenService(filepath.Join(projectPath, "service", "service.go"), data)
		if err != nil {
			return err
		}

		// create project/service/health.go
		err = gokit_tool.GenHealth(filepath.Join(projectPath, "service", "health.go"), data)
		if err != nil {
			return err
		}

		// create project/main.go
		err = gokit_tool.GenMain(filepath.Join(projectPath, "main.go"), data)
		if err != nil {
			return err
		}

		// create build.sh
		err = gokit_tool.GenBuild(filepath.Join(projectPath, "build.sh"), data)
		if err != nil {
			return errors.WithStack(err)
		}

		// create Dockerfile
		err = gokit_tool.GenDockerfile(filepath.Join(projectPath, "Dockerfile"), data)
		if err != nil {
			return errors.WithStack(err)
		}

		// create Makefile
		err = gokit_tool.GenMakefile(filepath.Join(projectPath, "Makefile"), data)
		if err != nil {
			return errors.WithStack(err)
		}

		// compile protos
		err = gokit_tool.CompileProto(projectPath, serviceName, data.PkgName)
		if err != nil {
			return err
		}
	}

	if reGen || create {
		var err error
		// parse project/grpc/prject.pb.go file
		data.Methods, err = gokit_tool.ParseProtoPBFile(protoFilePath, serviceName)
		if err != nil {
			return err
		}

		// create /project/api/api.go file
		err = gokit_tool.GenAPI(filepath.Join(projectPath, "api", "api.go"), data)
		if err != nil {
			return err
		}

		// create project/grpc/endpoints/endpoints.go
		err = gokit_tool.GenEndpoints(filepath.Join(grpcPath, "endpoints", "endpoints.go"), data)
		if err != nil {
			return err
		}

		// create project/grpc/transport/transport.go
		err = gokit_tool.GenTransport(filepath.Join(grpcPath, "transport", "transport.go"), data)
		if err != nil {
			return err
		}

		// create project/grpc/client/client.go
		err = gokit_tool.GenClient(filepath.Join(grpcPath, "client", "client.go"), data)
		if err != nil {
			return err
		}
	}
	return nil
}
