// Package store provides ...
package store

var store Store

func init() {
	var err error
	store, err = NewStore("mongodb", "mongodb://127.0.0.1:27017")
	if err != nil {
		panic(err)
	}
}
