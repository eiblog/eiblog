// Package store provides ...
package store

import (
	"context"
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
	collectionSerie   = "serie"

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
	db.Database(mongoDBName).Collection(collectionSerie).
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

	filter := bson.M{"username": config.Conf.EiBlogApp.Account.Username}
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

// InsertSerie 创建专题
func (db *mongodb) InsertSerie(ctx context.Context, serie *model.Serie) error {
	collection := db.Database(mongoDBName).Collection(collectionSerie)

	serie.ID = db.nextValue(ctx, counterNameSerie)
	_, err := collection.InsertOne(ctx, serie)
	return err
}

// RemoveSerie 删除专题
func (db *mongodb) RemoveSerie(ctx context.Context, id int) error {
	collection := db.Database(mongoDBName).Collection(collectionSerie)

	filter := bson.M{"id": id}
	_, err := collection.DeleteOne(ctx, filter)
	return err
}

// UpdateSerie 更新专题
func (db *mongodb) UpdateSerie(ctx context.Context, id int,
	fields map[string]interface{}) error {

	collection := db.Database(mongoDBName).Collection(collectionSerie)

	filter := bson.M{"id": id}
	params := bson.M{}
	for k, v := range fields {
		params[k] = v
	}
	update := bson.M{"$set": params}
	_, err := collection.UpdateOne(ctx, filter, update)
	return err
}

// LoadAllSerie 查询所有专题
func (db *mongodb) LoadAllSerie(ctx context.Context) (model.SortedSeries, error) {
	collection := db.Database(mongoDBName).Collection(collectionSerie)

	filter := bson.M{}
	cur, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var series model.SortedSeries
	for cur.Next(ctx) {
		obj := model.Serie{}
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
		if id < config.Conf.EiBlogApp.General.StartID {
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

// CleanArticles 清理回收站文章
func (db *mongodb) CleanArticles(ctx context.Context) error {
	collection := db.Database(mongoDBName).Collection(collectionArticle)

	exp := time.Now().Add(time.Duration(config.Conf.EiBlogApp.General.Trash) * time.Hour)
	filter := bson.M{"deleted_at": bson.M{"$gt": time.Time{}, "$lt": exp}}
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
	_, err := collection.UpdateOne(ctx, filter, update)
	return err
}

// LoadArticle 查找文章
func (db *mongodb) LoadArticle(ctx context.Context, id int) (*model.Article, error) {
	collection := db.Database(mongoDBName).Collection(collectionArticle)

	filter := bson.M{"id": id}
	result := collection.FindOne(ctx, filter)
	err := result.Err()
	if err != nil {
		return nil, err
	}
	article := &model.Article{}
	err = result.Decode(article)
	return article, err
}

// LoadArticleList 获取文章列表
func (db *mongodb) LoadArticleList(ctx context.Context, search SearchArticles) (
	model.SortedArticles, int, error) {
	collection := db.Database(mongoDBName).Collection(collectionArticle)

	filter := bson.M{}
	for k, v := range search.Fields {
		switch k {
		case SearchArticleDraft:
			if ok := v.(bool); ok {
				filter["is_draft"] = true
			} else {
				filter["is_draft"] = false
				filter["deleted_at"] = bson.M{"$eq": time.Time{}}
			}
		case SearchArticleTitle:
			filter["title"] = bson.M{
				"$regex":   v.(string),
				"$options": "$i",
			}
		case SearchArticleSerieID:
			filter["serie_id"] = v.(int)
		case SearchArticleTrash:
			filter["deleted_at"] = bson.M{"$ne": time.Time{}}
		}
	}
	// search count
	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	opts := options.Find().SetLimit(int64(search.Limit)).
		SetSkip(int64((search.Page - 1) * search.Limit)).
		SetSort(bson.M{"created_at": -1})
	cur, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	var articles model.SortedArticles
	for cur.Next(ctx) {
		obj := model.Article{}
		err = cur.Decode(&obj)
		if err != nil {
			return nil, 0, err
		}
		articles = append(articles, &obj)
	}
	sort.Sort(articles)
	return articles, int(count), nil
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
