package cmd

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/liyanbing/my-gokit/gokit_tool"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	create        bool
	port          int
	reGen         bool
	projectPath   string
	serviceName   string
	namespace     string
	imageRegistry string
	metricPort    int
)

var rootCmd = &cobra.Command{
	Use:   "gokit",
	Short: "gokit gen micr service",
	Run: func(cmd *cobra.Command, args []string) {
		work()
	},
}

func init() {
	rootCmd.Flags().BoolVarP(&create, "create", "c", false, "创建项目")
	rootCmd.Flags().IntVarP(&port, "port", "P", 2048, "服务端口(默认2048)")
	// TODO metrics 上报注册中心
	rootCmd.Flags().IntVarP(&metricPort, "metric_port", "m", 0, "prometheus 抓取端口，如果不填则不会暴露该端口")
	rootCmd.Flags().StringVarP(&projectPath, "project-path", "p", "", "生成项目路径(需要包含项目名称例如./app)")
	rootCmd.Flags().StringVarP(&serviceName, "service-name", "s", "", "proto服务名称(默认跟项目名称相同)")
	rootCmd.Flags().BoolVarP(&reGen, "generate", "g", false, "重新生成 api grpc/client、endpoints、transport文件")
	rootCmd.Flags().StringVarP(&namespace, "namespace", "n", "sobe", "k8s对象的namespace(默认值sobe)")
	rootCmd.Flags().StringVarP(&imageRegistry, "check-registry", "i", "", "镜像仓库")
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
		fmt.Println("请输入完整的文件路径(例如 -p=./user)")
		return
	}

	if create && len(strings.TrimSpace(imageRegistry)) == 0 {
		fmt.Println("请输入镜像仓库(例如 -i=hub.docker.com/sobe/)")
		return
	}

	serviceName = FirstUpper(serviceName)
	err := genGRPCServerAndClient(projectPath)
	if err != nil {
		log.Fatal(err)
	}
}

func genGRPCServerAndClient(projectPath string) error {
	importPrefix := gokit_tool.ParseProjectImportPath(projectPath)
	grpcPath := filepath.Join(projectPath, "grpc")

	data := &gokit_tool.Data{
		ImportPrefix: importPrefix,
		PkgName:      filepath.Base(projectPath),
		ServiceName:  serviceName,
		ImportPath:   filepath.Join(importPrefix, "grpc", "pb"),
		Quote:        "`",
		ProjectPath:  projectPath,
		Port:         port,
		Namespace:    namespace,
		Registry:     imageRegistry,
		MetricPort:   metricPort,
	}

	if create {
		// create proto
		err := gokit_tool.GenProto(filepath.Join(projectPath, "grpc", fmt.Sprintf("%v.proto", data.PkgName)), data)
		if err != nil {
			log.Fatal(err)
		}

		// create /conf/xx.conf
		err = gokit_tool.GenConfig(filepath.Join(projectPath, "conf", fmt.Sprintf("%v.conf", data.PkgName)), data)
		if err != nil {
			return err
		}

		// create /conf/deployment.yaml
		err = gokit_tool.GenK8sDeployment(filepath.Join(projectPath, "conf", fmt.Sprintf("%s-deploy.yaml", data.PkgName)), data)
		if err != nil {
			return err
		}

		// create /conf/service.yaml
		err = gokit_tool.GenK8sService(filepath.Join(projectPath, "conf", fmt.Sprintf("%s-svc.yaml", data.PkgName)), data)
		if err != nil {
			return err
		}

		err = gokit_tool.GenK8sConfigMap(filepath.Join(projectPath, "conf", fmt.Sprintf("%s-config.yaml", data.PkgName)), data)
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
		data.Methods, err = gokit_tool.ParseProtoPBFile(filepath.Join(grpcPath, "pb", fmt.Sprintf("%v.pb.go", data.PkgName)), serviceName)
		if err != nil {
			return err
		}

		// create  project/grpc/pb/constant.go
		err = gokit_tool.GenConstant(filepath.Join(grpcPath, "pb", "constant.go"), data)
		if err != nil {
			return err
		}

		// create project/api/api.go file
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
