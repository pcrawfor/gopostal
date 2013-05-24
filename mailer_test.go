package gopostal

import (
	"bytes"
	"net/mail"
	"testing"
)

// test Mailer
// test that mailer object is configured properly
func TestMailer(t *testing.T) {
	m := NewMailer("", "username", "password", "host", "100")
	// verify attrs
	if m.Username != "username" {
		t.Error("Expected username to be correct")
	}
	if m.Password != "password" {
		t.Error("Expected password to be correct")
	}
	if m.Host != "host" {
		t.Error("Expected host to be correct")
	}
	if m.Port != "100" {
		t.Error("Expected port to be correct")
	}
}

// test that gmail mailer is correctly configured
func TestGmailMailer(t *testing.T) {
	m := NewGmailMailer("username", "password")
	if m.Host != "smtp.gmail.com" || m.Port != "587" {
		t.Error("Expected gmail config got: ", m)
	}
}

// test that sendgrid mailer is correctly configured
func TestSendGridMailer(t *testing.T) {
	m := NewSendGridMailer("username", "password")
	if m.Host != "smtp.sendgrid.net" || m.Port != "25" {
		t.Error("Expected sendgrid config got: ", m)
	}
}

// test Message
type testpair struct {
	messageValues  []string
	messageContent []byte
	isText         bool
	isHtml         bool
}

var tests = []testpair{
	// text only
	{[]string{"testrep@test.com", "testfrom@test.com", "Testing", "It's the text body of the mail", ""},
		[]byte("To: <testrep@test.com>\r\nFrom: <testfrom@test.com>\r\nSubject: Testing\r\nDate: 24 May 13 16:32 UTC\r\nMIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\n\nIt's the text body of the mail"),
		true, false},
	// text and html
	{[]string{"testrep1@test.com", "testfrom1@test.com", "Testing some more: here", "It's the text body of the mail", "<h1>Hi!</h1> <p>This is some more testing with html</p> <p>Check out <a href='http://www.github.com'>Github!</a></p>"},
		[]byte("To: <testrep1@test.com>\r\nFrom: <testfrom1@test.com>\r\nSubject: Testing some more: here\r\nDate: 24 May 13 16:32 UTC\r\nContent-Type: multipart/alternative; boundary=faaac39e6bf197bd7ae820c49d5f9958\r\n\r\n--faaac39e6bf197bd7ae820c49d5f9958\r\nMIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\n\n\r\nIt's the text body of the mail\r\n\r\n--faaac39e6bf197bd7ae820c49d5f9958\r\nMIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n\r\n<h1>Hi!</h1> <p>This is some more testing with html</p> <p>Check out <a href='http://www.github.com'>Github!</a></p>\r\n\r\n--faaac39e6bf197bd7ae820c49d5f9958--\r\n"),
		true, true},
	// html only
	{[]string{"testrep2@test.com", "testfrom2@test.com", "Testing html only", "", "<h1>Hi!</h1> <p>This is some more testing with html</p> <p>Check out <a href='http://www.github.com'>Github!</a></p>"},
		[]byte("To: <testrep2@test.com>\r\nFrom: <testfrom2@test.com>\r\nSubject: Testing html only\r\nDate: 24 May 13 16:32 UTC\r\nMIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n<h1>Hi!</h1> <p>This is some more testing with html</p> <p>Check out <a href='http://www.github.com'>Github!</a></p>"),
		false, true},
}

func TestTextEmail(t *testing.T) {
	m := NewMailer("", "username", "password", "host", "587")
	for _, info := range tests {

		msg := m.NewMessage(info.messageValues[0], info.messageValues[1], info.messageValues[2], info.messageValues[3], info.messageValues[4])

		// verifying length since date & boundary values will mess up the direct byte comparison
		if len(info.messageContent) != len(msg.Bytes()) {
			t.Error("Expected content length to match: \n" + string(msg.Bytes()) + "\n == \n" + string(info.messageContent))
		}

		if msg.IsText != info.isText || msg.IsHtml != info.isHtml {
			t.Error("Email should match text/html content expectations.")
		}

		// verify parts of the message explicitly
		check, _ := mail.ReadMessage(bytes.NewBuffer(msg.Bytes()))
		from, _ := check.Header.AddressList("From")
		if len(from) > 0 {
			if from[0].Name != "" || from[0].Address != info.messageValues[1] {
				t.Error("From address to be: " + info.messageValues[1] + " got: " + from[0].Address)
			}
		}

		subject := check.Header.Get("Subject")
		if len(from) > 0 {
			if subject != info.messageValues[2] {
				t.Error("From address to be: " + info.messageValues[2] + " got: " + subject)
			}
		}

	}
}
