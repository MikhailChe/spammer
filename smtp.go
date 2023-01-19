package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/mail"
	"net/smtp"
	"strings"
)

type Mail struct {
	Connection Connection
	From       mail.Address
	Recipients []string
	Message    Message
}

type Connection struct {
	Host     string
	Port     int
	Username string
	Password string
}

type Message struct {
	Subject     string
	Body        MessageBody
	Attachments []Attachment
}

type MessageBody struct {
	ContentType string
	Body        string
}

type Attachment struct {
	Filename    string
	ContentType string
	Body        []byte
}

func sendSmtp(m Mail) {
	server := fmt.Sprintf("%s:%d", m.Connection.Host, m.Connection.Port)

	auth := smtp.PlainAuth("", m.Connection.Username, m.Connection.Password, m.Connection.Host)

	buf := new(bytes.Buffer)
	mpart := multipart.NewWriter(buf)

	headers := make(http.Header)
	headers.Set("From", m.From.String())
	headers.Set("Subject", m.Message.Subject)
	headers.Set("Content-Type", fmt.Sprintf("multipart/mixed; boundary=%s", mpart.Boundary()))

	headers.Write(buf)
	buf.WriteString("\n")
	m.Message.WriteMultipart(buf, mpart)

	Message := buf.Bytes()

	tlsConfig := tls.Config{
		InsecureSkipVerify: true,
		ServerName:         m.Connection.Host,
	}

	fmt.Println("Устанавливаю безопасное соединение по адресу", server)
	conn, err := tls.Dial("tcp", server, &tlsConfig)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Начинаю общение по SMTP")
	c, err := smtp.NewClient(conn, m.Connection.Host)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Прохожу авторизацию")
	err = c.Auth(auth)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Отправляю письмо от", m.From.Address)
	err = c.Mail(m.From.Address)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Письма будут отправлены следующим получателям:", strings.Join(m.Recipients, ","))
	for _, to := range m.Recipients {
		err = c.Rcpt(to)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	w, err := c.Data()
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = w.Write([]byte(Message))
	if err != nil {
		fmt.Println(err)
		return
	}

	err = w.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Закрываю соединение с SMTP сервером")

	err = c.Quit()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Сообщения отправлены")
}
