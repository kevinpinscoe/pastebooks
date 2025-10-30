# Proxy setup guidance

You should put a TLS enabled and signed proxy in front of this application.

It still require much work on securing it and although a proxy will not make it
completely secure it should TLS enabled for obvious reason. I do not plan on implementing TLS natively in this webapp at least not for a long time.

For that reason you should put a TLS enabled proxy in front of it such as Nginx or Apache.

If you are sing it as a simple internal app that is fine but remember this app is designed to fill your paste buffer so the possibility of nefarious action does exist.

## Apache 

In my case I use Apache as a proxy on my main internal site.

I modified ssl.conf with the following:

```
<VirtualHost _default_:443>

...
<Proxy "http://127.0.0.1:8080/">
    ProxySet retry=0 acquire=3000 timeout=600 keepalive=On
</Proxy>

RewriteEngine On
RewriteRule ^/pastebooks$ /pastebooks/ [R=301,L]

ProxyPreserveHost On
ProxyRequests Off

<Proxy "http://127.0.0.1:8080/">
    ProxySet retry=0 acquire=3000 timeout=600 keepalive=On
</Proxy>

ProxyPass        /pastebooks/ http://127.0.0.1:8080/ 
ProxyPassReverse /pastebooks/ http://127.0.0.1:8080/

ProxyPass        /static/  http://127.0.0.1:8080/static/ 
ProxyPassReverse /static/  http://127.0.0.1:8080/static/

ProxyPass        /api/        http://127.0.0.1:8080/api/
ProxyPassReverse /api/        http://127.0.0.1:8080/api/

ProxyPassReverseCookiePath / /pastebooks/

RequestHeader set X-Forwarded-Prefix "/pastebooks"

RewriteCond %{HTTP:Upgrade} =websocket [NC]
RewriteCond %{HTTP:Connection} upgrade [NC]
RewriteRule ^/pastebooks/(.*)$  wss://127.0.0.1:8080/$1  [P,L]

</VirtualHost>
```

## Nginx

TBA

