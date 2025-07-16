package disqus

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/eiblog/eiblog/pkg/config"
	"github.com/eiblog/eiblog/pkg/model"
)

// disqus api
const (
	apiPostsCount    = "https://disqus.com/api/3.0/threads/set.json"
	apiPostsList     = "https://disqus.com/api/3.0/threads/listPostsThreaded"
	apiPostCreate    = "https://disqus.com/api/3.0/posts/create.json"
	apiPostApprove   = "https://disqus.com/api/3.0/posts/approve.json"
	apiThreadCreate  = "https://disqus.com/api/3.0/threads/create.json"
	apiThreadDetails = "https://disqus.com/api/3.0/threads/details.json"

	disqusAPIKey = "E8Uh5l5fHZ6gD8U3KycjAIAk46f68Zw7C6eW8WSjZvCLXebZ7p0r1yrYDrLilk2F"
)

// DisqusClient disqus client
type DisqusClient struct {
	Host string

	Conf config.Disqus
}

// NewDisqusClient new disqus client
func NewDisqusClient(host string, conf config.Disqus) (*DisqusClient, error) {
	if conf.ShortName == "" || conf.PublicKey == "" || conf.AccessToken == "" {
		return nil, errors.New("disqus: config incompleted")
	}
	return &DisqusClient{Host: host, Conf: conf}, nil
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
func (cli *DisqusClient) PostsCount(articles map[string]*model.Article) error {
	vals := url.Values{}
	vals.Set("api_key", cli.Conf.PublicKey)
	vals.Set("forum", cli.Conf.ShortName)
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
		resp, err := http.DefaultClient.Get(apiPostsCount + "?" + vals.Encode())
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)
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
				if article.Thread == "" && v.ID != "" {
					article.Thread = v.ID
				}
			}
		}
	}
	return nil
}

// PostsListResp 获取评论列表
type PostsListResp struct {
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
		ProfileURL string
		Avatar     struct {
			Cache string
		}
	}
	Thread string
}

// PostsList 评论列表
func (cli *DisqusClient) PostsList(article *model.Article, cursor string) (*PostsListResp, error) {
	vals := url.Values{}
	vals.Set("api_key", disqusAPIKey)
	vals.Set("forum", cli.Conf.ShortName)
	vals.Set("thread:ident", "post-"+article.Slug)
	vals.Set("cursor", cursor)
	vals.Set("order", "popular")
	vals.Set("limit", "50")

	resp, err := http.DefaultClient.Get(apiPostsList + "?" + vals.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(b))
	}

	result := &PostsListResp{}
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
	IPAddress   string
	Identifier  string
	UserAgent   string
}

// PostCreateResp create comments resp
type PostCreateResp struct {
	Code     int
	Response postDetail
}

// PostCreate 评论文章
func (cli *DisqusClient) PostCreate(pc *PostComment) (*PostCreateResp, error) {
	vals := url.Values{}
	vals.Set("api_key", disqusAPIKey)
	vals.Set("message", pc.Message)
	vals.Set("parent", pc.Parent)
	vals.Set("thread", pc.Thread)
	vals.Set("author_email", pc.AuthorEmail)
	vals.Set("author_name", pc.AuthorName)
	// vals.Set("state", "approved")

	req, err := http.NewRequest(http.MethodPost, apiPostCreate, strings.NewReader(vals.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", "https://disqus.com")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(b))
	}
	result := &PostCreateResp{}
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
func (cli *DisqusClient) PostApprove(post string) error {
	vals := url.Values{}
	vals.Set("api_key", disqusAPIKey)
	vals.Set("access_token", cli.Conf.AccessToken)
	vals.Set("post", post)

	req, err := http.NewRequest(http.MethodPost, apiPostApprove, strings.NewReader(vals.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", "https://disqus.com")
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
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
func (cli *DisqusClient) ThreadCreate(article *model.Article, btitle string) error {
	vals := url.Values{}
	vals.Set("api_key", disqusAPIKey)
	vals.Set("access_token", cli.Conf.AccessToken)
	vals.Set("forum", cli.Conf.ShortName)
	vals.Set("title", article.Title+" | "+btitle)
	vals.Set("identifier", "post-"+article.Slug)

	urlPath := fmt.Sprintf("https://%s/post/%s.html", cli.Host, article.Slug)
	vals.Set("url", urlPath)

	req, err := http.NewRequest(http.MethodPost, apiThreadCreate, strings.NewReader(vals.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", "https://disqus.com")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	b, err := io.ReadAll(resp.Body)
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

// threadDetailsResp thread info
type threadDetailsResp struct {
	Code     int
	Response struct {
		ID string
	}
}

// ThreadDetails thread详细
func (cli *DisqusClient) ThreadDetails(article *model.Article) error {
	vals := url.Values{}
	vals.Set("api_key", disqusAPIKey)
	vals.Set("access_token", cli.Conf.AccessToken)
	vals.Set("forum", cli.Conf.ShortName)
	vals.Set("thread:ident", "post-"+article.Slug)

	resp, err := http.DefaultClient.Get(apiThreadDetails + "?" + vals.Encode())
	if err != nil {
		return err
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(string(b))
	}

	result := &threadDetailsResp{}
	err = json.Unmarshal(b, result)
	if err != nil {
		return err
	}
	article.Thread = result.Response.ID
	return nil
}
