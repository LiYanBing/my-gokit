package alert

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/viki-org/dnscache"
)

type DingTalkConf struct {
	Secret    string   `json:"secret"`     // 加签时的密钥
	BatchSize int      `json:"batch_size"` // 批量发送个数
	Interval  int64    `json:"interval"`   // 发送间隔时间
	WebHook   string   `json:"web_hook"`   // webHook地址
	Receivers []string `json:"receivers"`  // 需要指定at的联系人
}

type dingTalk struct {
	cnf          *DingTalkConf
	client       *http.Client
	buff         []string
	enable       bool
	c            chan struct{}
	quit         chan struct{}
	lock         sync.Mutex
	wg           sync.WaitGroup
	sign         string
	lastSignTime time.Time
}

type DingTalkMessageType string

const (
	DingTalkMessageTypeMarkdown DingTalkMessageType = "markdown"
	DingTalkMessageTypeText     DingTalkMessageType = "text"
	DingTalkMessageTypeLink     DingTalkMessageType = "link"
)

type DingTalkMessage struct {
	MsgType DingTalkMessageType `json:"msgtype,omitempty"`
	Text    struct {
		Content string `json:"content,omitempty"`
	} `json:"text,omitempty"`
	Link struct {
		Text       string `json:"text,omitempty"`
		Title      string `json:"title,omitempty"`
		PicUrl     string `json:"picUrl,omitempty"`
		MessageUrl string `json:"messageUrl,omitempty"`
	}
	Markdown struct {
		Title string `json:"title,omitempty"`
		Text  string `json:"text,omitempty"`
	} `json:"markdown,omitempty"`
	At struct {
		AtMobiles []string `json:"atMobiles,omitempty"`
		IsAtAll   bool     `json:"isAtAll,omitempty"`
	} `json:"at,omitempty"`
}

func NewAlertWithDingTalk(cnf *DingTalkConf) Alert {
	s := &dingTalk{
		cnf: cnf,
		client: &http.Client{
			Transport: Transport(),
			Timeout:   time.Second * 5,
		},
		enable: true,
		c:      make(chan struct{}),
		quit:   make(chan struct{}),
	}
	s.asyncLoop()
	return s
}

func NewDefaultDingTalk(cnf *DingTalkConf) Alert {
	dingTalkAlert := &dingTalk{
		cnf: cnf,
		client: &http.Client{
			Transport: Transport(),
			Timeout:   time.Second * 5,
		},
		enable: true,
		c:      make(chan struct{}),
		quit:   make(chan struct{}),
	}
	dingTalkAlert.asyncLoop()
	return dingTalkAlert
}

func (s *dingTalk) Send(msg string) error {
	webhook := s.cnf.WebHook
	if s.cnf.Secret != "" {
		webhook = s.signedWebhook()
	}

	return s.send(webhook, msg)
}

func (s *dingTalk) AsyncSend(msg string) {
	s.lock.Lock()
	s.buff = append(s.buff, msg)
	defer s.lock.Unlock()

	if len(s.buff) >= s.cnf.BatchSize {
		s.c <- struct{}{}
	}
}

func (s *dingTalk) Close() error {
	s.c <- struct{}{}
	close(s.quit)
	s.wg.Wait()
	return nil
}

func (s *dingTalk) asyncLoop() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-time.After(time.Second * time.Duration(s.cnf.Interval)):
			case <-s.c:
			case <-s.quit:
				return
			}

			s.lock.Lock()
			if len(s.buff) > 0 {
				sendMsg := s.buff[:]
				s.buff = s.buff[:0]
				s.lock.Unlock()
				data := strings.Join(sendMsg, "\n\n")
				_ = s.Send(data)
			} else {
				s.lock.Unlock()
			}
		}
	}()
}

func (s *dingTalk) send(target, msg string) error {
	body := DingTalkMessage{
		MsgType: DingTalkMessageTypeText,
		Text: struct {
			Content string `json:"content,omitempty"`
		}{
			Content: msg,
		},
		At: struct {
			AtMobiles []string `json:"atMobiles,omitempty"`
			IsAtAll   bool     `json:"isAtAll,omitempty"`
		}{
			AtMobiles: s.cnf.Receivers,
		},
	}

	buf := bytes.NewBuffer(nil)
	err := json.NewEncoder(buf).Encode(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, target, buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	type temp struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}

	var ret temp
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return err
	}

	if ret.ErrCode != 0 && ret.ErrMsg != "ok" {
		return fmt.Errorf("send errcode %v errmsg %v", ret.ErrCode, ret.ErrMsg)
	}
	return nil
}

func (s *dingTalk) signedWebhook() string {
	var sign string
	now := time.Now()
	target, _ := url.Parse(s.cnf.WebHook)
	timestamp := now.UnixNano() / 1000000

	if s.sign != "" && s.lastSignTime.Add(time.Minute*50).After(time.Now()) {
		timestamp = s.lastSignTime.UnixNano() / 1000000
		sign = s.sign
	} else {
		data := fmt.Sprintf("%v\n%v", timestamp, s.cnf.Secret)
		h := hmac.New(sha256.New, []byte(s.cnf.Secret))
		h.Write([]byte(data))
		sign = base64.StdEncoding.EncodeToString(h.Sum(nil))
		s.sign = sign
		s.lastSignTime = now
	}

	query := target.Query()
	query.Set("timestamp", fmt.Sprintf("%v", timestamp))
	query.Set("sign", sign)
	target.RawQuery = query.Encode()

	return target.String()
}

func Transport() *http.Transport {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
			Resolver: &net.Resolver{
				Dial: func(ctx context.Context, network string, addr string) (net.Conn, error) {
					separator := strings.LastIndex(addr, ":")
					resolver := dnscache.New(time.Minute * 3)
					ip, err := resolver.FetchOneString(addr[:separator])
					if err != nil {
						return nil, err
					}
					return net.Dial(network, ip+addr[separator:])
				},
			},
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}
