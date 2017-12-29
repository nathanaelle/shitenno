# Shitenno


## What is Shitenno ?

Shitenno is an unification proxy for postfix, dovecot and nginx mail.

### Use Case

The use case is mainly for distributed servers as VPS with imap nginx proxy on some VPS, postfix SMTP on others with a virtual users.

The common solution are :

  * cron replication of user file backed database
  * a common database like MySQL, LDAP

These solutions implies :

  * a lot of scripts
  * a lot of SPoF and failure sources
  * Strong coupling between the database and each app
  * With great complexity comes lots of bugs and security issues

with shitenno, you have :

  * a proxy for each 3rd party's api, which abstract them in a single HTTP API
  * upstart, runit or systemd that can spawn the process
  * a loose coupling between database schema and each 3rd party schema constraint
  * less complexity
  * 1 SPoF that can be respawned by upstart, runit or systemd


### Main goals

  * [x] postfix socketmap
  * [x] dovecot proxy dict
  * [x] nginx mail auth
  * [x] HTTP client
  * [x] HTTPS OCSP verification
  * [x] HTTPS SNI verification
  * [x] HTTPS HPKP verification
  * [ ] Logging with syslog5424
  * [ ] Health monitoring
  * [ ] HTTP Caching
  * [ ] TLS Client Certificate
  * [ ] Custom CA pool


## Security Statements

  * this code was audited 0 time

## Installation

### with go install

```
go install -u github.com/nathanaelle/shitenno/cmd/shitenno
```


## Shitenno Configuration

For the full options list see [conf/shitenno.conf](conf/shitenno.conf)


### Minimal config for postfix

```
RemoteURL = "https://remote.tld/path/"

[Postfix]

```

### Minimal config for postfix and nginx

```
RemoteURL = "https://remote.tld/path/"

[Postfix]

[Nginx]

```

## Tierce Configuration

### Postfix

```
transport_maps		= proxy:socketmap:unix:/var/run/shitenno-postfix:verb1
virtual_alias_maps	= proxy:socketmap:unix:/var/run/shitenno-postfix:verb2
```

### Dovecot

```
uri = proxy:/var/run/shitenno-dovecot:somewhere
```

### Nginx

```
http {
	upstream shitenno {
		server	unix:/var/run/shitenno-nginx;
	}

	server {
		listen		127.0.0.1:1234;
		location	/auth_imap {
			proxy_pass      http://shitenno;
		}
	}
}

mail {
	auth_http	localhost:1234/auth_imap;
}
```

## HTTP backend

the HTTP Backend is declared in field `RemoteURL`.

for each request from nginx, dovecot, postfix, a the request is rewritten as a `JSON` and `POST`ed to the backend.


### Common Request
```
{
        "Verb": "verb1",
        "Object": query_payload
}
```


### Common Reply
```
{
        "Verb": "verb1",
        "Object": query_payload,
        "Status": "OK" or "KO",
        "Data": reply_payload
}
```


### Postfix

the verb is the table name in the configuration.

the `query payload` and `reply payload` are always a string as described in [http://www.postfix.org/postmap.1.html](http://www.postfix.org/postmap.1.html).


### Dovecot

the verb is : `userdb` or `passdb`.

the `query payload` are :
```
{
        "Context": some dovecot context,
        "Object": object requested
}
```

the `reply payload` are ad-hoc reply for `userdb` or `passdb` query verb as in [http://wiki2.dovecot.org/AuthDatabase/Dict](http://wiki2.dovecot.org/AuthDatabase/Dict).


### Nginx

the verb is always `nginx`.

the `query payload` and `reply payload` are `JSON` as described in [http://nginx.org/en/docs/mail/ngx_mail_auth_http_module.html#protocol](http://nginx.org/en/docs/mail/ngx_mail_auth_http_module.html#protocol).


## License
2-Clause BSD


## Todo

  * write comments
