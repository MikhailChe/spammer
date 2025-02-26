package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
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
	"time"

	"github.com/pkg/errors"
)

var version = "dev-build"

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

func splitHeadersBody(contentType string, msg string) (string, string) {
	// Пытаемся найти загловки в теле письма
	rnrn := strings.Index(msg, "\r\n\r\n")
	nn := strings.Index(msg, "\n\n")

	var firstSplit = -1
	if rnrn > 0 {
		firstSplit = rnrn
	}
	if nn > 0 && nn < rnrn {
		firstSplit = nn
	}

	if firstSplit < 0 {
		return contentType, msg
	}

	tp := textproto.NewReader(bufio.NewReader(strings.NewReader(msg)))
	headers, err := tp.ReadMIMEHeader()
	if err != nil {
		return contentType, msg
	}
	if len(headers.Get("Content-Type")) == 0 {
		return contentType, msg
	}
	contentType = headers.Get("Content-Type")
	msg = msg[firstSplit:]
	return contentType, msg
}

func main() {
	fmt.Println(`Версия приложения:`, version)
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

	if conf.Delay != 0 {
		fmt.Printf("Письма будут отправлены по одному с интервалом %v\n", conf.Delay)
	}

	fmt.Println("Конфигурация для отладки")

	if bb, err := json.MarshalIndent(conf, "  ", "  "); err == nil {
		fmt.Println(string(bb))
		fmt.Println()
	}

	// Body
	conf.BodyFile, err = filepath.Abs(conf.BodyFile)
	if err != nil {
		ERR(err)
	}
	var msgBody MessageBody
	{
		f, err := os.Open(conf.BodyFile)
		if err != nil {
			ERR(errors.Wrap(err, "Не могу открыть файл с телом письма"))
		}
		bb, err := io.ReadAll(f)
		if err != nil {
			ERR(err)
		}
		_ = f.Close()
		msgBody.ContentType = conf.ContentType
		msgBody.Body = string(bb)

		if strings.HasSuffix(conf.BodyFile, ".mhtml") || strings.HasSuffix(conf.BodyFile, ".mht") {
			msgBody.ContentType, msgBody.Body = splitHeadersBody(msgBody.ContentType, msgBody.Body)
		}

		if msgBody.ContentType == "" && (strings.HasSuffix(conf.BodyFile, ".html") || strings.HasSuffix(conf.BodyFile, ".htm")) {
			msgBody.ContentType = "text/html"
		}

		if msgBody.ContentType == "" && strings.HasSuffix(conf.BodyFile, ".txt") {
			msgBody.ContentType = "text/plain"
		}
	}

	if len(msgBody.Body) == 0 {
		ERR("Тело письма не должно быть пустым")
	} else {
		fmt.Println("Нашел тело письма в файле", conf.BodyFile)
	}

	if len(msgBody.ContentType) == 0 {
		ERR("Не получилось определить тип сообщения. Укажите его в параметре content-type. Например, text/plain или text/html")
	}

	// Attachments
	var attachments []Attachment
	for _, attFile := range conf.Attachments {
		attFile.File, err = filepath.Abs(attFile.File)
		if err != nil {
			ERR(err)
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
		_ = f.Close()
	}

	// Получатели
	conf.ToFile, err = filepath.Abs(conf.ToFile)
	if err != nil {
		ERR(err)
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

		emailSet := map[string]struct{}{}

		for _, email := range dirtyEmails {
			email = strings.Trim(email, "\r\n\t")
			email = strings.TrimSpace(email)
			if email == "" {
				continue
			}
			if _, exists := emailSet[email]; exists {
				ERR(fmt.Sprintln("Найден дубликат почтового ящика в списке получателей:", email))
			} else {
				emailSet[email] = struct{}{}
			}
			recipients = append(recipients, email)
		}
		_ = f.Close()
	}

	if len(recipients) == 0 {
		ERR("Нужно добавить хотя бы одного получателя")
	}

	for _, recipient := range recipients {
		err = sendSmtp(Mail{
			Connection: Connection{
				Host:     conf.Connection.Host,
				Port:     conf.Connection.Port,
				Username: conf.Connection.User,
				Password: conf.Connection.Password,
				SSL:      conf.Connection.SSL,
			},
			From: mail.Address{
				Name:    conf.From.Name,
				Address: conf.From.Address,
			},
			Recipients: []string{recipient},
			Message: Message{
				Subject:     conf.Subject,
				Body:        msgBody,
				Attachments: attachments,
			},
		})
		if err != nil {
			ERR(err)
		}
		fmt.Printf("Ожидаю %v перед следующей отправкой\n", conf.Delay)
		time.Sleep(conf.Delay)
	}
	fmt.Println(`Приложение для отправки почты нескольким адресатам.
Автор: Черноскутов Михаил <mikhail-chernoskutov@yandex.ru>`)
}
