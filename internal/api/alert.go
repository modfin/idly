package api

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"github.com/google/uuid"
	goip "github.com/jpiontek/go-ip-api"
	"github.com/mileusna/useragent"
	"github.com/modfin/idly"
	"github.com/modfin/idly/internal/config"
	"github.com/modfin/mmailer"
	gomail "gopkg.in/mail.v2"
	"html/template"
	"time"
)

type email struct {
	Service   string
	UID       string
	Location  *goip.Location
	At        time.Time
	Device    string
	IPAddress string
	UserAgent string
}

func Alert(login idly.Login) error {

	fmt.Println("Creating alert for", login.Key())

	tmpltext := config.Get().AlertEmailTemplate

	e := email{
		Service:   login.Service,
		UID:       login.UID,
		At:        login.At,
		IPAddress: login.IPAddress,
	}

	if config.Get().IpApiEnable {
		ipcli := goip.NewClient()
		if len(config.Get().IpApiKey) > 0 {
			ipcli = goip.NewClientWithApiKey(config.Get().IpApiKey)
		}
		e.Location, _ = ipcli.GetLocationForIp(login.IPAddress)
	}

	if len(login.HttpHeaders) > 0 && len(login.HttpHeaders.Get("User-Agent")) > 0 {
		e.UserAgent = login.HttpHeaders.Get("User-Agent")
		ua := useragent.Parse(e.UserAgent)

		device := "Unknown"
		if ua.Mobile {
			device = "Phone"
		}
		if ua.Tablet {
			device = "Tablet"
		}
		if ua.Desktop {
			device = "Desktop"
		}
		e.Device = fmt.Sprintf("%s, %s, %s", device, ua.OS, ua.Name)
	}

	titltmpl, err := template.New("emailTitle").Parse(config.Get().AlertEmailTitle)
	if err != nil {
		return fmt.Errorf("could not parse email template, %w", err)
	}
	buf := bytes.NewBuffer(nil)
	err = titltmpl.Execute(buf, e)
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	titleContent := buf.String()

	emltmpl, err := template.New("email").Parse(tmpltext)
	if err != nil {
		return fmt.Errorf("could not parse email template, %w", err)
	}
	buf = bytes.NewBuffer(nil)
	err = emltmpl.Execute(buf, e)
	if err != nil {
		return err
	}

	emailContent := buf.String()

	go Send(login.Service, config.Get().AlertFromEmail, login.Email, titleContent, emailContent)

	return nil
}

func Send(service, from, to, subject, content string) {
	fmt.Println("[Sending]", subject, "[To]", to)

	if !config.Get().Production {
		fmt.Println()
		fmt.Println("=== DEV, not sending email ===")
		fmt.Println("From:", from)
		fmt.Println("To:", to)
		fmt.Println("Subject:", subject)
		fmt.Println(content)
		return
	}

	if len(config.Get().MMailerURL) > 0 {
		c := mmailer.NewClient(config.Get().MMailerURL, config.Get().MMailerKey)
		_, err := c.Send(context.Background(), mmailer.Email{
			From: mmailer.Address{
				Name:  service,
				Email: from,
			},
			To:      []mmailer.Address{{Email: to}},
			Subject: subject,
			Html:    content,
			Headers: map[string]string{
				"References": fmt.Sprintf("<%s@idly>", uuid.New().String()),
			},
		})
		if err != nil {
			fmt.Println("[ERR] error sending email", err)
			return
		}
		fmt.Println("[MMAILER] Alert email sent;", subject, "[To]", to)
		return
	}

	if len(config.Get().SMTPServer) > 0 {
		message := gomail.NewMessage()

		message.SetHeader("From", fmt.Sprintf("\"%s\" <%s>", service, from))
		message.SetHeader("To", to)
		message.SetHeader("Subject", subject)
		message.SetHeader("References", fmt.Sprintf("<%s@idly>", uuid.New().String()))
		message.SetBody("text/html", content)

		d := gomail.NewDialer(config.Get().SMTPServer, config.Get().SMTPPort, config.Get().SMTPUser, config.Get().SMTPPassword)
		d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

		err := d.DialAndSend(message)
		if err != nil {
			fmt.Println("[ERR] error sending email with smtp", config.Get().SMTPServer, config.Get().SMTPPort, "; error:", err)
			return
		}
		fmt.Println("[SMTP] Alert email sent;", subject, "[To]", to)
		return
	}
}
