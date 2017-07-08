// Package logd provides ...
package logd

import "testing"

func TestSmtpSendMail(t *testing.T) {
	s := &Smtp{
		From:    "120735581@qq.com",
		Key:     "peerdmnoqirqbiaa",
		Host:    "smtp.qq.com",
		Port:    "465",
		To:      []string{"a120735581@foxmail.com"},
		Subject: "test email from logd",
	}

	err := s.SendMail("test", []byte("hello world"))
	if err != nil {
		t.Error(err)
	}
}
