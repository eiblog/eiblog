// Package store provides ...
package store

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/eiblog/eiblog/pkg/config"
	"github.com/eiblog/eiblog/pkg/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// driver: mongodb
// source: mongodb://localhost:27017

const (
	mongoDBName       = "eiblog"
	collectionAccount = "account"
	collectionArticle = "article"
	collectionBlogger = "blogger"
	collectionCounter = "counter"
	collectionSeries  = "series"

	counterNameSerie   = "serie"
	counterNameArticle = "article"
)

type mongodb struct {
	*mongo.Client
}

// Init init mongodb client
func (db *mongodb) Init(source string) (Store, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Client().ApplyURI(source)
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}
	db.Client = client
	// create index
	indexModel := mongo.IndexModel{
		Keys:    bson.D{bson.E{Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true).SetSparse(true),
	}
	db.Database(mongoDBName).Collection(collectionAccount).
		Indexes().
		CreateOne(context.Background(), indexModel)
	indexModel = mongo.IndexModel{
		Keys:    bson.D{bson.E{Key: "slug", Value: 1}},
		Options: options.Index().SetUnique(true).SetSparse(true),
	}
	db.Database(mongoDBName).Collection(collectionArticle).
		Indexes().
		CreateOne(context.Background(), indexModel)
	indexModel = mongo.IndexModel{
		Keys:    bson.D{bson.E{Key: "slug", Value: 1}},
		Options: options.Index().SetUnique(true).SetSparse(true),
	}
	db.Database(mongoDBName).Collection(collectionSeries).
		Indexes().
		CreateOne(context.Background(), indexModel)
	return db, nil
}

// LoadInsertBlogger 读取或创建博客
func (db *mongodb) LoadInsertBlogger(ctx context.Context,
	blogger *model.Blogger) (created bool, err error) {

	collection := db.Database(mongoDBName).Collection(collectionBlogger)

	filter := bson.M{}
	result := collection.FindOne(ctx, filter)
	err = result.Err()
	if err != nil {
		if err != mongo.ErrNoDocuments {
			return
		}
		_, err = collection.InsertOne(ctx, blogger)
		created = true
	} else {
		err = result.Decode(blogger)
	}
	return
}

// UpdateBlogger 更新博客
func (db *mongodb) UpdateBlogger(ctx context.Context,
	fields map[string]interface{}) error {

	collection := db.Database(mongoDBName).Collection(collectionBlogger)

	filter := bson.M{}
	params := bson.M{}
	for k, v := range fields {
		params[k] = v
	}
	update := bson.M{"$set": params}
	_, err := collection.UpdateOne(ctx, filter, update)
	return err
}

// LoadInsertAccount 读取或创建账户
func (db *mongodb) LoadInsertAccount(ctx context.Context,
	acct *model.Account) (created bool, err error) {

	collection := db.Database(mongoDBName).Collection(collectionAccount)

	filter := bson.M{"username": config.Conf.BlogApp.Account.Username}
	result := collection.FindOne(ctx, filter)
	err = result.Err()
	if err != nil {
		if err != mongo.ErrNoDocuments {
			return
		}
		_, err = collection.InsertOne(ctx, acct)
		created = true
	} else {
		err = result.Decode(acct)
	}
	return
}

// UpdateAccount 更新账户
func (db *mongodb) UpdateAccount(ctx context.Context, name string,
	fields map[string]interface{}) error {

	collection := db.Database(mongoDBName).Collection(collectionAccount)

	filter := bson.M{"username": name}
	params := bson.M{}
	for k, v := range fields {
		params[k] = v
	}
	update := bson.M{"$set": params}
	_, err := collection.UpdateOne(ctx, filter, update)
	return err
}

// InsertSeries 创建专题
func (db *mongodb) InsertSeries(ctx context.Context, series *model.Series) error {
	collection := db.Database(mongoDBName).Collection(collectionSeries)

	series.ID = db.nextValue(ctx, counterNameSerie)
	_, err := collection.InsertOne(ctx, series)
	return err
}

// RemoveSeries 删除专题
func (db *mongodb) RemoveSeries(ctx context.Context, id int) error {
	collection := db.Database(mongoDBName).Collection(collectionSeries)

	filter := bson.M{"id": id}
	_, err := collection.DeleteOne(ctx, filter)
	return err
}

// UpdateSeries 更新专题
func (db *mongodb) UpdateSeries(ctx context.Context, id int,
	fields map[string]interface{}) error {

	collection := db.Database(mongoDBName).Collection(collectionSeries)

	filter := bson.M{"id": id}
	params := bson.M{}
	for k, v := range fields {
		params[k] = v
	}
	update := bson.M{"$set": params}
	_, err := collection.UpdateOne(ctx, filter, update)
	return err
}

// LoadAllSeries 查询所有专题
func (db *mongodb) LoadAllSeries(ctx context.Context) (model.SortedSeries, error) {
	collection := db.Database(mongoDBName).Collection(collectionSeries)

	filter := bson.M{}
	cur, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var series model.SortedSeries
	for cur.Next(ctx) {
		obj := model.Series{}
		err = cur.Decode(&obj)
		if err != nil {
			return nil, err
		}
		series = append(series, &obj)
	}
	sort.Sort(series)
	return series, nil
}

// InsertArticle 创建文章
func (db *mongodb) InsertArticle(ctx context.Context, article *model.Article) error {
	// 可手动分配ID或者分配ID, 占位至起始id
	for article.ID == 0 {
		id := db.nextValue(ctx, counterNameArticle)
		if id < config.Conf.BlogApp.General.StartID {
			continue
		} else {
			article.ID = id
		}
	}

	collection := db.Database(mongoDBName).Collection(collectionArticle)
	_, err := collection.InsertOne(ctx, article)
	return err
}

// RemoveArticle 硬删除文章
func (db *mongodb) RemoveArticle(ctx context.Context, id int) error {
	collection := db.Database(mongoDBName).Collection(collectionArticle)

	filter := bson.M{"id": id}
	_, err := collection.DeleteOne(ctx, filter)
	return err
}

// DeleteArticle 软删除文章,放入回收箱
func (db *mongodb) DeleteArticle(ctx context.Context, id int) error {
	collection := db.Database(mongoDBName).Collection(collectionArticle)

	filter := bson.M{"id": id}
	update := bson.M{"$set": bson.M{"deletetime": time.Now()}}
	_, err := collection.UpdateOne(ctx, filter, update)
	return err
}

// CleanArticles 清理回收站文章
func (db *mongodb) CleanArticles(ctx context.Context) error {
	collection := db.Database(mongoDBName).Collection(collectionArticle)

	exp := time.Now().Add(time.Duration(config.Conf.BlogApp.General.Trash) * time.Hour)
	filter := bson.M{"deletetime": bson.M{"$gt": time.Time{}, "$lt": exp}}
	_, err := collection.DeleteMany(ctx, filter)
	return err
}

// UpdateArticle 更新文章
func (db *mongodb) UpdateArticle(ctx context.Context, id int,
	fields map[string]interface{}) error {

	collection := db.Database(mongoDBName).Collection(collectionArticle)

	filter := bson.M{"id": id}
	params := bson.M{}
	for k, v := range fields {
		params[k] = v
	}
	update := bson.M{"$set": params}
	fmt.Println(update)
	_, err := collection.UpdateOne(ctx, filter, update)
	return err
}

// RecoverArticle 恢复文章到草稿
func (db *mongodb) RecoverArticle(ctx context.Context, id int) error {
	collection := db.Database(mongoDBName).Collection(collectionArticle)

	filter := bson.M{"id": id}
	update := bson.M{"$set": bson.M{"deletetime": time.Time{}, "isdraft": true}}
	_, err := collection.UpdateOne(ctx, filter, update)
	return err
}

// LoadAllArticle 读取所有文章
func (db *mongodb) LoadAllArticle(ctx context.Context) (model.SortedArticles, error) {
	collection := db.Database(mongoDBName).Collection(collectionArticle)

	filter := bson.M{"isdraft": false, "deletetime": bson.M{"$eq": time.Time{}}}
	cur, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var articles model.SortedArticles
	for cur.Next(ctx) {
		obj := model.Article{}
		err = cur.Decode(&obj)
		if err != nil {
			return nil, err
		}
		articles = append(articles, &obj)
	}
	sort.Sort(articles)
	return articles, nil
}

// LoadTrashArticles 读取回收箱
func (db *mongodb) LoadTrashArticles(ctx context.Context) (model.SortedArticles, error) {
	collection := db.Database(mongoDBName).Collection(collectionArticle)

	filter := bson.M{"deletetime": bson.M{"$ne": time.Time{}}}
	cur, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var articles model.SortedArticles
	for cur.Next(ctx) {
		obj := model.Article{}
		err = cur.Decode(&obj)
		if err != nil {
			return nil, err
		}
		articles = append(articles, &obj)
	}
	sort.Sort(articles)
	return articles, nil
}

// LoadDraftArticles 读取草稿箱
func (db *mongodb) LoadDraftArticles(ctx context.Context) (model.SortedArticles, error) {
	collection := db.Database(mongoDBName).Collection(collectionArticle)

	filter := bson.M{"isdraft": true}
	cur, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var articles model.SortedArticles
	for cur.Next(ctx) {
		obj := model.Article{}
		err = cur.Decode(&obj)
		if err != nil {
			return nil, err
		}
		articles = append(articles, &obj)
	}
	sort.Sort(articles)
	return articles, nil
}

// counter counter
type counter struct {
	Name    string
	NextVal int
}

// nextValue counter value
func (db *mongodb) nextValue(ctx context.Context, name string) int {
	collection := db.Database(mongoDBName).Collection(collectionCounter)

	opts := options.FindOneAndUpdate().SetUpsert(true).
		SetReturnDocument(options.After)
	filter := bson.M{"name": name}
	update := bson.M{"$inc": bson.M{"nextval": 1}}

	next := counter{}
	err := collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&next)
	if err != nil {
		return -1
	}
	return next.NextVal
}

// register store
func init() {
	Register("mongodb", &mongodb{})
}
