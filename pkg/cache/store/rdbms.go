// Package store provides ...
package store

import (
	"context"
	"time"

	"github.com/eiblog/eiblog/pkg/model"

	"gorm.io/driver/clickhouse"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

type rdbms struct {
	*gorm.DB
}

// Init 数据库初始化, 建表, 加索引操作等
// name 应该为具体的关系数据库驱动名
func (db *rdbms) Init(name, source string) (Store, error) {
	var (
		gormDB *gorm.DB
		err    error
	)
	switch name {
	case "mysql":
		// https://github.com/go-sql-driver/mysql
		gormDB, err = gorm.Open(mysql.Open(source), &gorm.Config{})
	case "postgres":
		// https://github.com/go-gorm/postgres
		gormDB, err = gorm.Open(postgres.Open(source), &gorm.Config{})
	case "sqlite":
		// github.com/mattn/go-sqlite3
		gormDB, err = gorm.Open(sqlite.Open(source), &gorm.Config{})
	case "sqlserver":
		// github.com/denisenkom/go-mssqldb
		gormDB, err = gorm.Open(sqlserver.Open(source), &gorm.Config{})
	case "clickhouse":
		gormDB, err = gorm.Open(clickhouse.Open(source), &gorm.Config{})
	}
	if err != nil {
		return nil, err
	}
	// auto migrate
	gormDB.AutoMigrate(
		&model.Account{},
		&model.Blogger{},
		&model.Article{},
		&model.Serie{},
	)
	db.DB = gormDB
	return db, nil
}

// LoadInsertBlogger 读取或创建博客
func (db *rdbms) LoadInsertBlogger(ctx context.Context, blogger *model.Blogger) (bool, error) {
	result := db.FirstOrCreate(blogger)
	return result.RowsAffected > 0, result.Error
}

// UpdateBlogger 更新博客
func (db *rdbms) UpdateBlogger(ctx context.Context, fields map[string]interface{}) error {
	return db.Model(model.Blogger{}).Updates(fields).Error
}

// LoadInsertAccount 读取或创建账户
func (db *rdbms) LoadInsertAccount(ctx context.Context, acct *model.Account) (bool, error) {
	result := db.Where("username=?", acct.Username).FirstOrCreate(acct)
	return result.RowsAffected > 0, result.Error
}

// UpdateAccount 更新账户
func (db *rdbms) UpdateAccount(ctx context.Context, name string, fields map[string]interface{}) error {
	return db.Model(model.Account{}).Where("username=?", name).Updates(fields).Error
}

// InsertSerie 创建专题
func (db *rdbms) InsertSerie(ctx context.Context, serie *model.Serie) error {
	return db.Create(serie).Error
}

// RemoveSerie 删除专题
func (db *rdbms) RemoveSerie(ctx context.Context, id int) error {
	return db.Where("id=?", id).Delete(model.Serie{}).Error
}

// UpdateSerie 更新专题
func (db *rdbms) UpdateSerie(ctx context.Context, id int, fields map[string]interface{}) error {
	return db.Model(model.Serie{}).Where("id=?", id).Updates(fields).Error
}

// LoadAllSerie 读取所有专题
func (db *rdbms) LoadAllSerie(ctx context.Context) (model.SortedSeries, error) {
	var series model.SortedSeries
	err := db.Order("id DESC").Find(&series).Error
	return series, err
}

// InsertArticle 创建文章
func (db *rdbms) InsertArticle(ctx context.Context, article *model.Article, startID int) error {
	// TODO id stting
	return db.Create(article).Error
}

// RemoveArticle 硬删除文章
func (db *rdbms) RemoveArticle(ctx context.Context, id int) error {
	return db.Where("id=?", id).Delete(model.Article{}).Error
}

// CleanArticles 清理回收站文章
func (db *rdbms) CleanArticles(ctx context.Context, exp time.Time) error {
	return db.Where("deleted_at BETWEEN ? AND ?", time.Time{}, exp).Delete(model.Article{}).Error
}

// UpdateArticle 更新文章
func (db *rdbms) UpdateArticle(ctx context.Context, id int, fields map[string]interface{}) error {
	return db.Model(model.Article{}).Where("id=?", id).Updates(fields).Error
}

// LoadArticle 查找文章
func (db *rdbms) LoadArticle(ctx context.Context, id int) (*model.Article, error) {
	article := &model.Article{}
	err := db.Where("id=?", id).First(article).Error
	return article, err
}

// LoadArticleList 查找文章列表
func (db *rdbms) LoadArticleList(ctx context.Context, search SearchArticles) (model.SortedArticles, int, error) {
	gormDB := db.Model(model.Article{})
	for k, v := range search.Fields {
		switch k {
		case SearchArticleDraft:
			if ok := v.(bool); ok {
				gormDB = gormDB.Where("is_draft=?", true)
			} else {
				gormDB = gormDB.Where("is_draft=? AND deleted_at=?", false, time.Time{})
			}
		case SearchArticleTitle:
			gormDB = gormDB.Where("title LIKE ?", "%"+v.(string)+"%")
		case SearchArticleSerieID:
			gormDB = gormDB.Where("serie_id=?", v.(int))
		case SearchArticleTrash:
			gormDB = gormDB.Where("deleted_at!=?", time.Time{})
		}
	}
	// search count
	var count int64
	err := gormDB.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}
	var articles model.SortedArticles
	err = db.Limit(search.Limit).
		Offset((search.Page - 1) * search.Limit).
		Order("created_at DESC").Find(&articles).Error
	return articles, int(count), err
}

// register store
func init() {
	Register("mysql", &rdbms{})
	Register("postgres", &rdbms{})
	Register("sqlite", &rdbms{})
	Register("sqlserver", &rdbms{})
	Register("clickhouse", &rdbms{})
}
