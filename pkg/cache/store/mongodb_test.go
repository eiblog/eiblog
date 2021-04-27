// Package store provides ...
package store

import (
	"context"
	"testing"
	"time"

	"github.com/eiblog/eiblog/pkg/model"
)

var (
	store   Store
	acct    *model.Account
	blogger *model.Blogger
	series  *model.Series
	article *model.Article
)

func init() {
	var err error
	store, err = NewStore("mongodb", "mongodb://127.0.0.1:27017")
	if err != nil {
		panic(err)
	}
	// account
	acct = &model.Account{
		Username:   "deepzz",
		Password:   "deepzz",
		Email:      "deepzz@example.com",
		PhoneN:     "12345678900",
		Address:    "address",
		CreateTime: time.Now(),
	}
	// blogger
	blogger = &model.Blogger{
		BlogName:  "Deepzz",
		SubTitle:  "不抛弃，不放弃",
		BeiAn:     "beian",
		BTitle:    "Deepzz's Blog",
		Copyright: "Copyright",
	}
	// series
	series = &model.Series{
		Slug:       "slug",
		Name:       "series name",
		Desc:       "series desc",
		CreateTime: time.Now(),
	}
	// article
	article = &model.Article{
		Author:  "deepzz",
		Slug:    "slug",
		Title:   "title",
		Count:   0,
		Content: "### count",
		SerieID: 0,
		Tags:    "",
		IsDraft: false,

		UpdateTime: time.Now(),
		CreateTime: time.Now(),
	}
}

func TestLoadInsertAccount(t *testing.T) {
	acct2, err := store.LoadInsertAccount(context.Background(), acct)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(acct2)
	t.Log(acct == acct2)
}

func TestUpdateAccount(t *testing.T) {
	err := store.UpdateAccount(context.Background(), "deepzz", map[string]interface{}{
		"phonn":      "09876543211",
		"loginua":    "chrome",
		"password":   "123456",
		"logintime":  time.Now(),
		"logouttime": time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestLoadInsertBlogger(t *testing.T) {
	blogger2, err := store.LoadInsertBlogger(context.Background(), blogger)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(blogger2)
	t.Log(blogger == blogger2)
}

func TestUpdateBlogger(t *testing.T) {
	err := store.UpdateBlogger(context.Background(), map[string]interface{}{
		"blogname": "blogname",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestInsertSeries(t *testing.T) {
	err := store.InsertSeries(context.Background(), series)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRemoveSeries(t *testing.T) {
	err := store.RemoveSeries(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUpdateSeries(t *testing.T) {
	err := store.UpdateSeries(context.Background(), 2, map[string]interface{}{
		"desc": "update desc",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestLoadAllSeries(t *testing.T) {
	series, err := store.LoadAllSeries(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("load all series: %d", len(series))
}

func TestInsertArticle(t *testing.T) {
	article.ID = 12
	err := store.InsertArticle(context.Background(), article)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRemoveArticle(t *testing.T) {
	err := store.RemoveArticle(context.Background(), 11)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDeleteArticle(t *testing.T) {
	err := store.DeleteArticle(context.Background(), 12)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCleanArticles(t *testing.T) {
	err := store.CleanArticles(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

func TestUpdateArticle(t *testing.T) {
	err := store.UpdateArticle(context.Background(), 13, map[string]interface{}{
		"title":      "new title",
		"updatetime": time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRecoverArticle(t *testing.T) {
	err := store.RecoverArticle(context.Background(), 12)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLoadAllArticle(t *testing.T) {
	articles, err := store.LoadAllArticle(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("load all articles: %d", len(articles))
}

func TestLoadTrashArticles(t *testing.T) {
	articles, err := store.LoadTrashArticles(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("load trash articles: %d", len(articles))
}

func TestLoadDraftArticles(t *testing.T) {
	articles, err := store.LoadDraftArticles(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("load draft articles: %d", len(articles))
}
