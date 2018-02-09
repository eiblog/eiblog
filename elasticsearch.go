package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/eiblog/utils/logd"
)

const (
	INDEX = "eiblog"
	TYPE  = "article"

	ES_FILTER = `"filter":{"bool":{"must":[%s]}}`
	ES_TERM   = `{"term":{"%s":"%s"}}`
	ES_DATE   = `{"range":{"date":{"gte":"%s","lte": "%s","format": "yyyy-MM-dd||yyyy-MM||yyyy"}}}` // 2016-10||/M
)

var (
	ErrUninitializedES = errors.New("uninitialized elasticsearch")

	es *ElasticService
)

// 初始化 Elasticsearch 服务器
func init() {
	_, err := net.LookupIP("elasticsearch")
	if err != nil {
		logd.Info(err)
		return
	}

	es = &ElasticService{url: "http://elasticsearch:9200", c: new(http.Client)}
	initIndex()
}

// 创建索引
func initIndex() {
	mappings := fmt.Sprintf(`{"mappings":{"%s":{"properties":{"content":{"analyzer":"ik_syno","search_analyzer":"ik_syno","term_vector":"with_positions_offsets","type":"string"},"date":{"index":"not_analyzed","type":"date"},"slug":{"type":"string"},"tag":{"index":"not_analyzed","type":"string"},"title":{"analyzer":"ik_syno","search_analyzer":"ik_syno","term_vector":"with_positions_offsets","type":"string"}}}}}`, TYPE)
	err := CreateIndexAndMappings(INDEX, TYPE, []byte(mappings))
	if err != nil {
		logd.Error(err)
	}
}

// 查询
func Elasticsearch(qStr string, size, from int) (*ESSearchResult, error) {
	if es == nil {
		return nil, ErrUninitializedES
	}

	// 分析查询字符串
	reg := regexp.MustCompile(`(tag|slug|date):`)
	indexs := reg.FindAllStringIndex(qStr, -1)
	length := len(indexs)
	var str, kw string
	var filter []string
	if length == 0 { // 全文搜索
		kw = qStr
	}
	// 字段搜索,检出 全文搜索
	for i, index := range indexs {
		if i == length-1 {
			str = qStr[index[0]:]
			if space := strings.Index(str, " "); space != -1 && space < len(str)-1 {
				kw = str[space+1:]
				str = str[:space]
			}
		} else {
			str = strings.TrimSpace(qStr[index[0]:indexs[i+1][0]])
		}
		kv := strings.Split(str, ":")
		switch kv[0] {
		case "slug":
			filter = append(filter, fmt.Sprintf(ES_TERM, kv[0], kv[1]))
		case "tag":
			filter = append(filter, fmt.Sprintf(ES_TERM, kv[0], kv[1]))
		case "date":
			var date string
			switch len(kv[1]) {
			case 4:
				date = fmt.Sprintf(ES_DATE, kv[1], kv[1]+"||/y")
			case 7:
				date = fmt.Sprintf(ES_DATE, kv[1], kv[1]+"||/M")
			case 10:
				date = fmt.Sprintf(ES_DATE, kv[1], kv[1]+"||/d")
			default:
				break
			}
			filter = append(filter, date)
		}
	}
	// 判断是否为空，选择搜索方式
	var dsl string
	if kw != "" {
		dsl = strings.Replace(strings.Replace(`{"highlight":{"fields":{"content":{},"title":{}},"post_tags":["\u003c/b\u003e"],"pre_tags":["\u003cb\u003e"]},"query":{"dis_max":{"queries":[{"match":{"title":{"boost":4,"minimum_should_match":"50%","query":"$1"}}},{"match":{"content":{"boost":4,"minimum_should_match":"75%","query":"$1"}}},{"match":{"tag":{"boost":2,"minimum_should_match":"100%","query":"$1"}}},{"match":{"slug":{"boost":1,"minimum_should_match":"100%","query":"$1"}}}],"tie_breaker":0.3}},$2}`, "$1", kw, -1), "$2", fmt.Sprintf(ES_FILTER, strings.Join(filter, ",")), -1)
	} else {
		dsl = fmt.Sprintf("{"+ES_FILTER+"}", strings.Join(filter, ","))
	}
	docs, err := IndexQueryDSL(INDEX, TYPE, size, from, []byte(dsl))
	if err != nil {
		return nil, err
	}
	return docs, nil
}

// 添加或更新索引
func ElasticIndex(artc *Article) error {
	if es == nil {
		return ErrUninitializedES
	}

	img := PickFirstImage(artc.Content)
	mapping := map[string]interface{}{
		"title":   artc.Title,
		"content": IgnoreHtmlTag(artc.Content),
		"slug":    artc.Slug,
		"tag":     artc.Tags,
		"img":     img,
		"date":    artc.CreateTime,
	}
	b, _ := json.Marshal(mapping)
	return IndexOrUpdateDocument(INDEX, TYPE, artc.ID, b)
}

// 删除索引
func ElasticDelIndex(ids []int32) error {
	if es == nil {
		return ErrUninitializedES
	}

	var target []string
	for _, id := range ids {
		target = append(target, fmt.Sprint(id))
	}
	return DeleteDocument(INDEX, TYPE, target)
}

///////////////////////////// Elasticsearch api /////////////////////////////
type ElasticService struct {
	c   *http.Client
	url string
}

type IndicesCreateResult struct {
	Acknowledged bool `json:"acknowledged"`
}

// 返回 url
func (s *ElasticService) ParseURL(format string, params ...interface{}) string {
	return fmt.Sprintf(s.url+format, params...)
}

// Elastic 相关操作请求
func (s *ElasticService) Do(req *http.Request) (interface{}, error) {
	resp, err := s.c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	switch req.Method {
	case "POST":
		fallthrough
	case "DELETE":
		fallthrough
	case "PUT":
		fallthrough
	case "GET":
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return b, nil
	case "HEAD":
		return resp.StatusCode, nil
	}
	return nil, errors.New("unknown methods")
}

func CreateIndexAndMappings(index, typ string, mappings []byte) (err error) {
	req, err := http.NewRequest("HEAD", es.ParseURL("/%s/%s", index, typ), nil)
	code, err := es.Do(req)
	if err != nil {
		return err
	}
	if code.(int) == http.StatusOK {
		return nil
	}
	req, err = http.NewRequest("PUT", es.ParseURL("/%s", index), bytes.NewReader(mappings))
	if err != nil {
		return err
	}
	data, err := es.Do(req)
	if err != nil {
		return err
	}
	var rst IndicesCreateResult
	err = json.Unmarshal(data.([]byte), &rst)
	if err != nil {
		return err
	}
	if !rst.Acknowledged {
		return errors.New(string(data.([]byte)))
	}
	return nil
}

// 创建或更新索引
func IndexOrUpdateDocument(index, typ string, id int32, doc []byte) (err error) {
	req, err := http.NewRequest("PUT", es.ParseURL("/%s/%s/%d", index, typ, id), bytes.NewReader(doc))
	if err != nil {
		return err
	}
	data, err := es.Do(req)
	if err != nil {
		return err
	}
	logd.Debug(string(data.([]byte)))
	return nil
}

type ESDeleteDocument struct {
	Index string `json:"_index"`
	Type  string `json:"_type"`
	ID    string `json:"_id"`
}

type ESDeleteResult struct {
	Errors bool `json:"errors"`
	Iterms []map[string]struct {
		Error string `json:"error"`
	} `json:"iterms"`
}

// 删除文档
func DeleteDocument(index, typ string, ids []string) error {
	var buff bytes.Buffer
	for _, id := range ids {
		dd := &ESDeleteDocument{Index: index, Type: typ, ID: id}
		m := map[string]*ESDeleteDocument{"delete": dd}
		b, _ := json.Marshal(m)
		buff.Write(b)
		buff.WriteByte('\n')
	}
	req, err := http.NewRequest("POST", es.ParseURL("/_bulk"), bytes.NewReader(buff.Bytes()))
	if err != nil {
		return err
	}
	data, err := es.Do(req)
	if err != nil {
		return err
	}
	var result ESDeleteResult
	err = json.Unmarshal(data.([]byte), &result)
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

// 查询结果
type ESSearchResult struct {
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

// DSL 语句查询文档
func IndexQueryDSL(index, typ string, size, from int, dsl []byte) (*ESSearchResult, error) {
	req, err := http.NewRequest("POST", es.ParseURL("/%s/%s/_search?size=%d&from=%d", index, typ, size, from), bytes.NewReader(dsl))
	if err != nil {
		return nil, err
	}
	data, err := es.Do(req)
	if err != nil {
		return nil, err
	}
	result := &ESSearchResult{}
	err = json.Unmarshal(data.([]byte), result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
