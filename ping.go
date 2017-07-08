package main

import (
	"bytes"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/eiblog/eiblog/setting"
	"github.com/eiblog/utils/logd"
)

type Pinger interface {
	PingFunc(url string)
}

// superfeedr
// http://<your-hub-name>.superfeedr.com/
type superfeedr struct{}

func (*superfeedr) PingFunc(slug string) {
	if setting.Conf.FeedrURL == "" {
		return
	}
	vals := url.Values{}
	vals.Set("hub.mode", "publish")
	vals.Add("hub.url", "https://"+setting.Conf.Mode.Domains[0]+"/post/"+slug+".html")
	res, err := http.PostForm(setting.Conf.FeedrURL, vals)
	if err != nil {
		logd.Error(err)
		return
	}
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logd.Error(err)
		return
	}
	if res.StatusCode != 204 {
		logd.Error(res.StatusCode, string(data))
	}
}

// google baidu
// params
// BlogName string `xml:"param>value>string"`
// HomePage string `xml:"param>value>string"`
// Article  string `xml:"param>value>string"`
// RSS_URL  string `xml:"param>value>string"`
type pingRPC struct {
	XMLName    xml.Name `xml:"methodCall"`
	MethodName string   `xml:"methodName"`
	Params     struct {
		Param [4]rpcValue `xml:"param"`
	} `xml:"params"`
}

type rpcValue struct {
	Value string `xml:"value>string"`
}

func (p *pingRPC) PingFunc(slug string) {
	if len(setting.Conf.PingRPCs) == 0 {
		return
	}
	p.Params.Param[1].Value = "https://" + setting.Conf.Mode.Domains[0] + "/post/" + slug + ".html"
	buf := &bytes.Buffer{}
	buf.WriteString(xml.Header)
	enc := xml.NewEncoder(buf)
	if err := enc.Encode(p); err != nil {
		logd.Error(err)
		return
	}
	for _, url := range setting.Conf.PingRPCs {
		rep, err := http.Post(url, "text/xml", buf)
		if err != nil {
			logd.Error(err)
			continue
		}
		defer rep.Body.Close()
		data, err := ioutil.ReadAll(rep.Body)
		if err != nil {
			logd.Error(err)
			continue
		}
		if rep.StatusCode != 200 {
			logd.Error(string(data))
			continue
		}
	}
}

var Pings []Pinger

func init() {
	Pings = append(Pings, &superfeedr{})
	pr := &pingRPC{MethodName: "weblogUpdates.extendedPing"}
	pr.Params.Param = [4]rpcValue{
		0: rpcValue{Value: Ei.BTitle},
		1: rpcValue{Value: "https://" + setting.Conf.Mode.Domains[0]},
		2: rpcValue{},
		3: rpcValue{Value: "https://" + setting.Conf.Mode.Domains[0] + "/rss.html"},
	}
	Pings = append(Pings, pr)
}

func DoPings(slug string) {
	for _, p := range Pings {
		go p.PingFunc(slug)
	}
}
