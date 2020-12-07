PROJECT_NAME=$(notdir $(shell pwd))

vpath $(PROJECT_NAME) ./_bin
vpath %.proto ./grpc
vpath %.pb.go ./grpc/pb

PROTO_PATH=./grpc
PB_TARGET=./grpc

export SERVER_ADDRESS=0.0.0.0:2048

$(PROJECT_NAME).pb.go: $(PROJECT_NAME).proto
	@echo "building *.pb.go"
	protoc --proto_path=$(PROTO_PATH) --go_out=plugins=grpc:$(PB_TARGET) $(PROTO_PATH)/*.proto
	sobe-kit -g -p=./../order

$(PROJECT_NAME):  $(PROJECT_NAME).pb.go
	@echo "build " $(PROJECT_NAME)
	if [ ! -e _bin ]; then mkdir _bin; fi;
	go build -i -o _bin/$(PROJECT_NAME)

.PHONY: gen 
gen:
	@echo "building *.pb.go"
	protoc --proto_path=$(PROTO_PATH) --go_out=plugins=grpc:$(PB_TARGET) $(PROTO_PATH)/*.proto
	sobe-kit -g -p=./../order

.PHONY: check
check:
	go vet ./...
	go test -race $$(go list ./...| grep -v pkg)

.PHONY: build
build: $(PROJECT_NAME)
	@echo successful

.PHONY: run
run: $(PROJECT_NAME)
	@echo "start "  $(PROJECT_NAME)
	for i in $$(ps -ef | grep _bin/$(PROJECT_NAME) | grep -v grep | awk '{print $$2}') ; do \
	   kill $$i;\
	done
	nohup _bin/$(PROJECT_NAME) >> _bin/nohup.out 2>&1 &

.PHONY: clean
clean:
	rm -rf _bin
	rm -rf ./api
	rm -rf ./grpc/pb
	rm -rf ./grpc/client
	rm -rf ./grpc/endpoints
	rm -rf ./grpc/transport

.PHONY: image
image:
	make proto
	rm -rf _bin
	sh ./build.sh