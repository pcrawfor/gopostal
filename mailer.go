/*

Mailer

Supports Sending Text and HTML based emails

Send email as plain text or thml
Add attachments - in progress
Send with cc or bcc recipients
Built in config for Gmail, Sendgrid

Based on work from:
* github.com/ungerik/go-mail
* github.com/marcw/ezmail

*/

package gopostal

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"net/mail"
	"net/smtp"
	"strings"
	"time"
)

// Mailer
type Mailer struct {
	Identity string
	Username string
	Password string
	Host     string
	Port     string
}

func NewMailer(identity, username, password, host, port string) *Mailer {
	return &Mailer{Identity: identity, Username: username, Password: password, Host: host, Port: port}
}

func NewGmailMailer(username, password string) *Mailer {
	return &Mailer{Identity: "", Username: username, Password: password, Host: "smtp.gmail.com", Port: "587"}
}

func NewSendGridMailer(username, password string) *Mailer {
	return &Mailer{Identity: "", Username: username, Password: password, Host: "smtp.sendgrid.net", Port: "25"}
}

/*
Send email may contain html and/or text content
*/

func (m *Mailer) SendMail(to, from, subject, body, htmlBody string) error {
	fmt.Println("Send mail with text/html body")
	msg := m.NewMessage(to, from, subject, body, htmlBody)
	return m.Send(*msg)
}

/*
Send email message via smtp
*/
func (m *Mailer) Send(msg Message) error {
	fmt.Println("Sending message.")

	// validate the Message object
	if verr := msg.validate(); verr != nil {
		return verr
	}

	auth := smtp.PlainAuth(m.Identity, m.Username, m.Password, m.Host)
	content := msg.Bytes()
	host := m.Host + ":" + m.Port

	err := smtp.SendMail(
		host,
		auth,
		msg.From.Address,
		msg.recipients(),
		content,
	)

	if err != nil {
		fmt.Println("Error sending: ", err)
		return err
	}

	return nil
}

// Message
type Message struct {
	From     mail.Address
	To       []mail.Address
	cc       []mail.Address
	bcc      []mail.Address
	Subject  string
	TextBody string
	IsText   bool
	HtmlBody string
	IsHtml   bool
}

func (m *Mailer) NewMessage(to, from, subject, body, htmlBody string) *Message {
	isText := body != ""
	isHtml := htmlBody != ""

	if isText {
		fmt.Println("has text: ", body)
	}

	if isHtml {
		fmt.Println("has html: ", htmlBody)
	}

	toAddr := mail.Address{"", to}

	return &Message{
		To:       []mail.Address{toAddr},
		From:     mail.Address{"", from},
		Subject:  subject,
		HtmlBody: htmlBody,
		IsHtml:   isHtml,
		TextBody: body,
		IsText:   isText,
	}
}

/*
Validate that the message has valid addresses for to/from and non-empty subject and content strings
*/
func (m *Message) validate() error {
	// if invalid return an error

	if len(m.To) == 0 {
		return errors.New("No to addressees for message")
	}

	if &m.From == nil {
		return errors.New("No from address for message")
	}

	if m.Subject == "" {
		return errors.New("Empty subject for message")
	}

	if m.TextBody == "" && m.HtmlBody == "" {
		return errors.New("No text or html content for message")
	}

	return nil
}

func (m *Message) AddTo(to string) {
	m.To = append(m.To, mail.Address{"", to})
}

func (m *Message) AddCc(cc string) {
	m.cc = append(m.cc, mail.Address{"", cc})
}

func (m *Message) AddBcc(bcc string) {
	m.bcc = append(m.bcc, mail.Address{"", bcc})
}

/*
Boundary key for use in multipart mail content

Based on implemenation from github.com/ungerik/go-mail
*/
func (m *Message) boundary() string {
	h := md5.New()
	io.WriteString(h, fmt.Sprintf("%s", time.Now().Nanosecond()))
	return fmt.Sprintf("%x", h.Sum(nil))
}

/*
Return the byte array for the email content - if there is text and html send both appropriately
*/

const crlf = "\r\n"

func (m *Message) Bytes() []byte {
	b := bytes.NewBuffer(nil)

	b.WriteString("To: " + addressListString(m.To) + crlf)
	b.WriteString("From: " + m.From.String() + crlf)

	// TODO handle cc and bcc lists
	if len(m.cc) > 0 {
		b.WriteString("BCC: " + addressListString(m.cc) + crlf)
	}
	if len(m.bcc) > 0 {
		b.WriteString("CC: " + addressListString(m.bcc) + crlf)
	}

	b.WriteString("Subject: " + m.Subject + crlf)

	b.WriteString("Date: " + time.Now().UTC().Format(time.RFC822) + crlf)

	if m.IsText && m.IsHtml {
		// for text and html content set multipart/alternative
		boundary := m.boundary()
		b.WriteString("Content-Type: multipart/alternative; boundary=" + boundary + crlf + crlf)
		b.WriteString("--" + boundary + crlf)
		b.WriteString("MIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\n\n")
		b.WriteString(crlf + m.TextBody + crlf)
		b.WriteString(crlf + "--" + boundary + crlf)
		b.WriteString("MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n")
		b.WriteString(crlf + m.HtmlBody + crlf)
		b.WriteString(crlf + "--" + boundary + "--" + crlf)
	} else if m.IsText {
		b.WriteString("MIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\n\n")
		b.WriteString(m.TextBody)
	} else if m.IsHtml {
		b.WriteString("MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n")
		b.WriteString(m.HtmlBody)
	}

	return b.Bytes()
}

/*
Return recipients - address list of To addresses
*/
func (m *Message) recipients() []string {
	var recipients []string
	for _, i := range m.To {
		recipients = append(recipients, i.Address)
	}
	return recipients
}

/*
String of comma separated addresses
*/
func addressListString(addresses []mail.Address) string {
	var addressStrings []string
	for _, i := range addresses {
		addressStrings = append(addressStrings, i.String())
	}
	return strings.Join(addressStrings, ",")
}
