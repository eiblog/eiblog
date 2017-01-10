// Package setting provides ...
package setting

import (
	"io/ioutil"

	"github.com/eiblog/utils/logd"
	"gopkg.in/yaml.v2"
)

const (
	DEV  = "dev"
	PROD = "prod"
)

var (
	Conf    = new(Config)
	BlackIP = make(map[string]bool)
)

type Config struct {
	StaticVersion int      // 当前静态文件版本
	RunMode       string   // 运行模式
	Trash         int      // 回收箱文章保留时间
	Clean         int      // 清理回收箱频率
	PageNum       int      // 前端每页文章数量
	PageSize      int      // 后台每页文章数量
	Length        int      // 自动截取预览长度
	Identifier    string   // 截取标示
	Description   string   // 文章描述前缀
	Favicon       string   // icon地址
	StartID       int32    // 文章起始id
	SearchURL     string   // elasticsearch 地址
	Disqus        struct { // 获取文章数量相关
		ShortName  string
		PublicKey  string
		PostsCount string
		PostsList  string
		PostCreate string
		Interval   int
	}
	HotWords []string // 热搜词
	Google   struct { // 谷歌统计
		Tid string
		V   string
		T   string
	}
	Kodo struct { // 七牛CDN
		Name      string
		Domain    string
		AccessKey string
		SecretKey string
	}
	Mode    RunMode  // 运行模式
	Twitter struct { // twitter信息
		Card    string
		Site    string
		Image   string
		Address string
	}
	FeedrURL string   // superfeedr url
	PingRPCs []string // ping rpc 地址
	Account  struct {
		Username    string // *
		Password    string // *
		Email       string
		PhoneNumber string
		Address     string
	}
	Blogger struct { // 初始化数据
		BlogName  string
		SubTitle  string
		BeiAn     string
		BTitle    string
		Copyright string
	}
}

type RunMode struct {
	EnableHttp  bool
	HttpPort    int
	EnableHttps bool
	HttpsPort   int
	CertFile    string
	KeyFile     string
	Domain      string
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
