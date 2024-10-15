package api

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	goip "github.com/jpiontek/go-ip-api"
	"github.com/mileusna/useragent"
	"github.com/modfin/idly"
	"github.com/modfin/idly/internal/config"
	"github.com/modfin/mmailer"
	gomail "gopkg.in/mail.v2"
	"html/template"
	"net/http"
	"strings"
)

func Alert(login idly.Login) error {

	fmt.Println("Creating alert for", login.Key())

	ai := idly.AlertInfo{
		Service:   login.Service,
		Email:     login.Email,
		UID:       login.UID,
		At:        login.At,
		IPAddress: login.IPAddress,
		Location:  "Unknown",
	}

	if config.Get().IpApiEnable {
		ipcli := goip.NewClient()
		if len(config.Get().IpApiKey) > 0 {
			ipcli = goip.NewClientWithApiKey(config.Get().IpApiKey)
		}
		loc, err := ipcli.GetLocationForIp(login.IPAddress)
		if loc != nil {
			ai.Location = loc.Country
		}
		if err != nil {
			fmt.Println("[ERR] could not get location for ip", err)
		}
	}

	if len(login.HttpHeaders) > 0 && len(login.HttpHeaders.Get("User-Agent")) > 0 {
		ai.UserAgent = login.HttpHeaders.Get("User-Agent")
		ua := useragent.Parse(ai.UserAgent)

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
		ai.Device = fmt.Sprintf("%s, %s, %s", device, ua.OS, ua.Name)
	}

	titltmpl, err := template.New("emailTitle").Parse(config.Get().AlertEmailTitle)
	if err != nil {
		return fmt.Errorf("could not parse email template, %w", err)
	}
	buf := bytes.NewBuffer(nil)
	err = titltmpl.Execute(buf, ai)
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	titleContent := buf.String()

	ai.Subject = titleContent

	if config.Get().AlertEmailPosthook != "" {
		// posthook overrides email if it is defined
		return AlertPosthook(config.Get().AlertEmailPosthook, ai)
	}

	return AlertEmail(config.Get().AlertFromEmail, ai)
}

func AlertPosthook(posthook string, ai idly.AlertInfo) error {
	fmt.Println("[Alerting posthook]", ai.Subject, "[To]", posthook)

	if !strings.HasPrefix(posthook, "http") {
		fmt.Println("[ERR] unsupported posthook scheme, only http/https allowed")
		return fmt.Errorf("unsupported posthook (%s) scheme, only http/https allowed", posthook)
	}

	payload, err := json.Marshal(ai)
	if err != nil {
		fmt.Println("[ERR] could not marshal alert info", err)
		return err
	}

	if !config.Get().Production {
		fmt.Println("[POSTHOOK] Sending", string(payload))
	}

	go SendPosthook(posthook, payload)

	return nil
}

func SendPosthook(posthook string, payload []byte) {
	request, err := http.NewRequest(http.MethodPost, posthook, bytes.NewReader(payload))
	if err != nil {
		fmt.Println("[ERR] could not create posthook request", err)
		return
	}
	request.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		fmt.Println("[ERR] could not send posthook request", err)
		return
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		fmt.Println("[ERR] non 200 statuscode response from posthook request")
		return
	}

	fmt.Println("[POSTHOOK] Alert succesfully sent and receieved to posthook", posthook)
}

func AlertEmail(from string, ai idly.AlertInfo) error {
	tmpltext := config.Get().AlertEmailTemplate

	emltmpl, err := template.New("email").Parse(tmpltext)
	if err != nil {
		return fmt.Errorf("could not parse email template, %w", err)
	}
	buf := bytes.NewBuffer(nil)
	err = emltmpl.Execute(buf, ai)
	if err != nil {
		return err
	}

	emailContent := buf.String()

	go SendEmail(ai.Service, from, ai.Email, ai.Subject, emailContent)

	return nil
}

func SendEmail(service, from, to, subject, content string) {
	fmt.Println("[Sending email]", subject, "[To]", to)

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
