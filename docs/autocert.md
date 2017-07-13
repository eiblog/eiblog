### 证书自动更新

本博客证书自动更新有两种方式：

* [acme/autocert](https://github.com/golang/crypto/tree/master/acme/autocert)，博客内部集成，通过 tls-sni 验证，实现全自动更新证书，一键开启关闭。请在裸服务器的情况下使用（不要使用代理）。单证书
* [acme.sh](https://github.com/Neilpang/acme.sh)，强大的 acme 脚本，多种自动更新证书方式，满足你各方面的需求。双证书

#### 方式一
什么是 autocert，简单点，你只需要两步操作：

1. 将域名解析到你的服务器。
2. 在服务器上运行开启 autocert 功能的程序（这里不需要配置证书），需要占用 443 端口。

其它过程你不需要过问，即会完成自动申请证书，自动更新证书的功能（默认 30 天）。这个是在 tcp/ip 层的操作，对用户完全透明，非常棒。

一键开启 autocert 功能，只需修改 `conf/app.yml` 文件内容：

```
# 运行模式
mode:
  # http server
  enablehttp: true
  httpport: 9000
  # https server
  enablehttps: true                     # 必须开启
  autocert: false                       # autocert 功能开关
  httpsport: 9001
  certfile: 
  keyfile: 
  domain: deepzz.com                    # 申请证书的域名，也是博客的域名
```

首先，使用 HTTPS 必须启用  `enablehttps`，它有两个作用：

* 如果 `enablehttp` 开启，会自动 301 重定向到 https。
* 作为开启 autocert 的前提条件。

其次， `autocert` 是否开启也有两个作用：

* false，服务器将使用 `httpsport`、`certfile`、`keyfile` 作为参数启动 https 服务器。
* true，服务器直接使用 443 端口启动 https 服务器，并且自动申请证书，且在证书只有 30 天有效期时自动更新证书。域名为 *运行模式* 下的 mode->domain。

#### 方式二

使用方式二，你需要了解 acme.sh 的具体使用方式，非常简单。选择适合自己的方式实现自动更新证书。

博主，这里实现了 aliyun dns 的自动验证，自动更新证书。详情参见 Makefile->gencert。这里实现了自动申请 ecc、rsa 双证书，并且自动申请 scts，自动安装，自动更新。

基本流程如下：

1. 创建相关目录：`/data/eiblog/conf/ssl`、`/data/eiblog/conf/scts/rsa`、`/data/eiblog/conf/scts/ecc`。
2. 自动下载安装 acme.sh 脚本。
3. 自动申请 RSA 证书并且自动获取 scts，并且自动安装到指定位置。
4. 自动申请 ECC 证书并且自动获取 scts，并且自动安装到指定位置。

##### 使用方式

导出环境变量，Aliyun dns 的环境变量为：

```
export Ali_Key="sdfsdfsdfljlbjkljlkjsdfoiwje"
export Ali_Secret="jlsdflanljkljlfdsaklkjflsa"
```

执行命令：

```
$ make gencert -cn=common_name -sans="-d example.com  -d example1.com"
```
