# EiBlog [![Build Status](https://travis-ci.org/eiblog/eiblog.svg?branch=master)](https://travis-ci.org/eiblog/eiblog)

> 系统根据[https://imququ.com](https://imququ.com)一系列文章和方向进行搭建，期间获得了QuQu的很大帮助，在此表示感谢。

用过其它博客系统，不喜欢，不够轻，不够快！自己做过共两款博客系统，完美主义的我（毕竟处女座）也实在是不想再在这件事情上过多纠结了。`Eiblog`应该是一个比较稳定的博客系统，且是博主以后使用的博客系统，稳定性和维护你是不用担心的，唯独该系统部署过程太过复杂，并且不推荐没有计算机知识的朋友搭建，欢迎咨询。该博客的个中优点（明显快，安全），等你体验。

<!--more-->

### 介绍
整个博客系统涉及到模块如下：

* `MongoDB`，博客采用 mongodb 作为存储数据库。
* `Elasticsearch`，采用`elasticsearch`作为博客的站内搜索，尽管占用内存稍高。
* `Disqus`，作为博客评论系统，国内大部分被墙，故实现两种评论方式。
* `Nginx`，作为反向代理服务器，并做相关`http header`和证书的设置。
* `Google Analytics`，作为博客系统的数据分析统计工具。
* `七牛 CDN`，作为博客系统的静态文件存储，博文的图片附件什么上传至这里。

相关技术有：

* `Golang`，博客系统后端采用golang编写，并开源至[Eiblog](https://github.com/eiblog/eiblog)。
* `HTML Javascript CSS`，博客系统的前端采用`html`和`jquery`编写，样式采用`CSS`。
* `Glide`， golang 编写。作为博客系统的包依赖管理器，其开源地址是[Glide](https://github.com/Masterminds/glide)。
* `Docker`，博客系统可 docker 部署，方便，快捷。
* `Docker Compose`，博客系统可完全 docker 运行，compose起到很好管理作用。
* `SSL 证书`，`https`是未来的趋势，整个博客系统都将围绕着`证书`进行，请事先准备好一张有效的 ssl 证书。
* `Travis`，作为博客系统的自动构建工具，自动构建docker镜像并推送到镜像仓库。
* `Yaml`，博客系统的配置文件使用`yaml`，请悉知。

作为博主之心血之作，`Eiblog`实现了什么功能，有什么特点，做了什么优化呢？

1. 系统目前只有`首页`、`专题`、`归档`、`友链`、`关于`、`搜索`界面。相信已经可以满足大部分用户的需求。
2. `.js`、`.css`等静态文件本地存储，小图片 base64 内置到 css 中，不会产生网络所带来的延迟，加速网页访问。版本控制方式，动态更新静态文件。
3. 采用谷歌统计，并实现异步（将访问信息发给后端，后端提交给谷歌）统计，加速访问速度。
4. 采用直接缓存 markdown 转过的 html 文档的方式，加速后端处理。响应速度均在 3ms 以内，真正极速。
5. 通过 Nginx 的配置，开启压缩缩小传输量，服务器传输证书链、开启`Session Resumption`、`Session Ticket`、`OCSP Stapling`等加速证书握手，再次提高速度。
6. 文章评论数量（不重要）后端跑定时脚本，定时更新，所以有时评论数是不对的。这样减少了 api 调用，又再次达到加速访问的目的。
7. 针对`disqus`被墙原因，实现[Jerry Qu](https://imququ.com)的另类评论方式，保证评论的流畅。
8. 开源`Typecho`完整后台系统，全功能`markdown`编辑器，让你体验什么是简洁清爽。
9. 博客后台直接对接`七牛 SDK`，实现后台上传文件和删除文件的简单功能。
10. 采用`elasticsearch`作为站内搜索，添加`google opensearch`功能，搜索更加自然。

当然，在信息安全方面也没少下功夫，虽然我们只是一个小小的博客系统。

1. `CDN`，使用七牛融合CDN，并`https`化，实现全站`https`。七牛可申请免费证书了。
2. `CT`，证书透明度检测，提供一个开放的审计和监控系统。可以让任何域名所有者或者 CA 确定证书是否被错误签发或者被恶意使用，从而提高 HTTPS 网站的安全性。
3. `OSCP`，在线证书状态协议。用来检验证书合法性的在线查询服务.
4. `HSTS`，强制客户端（如浏览器）使用 HTTPS 与服务器创建连接。可以很好的解决 HTTPS 降级攻击。
5. `HPKP`，HTTP公钥固定扩展，防范由「伪造或不正当手段获得网站证书」造成的中间人攻击。该功能让我们选择信任哪些`CA`。
5. `SSL Protocols`，罗列支持的`TLS`协议，SSLv3被证实是不安全的。
6. `SSL dhparam`，迪菲赫尔曼密钥交换。
7. `Cipher suite`，罗列服务器支持加密套件。

可以容易的看到[httpsecurityreport](https://httpsecurityreport.com/?report=deepzz.com)评分`96`，[ssllabs](https://www.ssllabs.com/ssltest/analyze.html?d=deepzz.com&latest)评分`A+`，堪称完美。这些安全的相关配置会在后面的部署过程中接触到。

相关图片展示：
![show-home](http://7xokm2.com1.z0.glb.clouddn.com/static/img/show-home1.png)

![show-home2](http://7xokm2.com1.z0.glb.clouddn.com/static/img/show-home2.png)

![show-admin](http://7xokm2.com1.z0.glb.clouddn.com/static/img/show-admin.png)

![show-time](http://7xokm2.com1.z0.glb.clouddn.com/static/img/eiblog-time.png)

> `注`：图片1，图片2是博客界面，图片3是后台界面，图片4是性能展示。

好了，说了那么多，吹了那么多，我们实际来动手搭建一个`Eiblog`吧。

### 安装
1、`Eiblog`提供多个平台的压缩包下载，可到[Eiblog release](https://github.com/eiblog/eiblog/releases)选择相应版本和平台下载。也可通过：
``` sh
$ curl -L https://github.com/eiblog/eiblog/releases/download/v0.1.0/eiblog-v0.1.0.`uname -s | tr '[A-Z]' '[a-z]'`-amd64.tar.gz > eiblog-v0.1.0.`uname -s | tr '[A-Z]' '[a-z]'`-amd64.tar.gz
```

2、如果有幸你也是`Gopher`，相信你会亲自动手，你可以通过：
``` sh
$ go get https://github.com/eiblog/eiblog
```
进行源码编译二进制文件运行。

3、如果你对`docker`技术也有研究的话，你也可以通过`docker`来安装：
``` sh
$ docker pull registry.cn-hangzhou.aliyuncs.com/deepzz/eiblog

```
镜像内部只提供了`eiblog`的二进制文件，因为其它内容定制化的需求过高。所以需要将`conf`、`static`、`views`目录映射出来，后面会具体说到。

### 本地测试
在我们下载好可执行程序之后，我们可以开始本地测试的工作了。

本地测试需要搭建两个服务`mongodb`和`elasticsearch2.4.1`（可选，搜索服务不可用）。

`Eiblog`默认会连接`hostname`为`eidb`和`eisearch`，因此你需要将信息填入`/etc/hosts`下。假如你搭建的`mongodb`地址为`127.0.0.1:27017`，`elasticsearch`地址为`192.168.99.100:9200`，如：
``` sh
$ sudo vi /etc/hosts

# 在末尾加上两行
127.0.0.1       eidb
192.168.99.100  eisearch
```

#### MongoDB 搭建
1、`MongoDB`搭建，Mac 可通过`brew install mongo`进行安装，其它平台请查询资料。
#### Elasticsearch 搭建
2、`Elasticsearch`搭建，它的搭建要些许复杂。博主尚未接触如何直接安装，因此建议通过`docker`搭建。需要注意的是 es 自带的分析器对中文分词是不友好的，这里采用了`elasticsearch-analysis-ik`分词器。如果你想了解更多[Github](https://github.com/medcl/elasticsearch-analysis-ik)或则如何实现[博客站内搜索](https://imququ.com/post/elasticsearch.html)。

* pull 镜像`docker pull elasticsearch:2.4.1`，必需使用该版本。
* 添加环境变量`ES_JAVA_OPTS: "-Xms512m -Xmx512m"`，除非你想让你的服务器爆掉。
* 映射相关目录：

  ```
  conf/es/config:/usr/share/elasticsearch/config
  conf/es/plugins:/usr/share/elasticsearch/plugins
  conf/es/data:/usr/share/elasticsearch/data
  conf/es/logs:/usr/share/elasticsearch/logs
  ```
  请将这四个目录映射至`eiblog`下的`conf`目录。如果你想查看更多，请查看`docker-compose.yml`文件。

总结一下，`docker`运行 es 的命令为：
``` sh
$ docker run -d --name eisearch \
    -p 9200:9200 \
    -e ES_JAVA_OPTS: "-Xms512m -Xmx512m" \
    -v conf/es/config:/usr/share/elasticsearch/config \
    -v conf/es/plugins:/usr/share/elasticsearch/plugins \
    -v conf/es/data:/usr/share/elasticsearch/data \
    -v conf/es/logs:/usr/share/elasticsearch/logs \
    elasticsearch:2.4.1
```

之后执行`./eiblog`，咱们的`eiblog`就可以运行起来了。

通过`127.0.0.1:9000`可以进入博客首页，`127.0.0.1:9000/admin/login`进入后台登陆，账号密码为`eiblog/conf/app.yml`下的`username`和`password`。也就是初始账号密码`deepz`、`deepzz`。

> `注意`，因为配置`conf/app.yml`均是博主自用配置。有些操作可能（如评论）会评论到我的博客，还请尽量避免，谢谢。

### 准备部署
如果你在感受了该博客的魅力了之后，仍然坚持想要搭建它。那么，恭喜你，获得的一款不想再更换的博客系统。下面，我们跟随步骤对部署流程进一步说明。

这里只提供`Docker`的相关部署说明。你如果需要其它方式部署，请参考该方式。

#### 前提准备
这里需要准备一些必要的东西，如果你已准备好。请跳过。

* `一台服务器`。
* `一个域名`，国内服务器需备案。
* `有效的证书`。一般使用免费的就可以。如：`Let‘s Encrypt`，另外`qcloud`、`七牛`也提供了免费证书的申请，均是全球可信。
* `七牛CDN`。博客只设计接入了七牛cdn，相信该CDN服务商不会让你失望。
* `Disqus`。作为博客评论系统，你得有翻墙的能力注册到该账号，具体配置我想又可以写一片博客了。简单说需要`shorname`和`public key`。
* `Google Analystic`。数据统计分析工具。
* `Superfeedr`。加速 RSS 订阅。
* `Twitter`。希望你能够有一个twitter账号。

是不是这么多要求，很费解。其实当初该博客系统只是为个人而设计的，是自己心中想要的那一款。博主些这篇文章不是想要多少人来用该博客，而是希望对那些追求至极的朋友说：你需要这款博客系统。
#### 文件准备
尽管大多数文件已经准备好。但有些默认的文件需要特别指出来，需要你在 CDN 上写特殊的路径。

假如你的 CDN 域名为`st.example.com`，那么：

* `favicon.ico`，其 URL 应该是`st.example.com/static/img/favicon.ico`。故你在 CDN 中的文件名为`static/img/favicon.ico`，以下如是。
* `左侧背景图片`，`500*1200`左右，CDN 中文件名：`static/img/bg04.jpg`。如需更改，请在`eiblog/view/st_blog.css`中替换该名称。
* `头像`，`160*160~256*256`之间，CDN 文件名：`static/img/avatar.jpg`。另外你需要将该图片 `Base64` 编码后替换掉`eiblog/views/st_blog.css`中合适位置的图片。
* `blank.gif`，CDN 文件名：`static/img/blank.gif`。该图片请从[这里](https://st.deepzz.com/static/img/blank.gif)下载并上传至你的 CDN。
* `default_avatar.png`，CDN 文件名：`static/img/default_avatar.png`，请从[这里](https://st.deepzz.com/static/img/default_avatar.png)下载并上传至你的 CDN。
* `disqus.js`，该文件名是会变的，每次更新如果没有提及就没有改变，更新说明在[这里](https://github.com/eiblog/eiblog/blob/master/CHANGELOG.md)。CDN 文件名格式是：`static/js/name.js`。在我写这篇文章是使用的是：`static/js/disqus_a9d3fd.js`，请从[这里](https://st.deepzz.com/static/js/disqus_a9d3fd.js)下载并上传至你的 CDN。

> `注意`：本人 CDN 做了防盗链处理，故请将这些资源上传至您的 CDN ，以免静态资源不能访问，请悉知。

#### 配置说明
走到这里，我相信只走到`60%`的路程。放弃还来得及。

这里会对`eiblog/conf`下的所有文件做说明，希望你做好准备。
```
├── app.yml                         # 博客配置文件
├── blackip.yml                     # 博客ip黑名单
├── es                              # elasticsearch配置
│   ├── config                      # 配置文件
│   │   ├── analysis                # 同义词
│   │   ├── elasticsearch.yml       # 具体配置
│   │   ├── logging.yml             # 日志配置
│   │   └── scripts                 # 脚本文件夹
│   └── plugins                     # 插件文件夹
│       └── ik1.10.1                # ik分词器
├── nginx                           # nginx配置
│   ├── domain                      # 域名配置，nginx会读区改文件夹下的.conf文件
│   │   └── deepzz.conf
│   ├── ip.blacklist                # nginx ip黑名单
│   └── nginx.conf                  # nginx配置，请替换原有配置
├── scts                            # ct文件
│   ├── aviator.sct
│   └── digicert.sct
├── ssl                             # 证书文件，具体请看deepzz.conf
│   ├── dhparams.pem
│   ├── domain.key
│   ├── domain.pem
│   ├── full_chained.pem
│   └── session_ticket.key
└── tpl                             # 模版文件
    ├── feedTpl.xml
    ├── opensearchTpl.xml
    └── sitemapTpl.xml

```
1、app.yml，整个程序的配置文件，里面已经列出了所有配置项的说明，这里不再阐述。  
2、blackip.yml，如果没有使用`Nginx`，博客内置`ip`过滤系统。  
3、`es`全名`elasticsearch`，非常强大的分布式搜索引擎，`github`用的就是它。里面的配置基本不用修改，但`es/analysis/synonym.txt`是同义词，你可以照着已有的随意增加。
```
├── es
│   ├── config
│   │   ├── analysis
│   │   │   └── synonym.txt                 #同义词配置
│   │   ├── elasticsearch.yml               #分词器配置
│   │   ├── logging.yml                     #日志配置
│   │   └── scripts                         #脚本
│   └── plugins                             #中文分词插件
│       └── ik1.10.0
│
```

> `注意`，scripts文件夹虽然是空的，但必需存在，不然elasticsearch报错。

4、`nginx`，系统采用`nginx`作为代理(相信博客系统也不会独占一台服务器～)。请使用`nginx.conf`替换原`nginx`的配置。博客系统的配置文件是`domain/deepzz.conf`，或则重命名(只要是满足`*.conf`)。`deepzz.conf`文件里面学问是最多的。或许你想一一弄懂，或许…。

> 注意本配置需要更新nginx到最新版，openssl更新到1.0.2j，具体请到 Jerry Qu 的[本博客 Nginx 配置之完整篇](https://imququ.com/post/my-nginx-conf.html)查看，了解详情。

5、`scts`，存放 ct 文件。

6、`ssl`，这里存放了所有证书相关的内容。
```
├── dhparams.pem                #参见eiblog/conf/nginx/domain/deepzz.conf
├── domain.key                  #证书私钥，一般颁发者处下载
├── domain.pem                  #证书链，一般从证书颁发者那可以直接下载到
├── full_chained.pem            #参见eiblog/conf/nginx/domain/deepzz.conf
└── session_ticket.key          #参见eiblog/conf/nginx/domain/deepzz.conf
```

7、`tpl`模版相关，不用修改。

### 开始部署

#### docker
请确定你已经完成了上面所说的所有步骤，在本地已经测试成功。服务器上`MognoDB`和`Elasticsearch`已经安装并已经运行成功。

首先，请将本地测试好的`conf`，`static`，`views`文件夹上传至服务器，建议存储到服务器`/data/eiblog`下。
``` sh
$ tree /data/eiblog -L 1

├── conf
├── static
├── views
```

然后，将镜像 PULL 到服务器本地。
``` sh
# PULL下Eiblog镜像
$ docker pull registry.cn-hangzhou.aliyuncs.com/deepzz/eiblog
```

最后，执行`docker run`命令，希望你能成功。
``` sh
$ docker run -d --name eiblog --restart=always \
    --add-host disqus.com:23.235.33.134 \
    --link eidb --link eisearch \
    -p 9000:9000 \
    -e GODEBUG=netdns=cgo \
    -v /data/eiblog/logdata:/eiblog/logdata \
    -v /data/eiblog/conf:/eiblog/conf \
    -v /data/eiblog/static:/eiblog/static \
    -v /data/eiblog/views:/eiblog/views \
    registry.cn-hangzhou.aliyuncs.com/deepzz/eiblog
```
这里默认`MongDB`和`Elasticsearch`均为`docker`部署，且名称为`eidb`，`eisearch`。

#### nginx + docker
通过`Nginx+docker`部署，是博主推荐的方式。这里采用`Docker Compose`管理我们整个博客系统。

请确认你已经成功安装好`Nginx`、`docker`、`docker-compose`。Nginx 请一定参照 Jerry Qu 的[Nginx 配置完整篇](https://imququ.com/post/my-nginx-conf.html)。

首先，请将本地测试好的`conf`，`static`，`views`，`docker-compose.yml`文件夹和文件上传至服务器。前三个文件夹建议存储到服务器`/data/eiblog`下，`docker-compose.yml`存放在你使用方便的地方。

> 注意`conf/es/config/scripts`空文件夹是否存在

``` sh
$ tree /data/eiblog -L 1

├── conf
├── static
├── views

$ ls ~/

docker-compose.yml
```

然后，执行：
``` sh
$ cd ~
$ docker-compose up -d
```

等待些许时间，成功运行。


### 成功搭建者博客

* [https://razeencheng.com/](https://razeencheng.com/) - Razeen's Blog

如果你的博客使用`Eiblog`搭建，你可以在[这里](https://github.com/eiblog/eiblog/issues/1)提交网址。
