module iglog

go 1.14

require (
	alert v0.0.0-00010101000000-000000000000
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	reqid v0.0.0-00010101000000-000000000000
)

replace (
	alert => ../alert
	reqid => ../reqid
)
