package iglog

import (
	"fmt"
	"net/url"
	"strings"
	"testing"
)

func TestLogger_Warningf(t *testing.T) {
	targetUrl := "http://www.baidu.com/test/1?name=zhangsan&age=18"
	target, err := url.Parse(targetUrl)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("rawPath:", target.RawPath, "rawQuery:", target.RawQuery, "URI", target.RequestURI())

	q := target.Query()
	q.Set("receiver", strings.Join([]string{"1", "2"}, ","))
	q.Set("group", "group")
	fmt.Println("22", q.Encode())
	target.RawPath = "666"
	target.RawQuery = q.Encode()
	t.Log(target.String())
}
