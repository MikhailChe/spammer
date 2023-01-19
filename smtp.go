package main

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"mime/multipart"
	"net"
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
	SSL      bool
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
type loginAuth struct {
	username, password string
}

func LoginAuth(username, password string) smtp.Auth {
	return &loginAuth{username, password}
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte{}, nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			return nil, errors.New("unkown fromServer")
		}
	}
	return nil, nil
}

func sendSmtp(m Mail) error {
	var err error
	server := fmt.Sprintf("%s:%d", m.Connection.Host, m.Connection.Port)

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

	var conn net.Conn
	if m.Connection.SSL {
		fmt.Println("Устанавливаю безопасное соединение по адресу", server)
		conn, err = tls.Dial("tcp", server, &tlsConfig)
		if err != nil {
			return err
		}
	} else {
		fmt.Println("Указан порт 25. Предполагаю, что соединение небезопасное. Устанавливаю соединение по адресу", server)
		conn, err = net.Dial("tcp", server)
		if err != nil {
			return err
		}
	}

	fmt.Println("Начинаю общение по SMTP")
	c, err := smtp.NewClient(conn, m.Connection.Host)
	if err != nil {
		return err
	}

	fmt.Println("Прохожу авторизацию")
	auth := &loginAuth{m.Connection.Username, m.Connection.Password}
	err = c.Auth(auth)
	if err != nil {
		return err
	}

	fmt.Println("Отправляю письмо от", m.From.Address)
	err = c.Mail(m.From.Address)
	if err != nil {
		return err
	}

	fmt.Println("Письма будут отправлены следующим получателям:", prettyPrintRecepients(m.Recipients))
	for _, to := range m.Recipients {
		err = c.Rcpt(to)
		if err != nil {
			return err
		}
	}

	w, err := c.Data()
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(Message))
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	fmt.Println("Закрываю соединение с SMTP сервером")

	err = c.Quit()
	if err != nil {
		return err
	}

	fmt.Println("Сообщения отправлены")
	return nil
}

func prettyPrintRecepients(s []string) string {
	if len(s) > 10 {
		return fmt.Sprintf("%s и ещё %d", strings.Join(s[:10], ","), len(s)-10)
	}
	return strings.Join(s, ",")
}
