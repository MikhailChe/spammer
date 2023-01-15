package main

import (
	"flag"
	"fmt"
	"io"
	"net/mail"
	"os"
	"strings"
)

func ERR(s any) {
	fmt.Println(s)
	flag.Usage()
	os.Exit(1)
}

func main() {
	var err error
	var conf = new(Config)

	configFile := flag.String("config", "", "Обязательный параметр. YAML файл с конфигурацией. Можно указать несуществующий файл, тогда программа создаст на его месте файл-пример, котоырй нужно будет отредактировать.")
	bodyFile := flag.String("body-file", "", "Файл с телом сообщения. Порядок: берём тело из конфигурации (параметр body), если его нет - берём из указанного файла. Тело должно быть, либо в конфигурации, либо в отдельном файле")
	toFile := flag.String("to-file", "", "Файл со списком получателей сообщения. По умолчанию берём список из конфигурации, если его нет - берём из указанного файла. Предполагается текстовый файл, где каждый email на новой строчке.")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Использование: %s [-body-file путь/до/файла] [-to-file путь/до/файла] -config <путь/до/конфигурации.yaml>\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if configFile == nil || len(*configFile) == 0 {
		ERR("Нужно обязательно передать путь до файла конфигурации.")
	}

	if bodyFile != nil && len(*bodyFile) > 0 {
		f, err := os.Open(*bodyFile)
		if err != nil {
			ERR(err)
		}
		bb, err := io.ReadAll(f)
		if err != nil {
			ERR(err)
		}
		conf.Body = string(bb)
	}

	if toFile != nil && len(*toFile) > 0 {
		f, err := os.Open(*toFile)
		if err != nil {
			ERR(err)
		}
		bb, err := io.ReadAll(f)
		if err != nil {
			ERR(err)
		}
		dirtyEmails := strings.Split(string(bb), "\n")
		var emails []string
		for _, email := range dirtyEmails {
			email = strings.Trim(email, "\r\n\t")
			email = strings.TrimSpace(email)
			if email == "" {
				continue
			}
			emails = append(emails, email)
		}
		conf.To = emails
	}

	err = loadConfig(conf, *configFile)
	if err != nil {
		ERR(err)
	}

	sendSmtp(Mail{
		Connection: Connection{
			Host:     conf.Connection.Host,
			Port:     conf.Connection.Port,
			Username: conf.Connection.User,
			Password: conf.Connection.Password,
		},
		From: mail.Address{
			Name:    conf.From.Name,
			Address: conf.From.Address,
		},
		To: conf.To,
		Message: Message{
			Subject:     conf.Subject,
			ContentType: conf.ContentType,
			Body:        conf.Body,
		},
	})
}
