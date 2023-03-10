# spammer
Sends text email to list of recepients.

Отправляет текстовые сообщения одного содержанию списку получаетелей. 
Использует возможность smtp отправить нескольким получателям не указывая принимающую сторону в заголовке To.

# Как запустить
## Простой способ
Запускаем приложение без параметров. В текущей рабочей директории будет создан файл конфигурации `config.yaml`. Его нужно исправть: изменить параметры подключения к smtp серверу, заголовок письма, имя и адрес отправителя.

Конфигурацию по умолчанию ожидает, что рядом должен быть `файл-с-получателями.txt` и `файл-с-телом-письма.html`.

## Файл с получателями (раздел `to-file`)

Ожидается текстовый файл, где каждый email поулчателя на новой строке
```
vasily@example.com
ujeen@example.com
```
Изменить файл с получателями можно в конфигурации: раздел `to-file`

## Файл с телом письма (раздел `body-file`)

Ожидается текстовый файл с содержимым письма. Тело письма можно форматировать html тегами.

```html
<!DOCTYPE html>
<html>
    <header>
        <meta charset="utf-8"/>
    </header>
    <body>
        <h1>Это</h1>
        <h2>Сообщение</h2>
        <h3>Содержит</h3>
        <p>Несколько строк</p>
        <b>И они отформатированы</b>
    </body>
</html>
```

Изменить файл с получателями можно в конфигурации: раздел `body-file`

## Приложения (раздел `attachments`)

К письму можно добавить приложения. На приложения действую стандартные ограничения по размерму файлов от вашего почтового сервера.

Обязательный параметр `file`: должен указывать на файл, который нужно прикрепить к письму.

Опциональный параметр `name`: так будет отображаться имя прикрепленного файла у получателя и с этим же именем будет скачиваться на ПК получателя. Если параметр name не указан, то будет использовано имя файла из раздела `file`.

Опциональный параметр `content-type`:  MIME-строка, указывающая на тип прикрепленного файла, например `application/pdf`. Если параметр не указан, приложение само постарается понять какой MIME тип соответствует данному файлу. Эта операция может быть неточной, поскольку использует магические байты в начале файлов для определения типа.


# Рекомендации

* Заполняем конфиг нужным телом письма и приложениями. 
* В файл с получателями вставляем свои email-адреса. 
* Проводим тест. 
* Если тест успешный, то меняем список emai-адресов на нужный и отправляем.
* Смотрим в консоль: если есть ошибки от сервера - уменьшаем количество получаетелей, пока сервер не перестанет ругаться.
* Повторяем, пока все email не будут отправлены.