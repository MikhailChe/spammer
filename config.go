package main

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/heetch/confita"
	"github.com/heetch/confita/backend"
	"github.com/heetch/confita/backend/file"
)

var exampleConfig = `connection:
  host: smtp.yandex.ru
  port: 465 # по умолчанию 465, расчёт на SSL/TLS SMTP
  user: здесь имя пользвоателя
  password: нужен пароль
from:
  name: Василий Пупкин
  address: youremail@example.com
to-file: файл-с-получателями.txt # путь относительно текущей рабочей директории или абсолютный путь
subject: Тема письма
content-type: text/html; charset=utf-8
body-file: файл-с-телом-письма.html  # путь относительно текущей рабочей директории или абсолютный путь
attachments:
  - file: правила участия 2023-01-19.pdf  # путь относительно текущей рабочей директории или абсолютный путь
    name: "2023 - правила участия.pdf" # имя файла, которое будет отображаться у получателя
	content-type: application/pdf # формат содержимого. Можно не указывать, тогда программа попытается определить автоматически

  - file: регламент 2023-01-19.pdf  # путь относительно текущей рабочей директории или абсолютный путь
    name: "2023 - регламент.pdf" # имя файла, которое будет отображаться у получателя
	content-type: application/pdf # формат содержимого. Можно не указывать, тогда программа попытается определить автоматически
`

type Config struct {
	Connection struct {
		Host     string `config:"host,required"`
		Port     int    `config:"port,required"`
		User     string `config:"user,required"`
		Password string `config:"password,required"`
	} `config:"connection"`
	From struct {
		Name    string `config:"name,required"`
		Address string `config:"addess,requried"`
	} `config:"from"`
	ToFile      string `config:"to,required" yaml:"to-file"`
	Subject     string `config:"subject"`
	ContentType string `config:"content-type,required" yaml:"content-type"`
	BodyFile    string `config:"body,required" yaml:"body-file"`
	Attachments []struct {
		File        string `config:"file,required" yaml:"file"`
		Name        string `config:"name" yaml:"name"`
		ContentType string `config:"content-type" yaml:"content-type"`
	} `config:"attachments" yaml:"attachments"`
}

func loadConfig(into *Config, filename string) error {
	var err error
	absPath, err := filepath.Abs(filename)
	if err != nil {
		return err
	}
	loader := confita.NewLoader(file.NewBackend(absPath))
	err = loader.Load(context.Background(), into)
	if err != nil {
		cause := errors.Cause(err)
		fmt.Println(cause)
		_, isPathError := cause.(*fs.PathError)
		if cause == backend.ErrNotFound || isPathError {
			fmt.Printf("Ожидал файл конфигурации %s, но не нашел его там. Создаю файл для примера. Его нужно будет исправить и запустить программу с теми же параметрами.\n", absPath)
			exerr := os.WriteFile(absPath, []byte(exampleConfig), 0644)
			if exerr != nil {
				fmt.Println("Не удалось записать пример конфигурации", exerr)
			}
		}
		return err
	}
	return nil
}
