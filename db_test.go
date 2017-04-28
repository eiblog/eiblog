package main

import (
	"html/template"
	"testing"
)

func TestPageListBack(t *testing.T) {
	_, artcs := PageListBack(0, "", false, false, 1, 20)
	for _, artc := range artcs {
		t.Log(*artc)
	}
	t.Log("------------------------------------------------------------")
	_, artcs = PageListBack(0, "", false, false, 2, 10)
	for _, artc := range artcs {
		t.Log(*artc)
	}
	t.Log("------------------------------------------------------------")
	_, artcs = PageListBack(3, "", false, false, 1, 20)
	for _, artc := range artcs {
		t.Log(*artc)
	}
	t.Log("------------------------------------------------------------")
	_, artcs = PageListBack(3, "19", false, false, 1, 20)
	for _, artc := range artcs {
		t.Log(*artc)
	}
	t.Log("------------------------------------------------------------")
	_, artcs = PageListBack(0, "", false, true, 1, 20)
	for _, artc := range artcs {
		t.Log(*artc)
	}
	t.Log("------------------------------------------------------------")
	_, artcs = PageListBack(0, "", true, false, 1, 20)
	for _, artc := range artcs {
		t.Log(*artc)
	}
}

func TestAddSerie(t *testing.T) {
	err := AddSerie("测试", "nothing", "这里是描述")
	if err != nil {
		t.Error(err)
	}
}

func TestRenderPage(t *testing.T) {
	data := []byte(`<ul class="links ssl">
<li><a href="https://yryz.net/">一人游走</a><span class="date">「不错的小伙子」</span></li>
<li><a href="https://hsulei.com/">Leo同学</a><span class="date">「小伙子，该干活了」</span></li>
<li><a href="https://razeencheng.com/">razeen同学</a><span class="date">「Stay hungry. Stay foolish.」</span></li>
</ul>

<ul class="links">
<li><a href="http://blog.mirreal.net/">Mirreal Ellison</a><span class="date">「kissing the fire」</span></li>
</ul>`)

	t.Log(IgnoreHtmlTag(string(data)))
	data = renderPage(data)

	t.Log(template.HTML(string(data)))
}
