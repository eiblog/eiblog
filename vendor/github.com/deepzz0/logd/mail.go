package logd

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
)

type Emailer interface {
	SendMail(fromname string, msg []byte) error
}

type Smtp struct {
	From    string   // 发件箱：number@qq.com
	Key     string   // 发件密钥：peerdmnoqirqbiaa
	Host    string   // 主机地址：smtp.example.com
	Port    string   // 主机端口：465
	To      []string // 发送给：object@163.com
	Subject string   // 标题：警告邮件[goblog]
}

func (s *Smtp) SendMail(fromname string, msg []byte) error {
	// 新建连接
	conn, err := tls.Dial("tcp", s.Host+":"+s.Port, nil)
	if err != nil {
		return err
	}

	// 新建客户端
	client, err := smtp.NewClient(conn, s.Host)
	if err != nil {
		return err
	}

	// 获取授权
	auth := smtp.PlainAuth("", s.From, s.Key, s.Host)
	if err = client.Auth(auth); err != nil {
		return err
	}

	// 向服务器发送MAIL命令
	if err = client.Mail(s.From); err != nil {
		return err
	}

	// 准备数据
	str := fmt.Sprint(
		"To:", strings.Join(s.To, ","),
		"\r\nFrom:", fmt.Sprintf("%s<%s>", fromname, s.From),
		"\r\nSubject:", s.Subject,
		"\r\n", "Content-Type:text/plain;charset=UTF-8",
		"\r\n\r\n",
	)
	data := make([]byte, len(str)+len(msg))
	copy(data, []byte(str))
	copy(data[len(str):], msg)

	// RCPT
	for _, d := range s.To {
		if err := client.Rcpt(d); err != nil {
			return err
		}
	}

	// 获取WriteCloser
	wc, err := client.Data()
	if err != nil {
		return err
	}

	// 写入数据
	_, err = wc.Write(data)
	if err != nil {
		return err
	}
	wc.Close()

	return client.Quit()
}
