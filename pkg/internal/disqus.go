// Package internal provides ...
package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/eiblog/eiblog/v2/pkg/config"
	"github.com/eiblog/eiblog/v2/pkg/model"
)

// disqus api
const (
	apiPostsCount   = "https://disqus.com/api/3.0/threads/set.json"
	apiPostsList    = "https://disqus.com/api/3.0/threads/listPosts.json"
	apiPostCreate   = "https://disqus.com/api/3.0/posts/create.json"
	apiPostApprove  = "https://disqus.com/api/3.0/posts/approve.json"
	apiThreadCreate = "https://disqus.com/api/3.0/threads/create.json"
)

func checkDisqusConfig() error {
	if config.Conf.BlogApp.Disqus.ShortName != "" &&
		config.Conf.BlogApp.Disqus.PublicKey != "" &&
		config.Conf.BlogApp.Disqus.AccessToken != "" {
		return nil
	}
	return errors.New("disqus: config incompleted")
}

// postsCountResp 评论数量响应
type postsCountResp struct {
	Code     int
	Response []struct {
		ID          string
		Posts       int
		Identifiers []string
	}
}

// PostsCount 获取文章评论数量
func PostsCount(articles map[string]*model.Article) error {
	if err := checkDisqusConfig(); err != nil {
		return err
	}

	vals := url.Values{}
	vals.Set("api_key", config.Conf.BlogApp.Disqus.PublicKey)
	vals.Set("forum", config.Conf.BlogApp.Disqus.ShortName)
	// batch get
	var count, index int
	for _, article := range articles {
		if index < len(articles) && count < 50 {
			count++
			index++

			vals.Add("thread:ident", "post-"+article.Slug)
			continue
		}
		count = 0
		resp, err := httpGet(apiPostsCount + "?" + vals.Encode())
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		// check http status code
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
			slug := v.Identifiers[0][i+1:]

			if article := articles[slug]; article != nil {
				article.Count = v.Posts
				article.Thread = v.ID
			}
		}
	}
	return nil
}

// postsListResp 获取评论列表
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
	ID        string
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

// PostsList 评论列表
func PostsList(slug, cursor string) (*postsListResp, error) {
	if err := checkDisqusConfig(); err != nil {
		return nil, err
	}

	vals := url.Values{}
	vals.Set("api_key", config.Conf.BlogApp.Disqus.PublicKey)
	vals.Set("forum", config.Conf.BlogApp.Disqus.ShortName)
	vals.Set("thread:ident", "post-"+slug)
	vals.Set("cursor", cursor)
	vals.Set("limit", "50")

	resp, err := httpGet(apiPostsList + "?" + vals.Encode())
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

// PostComment 评论
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

// PostCreate 评论文章
func PostCreate(pc *PostComment) (*postCreateResp, error) {
	if err := checkDisqusConfig(); err != nil {
		return nil, err
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
	resp, err := httpPostHeader(apiPostCreate, vals, header)
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

// approvedResp 批准评论通过
type approvedResp struct {
	Code     int
	Response []struct {
		ID string
	}
}

// PostApprove 批准评论
func PostApprove(post string) error {
	if err := checkDisqusConfig(); err != nil {
		return err
	}

	vals := url.Values{}
	vals.Set("api_key", config.Conf.BlogApp.Disqus.PublicKey)
	vals.Set("access_token", config.Conf.BlogApp.Disqus.AccessToken)
	vals.Set("post", post)

	header := http.Header{"Referer": {"https://disqus.com"}}
	resp, err := httpPostHeader(apiPostApprove, vals, header)
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
	return json.Unmarshal(b, result)
}

// threadCreateResp 创建thread
type threadCreateResp struct {
	Code     int
	Response struct {
		ID string
	}
}

// ThreadCreate 创建thread
func ThreadCreate(article *model.Article, btitle string) error {
	if err := checkDisqusConfig(); err != nil {
		return err
	}

	vals := url.Values{}
	vals.Set("api_key", config.Conf.BlogApp.Disqus.PublicKey)
	vals.Set("access_token", config.Conf.BlogApp.Disqus.AccessToken)
	vals.Set("forum", config.Conf.BlogApp.Disqus.ShortName)
	vals.Set("title", article.Title+" | "+btitle)
	vals.Set("identifier", "post-"+article.Slug)

	urlPath := fmt.Sprintf("https://%s/post/%s.html", config.Conf.BlogApp.Host, article.Slug)
	vals.Set("url", urlPath)

	resp, err := httpPost(apiThreadCreate, vals)
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
	article.Thread = result.Response.ID
	return nil
}
