package storage

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
)

var (
	testLocalFile = filepath.Join(os.Getenv("TRAVIS_BUILD_DIR"), "Makefile")
)

func TestFormUploadPutFile(t *testing.T) {
	var putRet PutRet
	ctx := context.TODO()
	putPolicy := PutPolicy{
		Scope:           testBucket,
		DeleteAfterDays: 7,
	}
	upToken := putPolicy.UploadToken(mac)
	testKey := fmt.Sprintf("testPutFileKey_%d", rand.Int())

	err := formUploader.PutFile(ctx, &putRet, upToken, testKey, testLocalFile, nil)
	if err != nil {
		t.Fatalf("FormUploader#PutFile() error, %s", err)
	}
	t.Logf("Key: %s, Hash:%s", putRet.Key, putRet.Hash)
}
