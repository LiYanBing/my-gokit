package alert

import (
	"fmt"
	"sync/atomic"
	"testing"
	"unsafe"
)

func TestDingTalk_Disable(t *testing.T) {
	var ok bool
	b := true
	pp := unsafe.Pointer(&ok)
	if *(*bool)(atomic.LoadPointer(&pp)) {
		fmt.Println("11")
	} else {
		fmt.Println(22)
		atomic.StorePointer((*unsafe.Pointer)(pp), unsafe.Pointer(&b))
	}

	fmt.Println("ok:", ok, "b:", b)
}
