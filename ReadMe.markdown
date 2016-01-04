# Shitenno


## What is Shitenno ?

Shitenno is an unification proxy for postfix, dovecot and nginx mail.


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

## License
2-Clause BSD


## Todo

  * write comments
