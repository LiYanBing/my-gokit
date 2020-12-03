package cmd

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"sobe-kit/gokit_tool"

	"github.com/spf13/cobra"
)

var (
	create      bool
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

	if create {
		err := gokit_tool.CreateProtoAndCompile(filepath.Join(projectPath, "grpc"), serviceName, filepath.Base(projectPath))
		if err != nil {
			log.Fatal(err)
		}
	} else {
		reGen = true
	}

	serviceName = FirstUpper(serviceName)
	err := genGRPCServerAndClient(projectPath)
	if err != nil {
		log.Fatal(err)
	}
}

func genGRPCServerAndClient(projectPath string) error {
	pkgName := filepath.Base(projectPath)
	importPath := filepath.Join(gokit_tool.ParseProjectImportPath(projectPath), "grpc")
	grpcPath := filepath.Join(projectPath, "grpc")
	protoFilePath := filepath.Join(grpcPath, fmt.Sprintf("%v.pb.go", pkgName))

	// compile protos
	err := gokit_tool.CompileProto(grpcPath, serviceName, pkgName)
	if err != nil {
		return err
	}

	// parse project/grpc/prject.pb.go file
	data, err := gokit_tool.ParseProtoPBFile(protoFilePath, serviceName, pkgName, importPath, projectPath)
	if err != nil {
		return err
	}

	if reGen || create {
		// create /project/api/api.go file
		err = gokit_tool.GenAPI(filepath.Join(projectPath, "api", "api.go"), data)
		if err != nil {
			return err
		}
	}

	if reGen || create {
		// create project/grpc/endpoints/endpoints.go
		err = gokit_tool.GenEndpoints(filepath.Join(grpcPath, "endpoints", "endpoints.go"), data)
		if err != nil {
			return err
		}
	}

	if reGen || create {
		// create project/grpc/transport/transport.go
		err = gokit_tool.GenTransport(filepath.Join(grpcPath, "transport", "transport.go"), data)
		if err != nil {
			return err
		}
	}

	if reGen || create {
		// create project/grpc/client/client.go
		err = gokit_tool.GenClient(filepath.Join(grpcPath, "client", "client.go"), data)
		if err != nil {
			return err
		}
	}

	if create {
		// create project/config/****-local.conf、****-test.conf、****-pro.conf
		err = gokit_tool.GenConfig(projectPath, pkgName)
		if err != nil {
			return err
		}

		// create project/service/service.go
		err = gokit_tool.GenService(filepath.Join(projectPath, "service", "service.go"), data)
		if err != nil {
			return err
		}

		// create project/server/server.go
		//err = gokit_tool.GenServer(filepath.Join(projectPath, "server", "server.go"), data)
		//if err != nil {
		//	return err
		//}

		// create project/client/client.go
		//err = gokit_tool.GenServerClient(filepath.Join(projectPath, "client", "client.go"), data)
		//if err != nil {
		//	return err
		//}

		// create project/cmd/cmd.go
		//err = gokit_tool.GenCmd(filepath.Join(projectPath, "cmd", "cmd.go"), data)
		//if err != nil {
		//	return err
		//}

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
	}
	return nil
}
