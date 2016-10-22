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

func PostsCount() {
	if setting.Conf.Disqus.PostsCount == "" || setting.Conf.Disqus.PublicKey == "" || setting.Conf.Disqus.ShortName == "" {
		return
	}
	baseUrl := setting.Conf.Disqus.PostsCount + "?api_key=" + setting.Conf.Disqus.PublicKey + "&forum=" + setting.Conf.Disqus.ShortName + "&"
	var count, index int
	for index < len(Ei.Articles) {
		logd.Debugf("count=====%d, index=======%d, length=======%d, bool=========%t", count, index, len(Ei.Articles), index < len(Ei.Articles) && count < 50)
		var threads []string
		for ; index < len(Ei.Articles) && count < 50; index++ {
			artc := Ei.Articles[index]
			threads = append(threads, fmt.Sprintf("thread:ident=post-%s", artc.Slug))
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
		if resp.StatusCode != http.StatusOK {
			logd.Error(string(b))
			break
		}
		rst := result{}
		err = json.Unmarshal(b, &rst)
		if err != nil {
			logd.Error(err)
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
	time.AfterFunc(time.Duration(setting.Conf.Disqus.Interval)*time.Hour, PostsCount)
}

type postsList struct {
	Cursor struct {
		HasNext bool
		Next    string
	}
	Code     int
	Response []struct {
		Parent    int
		Id        string
		CreatedAt string
		Message   string
		Author    struct {
			Name       string
			ProfileUrl string
			Avatar     struct {
				Cache string
			}
		}
	}
}

func PostsList(slug, cursor string) *postsList {
	if setting.Conf.Disqus.PostsList == "" || setting.Conf.Disqus.PublicKey == "" || setting.Conf.Disqus.ShortName == "" {
		return nil
	}
	url := setting.Conf.Disqus.PostsList + "?limit=50&api_key=" + setting.Conf.Disqus.PublicKey + "&forum=" + setting.Conf.Disqus.ShortName + "&cursor=" + cursor + "&thread:ident=post-" + slug
	resp, err := http.Get(url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logd.Error(err)
		return nil
	}
	if resp.StatusCode != http.StatusOK {
		logd.Error(string(b))
		return nil
	}
	pl := &postsList{}
	err = json.Unmarshal(b, pl)
	if err != nil {
		logd.Error(err)
		return nil
	}
	return pl
}
