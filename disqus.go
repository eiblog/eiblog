// Package main provides ...
// Get article' comments count
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/deepzz0/logd"
	"github.com/eiblog/eiblog/setting"
)

var ErrDisqusConfig = errors.New("disqus config incorrect")

// 定时获取所有文章评论数量
type postsCountResp struct {
	Code     int
	Response []struct {
		Id          string
		Posts       int
		Identifiers []string
	}
}

func PostsCount() error {
	if setting.Conf.Disqus.PostsCount == "" ||
		setting.Conf.Disqus.PublicKey == "" ||
		setting.Conf.Disqus.ShortName == "" {
		return ErrDisqusConfig
	}

	time.AfterFunc(time.Duration(setting.Conf.Disqus.Interval)*time.Hour, func() {
		err := PostsCount()
		if err != nil {
			logd.Error(err)
		}
	})

	baseUrl := setting.Conf.Disqus.PostsCount +
		"?api_key=" + setting.Conf.Disqus.PublicKey +
		"&forum=" + setting.Conf.Disqus.ShortName + "&"
	var count, index int
	for index < len(Ei.Articles) {
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
			return err
		}
		defer resp.Body.Close()

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return errors.New(string(b))
		}

		result := &postsCountResp{}
		err = json.Unmarshal(b, result)
		if err != nil {
			return err
		}
		for _, v := range result.Response {
			i := strings.Index(v.Identifiers[0], "-")
			artc := Ei.MapArticles[v.Identifiers[0][i+1:]]
			if artc != nil {
				artc.Count = v.Posts
				artc.Thread = v.Id
			}
		}
	}

	return nil
}

// 获取文章评论列表
type postsListResp struct {
	Cursor struct {
		HasNext bool
		Next    string
	}
	Code     int
	Response []postDetail
}

type postDetail struct {
	Parent    int
	Id        string
	CreatedAt string
	Message   string
	IsDeleted bool
	Author    struct {
		Name       string
		ProfileUrl string
		Avatar     struct {
			Cache string
		}
	}
	Thread string
}

func PostsList(slug, cursor string) (*postsListResp, error) {
	if setting.Conf.Disqus.PostsList == "" ||
		setting.Conf.Disqus.PublicKey == "" ||
		setting.Conf.Disqus.ShortName == "" {
		return nil, ErrDisqusConfig
	}

	url := setting.Conf.Disqus.PostsList + "?limit=50&api_key=" +
		setting.Conf.Disqus.PublicKey + "&forum=" + setting.Conf.Disqus.ShortName +
		"&cursor=" + cursor + "&thread:ident=post-" + slug
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(b))
	}

	result := &postsListResp{}
	err = json.Unmarshal(b, result)
	if err != nil {
		return nil, err
	}
	return result, nil
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

type postCreateResp struct {
	Code     int
	Response postDetail
}

// 评论文章
func PostComment(pc *PostCreate) (*postCreateResp, error) {
	if setting.Conf.Disqus.PostsList == "" ||
		setting.Conf.Disqus.PublicKey == "" ||
		setting.Conf.Disqus.ShortName == "" {
		return nil, ErrDisqusConfig
	}
	url := setting.Conf.Disqus.PostCreate +
		"?api_key=E8Uh5l5fHZ6gD8U3KycjAIAk46f68Zw7C6eW8WSjZvCLXebZ7p0r1yrYDrLilk2F" +
		"&message=" + pc.Message + "&parent=" + pc.Parent +
		"&thread=" + pc.Thread + "&author_email=" + pc.AuthorEmail +
		"&author_name=" + pc.AuthorName

	request, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Referer", "https://disqus.com")
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(b))
	}
	result := &postCreateResp{}
	err = json.Unmarshal(b, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// 批准评论通过
type approvedResp struct {
	Code     int `json:"code"`
	Response []struct {
		Id string `json:"id"`
	} `json:"response"`
}

func PostApprove(post string) error {
	if setting.Conf.Disqus.PostsList == "" ||
		setting.Conf.Disqus.PublicKey == "" ||
		setting.Conf.Disqus.ShortName == "" {
		return ErrDisqusConfig
	}

	url := setting.Conf.Disqus.PostApprove +
		"?api_key=" + setting.Conf.Disqus.PublicKey +
		"&access_token=" + setting.Conf.Disqus.AccessToken +
		"&post=" + post
	request, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Referer", "https://disqus.com")
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(string(b))
	}

	result := &approvedResp{}
	err = json.Unmarshal(b, result)
	if err != nil {
		return err
	}

	return nil
}
