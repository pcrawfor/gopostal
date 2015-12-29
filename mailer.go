// Package gopostal - Supports Sending Text and HTML based emails.  Send email as plain text or html.
package gopostal

/*

Send email as plain text or html
Add attachments - in progress
Send with cc or bcc recipients
Built in config for Gmail, Sendgrid

TODO: handle custom mail headers

Based on work from:
* github.com/ungerik/go-mail
* github.com/marcw/ezmail

*/

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

// Mailer represents a mail sender object
type Mailer struct {
	Identity string
	Username string
	Password string
	Host     string
	Port     string
}

// NewMailer returns an instance of Mailer with the passed in connnection configuration options
func NewMailer(identity, username, password, host, port string) *Mailer {
	return &Mailer{Identity: identity, Username: username, Password: password, Host: host, Port: port}
}

// NewGmailMailer is a shortcut constructor for a Mailer with a gmail connection
func NewGmailMailer(username, password string) *Mailer {
	return &Mailer{Identity: "", Username: username, Password: password, Host: "smtp.gmail.com", Port: "587"}
}

// NewSendGridMailer is a shortcut constructor for a Mailer with a sendgrid connection
func NewSendGridMailer(username, password string) *Mailer {
	return &Mailer{Identity: "", Username: username, Password: password, Host: "smtp.sendgrid.net", Port: "25"}
}

// SendMail sends email to the to address populating the from, body and htmlbody provided
func (m *Mailer) SendMail(to, from, subject, body, htmlBody string) error {
	msg, err := m.NewMessage(to, from, subject, body, htmlBody)
	if err != nil {
		return err
	}
	return m.Send(*msg)
}

// Send will send the given Message
func (m *Mailer) Send(msg Message) error {
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
		return err
	}

	return nil
}

// Message represents an email message containing the sender to, the from, cc and bcc recipients and all content for the message including text and html body content
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
	Headers  map[string]string
}

// NewMessage returns a new Message object build with the to, from, subject, body and html body provided
func (m *Mailer) NewMessage(to, from, subject, body, htmlBody string) (*Message, error) {
	isText := body != ""
	isHtml := htmlBody != ""

	toAddr, err := mail.ParseAddress(to)
	if err != nil {
		return nil, err
	}
	fromAddr, err := mail.ParseAddress(from)
	if err != nil {
		return nil, err
	}

	return &Message{
		To:       []mail.Address{*toAddr},
		From:     *fromAddr,
		Subject:  subject,
		HtmlBody: htmlBody,
		IsHtml:   isHtml,
		TextBody: body,
		IsText:   isText,
		Headers:  make(map[string]string),
	}, nil
}

// validate will verify that the message has valid addresses for to/from and non-empty subject and content strings
func (m *Message) validate() error {
	// if invalid return an error

	if len(m.To) == 0 {
		return errors.New("no to addressees for message")
	}

	if &m.From == nil {
		return errors.New("no from address for message")
	}

	if m.Subject == "" {
		return errors.New("empty subject for message")
	}

	if m.TextBody == "" && m.HtmlBody == "" {
		return errors.New("no text or html content for message")
	}

	return nil
}

// AddTo appends a to address to the message
func (m *Message) AddTo(to string) {
	a, e := mail.ParseAddress(to)
	if e == nil {
		m.To = append(m.To, *a)
	}
}

// AddCc appends a cc address to the message
func (m *Message) AddCc(cc string) {
	a, e := mail.ParseAddress(cc)
	if e == nil {
		m.cc = append(m.cc, *a)
	}
}

// AddBcc appends a bcc address to the message
func (m *Message) AddBcc(bcc string) {
	a, e := mail.ParseAddress(bcc)
	if e == nil {
		m.bcc = append(m.bcc, *a)
	}
}

// AddHeader appends a bcc address to the message
func (m *Message) AddHeader(name, value string) {
	m.Headers[name] = value
}

// boundary creates a boundary key for use in multipart mail content
// Based on implemenation from github.com/ungerik/go-mail
func (m *Message) boundary() string {
	h := md5.New()
	io.WriteString(h, fmt.Sprintf("%s", time.Now().Nanosecond()))
	return fmt.Sprintf("%x", h.Sum(nil))
}

const crlf = "\r\n"

// Bytes returns a byte array for the email content
func (m *Message) Bytes() []byte {
	b := bytes.NewBuffer(nil)

	b.WriteString("To: " + addressListString(m.To) + crlf)
	b.WriteString("From: " + m.From.String() + crlf)

	if len(m.cc) > 0 {
		b.WriteString("BCC: " + addressListString(m.cc) + crlf)
	}
	if len(m.bcc) > 0 {
		b.WriteString("CC: " + addressListString(m.bcc) + crlf)
	}

	// add custom headers
	for name, val := range m.Headers {
		res := name + ": " + val
		b.WriteString(res + crlf)
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

// recipients returns a slice of strings containing the list of recipients
func (m *Message) recipients() []string {
	var recipients []string
	for _, i := range m.To {
		recipients = append(recipients, i.Address)
	}
	return recipients
}

// addressListString returns a string of the addresses passed in concatenated together
func addressListString(addresses []mail.Address) string {
	var addressStrings []string
	for _, i := range addresses {
		addressStrings = append(addressStrings, i.String())
	}
	return strings.Join(addressStrings, ",")
}
