package messages

import (
	echo "github.com/labstack/echo"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"gopkg.in/go-playground/validator.v9"
)

// Message ...
type Message struct {
	localizer *i18n.Localizer
}

// NewMessage ...
func NewMessage(c echo.Context) Message {
	localizer, _ := c.Get("localizer").(*i18n.Localizer)
	m := Message{
		localizer: localizer,
	}
	return m
}

// GetMessage ...
func (m *Message) GetMessage(id string, templateData map[string]interface{}) string {
	msg, _ := m.localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID: id,
		},
		TemplateData: templateData,
	})
	return msg
}

// GetFieldTagMessage ...
func (m *Message) GetFieldTagMessage(e validator.FieldError, templateData map[string]interface{}) string {
	key := e.Field() + "." + e.Tag()
	msg := m.GetMessage(key, templateData)

	if len(msg) == 0 {
		msg = m.GetMessage(e.Tag(), templateData)
	}
	return msg
}
