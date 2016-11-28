# EiBlog [![Build Status](https://travis-ci.org/eiblog/eiblog.svg?branch=master)](https://travis-ci.org/eiblog/eiblog)

> 系统根据[https://imququ.com](https://imququ.com)一系列文章和方向进行搭建，期间获得了QuQu的很大帮助，在此表示感谢。

Eiblog的开发目的是自用博客系统，许多地方不适合直接使用部署。需要修改的地方很多，并且该系统要求比较苛刻，配置也比较复杂。有兴趣的朋友，可以参照下面的部署流程进行部署。

### 系统简介

Mongodb 数据库存储

Elasticsearch 站内文章搜索

Nginx 代理

### 前提准备

这里所提到的点都是必需提前准备的。

* 域名：对，域名。
* CDN：静态文件存储。
* HTTPS证书：系统根据https设计，你需要准备一张有效的证书。
* 服务器：当然是必需要有的，部署系统。
* Disqus：disqus评论系统账号（shortname）和disque application key。
* google-analystic: 谷歌分析

### 后续准备

#### 准备文件

1、CDN存储文件

```
static/img/bg04.jpg				#首页左侧背景图片，500*1200 左右
static/img/blank.gif			#空白图片
static/img/avatar.jpg			#头像，256*256左右
static/img/avatar_small.jpg		#小头像，128*128左右
static/img/default_avatar.png	#评论默认图片，92*92左右
static/img/favicon.ico			#网站icon，64*64左右
static/js/disqus_52ef5a.js		#disqus.js，国内背墙，你懂的
```

找到`views/st_blog_css.css`，搜索与`deepzz`相关的地方：

* 替换`left-col`内的url，即背景地址。
* 替换`profilepic a`内的url，即头像地址。
* 替换同样`profilepic a`内的地址，这里是小头像的地址（移动端用到该图片）。 

找到`views/st_blog_js.js`，搜索`deepzz`找到需要替换的地方：

* disqus_shortname：替换deepzz
* 替换与域名`deepzz.com`相关的域名
* 替换`/static/img/blank.gif`地址
* 替换`/static/js/disqus_52ef5a.js`地址

#### 配置文件

以下是配置文件目录，将一一进行说明。

```
├── app.yml
├── blackip.yml
├── es
│   ├── config
│   │   ├── analysis
│   │   │   └── synonym.txt
│   │   ├── elasticsearch.yml
│   │   ├── logging.yml
│   │   └── scripts
│   └── plugins
│       └── ik1.10.0
│           ├── commons-codec-1.9.jar
│           ├── commons-logging-1.2.jar
│           ├── config
│           ├── elasticsearch-analysis-ik-1.10.0.jar
│           ├── httpclient-4.5.2.jar
│           ├── httpcore-4.4.4.jar
│           └── plugin-descriptor.properties
├── nginx
│   ├── domain
│   │   └── deepzz.conf
│   ├── ip.blacklist
│   └── nginx.conf
├── scts
│   ├── aviator.sct
│   └── digicert.sct
├── ssl
│   ├── dhparams.pem
│   ├── domain.key
│   ├── domain.pem
│   ├── full_chained.pem
│   └── session_ticket.key
└── tpl
    ├── feedTpl.xml
    ├── opensearchTpl.xml
    └── sitemapTpl.xml
```

1、`app.yml`，整个程序的配置文件，里面已经列出了所有配置项的说明，这里不再阐述。

2、`es`全名`elasticsearch`，非常强大的分布式搜索引擎，`github`用的就是它。里面的配置基本不用修改，但`es/analysis/synonym.txt`是同义词，你可以照着已有的随意增加。

```
├── es
│   ├── config
│   │   ├── analysis
│   │   │   └── synonym.txt					#同义词配置
│   │   ├── elasticsearch.yml				#分词器配置
│   │   ├── logging.yml						#日志配置
│   │   └── scripts							#脚本
│   └── plugins								#中文分词插件
│       └── ik1.10.0
│
```

>  注意，scripts虽然是空的，但必需存在(空文件夹docker打包不了)，不然elasticsearch报错。

3、`nginx`，系统采用`nginx`作为代理(相信博客系统也不会独占一台服务器～)。该文件夹内容主要是修改`domain/deepzz.conf`，或则重命名(只要是满足`*.conf`)。ok，该`deepzz.conf`文件里面学问是最多的。或许你想一一弄懂，或许…。

> 注意本配置需要更新nginx到最新版，openssl更新到1.0.2j，具体请到`Jerry Qu`的[本博客 Nginx 配置之完整篇](https://imququ.com/post/my-nginx-conf.html)查看，了解详情。

```
server {
    listen               443 ssl http2 fastopen=3 reuseport;

    server_name          www.deepzz.com deepzz.com;
    server_tokens        off;

    include              /data/eiblog/conf/nginx/ip.blacklist;

    # 现在一般证书是内置的。可以注释该项
    # https://imququ.com/post/certificate-transparency.html#toc-2
    # ssl_ct               on;
    # ssl_ct_static_scts   /data/eiblog/conf/scts;

    # 中间证书 + 站点证书
    ssl_certificate      /data/eiblog/conf/ssl/domain.pem;

    # 创建 CSR 文件时用的密钥
    ssl_certificate_key  /data/eiblog/conf/ssl/domain.key;

    # openssl dhparam -out dhparams.pem 2048
    # https://weakdh.org/sysadmin.html
    ssl_dhparam          /data/eiblog/conf/ssl/dhparams.pem;

    # https://github.com/cloudflare/sslconfig/blob/master/conf
    ssl_ciphers          EECDH+CHACHA20:EECDH+CHACHA20-draft:EECDH+AES128:RSA+AES128:EECDH+AES256:RSA+AES256:EECDH+3DES:RSA+3DES:!MD5;

    # 如果启用了 RSA + ECDSA 双证书，Cipher Suite 可以参考以下配置：
    # ssl_ciphers              EECDH+CHACHA20:EECDH+CHACHA20-draft:EECDH+ECDSA+AES128:EECDH+aRSA+AES128:RSA+AES128:EECDH+ECDSA+AES256:EECDH+aRSA+AES256:RSA+AES256:EECDH+ECDSA+3DES:EECDH+aRSA+3DES:RSA+3DES:!MD5;

    ssl_prefer_server_ciphers  on;

    ssl_protocols              TLSv1 TLSv1.1 TLSv1.2;

    ssl_session_cache          shared:SSL:50m;
    ssl_session_timeout        1d;

    ssl_session_tickets        on;

    # openssl rand 48 > session_ticket.key
    # 单机部署可以不指定 ssl_session_ticket_key
    # ssl_session_ticket_key     /data/eiblog/conf/ssl/session_ticket.key;

    ssl_stapling               on;
    ssl_stapling_verify        on;

    # 根证书 + 中间证书
    # https://imququ.com/post/why-can-not-turn-on-ocsp-stapling.html
     ssl_trusted_certificate    /data/eiblog/conf/ssl/full_chained.pem;

    resolver                   114.114.114.114 valid=300s;
    resolver_timeout           10s;

    access_log                 /data/eiblog/logdata/nginx.log;

    if ($request_method !~ ^(GET|HEAD|POST|OPTIONS)$ ) {
        return           444;
    }

    if ($host != 'deepzz.com' ) {
        rewrite          ^/(.*)$  https://deepzz.com/$1 permanent;
    }

    # webmaster 站点验证相关
    location ~* (robots\.txt|favicon\.ico|crossdomain\.xml|google4c90d18e696bdcf8\.html|BingSiteAuth\.xml)$ {
        root             /data/eiblog/static;
        expires          1d;
    }

    location ^~ /static/ {
        root             /data/eiblog;
        add_header       Access-Control-Allow-Origin *;      
        expires          max;
    }

    location ^~ /admin/ {
        proxy_http_version       1.1;

        add_header               Strict-Transport-Security "max-age=31536000; includeSubDomains; preload";

        # DENY 将完全不允许页面被嵌套，可能会导致一些异常。如果遇到这样的问题，建议改成 SAMEORIGIN
        # https://imququ.com/post/web-security-and-response-header.html#toc-1
        add_header               X-Frame-Options DENY;

        add_header               X-Content-Type-Options nosniff;

        # proxy_set_header         X-Via            QingDao.Aliyun;
        proxy_set_header         Connection       "";
        proxy_set_header         Host             deepzz.com;
        proxy_set_header         X-Real_IP        $remote_addr;
        proxy_set_header         X-Forwarded-For  $proxy_add_x_forwarded_for;
        proxy_set_header         X-XSS-Protection 1; mode=block;

        proxy_pass               http://127.0.0.1:9000;
    }

    location / {
        proxy_http_version       1.1;

        add_header               Strict-Transport-Security "max-age=31536000; includeSubDomains; preload";
        add_header               X-Frame-Options deny;
        add_header               X-Content-Type-Options nosniff;
        # 改deepzz相关的
        add_header               Content-Security-Policy "default-src 'none'; script-src 'unsafe-inline' 'unsafe-eval' blob: https:; img-src data: https: https://st.deepzz.com; style-src 'unsafe-inline' https:; child-src https:; connect-src 'self' https://translate.googleapis.com; frame-src https://disqus.com https://www.slideshare.net";
        # 中间证书证书指纹
        # https://imququ.com/post/http-public-key-pinning.html
        add_header               Public-Key-Pins 'pin-sha256="lnsM2T/O9/J84sJFdnrpsFp3awZJ+ZZbYpCWhGloaHI="; pin-sha256="YLh1dUR9y6Kja30RrAn7JKnbQG/uEtLMkBgFF2Fuihg="; max-age=2592000; includeSubDomains';
        add_header               Cache-Control no-cache;

        proxy_ignore_headers     Set-Cookie;

        proxy_hide_header        Vary;
        proxy_hide_header        X-Powered-By;

        # proxy_set_header         X-Via            QingDao.Aliyun;
        proxy_set_header         Connection       "";
        proxy_set_header         Host             deepzz.com;
        proxy_set_header         X-Real_IP        $remote_addr;
        proxy_set_header         X-Forwarded-For  $proxy_add_x_forwarded_for;

        proxy_pass               http://127.0.0.1:9000;
    }
}

server {
    server_name       www.deepzz.com deepzz.com;
    server_tokens     off;

    access_log        /dev/null;

    if ($request_method !~ ^(GET|HEAD|POST)$ ) {
        return        444;
    }

    location ^~ /.well-known/acme-challenge/ {
        alias         /home/jerry/www/challenges/;
        try_files     $uri =404;
    }

    location / {
        rewrite       ^/(.*)$ https://deepzz.com/$1 permanent;
    }
}
```

需要修改项：

* 替换与deepzz有关的项。


* `ssl_trusted_certificate`：将你的证书的中间证书和根证书依次张贴到full_chained.pem内，方法见网址。
* `access_log`：nginx访问日志地址
* `location ~*`：站点验证或什么其它文件。如：你要验证百度站长，会下载到`*.xml`的一个文件，你需要将它放到`eiblog/static`下，并在该位置添加其文件名，那么通过`domain.com/*.xml`就可以访问到了。
* `location ^~ /static/uploads/ `：文件上传，暂时没有用到。
* `location /`：`add_header`处你需要注意的是`pin-sha256`，该功能详细请参见网址。不过建议添加你的证书的根证书的`pin-sha256`。

`ip.blacklist`自然就是ip黑名单了，本系统有两个ip黑名单，一个是nginx使用，另一个博客系统使用(不是nginx代理的用户)。

4、`scts`，现在的证书大都内置，可以不管。

5、`ssl`，这里存放了所有证书相关的内容。

```
├── dhparams.pem				#参见eiblog/conf/nginx/domain/deepzz.conf
├── domain.key					#证书私钥，一般颁发者处下载
├── domain.pem					#证书链，一般从证书颁发者那可以直接下载到
├── full_chained.pem			#参见eiblog/conf/nginx/domain/deepzz.conf
└── session_ticket.key			#参见eiblog/conf/nginx/domain/deepzz.conf
```

6、`tpl`模版相关，不用修改。

### 开始部署

#### 直接部署

或许你迫不及待的想体eiblog，或许你不愿折腾那么多的东西，或许你需要时间来慢慢理解。那么你可以直接部署该系统。直接部署将只会使用到三个配置文件：`conf/app.yml`、`conf/blackip.yml`、`es`。

1、运行`mongodb`数据库系统。获得数据库连接地址，如：127.0.0.1。修改`/etc/hosts`文件，末尾添加一条`127.0.0.1	mongodb`。默认使用`27017`端口，修改请到github的`eiblog`项目下utils/mgo修改。

2、运行`elasticsearch`搜索系统。将`conf/es`目录下的文件覆盖到`elasticsearch`配置文件目录。默认使用`9200`端口，修改请到`conf/app.yml`修改。

3、编译`go build`得到二进制`eiblog`文件，然后将二进制文件`eiblog`、静态文件`static`、网页模版`views`和配置文件`conf`目录拷贝到服务器与二进制文件`eiblog`相同的目录， 执行`./eiblog`即可运行。

```
# 运行模式
mode:
  # you can fix certfile, keyfile, domain
  enablehttp: true
  httpport: 9000
  enablehttps: false
  httpsport: 443
  certfile: conf/certs/domain.pem
  keyfile: conf/certs/domain.key
  domain: deepzz.com
```

可以看到`conf/app.yml`文件下的，默认只开启`http`模式。

#### Nginx Docker部署

前提你已经编译部署好`nginx`和`docker`，如未部署，请参考`Jerry Qu`的[本博客 Nginx 配置之完整篇](https://imququ.com/post/my-nginx-conf.html)。

1、将`conf`文件夹和`static`文件夹拷贝到服务器的`/data/eiblog`目录下，没有则创建。

2、将`docker-compose.yml`拷贝到服务器适当的地方，如：`~/eiblog`。

* 执行`docker-compose up -d`命令，启动博客系统。
* 执行`docker-compose logs -f`，查看日志是否有错(或许`scripts`文件夹不存在，到`/data/eiblog/es/config`创建即可)。
* 有错，根据错误信息修改。
* 部署成功。

### 镜像构建

1、`.travis.yml`，如果你`fork`了Eiblog，或许你会用上它。本系统不提供公共镜像，因为构建的镜像并不通用。

```
script:
  - glide up
  - CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build                      # 编译版本
  - docker build -t registry.cn-hangzhou.aliyuncs.com/deepzz/eiblog .   # 构建镜像

after_success:
  - if [ "$TRAVIS_BRANCH" == "master" ]; then
    docker login -u="$DOCKER_USERNAME" -p="$DOCKER_PASSWORD" registry.cn-hangzhou.aliyuncs.com;
    docker push registry.cn-hangzhou.aliyuncs.com/deepzz/eiblog;
    fi                                                                  # push到镜像仓库
```

首先你需要修改镜像的仓库地址`registry.cn-hangzhou.aliyuncs.com/deepzz/eiblog`。如果需要密码，请到[https://travis-ci.org/](https://travis-ci.org/)填写环境变量`DOCKER_USERNAME`，`DOCKER_PASSWORD`。没有则直接将`after_success`替换为：

```
after_success:
  - docker push registry.cn-hangzhou.aliyuncs.com/deepzz/eiblog
```

2、修改`build_docker.sh`文件中的`domain`为自己仓库地址。执行`./build_docker.sh`。

### 资料备忘

#### 创建mapping

```
	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			TYPE: map[string]interface{}{
				"properties": map[string]interface{}{
					"title": map[string]string{
						"type":            "string",
						"term_vector":     "with_positions_offsets",
						"analyzer":        "ik_syno",
						"search_analyzer": "ik_syno",
					},
					"content": map[string]string{
						"type":            "string",
						"term_vector":     "with_positions_offsets",
						"analyzer":        "ik_syno",
						"search_analyzer": "ik_syno",
					},
					"slug": map[string]string{
						"type": "string",
					},
					"tag": map[string]string{
						"type":  "string",
						"index": "not_analyzed",
					},
					"date": map[string]string{
						"type":  "date",
						"index": "not_analyzed",
					},
				},
			},
		},
	}
```

#### DSL高亮查询

```
fehelperFeHelper：JSON格式化查看

{"highlight":{"fields":{"content":{},"title":{}},"post_tags":["\u003c/b\u003e"],"pre_tags":["\u003cb\u003e"]},"query":{"dis_max":{"queries":[{"match":{"title":{"boost":4,"minimum_should_match":"50%","query":"天气"}}},{"match":{"content":{"boost":4,"minimum_should_match":"75%","query":"天气"}}},{"match":{"tag":{"boost":2,"minimum_should_match":"100%","query":"天气"}}},{"match":{"slug":{"boost":1,"minimum_should_match":"100%","query":"天气"}}}],"tie_breaker":0.3}},"filter":{"bool":{"must":[{"range":{"date":{"gte":"2016-10","lte": "2016-10||/M","format": "yyyy-MM-dd||yyyy-MM||yyyy"}}},{"term":{"tag":"tag3"}}]}}}
格式化
{
    "highlight": {
        "fields": {
            "content": {},
            "title": {}
        },
        "post_tags": [
            ""
        ],
        "pre_tags": [
            ""
        ]
    },
    "query": {
        "dis_max": {
            "queries": [
                {
                    "match": {
                        "title": {
                            "boost": 4,
                            "minimum_should_match": "50%",
                            "query": "天气"
                        }
                    }
                },
                {
                    "match": {
                        "content": {
                            "boost": 4,
                            "minimum_should_match": "75%",
                            "query": "天气"
                        }
                    }
                },
                {
                    "match": {
                        "tag": {
                            "boost": 2,
                            "minimum_should_match": "100%",
                            "query": "天气"
                        }
                    }
                },
                {
                    "match": {
                        "slug": {
                            "boost": 1,
                            "minimum_should_match": "100%",
                            "query": "天气"
                        }
                    }
                }
            ],
            "tie_breaker": 0.3
        }
    },
    "filter": {
        "bool": {
            "must": [
                {
                    "range": {
                        "date": {
                            "gte": "2016-10",
                            "lte": "2016-10||/M",
                            "format": "yyyy-MM-dd||yyyy-MM||yyyy"
                        }
                    }
                },
                {
                    "term": {
                        "tag": "tag3"
                    }
                }
            ]
        }
    }
}
```

#### term 查询

```
{
    "query": {
        "bool": {
            "must": [
                {
                    "term": {
                        "slug": "slug1"
                    }
                },{
                	"term": {
                		"tag": "tag1"
                	}
                }
            ]
        }
    },
    "filter": {
        "range": {
            "date": {
                "gte": "2016-10",
                "lte": "2016-10||/M",
                "format": "yyyy-MM||yyyy"
            }
        }
    }
}
```

