package send

import (
	_ "embed"
	"html/template"
	"net/http"

	"github.com/mikhailche/spammer/pages/base"
)

//go:embed index.html
var install_html string
var senderTemplate *template.Template

type FormField struct {
	ID    string
	Name  string
	Value string
}

func (f FormField) String() string {
	return f.Value
}

type SendForm struct {
	SMTP struct {
		Host     string
		Port     string
		SSL      string
		Login    string
		Password string
	}
	Sender struct {
		Name  string
		Email string
	}
	Recipient struct {
		Email string
	}
	Message struct {
		Subject string
		Text    string
	}
}

type TemplateData struct {
	Body SendForm
}

func init() {
	var err error
	senderTemplate, err = base.Page(install_html)
	if err != nil {
		panic(err)
	}
}

func Page(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(100 * 1024 * 1024) // 100 MB
	defer r.Body.Close()
	err := senderTemplate.ExecuteTemplate(w, "base", struct {
		PageTitle string
		Body      SendForm
	}{
		Body: SendForm{
			SMTP: struct {
				Host     string
				Port     string
				SSL      string
				Login    string
				Password string
			}{
				Host:     r.FormValue("app-config-smtp-host-name"),
				Port:     r.FormValue("app-config-smtp-port-name"),
				SSL:      r.FormValue("app-config-smtp-ssl-name"),
				Login:    r.FormValue("app-config-smtp-login-name"),
				Password: r.FormValue("app-config-smtp-password-name"),
			},
			Sender: struct {
				Name  string
				Email string
			}{
				Name:  r.FormValue("app-config-sender-name-name"),
				Email: r.FormValue("app-config-sender-email-name"),
			},
			Recipient: struct {
				Email string
			}{
				Email: r.FormValue("app-config-recipient-email-name"),
			},
			Message: struct {
				Subject string
				Text    string
			}{
				Subject: r.FormValue("app-config-message-subject-name"),
				Text:    r.FormValue("app-config-message-text-name"),
			},
		},
	})
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}
