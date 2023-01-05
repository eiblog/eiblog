// Package internal provides ...
package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/eiblog/eiblog/pkg/config"
	"github.com/eiblog/eiblog/pkg/model"
	"github.com/eiblog/eiblog/tools"

	"github.com/sirupsen/logrus"
)

// search mode
const (
	SearchFilter = `"filter":{"bool":{"must":[%s]}}`
	SearchTerm   = `{"term":{"%s":"%s"}}`
	SearchDate   = `{"range":{"date":{"gte":"%s","lte": "%s","format": "yyyy-MM-dd||yyyy-MM||yyyy"}}}` // 2016-10||/M

	ElasticIndex = "eiblog"
	ElasticType  = "article"
)

func init() {
	if checkESConfig() != nil {
		return
	}

	mappings := fmt.Sprintf(`{"mappings":{"%s":{"properties":{"content":{"analyzer":"ik_syno","search_analyzer":"ik_syno","term_vector":"with_positions_offsets","type":"string"},"date":{"index":"not_analyzed","type":"date"},"slug":{"type":"string"},"tag":{"index":"not_analyzed","type":"string"},"title":{"analyzer":"ik_syno","search_analyzer":"ik_syno","term_vector":"with_positions_offsets","type":"string"}}}}}`, "article")
	err := createIndexAndMappings(ElasticIndex, ElasticType, []byte(mappings))
	if err != nil {
		panic(err)
	}
}

func checkESConfig() error {
	if config.Conf.ESHost == "" {
		return errors.New("es: elasticsearch not config")
	}
	return nil
}

// ElasticSearch 搜索文章
func ElasticSearch(query string, size, from int) (*SearchIndexResult, error) {
	if err := checkESConfig(); err != nil {
		return nil, err
	}
	// 分析查询
	var (
		regTerm = regexp.MustCompile(`(tag|slug|date):`)
		idxs    = regTerm.FindAllStringIndex(query, -1)
		length  = len(idxs)
		str, kw string
		filter  []string
	)
	if length == 0 { // 全文搜索
		kw = query
	}
	// 字段搜索，检出，全文搜索
	for i, idx := range idxs {
		if i == length-1 {
			str = query[idx[0]:]
			if space := strings.Index(str, " "); space != -1 && space < len(str)-1 {
				kw = str[space+1:]
				str = str[:space]
			}

		} else {
			str = strings.TrimSpace(query[idx[0]:idxs[i+1][0]])
		}
		kv := strings.Split(str, ":")
		switch kv[0] {
		case "slug":
			filter = append(filter, fmt.Sprintf(SearchTerm, kv[0], kv[1]))
		case "tag":
			filter = append(filter, fmt.Sprintf(SearchTerm, kv[0], kv[1]))
		case "date":
			var date string
			switch len(kv[1]) {
			case 4:
				date = fmt.Sprintf(SearchDate, kv[1], kv[1]+"||/y")
			case 7:
				date = fmt.Sprintf(SearchDate, kv[1], kv[1]+"||/M")
			case 10:
				date = fmt.Sprintf(SearchDate, kv[1], kv[1]+"||/d")
			default:
				break
			}
			filter = append(filter, date)
		}
	}
	// 判断是否为空，判断搜索方式
	dsl := fmt.Sprintf("{"+SearchFilter+"}", strings.Join(filter, ","))
	if kw != "" {
		dsl = strings.Replace(strings.Replace(`{"highlight":{"fields":{"content":{},"title":{}},"post_tags":["\u003c/b\u003e"],"pre_tags":["\u003cb\u003e"]},"query":{"dis_max":{"queries":[{"match":{"title":{"boost":4,"minimum_should_match":"50%","query":"$1"}}},{"match":{"content":{"boost":4,"minimum_should_match":"75%","query":"$1"}}},{"match":{"tag":{"boost":2,"minimum_should_match":"100%","query":"$1"}}},{"match":{"slug":{"boost":1,"minimum_should_match":"100%","query":"$1"}}}],"tie_breaker":0.3}},$2}`, "$1", kw, -1), "$2", fmt.Sprintf(SearchFilter, strings.Join(filter, ",")), -1)
	}
	return indexQueryDSL(ElasticIndex, ElasticType, size, from, []byte(dsl))
}

// ElasticAddIndex 添加或更新索引
func ElasticAddIndex(article *model.Article) error {
	if err := checkESConfig(); err != nil {
		return err
	}

	img := tools.PickFirstImage(article.Content)
	mapping := map[string]interface{}{
		"title":   article.Title,
		"content": tools.IgnoreHTMLTag(article.Content),
		"slug":    article.Slug,
		"tag":     article.Tags,
		"img":     img,
		"date":    article.CreatedAt,
	}
	data, _ := json.Marshal(mapping)
	return indexOrUpdateDocument(ElasticIndex, ElasticType, article.ID, data)
}

// ElasticDelIndex 删除索引
func ElasticDelIndex(ids []int) error {
	if err := checkESConfig(); err != nil {
		return err
	}

	var target []string
	for _, id := range ids {
		target = append(target, fmt.Sprint(id))
	}
	return deleteIndexDocument(ElasticIndex, ElasticType, target)
}

// indicesCreateResult 索引创建结果
type indicesCreateResult struct {
	Acknowledged bool `json:"acknowledged"`
}

// createIndexAndMappings 创建索引和映射关系
func createIndexAndMappings(index, typ string, mappings []byte) error {
	rawurl := fmt.Sprintf("%s/%s/%s", config.Conf.ESHost, index, typ)
	resp, err := httpHead(rawurl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return nil
	}

	rawurl = fmt.Sprintf("%s/%s", config.Conf.ESHost, index)
	resp, err = httpPut(rawurl, mappings)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	result := indicesCreateResult{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return errors.New(string(data))
	}
	if !result.Acknowledged {
		return errors.New(string(data))
	}
	return nil
}

// indexOrUpdateDocument 创建或更新索引
func indexOrUpdateDocument(index, typ string, id int, doc []byte) (err error) {
	rawurl := fmt.Sprintf("%s/%s/%s/%d", config.Conf.ESHost, index, typ, id)
	resp, err := httpPut(rawurl, doc)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	logrus.Debug(string(data))
	return nil
}

type deleteIndexReq struct {
	Index string `json:"_index"`
	Type  string `json:"_type"`
	ID    string `json:"_id"`
}

type deleteIndexResult struct {
	Errors bool `json:"errors"`
	Iterms []map[string]struct {
		Error string `json:"error"`
	} `json:"iterms"`
}

// deleteIndexDocument 删除文档
func deleteIndexDocument(index, typ string, ids []string) error {
	buf := bytes.Buffer{}
	for _, id := range ids {
		dd := deleteIndexReq{Index: index, Type: typ, ID: id}
		m := map[string]deleteIndexReq{"delete": dd}
		b, _ := json.Marshal(m)
		buf.Write(b)
		buf.WriteByte('\n')
	}
	rawurl := fmt.Sprintf("%s/_bulk", config.Conf.ESHost)
	resp, err := httpPost(rawurl, buf.Bytes())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	result := deleteIndexResult{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return err
	}
	if result.Errors {
		for _, iterm := range result.Iterms {
			for _, s := range iterm {
				if s.Error != "" {
					return errors.New(s.Error)
				}
			}
		}
	}
	return nil
}

// SearchIndexResult 查询结果
type SearchIndexResult struct {
	Took float32 `json:"took"`
	Hits struct {
		Total int `json:"total"`
		Hits  []struct {
			ID     string `json:"_id"`
			Source struct {
				Slug    string    `json:"slug"`
				Content string    `json:"content"`
				Date    time.Time `json:"date"`
				Title   string    `json:"title"`
				Img     string    `json:"img"`
			} `json:"_source"`
			Highlight struct {
				Title   []string `json:"title"`
				Content []string `json:"content"`
			} `json:"highlight"`
		} `json:"hits"`
	} `json:"hits"`
}

// indexQueryDSL 语句查询文档
func indexQueryDSL(index, typ string, size, from int, dsl []byte) (*SearchIndexResult, error) {
	rawurl := fmt.Sprintf("%s/%s/%s/_search?size=%d&from=%d", config.Conf.ESHost,
		index, typ, size, from)
	resp, err := httpPost(rawurl, dsl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	result := &SearchIndexResult{}
	err = json.Unmarshal(data, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
