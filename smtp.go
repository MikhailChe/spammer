package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/mail"
	"net/smtp"
	"strings"
)

type Mail struct {
	Connection Connection
	From       mail.Address
	To         []string
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
	ContentType string
	Body        string
}

func sendSmtp(m Mail) {
	server := fmt.Sprintf("%s:%d", m.Connection.Host, m.Connection.Port)

	auth := smtp.PlainAuth("", m.Connection.Username, m.Connection.Password, m.Connection.Host)

	headers := make(http.Header)
	headers.Set("From", m.From.String())
	headers.Set("Subject", m.Message.Subject)
	headers.Set("Content-Type", m.Message.ContentType)

	buf := new(bytes.Buffer)
	headers.Write(buf)
	buf.WriteString("\r\n")
	buf.WriteString(m.Message.Body)
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

	fmt.Println("Письма будут отправлены следующим получателям:", strings.Join(m.To, ","))
	for _, to := range m.To {
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

	fmt.Println("===============\n", string(Message), "\n===============")
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
