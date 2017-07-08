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
	testStr := []string{
		time.Now().Format("2006-01-02T15:04:05"),
	}

	expectStr := []string{
		JUST_NOW,
	}

	for i, v := range testStr {
		assert.Equal(t, expectStr[i], ConvertStr(v))
	}
}
