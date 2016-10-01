// Package main provides ...
// Get article' comments count
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/eiblog/eiblog/setting"
	"github.com/eiblog/utils/logd"
)

type result struct {
	Code     int
	Response []struct {
		Posts       int
		Identifiers []string
	}
}

func CommentsCount() {
	if setting.Conf.Disqus.URL == "" || setting.Conf.Disqus.PublicKey == "" || setting.Conf.Disqus.ShortName == "" {
		return
	}
	baseUrl := setting.Conf.Disqus.URL + "?api_key=" + setting.Conf.Disqus.PublicKey + "&forum=" + setting.Conf.Disqus.ShortName + "&"
	domain := "http://" + runmode.Domain
	if runmode.EnableHttps {
		domain = "https://" + runmode.Domain
	}
	var count, index int
	for index < len(Ei.Articles) {
		logd.Debugf("count=====%d, index=======%d, length=======%d, bool=========%t", count, index, len(Ei.Articles), index < len(Ei.Articles) && count < 10)
		var threads []string
		for ; index < len(Ei.Articles) && count < 20; index++ {
			artc := Ei.Articles[index]
			threads = append(threads, fmt.Sprintf("thread=link:%s/post/%s.html", domain, artc.Slug))
			count++
		}
		count = 0
		url := baseUrl + strings.Join(threads, "&")
		resp, err := http.Get(url)
		if err != nil {
			logd.Error(err)
			break
		}
		defer resp.Body.Close()
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logd.Error(err)
			break
		}
		rst := result{}
		err = json.Unmarshal(b, &rst)
		if err != nil {
			logd.Error(err)
			break
		}
		if rst.Code != SUCCESS {
			logd.Error(rst.Code)
			break
		}
		for _, v := range rst.Response {
			i := strings.Index(v.Identifiers[0], "-")
			artc := Ei.MapArticles[v.Identifiers[0][i+1:]]
			if artc != nil {
				artc.Count = v.Posts
			}
		}
	}
	time.AfterFunc(time.Duration(setting.Conf.Disqus.Interval)*time.Hour, CommentsCount)
}
