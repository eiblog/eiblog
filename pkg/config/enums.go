package config

// APIMode 应用配置
type APIMode struct {
	// 运行模式，可根据不同模式做相关判断，如日志打印等
	RunMode RunMode

	// 服务名称，如 eiblog, backup
	Name string
	// 监听地址，如 0.0.0.0:9000
	Listen string
	// 所属域名，如 deepzz.com
	Host string
	// 一般的应用都会有个密钥，可以用这个字段
	Secret string
}

// Database 数据库配置
type Database struct {
	// 数据库驱动，如 sqlite, mysql, postgres 等
	Driver string
	// 数据库连接字符串，如
	//   sqlite:./db.sqlite
	//   mysql:root:123456@tcp(127.0.0.1:3306)/eiblog?charset=utf8mb4&parseTime=True&loc=Local
	Source string
}

// Disqus 评论配置
type Disqus struct {
	// 短名称，如 deepzz
	ShortName string
	// 公共密钥 wdSgxRm9rdGAlLKFcFdToBe3GT4SibmV7Y8EjJQ0r4GWXeKtxpopMAeIeoI2dTEg
	PublicKey string
	// 访问令牌, 需自行创建
	AccessToken string
}

// Twitter 社交配置
type Twitter struct {
	// 卡片, 可选 summary, summary_large_image, player
	Card string
	// id号, 如 deepzz02
	Site string
	// 图片, 如 st.deepzz.cn/static/img/avatar.jpg
	Image string
	// 地址, 如 twitter.com/deepzz02
	Address string
}

// Google 分析配置
type Google struct {
	// url, 如 https://www.google-analytics.com/g/collect
	URL string
	// tid, 如 G-S085VRC5PF
	Tid string
	// v, 如 "2"
	V string
	// 如果开启广发, 配置 <script async src="xxx" crossorigin="anonymous"></script>
	AdSense string
}

// Qiniu 对象存储配置
type Qiniu struct {
	// bucket, 如 eiblog
	Bucket string
	// domain, 如 st.deepzz.cn
	Domain string
	// accesskey, 如 1234567890
	AccessKey string
	// secretkey, 如 1234567890
	SecretKey string
}

// FeedRPC 订阅配置
type FeedRPC struct {
	// feedrurl, 如 https://deepzz.superfeedr.com/
	FeedrURL string
	// pingrpc, 如 http://ping.baidu.com/ping/RPC2, http://rpc.pingomatic.com/
	PingRPC []string
}

// General 博客通用配置
type General struct {
	// 前台每页文章数量, 一般配置为 10
	PageNum int
	// 后台每页文章数量, 一般配置为 20
	PageSize int
	// 文章描述前缀, 一般配置为 Desc:
	DescPrefix string
	// 文章截取标识, 一般配置为 <!--more-->
	Identifier string
	// 文章预览长度, 一般配置为 400
	Length int
	// 时区, 一般配置为 Asia/Shanghai
	Timezone string
	// 是否启用两步验证
	TwoFactor bool
}

// Account 账户配置
type Account struct {
	// *必须配置, 后台登录用户名
	Username string
	// *必须配置, 后台登录密码。登录后请后台立即修改
	Password string
}

// Blogger 博客配置, 无需配置，程序默认初始化，可在后台更改
type Blogger struct {
	// 博客名称, 如 deepzz
	BlogName string
	// 格言, 如 Rome was not built in one day.
	SubTitle string
	// 备案号, 不填则不显示在网站底部, 如 蜀ICP备xxxxxxxx号-1
	BeiAn string
	// 标题, 如 deepzz's Blog
	BTitle string
	// 版权, 如 本站使用「<a href="//creativecommons.org/licenses/by/4.0/">署名 4.0 国际</a>」创作共享协议，转载请注明作者及原网址。
	Copyright string // 版权
}
