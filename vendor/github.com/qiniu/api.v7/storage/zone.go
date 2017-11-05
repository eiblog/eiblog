package storage

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/qiniu/x/rpc.v7"
)

// Zone 为空间对应的机房属性，主要包括了上传，资源管理等操作的域名
type Zone struct {
	SrcUpHosts []string
	CdnUpHosts []string
	RsHost     string
	RsfHost    string
	ApiHost    string
	IovipHost  string
}

func (z *Zone) String() string {
	str := ""
	str += fmt.Sprintf("SrcUpHosts: %v\n", z.SrcUpHosts)
	str += fmt.Sprintf("CdnUpHosts: %v\n", z.CdnUpHosts)
	str += fmt.Sprintf("IovipHost: %s\n", z.IovipHost)
	str += fmt.Sprintf("RsHost: %s\n", z.RsHost)
	str += fmt.Sprintf("RsfHost: %s\n", z.RsfHost)
	str += fmt.Sprintf("ApiHost: %s\n", z.ApiHost)
	return str
}

// ZoneHuadong 表示华东机房
var ZoneHuadong = Zone{
	SrcUpHosts: []string{
		"up.qiniup.com",
		"up-nb.qiniup.com",
		"up-xs.qiniup.com",
	},
	CdnUpHosts: []string{
		"upload.qiniup.com",
		"upload-nb.qiniup.com",
		"upload-xs.qiniup.com",
	},
	RsHost:    "rs.qiniu.com",
	RsfHost:   "rsf.qiniu.com",
	ApiHost:   "api.qiniu.com",
	IovipHost: "iovip.qbox.me",
}

// ZoneHuabei 表示华北机房
var ZoneHuabei = Zone{
	SrcUpHosts: []string{
		"up-z1.qiniup.com",
	},
	CdnUpHosts: []string{
		"upload-z1.qiniup.com",
	},
	RsHost:    "rs-z1.qiniu.com",
	RsfHost:   "rsf-z1.qiniu.com",
	ApiHost:   "api-z1.qiniu.com",
	IovipHost: "iovip-z1.qbox.me",
}

// ZoneHuanan 表示华南机房
var ZoneHuanan = Zone{
	SrcUpHosts: []string{
		"up-z2.qiniup.com",
		"up-gz.qiniup.com",
		"up-fs.qiniup.com",
	},
	CdnUpHosts: []string{
		"upload-z2.qiniup.com",
		"upload-gz.qiniup.com",
		"upload-fs.qiniup.com",
	},
	RsHost:    "rs-z2.qiniu.com",
	RsfHost:   "rsf-z2.qiniu.com",
	ApiHost:   "api-z2.qiniu.com",
	IovipHost: "iovip-z2.qbox.me",
}

// ZoneBeimei 表示北美机房
var ZoneBeimei = Zone{
	SrcUpHosts: []string{
		"up-na0.qiniu.com",
	},
	CdnUpHosts: []string{
		"upload-na0.qiniu.com",
	},
	RsHost:    "rs-na0.qiniu.com",
	RsfHost:   "rsf-na0.qiniu.com",
	ApiHost:   "api-na0.qiniu.com",
	IovipHost: "iovip-na0.qbox.me",
}

// for programmers
var Zone_z0 = ZoneHuadong
var Zone_z1 = ZoneHuabei
var Zone_z2 = ZoneHuanan
var Zone_na0 = ZoneBeimei

// UcHost 为查询空间相关域名的API服务地址
const UcHost = "https://uc.qbox.me"

// UcQueryRet 为查询请求的回复
type UcQueryRet struct {
	TTL int                            `json:"ttl"`
	Io  map[string]map[string][]string `json:"io"`
	Up  map[string]UcQueryUp           `json:"up"`
}

// UcQueryUp 为查询请求回复中的上传域名信息
type UcQueryUp struct {
	Main   []string `json:"main,omitempty"`
	Backup []string `json:"backup,omitempty"`
	Info   string   `json:"info,omitempty"`
}

var (
	zoneMutext sync.RWMutex
	zoneCache  = make(map[string]*Zone)
)

// GetZone 用来根据ak和bucket来获取空间相关的机房信息
func GetZone(ak, bucket string) (zone *Zone, err error) {
	zoneID := fmt.Sprintf("%s:%s", ak, bucket)
	//check from cache
	zoneMutext.RLock()
	if v, ok := zoneCache[zoneID]; ok {
		zone = v
	}
	zoneMutext.RUnlock()
	if zone != nil {
		return
	}

	//query from server
	reqURL := fmt.Sprintf("%s/v2/query?ak=%s&bucket=%s", UcHost, ak, bucket)
	var ret UcQueryRet
	ctx := context.Background()
	qErr := rpc.DefaultClient.CallWithForm(ctx, &ret, "GET", reqURL, nil)
	if qErr != nil {
		err = fmt.Errorf("query zone error, %s", qErr.Error())
		return
	}

	ioHost := ret.Io["src"]["main"][0]
	srcUpHosts := ret.Up["src"].Main
	if ret.Up["src"].Backup != nil {
		srcUpHosts = append(srcUpHosts, ret.Up["src"].Backup...)
	}
	cdnUpHosts := ret.Up["acc"].Main
	if ret.Up["acc"].Backup != nil {
		cdnUpHosts = append(cdnUpHosts, ret.Up["acc"].Backup...)
	}

	zone = &Zone{
		SrcUpHosts: srcUpHosts,
		CdnUpHosts: cdnUpHosts,
		IovipHost:  ioHost,
		RsHost:     DefaultRsHost,
		RsfHost:    DefaultRsfHost,
		ApiHost:    DefaultAPIHost,
	}

	//set specific hosts if possible
	setSpecificHosts(ioHost, zone)

	zoneMutext.Lock()
	zoneCache[zoneID] = zone
	zoneMutext.Unlock()
	return
}

func setSpecificHosts(ioHost string, zone *Zone) {
	if strings.Contains(ioHost, "-z1") {
		zone.RsHost = "rs-z1.qiniu.com"
		zone.RsfHost = "rsf-z1.qiniu.com"
		zone.ApiHost = "api-z1.qiniu.com"
	} else if strings.Contains(ioHost, "-z2") {
		zone.RsHost = "rs-z2.qiniu.com"
		zone.RsfHost = "rsf-z2.qiniu.com"
		zone.ApiHost = "api-z2.qiniu.com"
	} else if strings.Contains(ioHost, "-na0") {
		zone.RsHost = "rs-na0.qiniu.com"
		zone.RsfHost = "rsf-na0.qiniu.com"
		zone.ApiHost = "api-na0.qiniu.com"
	}
}
