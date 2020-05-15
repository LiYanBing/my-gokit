#### 安装 
```
go get github.com/liyanbing/my-gokit
```

#### 使用
在指定目录生成项目，必须生成timi项目

```
my-gokit -c -p=/Users/Leo/Desktop/shuinfo/my-gokit/timi

-c：表示需要创建一个微服务项目（获取使用--create）
-p：需要生成的微服务项目路径，这里是我的要生成的项目地址，主要需要在目录中加上项目的名称（这里是timi）
```

#### 根据传递的项目路径生成当前项目的导入路径规则

 1.如果当前的 projectPath 已经存在于go mod 的项目下，则会以go mod 的路径+/projectName 为导入路径
 > 例如 /Home/lise/Desktop/app/login；login为新创建的登录分布式模块，而/Home/lise/Desktop/app已经是一个存在go.mod(github.com/app)文件的项目， 则返回 github.com/app/login
     
 2.如果当前项目在GOPATH下面，则返回以GOPATH为准的导入路径
 > 例如给定路径为 /Users/zhangsan/go/src/app/login；设置的GOPATH路径为 /Users/zhangsan/go 则返回的路径为 app/login
    
 3.如果项目处在GOPATH下面，且项目使用go.mod进行管理，则优先会使用 go.mod 的项目路径，跟情况1相同
 
 4.如果既不在go.mod 管理的项目下，也不在已经设置的gopath下面，则会自动创建一个 github.com/+projectname 的go mod依赖管理
 
 #### 生成项目结构如下
 
 ```
timi
├── api
│   └── api.go 
├── client
│   └── client.go
├── cmd
│   └── cmd.go
├── conf
│   ├── timi-local.conf
│   ├── timi-pro.conf
│   └── timi-test.conf
├── grpc
│   ├── client
│   │   └── client.go
│   ├── compile.sh
│   ├── constant.go
│   ├── endpoints
│   │   └── endpoints.go
│   ├── timi.pb.go
│   ├── timi.proto
│   └── transport
│       └── transport.go
├── main.go
├── server
│   └── server.go
└── service
    └── service.go

```
#### 结构说明
> api/api.go：整个timi项目 timi.proto 生成的 timi.pb.go 文件对应的 TimiServer 接口签名；这个文件没有实际意义；
因为存在微服务之间的相互调用，所以这里仅仅为了方便其他需要调用该服务时方便查看当前服务提供有哪些接口
注意：当前文件夹内容为自动生成

> client/client.go：grpc服务调用客户端，在项目创建成功之后可以调用该客户端做调试

> cmd/cmd.go：timi微服务模块启动时的命令参数集成

> grpc/**：grpc相关代码，包含了 proto文件，proto生成的pb.go文件、使用gokit包装好的 endpoints、tranport、client（用于连接grpc服务的客户端代码）

> conf/***：整个微服务的配置文件

> server/server.go：微服务的 服务启动代码块

> service/service.go：用于实现具体业务逻辑模块

#### 启动服务
```
    go run main.go server [--config-path=配置文件路径；默认使用conf/**-local.conf配置]
    会在本地 4096端口监听服务
```

#### 启动客户端验证服务是否正常
```
    go run main.go client [--config-path=配置文件路径；默认使用conf/**-local.conf配置]
```

