module github.com/liyanbing/my-gokit

go 1.12

replace (
	golang.org/x/crypto => github.com/golang/crypto v0.0.0-20200323165209-0ec3e9974c59
	golang.org/x/mod => github.com/golang/mod v0.2.0
	golang.org/x/net => github.com/golang/net v0.0.0-20200324143707-d3edc9973b7e
	golang.org/x/sync => github.com/golang/sync v0.0.0-20200317015054-43a5402ce75a
	golang.org/x/sys => github.com/golang/sys v0.0.0-20200331124033-c3d80250170d
	golang.org/x/text => github.com/golang/text v0.3.2
	golang.org/x/tools => github.com/golang/tools v0.0.0-20200331202046-9d5940d49312
	golang.org/x/xerrors => github.com/golang/xerrors v0.0.0-20191204190536-9bdfabe68543
	google.golang.org/genproto => github.com/googleapis/go-genproto v0.0.0-20200331122359-1ee6d9798940
	google.golang.org/grpc => github.com/grpc/grpc-go v1.28.0
)

require (
	github.com/go-kit/kit v0.10.0
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/hashicorp/consul/api v1.4.0
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/opentracing/opentracing-go v1.1.0
	github.com/openzipkin-contrib/zipkin-go-opentracing v0.4.5
	github.com/openzipkin/zipkin-go v0.2.2
	github.com/spf13/cobra v1.0.0
	google.golang.org/grpc v1.27.0
)
