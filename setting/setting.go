// Package setting provides ...
package setting

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/EiBlog/utils/logd"
	"gopkg.in/yaml.v2"
)

const (
	DEV  = "dev"
	PROD = "prod"
)

var (
	wd, _   = os.Getwd()
	Conf    = new(Config)
	BlackIP = make(map[string]bool)
)

type Config struct {
	StaticVersion int
	RunMode       string
	Trash         int
	Clean         int
	PageNum       int
	PageSize      int
	Length        int
	Identifier    string
	Favicon       string
	StartID       int32
	Static        string
	SearchURL     string
	Disqus        struct {
		ShortName string
		PublicKey string
		URL       string
		Interval  int
	}
	Modes   map[string]RunMode
	Twitter string
	RSS     string
	Search  string
	Blogger struct {
		BlogName  string
		SubTitle  string
		BeiAn     string
		BTitle    string
		Copyright string
	}
	Account struct {
		Username    string
		Password    string
		Email       string
		PhoneNumber string
		Address     string
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
	dir := wd + "/conf"
	data, err := ioutil.ReadFile(path.Join(dir, "app.yml"))
	checkError(err)
	err = yaml.Unmarshal(data, Conf)
	checkError(err)

	data, err = ioutil.ReadFile(path.Join(dir, "blackip.yml"))
	checkError(err)
	err = yaml.Unmarshal(data, BlackIP)
	checkError(err)
}

func checkError(err error) {
	if err != nil {
		logd.Fatal(err)
	}
}
