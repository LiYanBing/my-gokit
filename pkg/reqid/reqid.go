package reqid

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"os"
	"time"

	"github.com/spaolacci/murmur3"
)

var (
	// virtual process id
	pid = uint16(time.Now().UnixNano() & 65535)
	// machine id
	machineFlag uint16
)

type key int

const (
	requestIDKey key = 0 // used for Context
)

func init() {
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	machineFlag = uint16(murmur3.Sum32([]byte(hostname)) & 65535)
}

func GenRequestID() string {
	var b [12]byte
	binary.LittleEndian.PutUint16(b[:], pid)
	binary.LittleEndian.PutUint16(b[2:], machineFlag)
	binary.LittleEndian.PutUint64(b[4:], uint64(time.Now().UnixNano()))
	return base64.URLEncoding.EncodeToString(b[:])
}

func NewBlankContextWithRequestID(requestID string) context.Context {
	return context.WithValue(context.Background(), requestIDKey, requestID)
}

func FromContext(ctx context.Context) (requestID string, ok bool) {
	requestID, ok = ctx.Value(requestIDKey).(string)
	return
}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}
