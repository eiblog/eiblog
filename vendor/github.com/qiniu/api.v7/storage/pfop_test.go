package storage

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
)

var (
	testVideoKey = "qiniu.mp4"
)

func TestPrefop(t *testing.T) {
	pid := "na0.597802c092129336c20f3f91"
	prefopRet, err := operationManager.Prefop(pid)
	if err != nil {
		t.Fatalf("Prefop() error, %s", err)
	}
	t.Logf("%s", prefopRet.String())
}

func TestPfop(t *testing.T) {
	saveBucket := testBucket

	fopAvthumb := fmt.Sprintf("avthumb/mp4/s/480x320/vb/500k|saveas/%s",
		EncodedEntry(saveBucket, "pfop_test_qiniu.mp4"))
	fopVframe := fmt.Sprintf("vframe/jpg/offset/10|saveas/%s",
		EncodedEntry(saveBucket, "pfop_test_qiniu.jpg"))
	fopVsample := fmt.Sprintf("vsample/jpg/interval/20/pattern/%s",
		base64.URLEncoding.EncodeToString([]byte("pfop_test_$(count).jpg")))

	fopBatch := []string{fopAvthumb, fopVframe, fopVsample}
	fops := strings.Join(fopBatch, ";")

	force := true
	notifyURL := ""
	pid, err := operationManager.Pfop(testBucket, testVideoKey, fops,
		testPipeline, notifyURL, force)
	if err != nil {
		t.Fatalf("Pfop() error, %s", err)
	}
	t.Logf("persistentId: %s", pid)

}
