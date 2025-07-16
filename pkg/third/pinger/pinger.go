package pinger

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/eiblog/eiblog/pkg/config"

	"github.com/sirupsen/logrus"
)

// Pinger pinger
type Pinger struct {
	Host string

	Conf config.FeedRPC
}

// NewPinger new pinger
func NewPinger(host string, conf config.FeedRPC) (*Pinger, error) {
	if conf.FeedrURL == "" {
		return nil, fmt.Errorf("feedr url is empty")
	}
	return &Pinger{Host: host, Conf: conf}, nil
}

// PingFunc ping blog article to SE
func (p *Pinger) PingFunc(btitle, slug string) {
	err := p.feedrPingFunc(btitle, slug)
	if err != nil {
		logrus.Error("pinger: PingFunc feedr: ", err)
	}
	err = p.rpcPingFunc(btitle, slug)
	if err != nil {
		logrus.Error("pinger: PingFunc: rpc: ", err)
	}
}

// feedrPingFunc http://<your-hub-name>.superfeedr.com/
func (p *Pinger) feedrPingFunc(btitle, slug string) error {
	vals := url.Values{}
	vals.Set("hub.mode", "publish")
	vals.Add("hub.url", fmt.Sprintf("https://%s/post/%s.html", p.Host, slug))
	resp, err := http.DefaultClient.PostForm(p.Conf.FeedrURL, vals)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 204 {
		return fmt.Errorf("pinger: status code: %d, %s", resp.StatusCode, string(data))
	}
	return nil
}

// rpcPingParam ping to rpc, eg. google baidu
// params:
//
//	BlogName string `xml:"param>value>string"`
//	HomePage string `xml:"param>value>string"`
//	Article  string `xml:"param>value>string"`
//	RSS_URL  string `xml:"param>value>string"`
type rpcPingParam struct {
	XMLName    xml.Name `xml:"methodCall"`
	MethodName string   `xml:"methodName"`
	Params     struct {
		Param [4]rpcValue `xml:"param"`
	} `xml:"params"`
}

type rpcValue struct {
	Value string `xml:"value>string"`
}

// rpcPingFunc ping rpc
func (p *Pinger) rpcPingFunc(btitle, slug string) error {
	param := rpcPingParam{MethodName: "weblogUpdates.extendedPing"}
	param.Params.Param = [4]rpcValue{
		0: {Value: btitle},
		1: {Value: "https://" + p.Host},
		2: {Value: fmt.Sprintf("https://%s/post/%s.html", p.Host, slug)},
		3: {Value: "https://" + p.Host + "/rss.html"},
	}
	buf := bytes.Buffer{}
	buf.WriteString(xml.Header)
	enc := xml.NewEncoder(&buf)
	if err := enc.Encode(param); err != nil {
		return err
	}
	data := buf.Bytes()
	header := http.Header{}
	header.Set("Content-Type", "text/xml")

	for _, addr := range p.Conf.PingRPC {
		resp, err := http.DefaultClient.Post(addr, "text/xml", bytes.NewReader(data))
		if err != nil {
			logrus.Error("rpcPingFunc.httpPostHeader: ", err)
			continue
		}
		defer resp.Body.Close()

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			logrus.Error("rpcPingFunc.ReadAll: ", err)
			continue
		}
		if resp.StatusCode != 200 {
			logrus.Error("rpcPingFunc.failed: ", string(data), resp.StatusCode)
		}
	}
	return nil
}
