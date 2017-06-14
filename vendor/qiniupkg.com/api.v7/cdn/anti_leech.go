package cdn

import (
	"crypto/md5"
	"fmt"
	"net/url"
	"time"
)

// CreateTimestampAntileechURL 构建带时间戳防盗链的链接
// host需要加上 "http://" 或 "https://"
// encryptKey 七牛防盗链key
func CreateTimestampAntileechURL(host, fileName string, queryStr url.Values, encryptKey string, durationInSeconds int64) (antileechURL string, err error) {

	var urlStr string
	if queryStr != nil {
		urlStr = fmt.Sprintf("%s/%s?%s", host, fileName, queryStr.Encode())
	} else {
		urlStr = fmt.Sprintf("%s/%s", host, fileName)
	}

	u, parseErr := url.Parse(urlStr)
	if parseErr != nil {
		err = parseErr
		return
	}

	expireTime := time.Now().Add(time.Second * time.Duration(durationInSeconds)).Unix()
	toSignStr := fmt.Sprintf("%s%s%x", encryptKey, u.EscapedPath(), expireTime)
	signedStr := fmt.Sprintf("%x", md5.Sum([]byte(toSignStr)))

	q := u.Query()
	q.Add("sign", signedStr)
	q.Add("t", fmt.Sprintf("%x", expireTime))
	u.RawQuery = q.Encode()

	antileechURL = fmt.Sprintf("%s://%s%s?%s", u.Scheme, u.Host, u.EscapedPath(), u.Query().Encode())

	return

}
