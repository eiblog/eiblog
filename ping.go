package main

import (
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/eiblog/utils/logd"
)

type Feeder interface {
	PingFunc(url string)
}

// superfeedr
// http://<your-hub-name>.superfeedr.com/
type Superfeedr struct {
	URL string
}

func (f *Superfeedr) PingFunc(urls ...string) {
	vals := url.Values{}
	vals.Set("hub.mode", "publish")
	for _, u := range urls {
		vals.Add("hub.url", u)
	}
	res, err := http.PostForm(f.URL, vals)
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
	if res.StatusCode != 200 {
		logd.Error(string(data))
	}
}
