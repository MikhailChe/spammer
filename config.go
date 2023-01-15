package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

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
to:
  - mail2@example.com
subject: Темя письма
content-type: text/html; charset="utf-8"
body: >
  <html><body><h1>Сообщение</h1><h2>Которое</h2><h3>Нужно</h3><p>Отправить</p></body></html>
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
	To          []string `config:"to,required"`
	Subject     string   `config:"subject"`
	ContentType string   `config:"content-type,required" yaml:"content-type"`
	Body        string   `config:"body,required"`
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
		var pathError error = &fs.PathError{}
		if errors.Is(err, backend.ErrNotFound) || errors.As(err, &pathError) {
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
