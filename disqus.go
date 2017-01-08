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
	baseUrl := setting.Conf.Disqus.PostsCount +
		"?api_key=" + setting.Conf.Disqus.PublicKey +
		"&forum=" + setting.Conf.Disqus.ShortName + "&"
	var count, index int
	for index < len(Ei.Articles) {
		logd.Debugf("count=====%d, index=======%d, length=======%d, bool=========%t\n", count, index, len(Ei.Articles), index < len(Ei.Articles) && count < 50)
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
		Thread string
	}
}

func PostsList(slug, cursor string) *postsList {
	if setting.Conf.Disqus.PostsList == "" || setting.Conf.Disqus.PublicKey == "" || setting.Conf.Disqus.ShortName == "" {
		return nil
	}
	url := setting.Conf.Disqus.PostsList + "?limit=50&api_key=" +
		setting.Conf.Disqus.PublicKey + "&forum=" + setting.Conf.Disqus.ShortName +
		"&cursor=" + cursor + "&thread:ident=post-" + slug
	resp, err := http.Get(url)
	if err != nil {
		logd.Error(err)
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

type PostCreate struct {
	Message     string `json:"message"`
	Parent      string `json:"parent"`
	Thread      string `json:"thread"`
	AuthorEmail string `json:"author_email"`
	AuthorName  string `json:"autor_name"`
	IpAddress   string `json:"ip_address"`
	Identifier  string `json:"identifier"`
	UserAgent   string `json:"user_agent"`
}

type PostResponse struct {
	Code     int `json:"code"`
	Response struct {
		Id string `json:"id"`
	} `json:"response"`
}

func PostComment(pc *PostCreate) string {
	if setting.Conf.Disqus.PostsList == "" || setting.Conf.Disqus.PublicKey == "" || setting.Conf.Disqus.ShortName == "" {
		return ""
	}
	url := setting.Conf.Disqus.PostCreate +
		"?api_key=E8Uh5l5fHZ6gD8U3KycjAIAk46f68Zw7C6eW8WSjZvCLXebZ7p0r1yrYDrLilk2F" +
		"&message=" + pc.Message + "&parent=" + pc.Parent +
		"&thread=" + pc.Thread + "&author_email=" + pc.AuthorEmail +
		"&author_name=" + pc.AuthorName

	request, err := http.NewRequest("POST", url, nil)
	if err != nil {
		logd.Error(err)
		return ""
	}
	request.Header.Set("Referer", "https://disqus.com")
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		logd.Error(err)
		return ""
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logd.Error(err)
		return ""
	}
	if resp.StatusCode != http.StatusOK {
		logd.Error(string(b))
		return ""
	}
	pr := &PostResponse{}
	err = json.Unmarshal(b, pr)
	if err != nil {
		logd.Error(err)
		return ""
	}
	logd.Print(pr.Response.Id)
	return pr.Response.Id
}
