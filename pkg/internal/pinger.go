// Package internal provides ...
package internal

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/eiblog/eiblog/pkg/config"

	"github.com/sirupsen/logrus"
)

// feedrPingFunc http://<your-hub-name>.superfeedr.com/
var feedrPingFunc = func(btitle, slug string) error {
	feedrHost := config.Conf.EiBlogApp.FeedRPC.FeedrURL
	if feedrHost == "" {
		return nil
	}

	vals := url.Values{}
	vals.Set("hub.mode", "publish")
	vals.Add("hub.url", fmt.Sprintf("https://%s/post/%s.html",
		config.Conf.BackupApp.Host, slug))
	resp, err := httpPost(feedrHost, vals)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 204 {
		return fmt.Errorf("pinger: status code: %d, %s",
			resp.StatusCode, string(data))
	}
	return nil
}

// rpcPingParam ping to rpc, eg. google baidu
// params:
//	 BlogName string `xml:"param>value>string"`
//	 HomePage string `xml:"param>value>string"`
//	 Article  string `xml:"param>value>string"`
//	 RSS_URL  string `xml:"param>value>string"`
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
var rpcPingFunc = func(btitle, slug string) error {
	if len(config.Conf.EiBlogApp.FeedRPC.PingRPC) == 0 {
		return nil
	}
	param := rpcPingParam{MethodName: "weblogUpdates.extendedPing"}
	param.Params.Param = [4]rpcValue{
		0: rpcValue{Value: btitle},
		1: rpcValue{Value: "https://" + config.Conf.EiBlogApp.Host},
		2: rpcValue{Value: fmt.Sprintf("https://%s/post/%s.html", config.Conf.EiBlogApp.Host, slug)},
		3: rpcValue{Value: "https://" + config.Conf.EiBlogApp.Host + "/rss.html"},
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
	for _, addr := range config.Conf.EiBlogApp.FeedRPC.PingRPC {
		resp, err := httpPostHeader(addr, data, header)
		if err != nil {
			logrus.Error("rpcPingFunc.httpPostHeader: ", err)
			continue
		}
		defer resp.Body.Close()
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logrus.Error("rpcPingFunc.ReadAll: ", err)
			continue
		}
		if resp.StatusCode != 200 {
			logrus.Error("rpcPingFunc.failed: ", string(data))
		}
	}
	return nil
}

// PingFunc ping blog article to SE
func PingFunc(btitle, slug string) {
	err := feedrPingFunc(btitle, slug)
	if err != nil {
		logrus.Error("pinger: PingFunc feedr: ", err)
	}
	err = rpcPingFunc(btitle, slug)
	if err != nil {
		logrus.Error("pinger: PingFunc: rpc: ", err)
	}
}
