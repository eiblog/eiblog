// Package store provides ...
package store

import (
	"fmt"
	"sort"
	"sync"

	"github.com/eiblog/eiblog/v2/pkg/model"
)

var (
	storeMu sync.RWMutex
	stores  = make(map[string]Driver)
)

// Store 存储后端
type Store interface {
	LoadOrCreateAccount(acct *model.Account) (*model.Account, error)
	LoadOrCreateBlogger(blogger *model.Blogger) (*model.Blogger, error)
	LoadAllArticles() ([]*model.Article, error)

	UpdateAccount(name string, fields map[string]interface{}) error
	UpdateBlogger(fields map[string]interface{}) error
	UpdateArticle(article *model.Article) error
	CleanArticles() error
}

// Driver 存储驱动
type Driver interface {
	Init(source string) (Store, error)
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

	return driver.Init(source)
}
