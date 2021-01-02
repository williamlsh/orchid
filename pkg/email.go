/*
 * @Descripttion:
 * @version:
 * @Author: dengdajun
 * @Date: 2021-01-02 19:06:47
 * @LastEditors: dengdajun
 * @LastEditTime: 2021-01-03 01:04:57
 */
package pkg

import (
	"log"
	"net/smtp"
	"strings"
)

func SendMail(username, password, host, to, name, subject, body, mailType string) error {
	hp := strings.Split(host, ":")
	auth := smtp.PlainAuth("", username, password, hp[0])
	var contentType string
	if mailType == "html" {
		contentType = "Content-Type: text/" + mailType + "; charset=UTF-8"
	} else {
		contentType = "Content-Type: text/plain" + "; charset=UTF-8"
	}
	msg := []byte("To: " + to + "\r\nFrom: " + name + "<" + username + ">\r\nSubject: " + subject + "\r\n" + contentType + "\r\n\r\n" + body)
	sendTo := strings.Split(to, ";")
	err := smtp.SendMail(host, auth, username, sendTo, msg)
	return err
}

func main() {
	err := SendMail(
		"1163388086@qq.com",
		"hwyjaqjakqadhahb",
		"smtp.qq.com:25",
		"18000517398@163.com",
		"orchid",
		"测试",
		"这是一封测试邮件",
		"html",
	)
	if err != nil {
		log.Fatal(err)
	}
}
