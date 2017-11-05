package storage

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/x/rpc.v7"
)

// 资源管理相关的默认域名
const (
	DefaultRsHost  = "rs.qiniu.com"
	DefaultRsfHost = "rsf.qiniu.com"
	DefaultAPIHost = "api.qiniu.com"
	DefaultPubHost = "pu.qbox.me:10200"
)

// FileInfo 文件基本信息
type FileInfo struct {
	Hash     string `json:"hash"`
	Fsize    int64  `json:"fsize"`
	PutTime  int64  `json:"putTime"`
	MimeType string `json:"mimeType"`
	Type     int    `json:"type"`
}

func (f *FileInfo) String() string {
	str := ""
	str += fmt.Sprintf("Hash:     %s\n", f.Hash)
	str += fmt.Sprintf("Fsize:    %d\n", f.Fsize)
	str += fmt.Sprintf("PutTime:  %d\n", f.PutTime)
	str += fmt.Sprintf("MimeType: %s\n", f.MimeType)
	str += fmt.Sprintf("Type:     %d\n", f.Type)
	return str
}

// FetchRet 资源抓取的返回值
type FetchRet struct {
	Hash     string `json:"hash"`
	Fsize    int64  `json:"fsize"`
	MimeType string `json:"mimeType"`
	Key      string `json:"key"`
}

func (r *FetchRet) String() string {
	str := ""
	str += fmt.Sprintf("Key:      %s\n", r.Key)
	str += fmt.Sprintf("Hash:     %s\n", r.Hash)
	str += fmt.Sprintf("Fsize:    %d\n", r.Fsize)
	str += fmt.Sprintf("MimeType: %s\n", r.MimeType)
	return str
}

// ListItem 为文件列举的返回值
type ListItem struct {
	Key      string `json:"key"`
	Hash     string `json:"hash"`
	Fsize    int64  `json:"fsize"`
	PutTime  int64  `json:"putTime"`
	MimeType string `json:"mimeType"`
	Type     int    `json:"type"`
	EndUser  string `json:"endUser"`
}

func (l *ListItem) String() string {
	str := ""
	str += fmt.Sprintf("Hash:     %s\n", l.Hash)
	str += fmt.Sprintf("Fsize:    %d\n", l.Fsize)
	str += fmt.Sprintf("PutTime:  %d\n", l.PutTime)
	str += fmt.Sprintf("MimeType: %s\n", l.MimeType)
	str += fmt.Sprintf("Type:     %d\n", l.Type)
	str += fmt.Sprintf("EndUser:  %s\n", l.EndUser)
	return str
}

// BatchOpRet 为批量执行操作的返回值
// 批量操作支持 stat，copy，delete，move，chgm，chtype，deleteAfterDays几个操作
// 其中 stat 为获取文件的基本信息，如果文件存在则返回基本信息，如果文件不存在返回 error 。
// 其他的操作，如果成功，则返回 code，不成功会同时返回 error 信息，可以根据 error 信息来判断问题所在。
type BatchOpRet struct {
	Code int `json:"code,omitempty"`
	Data struct {
		Hash     string `json:"hash"`
		Fsize    int64  `json:"fsize"`
		PutTime  int64  `json:"putTime"`
		MimeType string `json:"mimeType"`
		Type     int    `json:"type"`
		Error    string `json:"error"`
	} `json:"data,omitempty"`
}

// BucketManager 提供了对资源进行管理的操作
type BucketManager struct {
	client *rpc.Client
	mac    *qbox.Mac
	cfg    *Config
}

// NewBucketManager 用来构建一个新的资源管理对象
func NewBucketManager(mac *qbox.Mac, cfg *Config) *BucketManager {
	if cfg == nil {
		cfg = &Config{}
	}

	return &BucketManager{
		client: NewClient(mac, nil),
		mac:    mac,
		cfg:    cfg,
	}
}

// NewBucketManagerEx 用来构建一个新的资源管理对象
func NewBucketManagerEx(mac *qbox.Mac, cfg *Config, client *rpc.Client) *BucketManager {
	if cfg == nil {
		cfg = &Config{}
	}

	if client == nil {
		client = NewClient(mac, nil)
	}

	return &BucketManager{
		client: client,
		mac:    mac,
		cfg:    cfg,
	}
}

// Buckets 用来获取空间列表，如果指定了 shared 参数为 true，那么一同列表被授权访问的空间
func (m *BucketManager) Buckets(shared bool) (buckets []string, err error) {
	ctx := context.TODO()
	var reqHost string

	scheme := "http://"
	if m.cfg.UseHTTPS {
		scheme = "https://"
	}

	reqHost = fmt.Sprintf("%s%s", scheme, DefaultRsHost)
	reqURL := fmt.Sprintf("%s/buckets?shared=%v", reqHost, shared)
	err = m.client.Call(ctx, &buckets, "POST", reqURL)
	return
}

// Stat 用来获取一个文件的基本信息
func (m *BucketManager) Stat(bucket, key string) (info FileInfo, err error) {
	ctx := context.TODO()
	reqHost, reqErr := m.rsHost(bucket)
	if reqErr != nil {
		err = reqErr
		return
	}

	reqURL := fmt.Sprintf("%s%s", reqHost, URIStat(bucket, key))
	err = m.client.Call(ctx, &info, "POST", reqURL)
	return
}

// Delete 用来删除空间中的一个文件
func (m *BucketManager) Delete(bucket, key string) (err error) {
	ctx := context.TODO()
	reqHost, reqErr := m.rsHost(bucket)
	if reqErr != nil {
		err = reqErr
		return
	}
	reqURL := fmt.Sprintf("%s%s", reqHost, URIDelete(bucket, key))
	err = m.client.Call(ctx, nil, "POST", reqURL)
	return
}

// Copy 用来创建已有空间中的文件的一个新的副本
func (m *BucketManager) Copy(srcBucket, srcKey, destBucket, destKey string, force bool) (err error) {
	ctx := context.TODO()
	reqHost, reqErr := m.rsHost(srcBucket)
	if reqErr != nil {
		err = reqErr
		return
	}

	reqURL := fmt.Sprintf("%s%s", reqHost, URICopy(srcBucket, srcKey, destBucket, destKey, force))
	err = m.client.Call(ctx, nil, "POST", reqURL)
	return
}

// Move 用来将空间中的一个文件移动到新的空间或者重命名
func (m *BucketManager) Move(srcBucket, srcKey, destBucket, destKey string, force bool) (err error) {
	ctx := context.TODO()
	reqHost, reqErr := m.rsHost(srcBucket)
	if reqErr != nil {
		err = reqErr
		return
	}

	reqURL := fmt.Sprintf("%s%s", reqHost, URIMove(srcBucket, srcKey, destBucket, destKey, force))
	err = m.client.Call(ctx, nil, "POST", reqURL)
	return
}

// ChangeMime 用来更新文件的MimeType
func (m *BucketManager) ChangeMime(bucket, key, newMime string) (err error) {
	ctx := context.TODO()
	reqHost, reqErr := m.rsHost(bucket)
	if reqErr != nil {
		err = reqErr
		return
	}
	reqURL := fmt.Sprintf("%s%s", reqHost, URIChangeMime(bucket, key, newMime))
	err = m.client.Call(ctx, nil, "POST", reqURL)
	return
}

// ChangeType 用来更新文件的存储类型，0表示普通存储，1表示低频存储
func (m *BucketManager) ChangeType(bucket, key string, fileType int) (err error) {
	ctx := context.TODO()
	reqHost, reqErr := m.rsHost(bucket)
	if reqErr != nil {
		err = reqErr
		return
	}
	reqURL := fmt.Sprintf("%s%s", reqHost, URIChangeType(bucket, key, fileType))
	err = m.client.Call(ctx, nil, "POST", reqURL)
	return
}

// DeleteAfterDays 用来更新文件生命周期，如果 days 设置为0，则表示取消文件的定期删除功能，永久存储
func (m *BucketManager) DeleteAfterDays(bucket, key string, days int) (err error) {
	ctx := context.TODO()
	reqHost, reqErr := m.rsHost(bucket)
	if reqErr != nil {
		err = reqErr
		return
	}

	reqURL := fmt.Sprintf("%s%s", reqHost, URIDeleteAfterDays(bucket, key, days))
	err = m.client.Call(ctx, nil, "POST", reqURL)
	return
}

// Batch 接口提供了资源管理的批量操作，支持 stat，copy，move，delete，chgm，chtype，deleteAfterDays几个接口
func (m *BucketManager) Batch(operations []string) (batchOpRet []BatchOpRet, err error) {
	if len(operations) > 1000 {
		err = errors.New("batch operation count exceeds the limit of 1000")
		return
	}
	ctx := context.TODO()
	scheme := "http://"
	if m.cfg.UseHTTPS {
		scheme = "https://"
	}
	reqURL := fmt.Sprintf("%s%s/batch", scheme, DefaultRsHost)
	params := map[string][]string{
		"op": operations,
	}
	err = m.client.CallWithForm(ctx, &batchOpRet, "POST", reqURL, params)
	return
}

// Fetch 根据提供的远程资源链接来抓取一个文件到空间并已指定文件名保存
func (m *BucketManager) Fetch(resURL, bucket, key string) (fetchRet FetchRet, err error) {
	ctx := context.TODO()
	reqHost, reqErr := m.iovipHost(bucket)
	if reqErr != nil {
		err = reqErr
		return
	}
	reqURL := fmt.Sprintf("%s%s", reqHost, uriFetch(resURL, bucket, key))
	err = m.client.Call(ctx, &fetchRet, "POST", reqURL)
	return
}

// FetchWithoutKey 根据提供的远程资源链接来抓取一个文件到空间并以文件的内容hash作为文件名
func (m *BucketManager) FetchWithoutKey(resURL, bucket string) (fetchRet FetchRet, err error) {
	ctx := context.TODO()
	reqHost, reqErr := m.iovipHost(bucket)
	if reqErr != nil {
		err = reqErr
		return
	}
	reqURL := fmt.Sprintf("%s%s", reqHost, uriFetchWithoutKey(resURL, bucket))
	err = m.client.Call(ctx, &fetchRet, "POST", reqURL)
	return
}

// Prefetch 用来同步镜像空间的资源和镜像源资源内容
func (m *BucketManager) Prefetch(bucket, key string) (err error) {
	ctx := context.TODO()
	reqHost, reqErr := m.iovipHost(bucket)
	if reqErr != nil {
		err = reqErr
		return
	}
	reqURL := fmt.Sprintf("%s%s", reqHost, uriPrefetch(bucket, key))
	err = m.client.Call(ctx, nil, "POST", reqURL)
	return
}

// SetImage 用来设置空间镜像源
func (m *BucketManager) SetImage(siteURL, bucket string) (err error) {
	ctx := context.TODO()
	reqURL := fmt.Sprintf("http://%s%s", DefaultPubHost, uriSetImage(siteURL, bucket))
	err = m.client.Call(ctx, nil, "POST", reqURL)
	return
}

// SetImageWithHost 用来设置空间镜像源，额外添加回源Host头部
func (m *BucketManager) SetImageWithHost(siteURL, bucket, host string) (err error) {
	ctx := context.TODO()
	reqURL := fmt.Sprintf("http://%s%s", DefaultPubHost,
		uriSetImageWithHost(siteURL, bucket, host))
	err = m.client.Call(ctx, nil, "POST", reqURL)
	return
}

// UnsetImage 用来取消空间镜像源设置
func (m *BucketManager) UnsetImage(bucket string) (err error) {
	ctx := context.TODO()
	reqURL := fmt.Sprintf("http://%s%s", DefaultPubHost, uriUnsetImage(bucket))
	err = m.client.Call(ctx, nil, "POST", reqURL)
	return err
}

type listFilesRet struct {
	Marker         string     `json:"marker"`
	Items          []ListItem `json:"items"`
	CommonPrefixes []string   `json:"commonPrefixes"`
}

// ListFiles 用来获取空间文件列表，可以根据需要指定文件的前缀 prefix，文件的目录 delimiter，循环列举的时候下次
// 列举的位置 marker，以及每次返回的文件的最大数量limit，其中limit最大为1000。
func (m *BucketManager) ListFiles(bucket, prefix, delimiter, marker string,
	limit int) (entries []ListItem, commonPrefixes []string, nextMarker string, hasNext bool, err error) {
	if limit <= 0 || limit > 1000 {
		err = errors.New("invalid list limit, only allow [1, 1000]")
		return
	}

	ctx := context.TODO()
	reqHost, reqErr := m.rsfHost(bucket)
	if reqErr != nil {
		err = reqErr
		return
	}

	ret := listFilesRet{}
	reqURL := fmt.Sprintf("%s%s", reqHost, uriListFiles(bucket, prefix, delimiter, marker, limit))
	err = m.client.Call(ctx, &ret, "POST", reqURL)
	if err != nil {
		return
	}

	commonPrefixes = ret.CommonPrefixes
	nextMarker = ret.Marker
	entries = ret.Items
	if ret.Marker != "" {
		hasNext = true
	}

	return
}

func (m *BucketManager) rsHost(bucket string) (rsHost string, err error) {
	var zone *Zone
	if m.cfg.Zone != nil {
		zone = m.cfg.Zone
	} else {
		if v, zoneErr := GetZone(m.mac.AccessKey, bucket); zoneErr != nil {
			err = zoneErr
			return
		} else {
			zone = v
		}
	}

	scheme := "http://"
	if m.cfg.UseHTTPS {
		scheme = "https://"
	}

	rsHost = fmt.Sprintf("%s%s", scheme, zone.RsHost)
	return
}

func (m *BucketManager) rsfHost(bucket string) (rsfHost string, err error) {
	var zone *Zone
	if m.cfg.Zone != nil {
		zone = m.cfg.Zone
	} else {
		if v, zoneErr := GetZone(m.mac.AccessKey, bucket); zoneErr != nil {
			err = zoneErr
			return
		} else {
			zone = v
		}
	}

	scheme := "http://"
	if m.cfg.UseHTTPS {
		scheme = "https://"
	}

	rsfHost = fmt.Sprintf("%s%s", scheme, zone.RsfHost)
	return
}

func (m *BucketManager) iovipHost(bucket string) (iovipHost string, err error) {
	var zone *Zone
	if m.cfg.Zone != nil {
		zone = m.cfg.Zone
	} else {
		if v, zoneErr := GetZone(m.mac.AccessKey, bucket); zoneErr != nil {
			err = zoneErr
			return
		} else {
			zone = v
		}
	}

	scheme := "http://"
	if m.cfg.UseHTTPS {
		scheme = "https://"
	}

	iovipHost = fmt.Sprintf("%s%s", scheme, zone.IovipHost)
	return
}

// 构建op的方法，导出的方法支持在Batch操作中使用

// URIStat 构建 stat 接口的请求命令
func URIStat(bucket, key string) string {
	return fmt.Sprintf("/stat/%s", EncodedEntry(bucket, key))
}

// URIDelete 构建 delete 接口的请求命令
func URIDelete(bucket, key string) string {
	return fmt.Sprintf("/delete/%s", EncodedEntry(bucket, key))
}

// URICopy 构建 copy 接口的请求命令
func URICopy(srcBucket, srcKey, destBucket, destKey string, force bool) string {
	return fmt.Sprintf("/copy/%s/%s/force/%v", EncodedEntry(srcBucket, srcKey),
		EncodedEntry(destBucket, destKey), force)
}

// URIMove 构建 move 接口的请求命令
func URIMove(srcBucket, srcKey, destBucket, destKey string, force bool) string {
	return fmt.Sprintf("/move/%s/%s/force/%v", EncodedEntry(srcBucket, srcKey),
		EncodedEntry(destBucket, destKey), force)
}

// URIDeleteAfterDays 构建 deleteAfterDays 接口的请求命令
func URIDeleteAfterDays(bucket, key string, days int) string {
	return fmt.Sprintf("/deleteAfterDays/%s/%d", EncodedEntry(bucket, key), days)
}

// URIChangeMime 构建 chgm 接口的请求命令
func URIChangeMime(bucket, key, newMime string) string {
	return fmt.Sprintf("/chgm/%s/mime/%s", EncodedEntry(bucket, key),
		base64.URLEncoding.EncodeToString([]byte(newMime)))
}

// URIChangeType 构建 chtype 接口的请求命令
func URIChangeType(bucket, key string, fileType int) string {
	return fmt.Sprintf("/chtype/%s/type/%d", EncodedEntry(bucket, key), fileType)
}

// 构建op的方法，非导出的方法无法用在Batch操作中
func uriFetch(resURL, bucket, key string) string {
	return fmt.Sprintf("/fetch/%s/to/%s",
		base64.URLEncoding.EncodeToString([]byte(resURL)), EncodedEntry(bucket, key))
}

func uriFetchWithoutKey(resURL, bucket string) string {
	return fmt.Sprintf("/fetch/%s/to/%s",
		base64.URLEncoding.EncodeToString([]byte(resURL)), EncodedEntryWithoutKey(bucket))
}

func uriPrefetch(bucket, key string) string {
	return fmt.Sprintf("/prefetch/%s", EncodedEntry(bucket, key))
}

func uriSetImage(siteURL, bucket string) string {
	return fmt.Sprintf("/image/%s/from/%s", bucket,
		base64.URLEncoding.EncodeToString([]byte(siteURL)))
}

func uriSetImageWithHost(siteURL, bucket, host string) string {
	return fmt.Sprintf("/image/%s/from/%s/host/%s", bucket,
		base64.URLEncoding.EncodeToString([]byte(siteURL)),
		base64.URLEncoding.EncodeToString([]byte(host)))
}

func uriUnsetImage(bucket string) string {
	return fmt.Sprintf("/unimage/%s", bucket)
}

func uriListFiles(bucket, prefix, delimiter, marker string, limit int) string {
	query := make(url.Values)
	query.Add("bucket", bucket)
	if prefix != "" {
		query.Add("prefix", prefix)
	}
	if delimiter != "" {
		query.Add("delimiter", delimiter)
	}
	if marker != "" {
		query.Add("marker", marker)
	}
	if limit > 0 {
		query.Add("limit", strconv.FormatInt(int64(limit), 10))
	}
	return fmt.Sprintf("/list?%s", query.Encode())
}

// EncodedEntry 生成URL Safe Base64编码的 Entry
func EncodedEntry(bucket, key string) string {
	entry := fmt.Sprintf("%s:%s", bucket, key)
	return base64.URLEncoding.EncodeToString([]byte(entry))
}

// EncodedEntryWithoutKey 生成 key 为null的情况下 URL Safe Base64编码的Entry
func EncodedEntryWithoutKey(bucket string) string {
	return base64.URLEncoding.EncodeToString([]byte(bucket))
}

// MakePublicURL 用来生成公开空间资源下载链接
func MakePublicURL(domain, key string) (finalUrl string) {
	srcUrl := fmt.Sprintf("%s/%s", domain, key)
	srcUri, _ := url.Parse(srcUrl)
	finalUrl = srcUri.String()
	return
}

// MakePrivateURL 用来生成私有空间资源下载链接
func MakePrivateURL(mac *qbox.Mac, domain, key string, deadline int64) (privateURL string) {
	publicURL := MakePublicURL(domain, key)
	urlToSign := publicURL
	if strings.Contains(publicURL, "?") {
		urlToSign = fmt.Sprintf("%s&e=%d", urlToSign, deadline)
	} else {
		urlToSign = fmt.Sprintf("%s?e=%d", urlToSign, deadline)
	}
	token := mac.Sign([]byte(urlToSign))
	privateURL = fmt.Sprintf("%s&token=%s", urlToSign, token)
	return
}
