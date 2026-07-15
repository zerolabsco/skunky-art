# Units
Maximum file size in megabytes, requires numeric value.<br>
Time units:
* `i` — minutes
* `h` — hours
* `w` — weeks
* `m` — months
* `y` — years

# Config
* `listen` — IP and port to listen on in the following form: ip:port
* `uri` — Instance URI. Example: `"uri":"/art/"` -> https://skunky.ebloid.ru/art/
* `cache` — Caching system; default is off.
  * `enabled` — Caching system state, requires boolean value
  * `path` — Path to cache directory, requires absolute filesystem path
  * `lifetime` — Cached file life time, requires numeric value, followed by multiplicative suffix (see Time Units for details)
  * `max-size` — Maximum file size in megabytes
  * `update-interval` — Automatic rotation interval
* `static-path` — This setting determines path to static, which will be copied to RAM when SkunkyArt is started. Useless if you're use binary compiled with 'embed' tag.
* `download-proxy` — Proxy address for downloading files.
* `user-agent` — String, which SkunkyArt uses as UA

# Setting up reverse proxy
Pretty much business as usual, except for the [`X-Forwarded-Proto`](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-Proto) header setting.

Nginx example configuration:
```apache
server {
    listen 443 ssl;
    server_name skunky.example.com;
    
    # In case of subdomain, use / instend of ((BASE_URL))
    location ((BASE_URL)) {
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header Host $host;
        proxy_http_version 1.1;
        proxy_pass http://((IP)):((PORT));
    }
}
```