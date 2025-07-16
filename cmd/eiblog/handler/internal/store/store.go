// Package store provides ...
package store

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/eiblog/eiblog/pkg/model"
)

var (
	storeMu sync.RWMutex
	stores  = make(map[string]Driver)
)

// search field
const (
	SearchArticleDraft   = "draft"
	SearchArticleTrash   = "trash"
	SearchArticleTitle   = "title"
	SearchArticleSerieID = "serieid"
)

// SearchArticles 搜索字段
type SearchArticles struct {
	Page   int                    // 第几页/1
	Limit  int                    // 每页大小
	Fields map[string]interface{} // 字段:值
}

// Store 存储后端
type Store interface {
	// LoadInsertBlogger 读取或创建博客
	LoadInsertBlogger(ctx context.Context, blogger *model.Blogger) (bool, error)
	// UpdateBlogger 更新博客
	UpdateBlogger(ctx context.Context, fields map[string]interface{}) error

	// LoadInsertAccount 读取或创建账户
	LoadInsertAccount(ctx context.Context, acct *model.Account) (bool, error)
	// UpdateAccount 更新账户
	UpdateAccount(ctx context.Context, name string, fields map[string]interface{}) error

	// InsertSerie 创建专题
	InsertSerie(ctx context.Context, serie *model.Serie) error
	// RemoveSerie 删除专题
	RemoveSerie(ctx context.Context, id int) error
	// UpdateSerie 更新专题
	UpdateSerie(ctx context.Context, id int, fields map[string]interface{}) error
	// LoadAllSerie 读取所有专题
	LoadAllSerie(ctx context.Context) (model.SortedSeries, error)

	// InsertArticle 创建文章
	InsertArticle(ctx context.Context, article *model.Article, startID int) error
	// RemoveArticle 硬删除文章
	RemoveArticle(ctx context.Context, id int) error
	// CleanArticles 清理回收站文章
	CleanArticles(ctx context.Context, exp time.Time) error
	// UpdateArticle 更新文章
	UpdateArticle(ctx context.Context, id int, fields map[string]interface{}) error
	// LoadArticle 查找文章
	LoadArticle(ctx context.Context, id int) (*model.Article, error)
	// LoadArticleList 查找文章列表
	LoadArticleList(ctx context.Context, search SearchArticles) (model.SortedArticles, int, error)
}

// Driver 存储驱动
type Driver interface {
	// Init 数据库初始化, 建表, 加索引操作等
	Init(name, source string) (Store, error)
}

// Register 注册驱动
func Register(name string, driver Driver) {
	storeMu.Lock()
	defer storeMu.Unlock()
	if driver == nil {
		panic("store: register driver is nil")
	}
	if _, dup := stores[name]; dup {
		panic("store: register called twice for driver " + name)
	}
	stores[name] = driver
}

// Drivers 获取所有
func Drivers() []string {
	storeMu.Lock()
	defer storeMu.Unlock()

	list := make([]string, 0, len(stores))
	for name := range stores {
		list = append(list, name)
	}
	sort.Strings(list)
	return list
}

// NewStore 新建存储
func NewStore(name string, source string) (Store, error) {
	storeMu.RLock()
	driver, ok := stores[name]
	storeMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("store: unknown driver %q (forgotten import?)", name)
	}

	return driver.Init(name, source)
}
