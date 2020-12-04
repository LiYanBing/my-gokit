package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

var serverAddress string
var serverName string
var healthAddress string

func init() {
	flag.StringVar(&serverAddress, "server-address", "127.0.0.1:2048", "server address of grpc health")
	flag.StringVar(&serverName, "server-name", "name", "server name of grpc server")
	flag.StringVar(&healthAddress, "health-address", "0.0.0.0:8080", "health address of http")
}

func main() {
	flag.Parse()
	grpcConn, err := GetGRPCConn(serverAddress)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if grpcConn != nil {
			grpcConn.Close()
		}
	}()

	method := fmt.Sprintf("/%s/Health", serverName)
	err = http.ListenAndServe(healthAddress, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		var reply HealthResponse
		err = grpcConn.Invoke(context.Background(), method, &HealthRequest{
			Service: serverName,
		}, &reply)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		writer.WriteHeader(http.StatusOK)
	}))
	if err != nil {
		log.Fatal(err)
	}
}

// GRPC connection
func GetGRPCConn(addr string) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                time.Hour,
			Timeout:             time.Second * 5,
			PermitWithoutStream: false,
		}),
	}
	return grpc.Dial(addr, opts...)
}

type HealthRequest struct {
	Service              string   `protobuf:"bytes,1,opt,name=service,proto3" json:"service,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *HealthRequest) Reset()         { *m = HealthRequest{} }
func (m *HealthRequest) String() string { return proto.CompactTextString(m) }
func (*HealthRequest) ProtoMessage()    {}
func (*HealthRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_order_04bd29bcc01b9a2e, []int{0}
}
func (m *HealthRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_HealthRequest.Unmarshal(m, b)
}
func (m *HealthRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_HealthRequest.Marshal(b, m, deterministic)
}
func (dst *HealthRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_HealthRequest.Merge(dst, src)
}
func (m *HealthRequest) XXX_Size() int {
	return xxx_messageInfo_HealthRequest.Size(m)
}
func (m *HealthRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_HealthRequest.DiscardUnknown(m)
}

var xxx_messageInfo_HealthRequest proto.InternalMessageInfo

func (m *HealthRequest) GetService() string {
	if m != nil {
		return m.Service
	}
	return ""
}

type HealthResponse struct {
	Status               int64    `protobuf:"varint,1,opt,name=status,proto3" json:"status,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *HealthResponse) Reset()         { *m = HealthResponse{} }
func (m *HealthResponse) String() string { return proto.CompactTextString(m) }
func (*HealthResponse) ProtoMessage()    {}
func (*HealthResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_order_04bd29bcc01b9a2e, []int{1}
}
func (m *HealthResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_HealthResponse.Unmarshal(m, b)
}
func (m *HealthResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_HealthResponse.Marshal(b, m, deterministic)
}
func (dst *HealthResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_HealthResponse.Merge(dst, src)
}
func (m *HealthResponse) XXX_Size() int {
	return xxx_messageInfo_HealthResponse.Size(m)
}
func (m *HealthResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_HealthResponse.DiscardUnknown(m)
}

var xxx_messageInfo_HealthResponse proto.InternalMessageInfo

func (m *HealthResponse) GetStatus() int64 {
	if m != nil {
		return m.Status
	}
	return 0
}

var fileDescriptor_order_04bd29bcc01b9a2e = []byte{
	// 139 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0xce, 0x2f, 0x4a, 0x49,
	0x2d, 0xd2, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x05, 0x73, 0x94, 0x34, 0xb9, 0x78, 0x3d,
	0x52, 0x13, 0x73, 0x4a, 0x32, 0x82, 0x52, 0x0b, 0x4b, 0x53, 0x8b, 0x4b, 0x84, 0x24, 0xb8, 0xd8,
	0x8b, 0x53, 0x8b, 0xca, 0x32, 0x93, 0x53, 0x25, 0x18, 0x15, 0x18, 0x35, 0x38, 0x83, 0x60, 0x5c,
	0x25, 0x0d, 0x2e, 0x3e, 0x98, 0xd2, 0xe2, 0x82, 0xfc, 0xbc, 0xe2, 0x54, 0x21, 0x31, 0x2e, 0xb6,
	0xe2, 0x92, 0xc4, 0x92, 0xd2, 0x62, 0xb0, 0x52, 0xe6, 0x20, 0x28, 0xcf, 0xc8, 0x81, 0x8b, 0xd5,
	0x1f, 0x64, 0xba, 0x90, 0x39, 0x17, 0x1b, 0x44, 0x8b, 0x90, 0x88, 0x1e, 0xc4, 0x72, 0x14, 0xcb,
	0xa4, 0x44, 0xd1, 0x44, 0x21, 0xe6, 0x2a, 0x31, 0x24, 0xb1, 0x81, 0x1d, 0x69, 0x0c, 0x08, 0x00,
	0x00, 0xff, 0xff, 0xe4, 0x41, 0x65, 0x09, 0xb3, 0x00, 0x00, 0x00,
}
