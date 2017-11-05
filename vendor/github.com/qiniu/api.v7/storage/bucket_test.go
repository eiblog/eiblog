package storage

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/qiniu/api.v7/auth/qbox"
)

var (
	testAK                  = os.Getenv("QINIU_ACCESS_KEY")
	testSK                  = os.Getenv("QINIU_SECRET_KEY")
	testBucket              = os.Getenv("QINIU_TEST_BUCKET")
	testBucketPrivate       = os.Getenv("QINIU_TEST_BUCKET_PRIVATE")
	testBucketPrivateDomain = os.Getenv("QINIU_TEST_DOMAIN_PRIVATE")
	testPipeline            = os.Getenv("QINIU_TEST_PIPELINE")

	testKey      = "qiniu.png"
	testFetchUrl = "http://devtools.qiniu.com/qiniu.png"
	testSiteUrl  = "http://devtools.qiniu.com"
)

var mac *qbox.Mac
var bucketManager *BucketManager
var operationManager *OperationManager
var formUploader *FormUploader
var resumeUploader *ResumeUploader
var base64Uploader *Base64Uploader

func init() {
	if testAK == "" || testSK == "" {
		panic("please run ./test-env.sh first")
	}
	mac = qbox.NewMac(testAK, testSK)
	cfg := Config{}
	cfg.Zone = &Zone_z0
	cfg.UseCdnDomains = true
	bucketManager = NewBucketManager(mac, &cfg)
	operationManager = NewOperationManager(mac, &cfg)
	formUploader = NewFormUploader(&cfg)
	resumeUploader = NewResumeUploader(&cfg)
	base64Uploader = NewBase64Uploader(&cfg)
	rand.Seed(time.Now().Unix())
}

//Test get zone
func TestGetZone(t *testing.T) {
	zone, err := GetZone(testAK, testBucket)
	if err != nil {
		t.Fatalf("GetZone() error, %s", err)
	}
	t.Log(zone.String())
}

//Test get bucket list
func TestBuckets(t *testing.T) {
	shared := true
	buckets, err := bucketManager.Buckets(shared)
	if err != nil {
		t.Fatalf("Buckets() error, %s", err)
	}

	for _, bucket := range buckets {
		t.Log(bucket)
	}
}

//Test get file info
func TestStat(t *testing.T) {
	keysToStat := []string{"qiniu.png"}

	for _, eachKey := range keysToStat {
		info, err := bucketManager.Stat(testBucket, eachKey)
		if err != nil {
			t.Logf("Stat() error, %s", err)
			t.Fail()
		} else {
			t.Logf("FileInfo:\n %s", info.String())
		}
	}
}

func TestCopyMoveDelete(t *testing.T) {
	keysCopyTarget := []string{"qiniu_1.png", "qiniu_2.png", "qiniu_3.png"}
	keysToDelete := make([]string, 0, len(keysCopyTarget))
	for _, eachKey := range keysCopyTarget {
		err := bucketManager.Copy(testBucket, testKey, testBucket, eachKey, true)
		if err != nil {
			t.Logf("Copy() error, %s", err)
			t.Fail()
		}
	}

	for _, eachKey := range keysCopyTarget {
		keyToDelete := eachKey + "_move"
		err := bucketManager.Move(testBucket, eachKey, testBucket, keyToDelete, true)
		if err != nil {
			t.Logf("Move() error, %s", err)
			t.Fail()
		} else {
			keysToDelete = append(keysToDelete, keyToDelete)
		}
	}

	for _, eachKey := range keysToDelete {
		err := bucketManager.Delete(testBucket, eachKey)
		if err != nil {
			t.Logf("Delete() error, %s", err)
			t.Fail()
		}
	}
}

func TestFetch(t *testing.T) {
	ret, err := bucketManager.Fetch(testFetchUrl, testBucket, "qiniu-fetch.png")
	if err != nil {
		t.Logf("Fetch() error, %s", err)
		t.Fail()
	} else {
		t.Logf("FetchRet:\n %s", ret.String())
	}
}

func TestFetchWithoutKey(t *testing.T) {
	ret, err := bucketManager.FetchWithoutKey(testFetchUrl, testBucket)
	if err != nil {
		t.Logf("FetchWithoutKey() error, %s", err)
		t.Fail()
	} else {
		t.Logf("FetchRet:\n %s", ret.String())
	}
}

func TestDeleteAfterDays(t *testing.T) {
	deleteKey := testKey + "_deleteAfterDays"
	days := 7
	bucketManager.Copy(testBucket, testKey, testBucket, deleteKey, true)
	err := bucketManager.DeleteAfterDays(testBucket, deleteKey, days)
	if err != nil {
		t.Logf("DeleteAfterDays() error, %s", err)
		t.Fail()
	}
}

func TestChangeMime(t *testing.T) {
	toChangeKey := testKey + "_changeMime"
	bucketManager.Copy(testBucket, testKey, testBucket, toChangeKey, true)
	newMime := "text/plain"
	err := bucketManager.ChangeMime(testBucket, toChangeKey, newMime)
	if err != nil {
		t.Fatalf("ChangeMime() error, %s", err)
	}

	info, err := bucketManager.Stat(testBucket, toChangeKey)
	if err != nil || info.MimeType != newMime {
		t.Fatalf("ChangeMime() failed, %s", err)
	}
	bucketManager.Delete(testBucket, toChangeKey)
}

func TestChangeType(t *testing.T) {
	toChangeKey := fmt.Sprintf("%s_changeType_%d", testKey, rand.Int())
	bucketManager.Copy(testBucket, testKey, testBucket, toChangeKey, true)
	fileType := 1
	err := bucketManager.ChangeType(testBucket, toChangeKey, fileType)
	if err != nil {
		t.Fatalf("ChangeType() error, %s", err)
	}

	info, err := bucketManager.Stat(testBucket, toChangeKey)
	if err != nil || info.Type != fileType {
		t.Fatalf("ChangeMime() failed, %s", err)
	}
	bucketManager.Delete(testBucket, toChangeKey)
}

func TestPrefetchAndImage(t *testing.T) {
	err := bucketManager.SetImage(testSiteUrl, testBucket)
	if err != nil {
		t.Fatalf("SetImage() error, %s", err)
	}

	err = bucketManager.Prefetch(testBucket, testKey)
	if err != nil {
		t.Fatalf("Prefetch() error, %s", err)
	}

	err = bucketManager.UnsetImage(testBucket)
	if err != nil {
		t.Fatalf("UnsetImage() error, %s", err)
	}
}

func TestListFiles(t *testing.T) {
	limit := 100
	prefix := "listfiles/"
	for i := 0; i < limit; i++ {
		newKey := fmt.Sprintf("%s%s/%d", prefix, testKey, i)
		bucketManager.Copy(testBucket, testKey, testBucket, newKey, true)
	}
	entries, _, _, hasNext, err := bucketManager.ListFiles(testBucket, prefix, "", "", limit)
	if err != nil {
		t.Fatalf("ListFiles() error, %s", err)
	}

	if hasNext {
		t.Fatalf("ListFiles() failed, unexpected hasNext")
	}

	if len(entries) != limit {
		t.Fatalf("ListFiles() failed, unexpected items count, expected: %d, actual: %d", limit, len(entries))
	}

	for _, entry := range entries {
		t.Logf("ListItem:\n%s", entry.String())
	}
}

func TestMakePrivateUrl(t *testing.T) {
	deadline := time.Now().Add(time.Second * 3600).Unix()
	privateURL := MakePrivateURL(mac, "http://"+testBucketPrivateDomain, testKey, deadline)
	t.Logf("PrivateUrl: %s", privateURL)
	resp, respErr := http.Get(privateURL)
	if respErr != nil {
		t.Fatalf("MakePrivateUrl() error, %s", respErr)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("MakePrivateUrl() error, %s", resp.Status)
	}
}

func TestBatch(t *testing.T) {
	copyCnt := 100
	copyOps := make([]string, 0, copyCnt)
	testKeys := make([]string, 0, copyCnt)
	for i := 0; i < copyCnt; i++ {
		cpKey := fmt.Sprintf("%s_batchcopy_%d", testKey, i)
		testKeys = append(testKeys, cpKey)
		copyOps = append(copyOps, URICopy(testBucket, testKey, testBucket, cpKey, true))
	}

	_, bErr := bucketManager.Batch(copyOps)
	if bErr != nil {
		t.Fatalf("BatchCopy error, %s", bErr)
	}

	statOps := make([]string, 0, copyCnt)
	for _, k := range testKeys {
		statOps = append(statOps, URIStat(testBucket, k))
	}
	batchOpRets, bErr := bucketManager.Batch(statOps)
	_, bErr = bucketManager.Batch(copyOps)
	if bErr != nil {
		t.Fatalf("BatchStat error, %s", bErr)
	}

	t.Logf("BatchStat: %v", batchOpRets)
}
