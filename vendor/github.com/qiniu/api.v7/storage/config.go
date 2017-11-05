package storage

// Config 为文件上传，资源管理等配置
type Config struct {
	Zone          *Zone //空间所在的机房
	UseHTTPS      bool  //是否使用https域名
	UseCdnDomains bool  //是否使用cdn加速域名
}
