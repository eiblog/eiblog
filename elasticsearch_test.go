package main

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestCreateIndexAndMappings(t *testing.T) {
	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"article": map[string]interface{}{
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
					"update_time": map[string]string{
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
		t.Error(err)
	}
}

func TestIndexDocument(t *testing.T) {
	mapping := map[string]interface{}{
		"title": "简单到不知道为什么",
		"content": `最近有很多朋友邮件或者留言询问本博客服务端配置相关问题，基本都是关于 HTTPS 和 HTTP/2 的，其实我的 Nginx 配置在之前的文章中多次提到过，不过都比较分散。为了方便大家参考，本文贴出完整配置。本文内容会随时调整或更新，请大家不要把本文内容全文转载到第三方平台，以免给他人造成困扰或误导。另外限于篇幅，本文不会对配置做过多说明，如有疑问或不同意见，欢迎留言指出。
`,
		"slug":        "vim3",
		"tags":        []string{"js", "javascript", "test"},
		"update_time": "2015-12-15T13:05:55Z",
	}
	b, _ := json.Marshal(mapping)
	err := IndexOrUpdateDocument(INDEX, TYPE, int32(11), b)
	if err != nil {
		t.Error(err)
	}
}

func TestIndexQueryDSL(t *testing.T) {
	kw := "实现访问限制"
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
	fmt.Println(string(b))
	_, err := IndexQueryDSL(INDEX, TYPE, 10, 1, b)
	if err != nil {
		t.Error(err)
	}
}
