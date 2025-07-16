package db

import (
	"context"
	"errors"
	"net/url"
	"time"

	"github.com/eiblog/eiblog/pkg/config"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// NewMDB new mongodb
// https://docs.mongodb.com/manual/reference/connection-string/
// mongodb://db0.example.com,db1.example.com,db2.example.com/dbname?replicaSet=myRepl&w=majority&wtimeoutMS=5000
func NewMDB(opts config.Database) (*mongo.Database, error) {
	if opts.Driver != "mongodb" {
		return nil, errors.New("db: driver must be mongodb, but " + opts.Driver)
	}
	u, err := url.Parse(opts.Source)
	if err != nil {
		return nil, errors.New("db: " + err.Error())
	}
	database := u.Path
	if database == "" {
		return nil, errors.New("db: please specify a database")
	}
	database = database[1:]
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	// remove database
	u.Path = ""
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(u.String()))
	if err != nil {
		return nil, err
	}
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}
	return client.Database(database), nil
}
