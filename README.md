# gopostal

Simple email library for creating and sending smtp mail - plain text, html, cc/bcc recipients and attachments are supported.

Default configurations for Gmail and Sendgrid are supported.

## Installation

Install via the go get tool:

    go get github.com/pcrawfor/gopostal

## Usage

### Quick use:

Sending basic mail via Gmail:

    import (
      "github.com/pcrawfor/gopostal"
      "log"
    )

    // Send mail via Gmail with both and text and html content    
    m := gopostal.NewGmailMailer("some_user", "some_pass")
    
    subject := "Hi there!"
    content := "Hi Paul, This is email sent from go code how are you today?"
    html_content := "<h2>Hi Paul,</h2> <p>This is email sent from go code how are you today?</p>"
    
    err := m.SendMail("some_to@test.com", "some_from@othertest.com", subject, content, html_content)
    if err != nil {
      log.Fatal(err)
    }

Sending basic email via any service:
  
    import (
      "github.com/pcrawfor/gopostal"
      "log"
    )

    // Similarly you can configure the mail service info:
    // using: NewMailer(identity, username, password, host, port string)
    m := gopostal.NewMailer("", "some_user", "some_password", "smtp.sendgrid.net", "25")
    
    subject := "Hi there!"
    content := "Hi Paul, This is email sent from go code how are you today?"
    html_content := "<h2>Hi Paul,</h2> <p>This is email sent from go code how are you today?</p>"
    
    err := m.SendMail("some_to@test.com", "some_from@othertest.com", subject, content, html_content)    
    
    if err != nil {
      log.Fatal(err)
    }

### Building messages and sending them

Building a more complicated message and sending it:

    // Create a mailer and a message with more recipients/info
    m := gopostal.NewMailer("", "some_user", "some_password", "smtp.sendgrid.net", "25")
    
    subject := "Hi there!"
    content := "Hi Paul, This is email sent from go code how are you today?"
    html_content := "<h2>Hi Paul,</h2> <p>This is email sent from go code how are you today?</p>"

    // build the message
    msg := m.NewMessage("some_to@test.com", "some_from@othertest.com", subject, content, html_content)
    msg.AddTo("other_recipient@test.com")
    msg.AddCc("somecc@test.com")
    msg.AddBcc("somebcc@test.com")

    // send the message
    err := m.Send(msg)
    if err != nil {
      log.Fatal(err)
    }

### Adding attachments

In Progress.

## Credits

Based in part on work done by:

* [ezmail](https://github.com/marcw/ezmail)
* [gomail](github.com/ungerik/go-mail)

## License

See repo LICENSE file.