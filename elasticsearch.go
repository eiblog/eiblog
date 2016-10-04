package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/eiblog/eiblog/setting"
	"github.com/eiblog/utils/logd"
)

const (
	INDEX = "eiblog"
	TYPE  = "article"
)

var es *ElasticService

func init() {
	es = &ElasticService{url: setting.Conf.SearchURL, c: new(http.Client)}
	initIndex()
}

func initIndex() {
	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			TYPE: map[string]interface{}{
				"properties": map[string]interface{}{
					"title": map[string]string{
						"type":            "string",
						"term_vector":     "with_positions_offsets",
						"analyzer":        "ik_syno",
						"search_analyzer": "ik_syno",
					},
					"content": map[string]string{
						"type":            "string",
						"term_vector":     "with_positions_offsets",
						"analyzer":        "ik_syno",
						"search_analyzer": "ik_syno",
					},
					"slug": map[string]string{
						"type": "string",
					},
					"tags": map[string]string{
						"type":  "string",
						"index": "not_analyzed",
					},
					"create_time": map[string]string{
						"type":  "date",
						"index": "not_analyzed",
					},
				},
			},
		},
	}
	b, _ := json.Marshal(mapping)
	err := CreateIndexAndMappings(INDEX, TYPE, b)
	if err != nil {
		logd.Fatal(err)
	}
}

func Elasticsearch(kw string, size, from int) *ESSearchResult {
	dsl := map[string]interface{}{
		"query": map[string]interface{}{
			"dis_max": map[string]interface{}{
				"queries": []map[string]interface{}{
					map[string]interface{}{
						"match": map[string]interface{}{
							"title": map[string]interface{}{
								"query":                kw,
								"minimum_should_match": "50%",
								"boost":                4,
							},
						},
					},
					map[string]interface{}{
						"match": map[string]interface{}{
							"content": map[string]interface{}{
								"query":                kw,
								"minimum_should_match": "75%",
								"boost":                4,
							},
						},
					},
					map[string]interface{}{
						"match": map[string]interface{}{
							"tags": map[string]interface{}{
								"query":                kw,
								"minimum_should_match": "100%",
								"boost":                2,
							},
						},
					},
					map[string]interface{}{
						"match": map[string]interface{}{
							"slug": map[string]interface{}{
								"query":                kw,
								"minimum_should_match": "100%",
								"boost":                1,
							},
						},
					},
				},
				"tie_breaker": 0.3,
			},
		},
		"highlight": map[string]interface{}{
			"pre_tags":  []string{"<b>"},
			"post_tags": []string{"</b>"},
			"fields": map[string]interface{}{
				"title":   map[string]string{},
				"content": map[string]string{
				// "fragment_size":       150,
				// "number_of_fragments": "3",
				},
			},
		},
	}
	b, _ := json.Marshal(dsl)
	docs, err := IndexQueryDSL(INDEX, TYPE, size, from, b)
	if err != nil {
		logd.Error(err)
		return nil
	}
	return docs
}

func ElasticsearchSimple(q string, size, from int) *ESSearchResult {
	docs, err := IndexQuerySimple(INDEX, TYPE, size, from, q)
	if err != nil {
		logd.Error(err)
		return nil
	}
	return docs
}

func ElasticIndex(artc *Article) error {
	mapping := map[string]interface{}{
		"title":       artc.Title,
		"content":     IgnoreHtmlTag(artc.Content),
		"slug":        artc.Slug,
		"tags":        artc.Tags,
		"create_time": artc.CreateTime,
	}
	b, _ := json.Marshal(mapping)
	return IndexOrUpdateDocument(INDEX, TYPE, artc.ID, b)
}

func ElasticDelIndex(ids []int32) error {
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

func (s *ElasticService) ParseURL(format string, params ...interface{}) string {
	return fmt.Sprintf(s.url+format, params...)
}

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

	default:
		return nil, errors.New("unknown methods")
	}
	return nil, nil
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

func IndexOrUpdateDocument(index, typ string, id int32, doc []byte) (err error) {
	req, err := http.NewRequest("PUT", es.ParseURL("/%s/%s/%d", index, typ, id), bytes.NewReader(doc))
	if err != nil {
		return err
	}
	data, err := es.Do(req)
	if err != nil {
		return err
	}
	fmt.Println(string(data.([]byte)))
	return nil
}

type ESDeleteDocument struct {
	_Index string `json:"_index"`
	_Type  string `json:"_type"`
	_ID    string `json:"_id"`
}

type ESDeleteResult struct {
	Errors bool `json:"errors"`
	Iterms []map[string]struct {
		Error string `json:"error"`
	} `json:"iterms"`
}

func DeleteDocument(index, typ string, ids []string) error {
	var buff bytes.Buffer
	for _, id := range ids {
		dd := &ESDeleteDocument{_Index: index, _Type: typ, _ID: id}
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

type ESSearchResult struct {
	Took float32 `json:"took"`
	Hits struct {
		Total int `json:"total"`
		Hits  []struct {
			ID     string `json:"_id"`
			Source struct {
				Slug       string    `json:"slug"`
				Content    string    `json:"content"`
				CreateTime time.Time `json:"create_time"`
				Title      string    `json:"title"`
			} `json:"_source"`
			Highlight struct {
				Title   []string `json:"title"`
				Content []string `json:"content"`
			} `json:"highlight"`
		} `json:"hits"`
	} `json:"hits"`
}

func IndexQuerySimple(index, typ string, size, from int, q string) (*ESSearchResult, error) {
	req, err := http.NewRequest("GET", es.ParseURL("/%s/%s/_search?size=%d&from=%d&q=%s", index, typ, size, from, q), nil)
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
