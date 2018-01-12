// Package main provides ...
// Get article' comments count
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/deepzz0/logd"
	"github.com/eiblog/eiblog/setting"
)

var ErrDisqusConfig = errors.New("disqus config incorrect")

func correctDisqusConfig() bool {
	return setting.Conf.Disqus.PostsCount != "" &&
		setting.Conf.Disqus.PublicKey != "" &&
		setting.Conf.Disqus.ShortName != ""
}

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
	if !correctDisqusConfig() {
		return ErrDisqusConfig
	}

	time.AfterFunc(time.Duration(setting.Conf.Disqus.Interval)*time.Hour, func() {
		err := PostsCount()
		if err != nil {
			logd.Error(err)
		}
	})

	vals := url.Values{}
	vals.Set("api_key", setting.Conf.Disqus.PublicKey)
	vals.Set("forum", setting.Conf.Disqus.ShortName)

	var count, index int
	for index < len(Ei.Articles) {
		for ; index < len(Ei.Articles) && count < 50; index++ {
			artc := Ei.Articles[index]
			vals.Add("thread:ident", "post-"+artc.Slug)
			count++
		}
		count = 0
		resp, err := http.Get(setting.Conf.Disqus.PostsCount + "?" + vals.Encode())
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
	if !correctDisqusConfig() {
		return nil, ErrDisqusConfig
	}

	vals := url.Values{}
	vals.Set("api_key", setting.Conf.Disqus.PublicKey)
	vals.Set("forum", setting.Conf.Disqus.ShortName)
	vals.Set("thread:ident", "post-"+slug)
	vals.Set("cursor", cursor)
	vals.Set("limit", "50")

	resp, err := http.Get(setting.Conf.Disqus.PostsList + "?" + vals.Encode())
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

type PostComment struct {
	Message     string
	Parent      string
	Thread      string
	AuthorEmail string
	AuthorName  string
	IpAddress   string
	Identifier  string
	UserAgent   string
}

type postCreateResp struct {
	Code     int
	Response postDetail
}

// 评论文章
func PostCreate(pc *PostComment) (*postCreateResp, error) {
	if !correctDisqusConfig() {
		return nil, ErrDisqusConfig
	}

	vals := url.Values{}
	vals.Set("api_key", "E8Uh5l5fHZ6gD8U3KycjAIAk46f68Zw7C6eW8WSjZvCLXebZ7p0r1yrYDrLilk2F")
	vals.Set("message", pc.Message)
	vals.Set("parent", pc.Parent)
	vals.Set("thread", pc.Thread)
	vals.Set("author_email", pc.AuthorEmail)
	vals.Set("author_name", pc.AuthorName)
	// vals.Set("state", "approved")

	request, err := http.NewRequest("POST", setting.Conf.Disqus.PostCreate, strings.NewReader(vals.Encode()))
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
	Code     int
	Response []struct {
		Id string
	}
}

func PostApprove(post string) error {
	if !correctDisqusConfig() {
		return ErrDisqusConfig
	}

	vals := url.Values{}
	vals.Set("api_key", setting.Conf.Disqus.PublicKey)
	vals.Set("access_token", setting.Conf.Disqus.AccessToken)
	vals.Set("post", post)

	request, err := http.NewRequest("POST", setting.Conf.Disqus.PostApprove, strings.NewReader(vals.Encode()))
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

// 创建thread
type threadCreateResp struct {
	Code     int
	Response struct {
		Id string
	}
}

func ThreadCreate(artc *Article) error {
	if !correctDisqusConfig() {
		return ErrDisqusConfig
	}

	vals := url.Values{}
	vals.Set("api_key", setting.Conf.Disqus.PublicKey)
	vals.Set("access_token", setting.Conf.Disqus.AccessToken)
	vals.Set("forum", setting.Conf.Disqus.ShortName)
	vals.Set("title", artc.Title+" | "+Ei.BTitle)
	vals.Set("identifier", "post-"+artc.Slug)
	urlPath := fmt.Sprintf("https://%s/post/%s.html", setting.Conf.Mode.Domain, artc.Slug)
	vals.Set("url", urlPath)

	resp, err := http.PostForm(setting.Conf.Disqus.ThreadCreate, vals)
	if err != nil {
		return err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(string(b))
	}

	result := &threadCreateResp{}
	err = json.Unmarshal(b, result)
	if err != nil {
		return err
	}

	artc.Thread = result.Response.Id
	return nil
}
