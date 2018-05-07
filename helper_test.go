// Package main provides ...
package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReadDir(t *testing.T) {
	files := ReadDir("setting", func(name string) bool { return false })
	assert.Len(t, files, 2)
}

func TestIgnoreHtmlTag(t *testing.T) {
	testStr := []string{
		"<script>hello</script>",
	}

	expectStr := []string{
		"hello",
	}

	for i, v := range testStr {
		assert.Equal(t, expectStr[i], IgnoreHtmlTag(v))
	}
}

func TestPickFirstImage(t *testing.T) {
	testStr := []string{
		`<img width="480" height="310" alt="acme_aliyun_1" src="https://st.deepzz.com/blog/img/acme_aliyun_1.jpg">`,
		`<img width="480" height="310" alt="acme_aliyun_1" data-src="https://st.deepzz.com/acme_aliyun_1.jpg"><img width="480" height="310" alt="acme_aliyun_1" src="https://st.deepzz.com/blog/img/acme_aliyun_1.jpg">`,
	}

	expectStr := []string{
		"",
		"https://st.deepzz.com/acme_aliyun_1.jpg",
	}

	for i, v := range testStr {
		assert.Equal(t, expectStr[i], PickFirstImage(v))
	}
}

func TestCovertStr(t *testing.T) {
	now := time.Now().UTC()
	testStr := []string{
		now.Format("2006-01-02T15:04:05"),
		now.Add(-time.Second * 20).Format("2006-01-02T15:04:05"),
		now.Add(-time.Minute).Format("2006-01-02T15:04:05"),
		now.Add(-time.Minute * 2).Format("2006-01-02T15:04:05"),
		now.Add(-time.Minute * 20).Format("2006-01-02T15:04:05"),
		now.Add(-time.Hour).Format("2006-01-02T15:04:05"),
		now.Add(-time.Hour * 2).Format("2006-01-02T15:04:05"),
		now.Add(-time.Hour * 24).Format("2006-01-02T15:04:05"),
	}

	time.Sleep(time.Second)
	t.Log(now.Format("2006-01-02T15:04:05"))
	for _, v := range testStr {
		t.Log(v, ConvertStr(v))
	}
}
