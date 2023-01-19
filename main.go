package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/mail"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

func ERR(s any) {
	fmt.Println(s)
	flag.Usage()
	os.Exit(1)
}

func (m *Message) WriteMultipart(buf io.Writer, mpart *multipart.Writer) error {
	mainPartHeader := make(textproto.MIMEHeader)
	mainPartHeader.Set("Content-Type", m.Body.ContentType)

	part, err := mpart.CreatePart(mainPartHeader)
	if err != nil {
		return err
	}

	part.Write([]byte(m.Body.Body))

	for _, att := range m.Attachments {
		if err := attachmentAsPart(mpart, att); err != nil {
			return err
		}
	}

	return mpart.Close()
}

func attachmentAsPart(mm *multipart.Writer, a Attachment) error {
	mimeHeaders := make(textproto.MIMEHeader)
	mimeHeaders.Set("Content-Type", a.ContentType)
	mimeHeaders.Set("Content-Transfer-Encoding", "base64")
	filenameEscaped := a.Filename
	filenameEscaped = strings.ReplaceAll(filenameEscaped, `\`, `\\`)
	filenameEscaped = strings.ReplaceAll(filenameEscaped, `"`, `\"`)
	mimeHeaders.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filenameEscaped))
	part, err := mm.CreatePart(mimeHeaders)
	if err != nil {
		return err
	}

	b64enc := base64.NewEncoder(base64.StdEncoding, part)
	b64enc.Write(a.Body)

	return nil
}

func main() {
	var err error
	var conf = new(Config)

	configFile := flag.String("config", "config.yaml", "YAML файл с конфигурацией. Можно указать несуществующий файл, тогда программа создаст на его месте файл-пример, котоырй нужно будет отредактировать.")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Использование: %s -config [путь/до/конфигурации.yaml]\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if configFile == nil || len(*configFile) == 0 {
		ERR("Нужно обязательно передать путь до файла конфигурации.")
	}

	err = loadConfig(conf, *configFile)
	if err != nil {
		ERR(err)
	}

	// Body
	conf.BodyFile, err = filepath.Abs(conf.BodyFile)
	if err != nil {
		ERR(err)
	}
	var body string
	{
		f, err := os.Open(conf.BodyFile)
		if err != nil {
			ERR(errors.Wrap(err, "Не могу открыть файл с телом письма"))
		}
		bb, err := io.ReadAll(f)
		if err != nil {
			ERR(err)
		}
		body = string(bb)
		f.Close()
	}

	if len(body) == 0 {
		ERR("Тело письма не должно быть пустым")
	} else {
		fmt.Println("Нашел тело письма в файле", conf.BodyFile)
	}

	// Attachments
	var attachments []Attachment
	for _, attFile := range conf.Attachments {
		attFile.File, err = filepath.Abs(attFile.File)
		if err != nil {
			ERR(errors.Wrapf(err, "Проблемы с приложением %s", attFile.File))
		}
		if len(attFile.File) == 0 {
			continue
		}
		f, err := os.Open(attFile.File)
		if err != nil {
			ERR(errors.Wrap(err, "Не могу открыть файл с приложением"))
		}
		bb, err := io.ReadAll(f)
		if err != nil {
			ERR(errors.Wrap(err, "Не могу прочитать файл с приложением"))
		}
		fmt.Println("Нашел приложение в файле", attFile.File)
		var filename string = attFile.Name
		if len(filename) == 0 {
			_, filename = filepath.Split(attFile.File)
		}
		fmt.Println("В письме будет отображаться как", filename)
		var contentType string = attFile.ContentType
		if len(contentType) == 0 {
			contentType = http.DetectContentType(bb)
			fmt.Println("Автоматически определенный тип файла:", contentType)
		}
		attachments = append(attachments, Attachment{filename, contentType, bb})
		f.Close()
	}

	var recipients []string
	{
		f, err := os.Open(conf.ToFile)
		if err != nil {
			ERR(errors.Wrap(err, "Не могу открыть файл со списком получаетелей"))
		}
		bb, err := io.ReadAll(f)
		if err != nil {
			ERR(err)
		}
		dirtyEmails := strings.Split(string(bb), "\n")

		for _, email := range dirtyEmails {
			email = strings.Trim(email, "\r\n\t")
			email = strings.TrimSpace(email)
			if email == "" {
				continue
			}
			recipients = append(recipients, email)
		}
		f.Close()
	}

	if len(recipients) == 0 {
		ERR("Нужно добавить хотя бы одного получателя")
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
		Recipients: recipients,
		Message: Message{
			Subject: conf.Subject,
			Body: MessageBody{
				ContentType: conf.ContentType,
				Body:        body,
			},
			Attachments: attachments,
		},
	})
}
