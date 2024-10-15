package idly

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"regexp"
	"time"
)

type Login struct {
	Service     string      `json:"service,omitempty"`
	UID         string      `json:"uid,omitempty"`
	At          time.Time   `json:"at,omitempty"`
	Email       string      `json:"email,omitempty"`
	IPAddress   string      `json:"ip_address,omitempty"`
	HttpHeaders http.Header `json:"http_headers,omitempty"`
	SentAlert   bool        `json:"sent_alert"`
}

var keyRegex = regexp.MustCompile(`[^\s^\/]+\/[^\s^\/]+`)
var IsLoginKeyPrefix = keyRegex.MatchString

var ipRegex = regexp.MustCompile("^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$")
var IsIPAddress = ipRegex.MatchString

func IsEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func (l Login) Prefix() string {
	return fmt.Sprintf("%s/%s", l.Service, l.UID)
}

func (l Login) Key() string {
	return fmt.Sprintf("%s/%s/%s", l.Service, l.UID, l.At.In(time.UTC).Format(time.RFC3339))
}
func (l Login) Value() (string, error) {
	b, err := json.Marshal(l)
	return string(b), err
}

type AlertInfo struct {
	Service   string    `json:"service,omitempty"`
	UID       string    `json:"uid,omitempty"`
	Location  string    `json:"location,omitempty"`
	At        time.Time `json:"at"`
	Device    string    `json:"device"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	Email     string    `json:"email"`
	Subject   string    `json:"subject"`
}
