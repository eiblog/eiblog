package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/storage"
)

var (
	accessKey = os.Getenv("QINIU_ACCESS_KEY")
	secretKey = os.Getenv("QINIU_SECRET_KEY")
	bucket    = os.Getenv("QINIU_TEST_BUCKET")
)

func main() {

	// 简单上传凭证
	putPolicy := storage.PutPolicy{
		Scope: bucket,
	}
	mac := qbox.NewMac(accessKey, secretKey)
	upToken := putPolicy.UploadToken(mac)
	fmt.Println(upToken)

	// 设置上传凭证有效期
	putPolicy = storage.PutPolicy{
		Scope: bucket,
	}
	putPolicy.Expires = 7200 //示例2小时有效期

	upToken = putPolicy.UploadToken(mac)
	fmt.Println(upToken)

	// 覆盖上传凭证
	// 需要覆盖的文件名
	keyToOverwrite := "qiniu.mp4"
	putPolicy = storage.PutPolicy{
		Scope: fmt.Sprintf("%s:%s", bucket, keyToOverwrite),
	}
	upToken = putPolicy.UploadToken(mac)
	fmt.Println(upToken)

	// 自定义上传回复凭证
	putPolicy = storage.PutPolicy{
		Scope:      bucket,
		ReturnBody: `{"key":"$(key)","hash":"$(etag)","fsize":$(fsize),"bucket":"$(bucket)","name":"$(x:name)"}`,
	}
	upToken = putPolicy.UploadToken(mac)
	fmt.Println(upToken)

	// 带回调业务服务器的凭证(JSON方式)
	putPolicy = storage.PutPolicy{
		Scope:            bucket,
		CallbackURL:      "http://api.example.com/qiniu/upload/callback",
		CallbackBody:     `{"key":"$(key)","hash":"$(etag)","fsize":$(fsize),"bucket":"$(bucket)","name":"$(x:name)"}`,
		CallbackBodyType: "application/json",
	}
	upToken = putPolicy.UploadToken(mac)
	fmt.Println(upToken)

	// 带回调业务服务器的凭证（URL方式）
	putPolicy = storage.PutPolicy{
		Scope:        bucket,
		CallbackURL:  "http://api.example.com/qiniu/upload/callback",
		CallbackBody: "key=$(key)&hash=$(etag)&bucket=$(bucket)&fsize=$(fsize)&name=$(x:name)",
	}
	upToken = putPolicy.UploadToken(mac)
	fmt.Println(upToken)

	// 带数据处理的凭证
	saveMp4Entry := base64.URLEncoding.EncodeToString([]byte(bucket + ":avthumb_test_target.mp4"))
	saveJpgEntry := base64.URLEncoding.EncodeToString([]byte(bucket + ":vframe_test_target.jpg"))
	//数据处理指令，支持多个指令
	avthumbMp4Fop := "avthumb/mp4|saveas/" + saveMp4Entry
	vframeJpgFop := "vframe/jpg/offset/1|saveas/" + saveJpgEntry
	//连接多个操作指令
	persistentOps := strings.Join([]string{avthumbMp4Fop, vframeJpgFop}, ";")
	pipeline := "test"
	putPolicy = storage.PutPolicy{
		Scope:               bucket,
		PersistentOps:       persistentOps,
		PersistentPipeline:  pipeline,
		PersistentNotifyURL: "http://api.example.com/qiniu/pfop/notify",
	}
	upToken = putPolicy.UploadToken(mac)
	fmt.Println(upToken)
}
