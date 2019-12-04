// Package main provides ...
// Get article' comments count
package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
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
		resp, err := Get(setting.Conf.Disqus.PostsCount + "?" + vals.Encode())
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

	resp, err := Get(setting.Conf.Disqus.PostsList + "?" + vals.Encode())
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

	header := http.Header{"Referer": {"https://disqus.com"}}
	resp, err := PostWithHeader(setting.Conf.Disqus.PostCreate, vals, header)
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

	header := http.Header{"Referer": {"https://disqus.com"}}
	resp, err := PostWithHeader(setting.Conf.Disqus.PostApprove, vals, header)
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

	resp, err := PostForm(setting.Conf.Disqus.ThreadCreate, vals)
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

///////////////////////////// HTTP 请求 /////////////////////////////

var httpClient = &http.Client{
	Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	},
}

func newRequest(method, rawurl string, vals url.Values) (*http.Request, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	host := u.Host
	// 获取主机IP
	ips, err := net.LookupHost(u.Host)
	if err != nil {
		return nil, err
	}
	if len(ips) == 0 {
		return nil, errors.New("not found ip: " + u.Host)
	}
	// 设置ServerName
	httpClient.Transport.(*http.Transport).TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	u.Host = ips[0]
	// 创建HTTP Request
	var req *http.Request
	if vals != nil {
		req, err = http.NewRequest(method, u.String(), strings.NewReader(vals.Encode()))
	} else {
		req, err = http.NewRequest(method, u.String(), nil)
	}
	if err != nil {
		return nil, err
	}
	// 改变Host
	req.Host = host
	return req, nil
}

// Get HTTP Get请求
func Get(rawurl string) (*http.Response, error) {
	req, err := newRequest(http.MethodGet, rawurl, nil)
	if err != nil {
		return nil, err
	}
	// 发起请求
	return httpClient.Do(req)
}

// PostForm HTTP Post请求
func PostForm(rawurl string, vals url.Values) (*http.Response, error) {
	req, err := newRequest(http.MethodPost, rawurl, vals)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// 发起请求
	return httpClient.Do(req)
}

// PostWithHeader HTTP Post请求，自定义Header
func PostWithHeader(rawurl string, vals url.Values, header http.Header) (*http.Response, error) {
	req, err := newRequest(http.MethodPost, rawurl, vals)
	if err != nil {
		return nil, err
	}
	// set header
	req.Header = header
	// 发起请求
	return httpClient.Do(req)
}
