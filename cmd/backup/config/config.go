package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/eiblog/eiblog/pkg/config"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// Config config
type Config struct {
	config.APIMode

	Database config.Database // 数据库配置
	BackupTo string          // 备份到, default: qiniu
	Interval string          // 备份周期, default: 7d
	Validity int             // 备份保留时间, default: 60
	Qiniu    config.Qiniu    // 七牛OSS配置
}

// Conf 配置
var Conf Config

// load config file
func init() {
	// run mode
	mode := config.RunMode(os.Getenv("RUN_MODE"))
	if !mode.IsRunMode() {
		panic("config: unsupported env RUN_MODE" + mode)
	}
	logrus.Infof("Run mode:%s", mode)

	// 加载配置文件
	dir, err := config.WalkWorkDir()
	if err != nil {
		panic(err)
	}

	path := filepath.Join(dir, "etc", "app.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(data, &Conf)
	if err != nil {
		panic(err)
	}
	Conf.RunMode = mode

	// read env
	readDatabaseEnv()
}

func readDatabaseEnv() {
	key := strings.ToUpper(Conf.Name) + "_DB_DRIVER"
	if d := os.Getenv(key); d != "" {
		Conf.Database.Driver = d
	}
	key = strings.ToUpper(Conf.Name) + "_DB_SOURCE"
	if s := os.Getenv(key); s != "" {
		Conf.Database.Source = s
	}
}
