package kodo

import (
	"math/rand"
	"strconv"
	"testing"
	"time"
)

var (
	bkey     = "abatch"
	bnewkey1 = "abatch/newkey1"
	bnewkey2 = "abatch/newkey2"
)

func init() {

	if skipTest() {
		return
	}

	rand.Seed(time.Now().UnixNano())
	bkey += strconv.Itoa(rand.Int())
	bnewkey1 += strconv.Itoa(rand.Int())
	bnewkey2 += strconv.Itoa(rand.Int())
	// 删除 可能存在的 key
	bucket.BatchDelete(nil, bkey, bnewkey1, bnewkey2)
}

func TestAll(t *testing.T) {

	if skipTest() {
		return
	}

	//上传一个文件用用于测试
	err := upFile("bucket_test.go", bkey)
	if err != nil {
		t.Fatal(err)
	}
	defer bucket.Delete(nil, bkey)

	testBatchStat(t)
	testBatchCopy(t)
	testBatchMove(t)
	testBatchDelete(t)
	testBatch(t)
	testClient_MakeUptokenBucket(t)
}

func testBatchStat(t *testing.T) {

	rets, err := bucket.BatchStat(nil, bkey, bkey, bkey)
	if err != nil {
		t.Fatal("bucket.BatchStat failed:", err)
	}

	if len(rets) != 3 {
		t.Fatal("BatchStat failed: len(rets) = ", 3)
	}

	stat, err := bucket.Stat(nil, bkey)
	if err != nil {
		t.Fatal("bucket.Stat failed:", err)
	}

	if rets[0].Data != stat || rets[1].Data != stat || rets[2].Data != stat {
		t.Fatal("BatchStat failed : returns err")
	}
}

func testBatchMove(t *testing.T) {

	stat0, err := bucket.Stat(nil, bkey)
	if err != nil {
		t.Fatal("BathMove get stat failed:", err)
	}

	_, err = bucket.BatchMove(nil, KeyPair{bkey, bnewkey1}, KeyPair{bnewkey1, bnewkey2})
	if err != nil {
		t.Fatal("bucket.BatchMove failed:", err)
	}
	defer bucket.Move(nil, bnewkey2, bkey)

	stat1, err := bucket.Stat(nil, bnewkey2)
	if err != nil {
		t.Fatal("BathMove get stat failed:", err)
	}

	if stat0.Hash != stat1.Hash {
		t.Fatal("BatchMove failed : Move err", stat0, stat1)
	}
}

func testBatchCopy(t *testing.T) {

	_, err := bucket.BatchCopy(nil, KeyPair{bkey, bnewkey1}, KeyPair{bkey, bnewkey2})
	if err != nil {
		t.Fatal(err)
	}
	defer bucket.Delete(nil, bnewkey1)
	defer bucket.Delete(nil, bnewkey2)

	stat0, _ := bucket.Stat(nil, bkey)
	stat1, _ := bucket.Stat(nil, bnewkey1)
	stat2, _ := bucket.Stat(nil, bnewkey2)
	if stat0.Hash != stat1.Hash || stat0.Hash != stat2.Hash {
		t.Fatal("BatchCopy failed : Copy err")
	}
}

func testBatchDelete(t *testing.T) {

	bucket.Copy(nil, bkey, bnewkey1)
	bucket.Copy(nil, bkey, bnewkey2)

	_, err := bucket.BatchDelete(nil, bnewkey1, bnewkey2)
	if err != nil {
		t.Fatal(err)
	}

	_, err1 := bucket.Stat(nil, bnewkey1)
	_, err2 := bucket.Stat(nil, bnewkey2)

	//这里 err1 != nil，否则文件没被成功删除
	if err1 == nil || err2 == nil {
		t.Fatal("BatchDelete failed : File do not delete")
	}
}

func testBatch(t *testing.T) {

	ops := []string{
		URICopy(bucketName, bkey, bucketName, bnewkey1),
		URIDelete(bucketName, bkey),
		URIMove(bucketName, bnewkey1, bucketName, bkey),
	}

	var rets []BatchItemRet
	err := client.Batch(nil, &rets, ops)
	if err != nil {
		t.Fatal(err)
	}
	if len(rets) != 3 {
		t.Fatal("len(rets) != 3")
	}
}
