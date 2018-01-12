// Package setting provides ...
package setting

import (
	"io/ioutil"

	"github.com/eiblog/utils/logd"
	"gopkg.in/yaml.v2"
)

const (
	DEV  = "dev"  // 该模式会输出 debug 等信息
	PROD = "prod" // 该模式用于生产环境
)

var (
	Conf    = new(Config)
	BlackIP = make(map[string]bool)
)

type Config struct {
	RunMode       string   // 运行模式
	StaticVersion int      // 当前静态文件版本
	FeedrURL      string   // superfeedr url
	HotWords      []string // 热搜词
	PingRPCs      []string // ping rpc 地址
	General       struct {
		PageNum    int    // 前端每页文章数量
		PageSize   int    // 后台每页文章数量
		StartID    int32  // 文章起始id
		DescPrefix string // 文章描述前缀
		Identifier string // 文章截取标示
		Length     int    // 文章自动截取预览长度
		Trash      int    // 回收箱文章保留时间
		Clean      int    // 清理回收箱频率
	}
	Disqus struct { // 获取文章数量相关
		ShortName    string
		PublicKey    string
		AccessToken  string
		PostsCount   string
		PostsList    string
		PostCreate   string
		PostApprove  string
		ThreadCreate string
		Embed        string
		Interval     int
	}
	Google struct { // 谷歌统计
		URL string
		Tid string
		V   string
		T   string
	}
	Qiniu struct { // 七牛CDN
		Bucket    string
		Domain    string
		AccessKey string
		SecretKey string
	}
	Mode struct { // 运行模式
		EnableHttp  bool
		HttpPort    int
		EnableHttps bool
		HttpsPort   int
		CertFile    string
		KeyFile     string
		AutoCert    bool
		Domain      string
	}
	Twitter struct { // twitter信息
		Card    string
		Site    string
		Image   string
		Address string
	}
	Account struct { // account 账户
		Username    string // *
		Password    string // *
		Email       string
		PhoneNumber string
		Address     string
	}
	Blogger struct { // blog info 博客信息
		BlogName  string
		SubTitle  string
		BeiAn     string
		BTitle    string
		Copyright string
	}
}

func init() {
	// 初始化配置
	data, err := ioutil.ReadFile("conf/app.yml")
	checkError(err)
	err = yaml.Unmarshal(data, Conf)
	checkError(err)

	data, err = ioutil.ReadFile("conf/blackip.yml")
	checkError(err)
	err = yaml.Unmarshal(data, BlackIP)
	checkError(err)
}

func checkError(err error) {
	if err != nil {
		logd.Fatal(err)
	}
}
