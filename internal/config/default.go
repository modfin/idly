package config

const DefaultEmailTemplate = `
<html>
<body>
<h1>New login to {{.Service}}</h1> 

<p>
    Your security is very important to us. <br/>
	We noticed you logged in from a new device or location. If this was you, you can ignore this alert.
</p>
<p>
  User <strong>{{.UID}}</strong> <br />
  {{if ne .Location nil}}
  Location <strong>{{.Location.Country}}</strong> <br/>
  {{end}}
  Time <strong>{{.At.Format "Mon, 02 Jan 2006 15:04:05 MST"}}</strong> <br/>
  Device <strong>{{.Device}}</strong> <br/>
  IP address <strong>{{.IPAddress}}</strong> <br/>
  User-Agent <strong>{{.UserAgent}}</strong> 
</p>

<h2>Not you?</h2>

If you suspect any suspicious activity on your account, please change your password and enable two-factor authentication

</body>
</html>
`
