# idly
A small service to send IDS type emails

It is a common practice in industry to have a service report back to a users registered email address when 
a login is detected from a new source, eg a new ip address, in order to facilitate instruction detection. Idly is a 
small service ment to run in a k8s cluster or similarly to notify users when a new ip is used for login.

Simply put, idly sends emails to user notifying them that their account was used to login to a service.

## Env Config
* `PRODUCTION` if true, mmailer or smtp will be used to send the email. if false, the email will just be logged to std out
* `LOGIN_TTL` The TTL for how long a login record will be kept. If an ip does not exist in the records an email will be sent.
default 720h (30 days)
* `ALERT_ON_INIT` if true, an alert will be sent on the very first login. 
* `HTTP_PORT` port used for HTTP
* `BADGER_URI` the path to the location of the badger kv store user for storing data. default ./badger 

* `IP_API_ENABLE` if true, the idly uses ip-api.com to lookup metadata about the IP address. default true
* `IP_API_KEY` you can subscribe to ip api and get an api key for higher volumes

* `ALERT_EMAIL_FROM` from who will the service send the email, 
* `ALERT_EMAIL_TITLE` The title of the email as a go template. default `[{{.Service}}]: Your {{.Service}} account has been accessed from a new IP Address`
* `ALERT_EMAIL_TEMPLATE` The path to a file containing a template for the email.

* `MMAILER_URL` MMailer API base URL (https://github.com/modfin/mmailer)
* `MMAILER_KEY` MMailer API key 

* `SMTP_SERVER` smtp server address
* `SMTP_PORT` smtp server port
* `SMTP_USER` smtp user for auth
* `SMTP_PASSWORD` smtp user password for auth


## Usage
When a login attempt fail or succeeds in you application/service, use the provided client or make a http request to idly

```go 

func login(email, password string, r http.Request) bool{
	
	ids := idly.NewClient("ServiceName", "http://idly:8080").
	    Request(email, r.Header.Get("X-Real-IP")).
	    WithUserAgent(r.Header.Get("User-Agent"))
	
    if email != "luke@example.com" || password != "skywalker"{
        ids.Fail()
        return false
    }
	
    ids.Success()
    return true
}


```


## Monitoring

http://idly:8080/metrics provides prometheus metrics that can be monitored for successfully and failed logins along with 
other metrics. This can be used in conjunction with eg. Grafana to set up alerts on failed logins to raise notice of a 
potential instruction.

## Example k8s config
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: idly
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: idly-pod
    spec:
      volumes:
        - name: data
          persistentVolumeClaim:
            claimName: data-idly
      containers:
        - name: idly
          image: modfin/idly:latest
          volumeMounts:
            - mountPath: /badger
              name: data
          env:
            - name: PRODUCTION
              value: "true"
            - name: ALERT_ON_INIT
              value: "true"
            - name: ALERT_EMAIL_FROM
              value: the_from_email@exampl.com
            - name: BADGER_URI
              value: /badger
            - name: SMTP_SERVER
              value: smtp.example.com
            - name: SMTP_PORT
              value: 587
            - name: SMTP_USER
              value: the_user@example.com
            - name: SMTP_PASSWORD
              value: the_password
              
---
kind: Service
metadata:
  name: idly
spec:
  ports:
    - name: http
      port: 8080
      targetPort: 8080
  selector:
    app: idly-pod

```
