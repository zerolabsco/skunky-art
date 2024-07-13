[![Matrix room](https://img.shields.io/badge/matrix-000000?style=for-the-badge&logo=Matrix&logoColor=white)]("https://go.kde.org/matrix/#/#skunkyart:ebloid.ru")
# Instances
|Инстанс|Yggdrasil|I2P|Tor|NSFW|Proxifying|Country|
|:-----:|:-------:|:-:|:-:|:--:|:--------:|:-----:|
|[skunky.ebloid.ru](https://skunky.ebloid.ru/art)|[Yes](http://[201:eba5:d1fc:bf7b:cfcb:a811:4b8b:7ea3]/art)|No|No| No | No | Russia |
|[clovius.club](https://skunky.clovius.club)|No|No|No| Yes | Yes | Sweden |
|[bloat.cat](https://skunky.bloat.cat)|No|No|No| Yes | Yes | Romania |
|[frontendfriendly.xyz](https://skunkyart.frontendfriendly.xyz)|No|No|No| Yes | Yes | Finland |

# EN 🇺🇸
## Description
SkunkyArt 🦨 -- alternative frontend to DeviantArt, which will work without problems even on quite old hardware, due to the lack of JavaScript.
## Config
The sample config is in the `config.example.json` file. To specify your own path to the config, use the CLI argument `-c` or `--config`.
* `listen` -- the address and port on which SkunkyArt will listen
* `base-path` -- the path to the instance. Example: "`base-path`:"/art/" -> https://skunky.ebloid.ru/art/
* `cache` -- caching system; default is off.
* * `path` -- the path to the cache
* * `lifetime` -- cache file lifetime; measured in Unix milliseconds.
* * `max-size` -- maximum file size in bytes.
* `dirs-to-memory` -- this setting determines which directories will be copied to RAM when SkunkyArt is started. Required
* `download-proxy` -- proxy address for downloading files.
## Examples of reverse proxies
Nginx:
```apache
server {
    listen 443 ssl;
    server_name skunky.example.com;
    
    location ((BASE URL)) { # if you have a separate subdomain for the frontend, insert '/' instead of '((BASE URL))'.
        proxy_set_header Scheme $scheme;
        proxy_http_version 1.1;
        proxy_pass http://((IP)):((PORT));
    }
}
```
## How do I add my instance to the list?
To do this, you must either make a PR by adding your instance to the `instances.json` file, or report it to the room in Matrix. I don't think it needs any description. However, be aware, this list has a couple rules:
1. the instance must not use Cloudflare.
2. If your instance has modified source code, you need to publish it to any free platform. For example, Github and Gitlab are not.
## Acknowledgements
* [Лис⚛](https://go.kde.org/matrix/#/@fox:matrix.org) -- helped me understand Go and gave me a lot of useful advice on this language.

# RU 🇷🇺
## Описание
SkunkyArt 🦨 -- альтернативный фронтенд к DeviantArt, который будет работать без проблем даже на довольно старом оборудовании, за счёт отсутствия JavaScript.
## Конфиг
Пример конфига находится в файле `config.example.json`. Чтобы указать свой путь до конфига, используйте CLI-аргумент `-c` или `--config`.
* `listen` -- адрес и порт, на котором будет слушать SkunkyArt
* `base-path` -- путь к инстансу. Пример: "base-path": "/art/" -> https://skunky.ebloid.ru/art/
* `cache` -- система кеширования; по умолчанию - выкл.
* * `path` -- путь до кеша
* * `lifetime` -- время жизни файла в кеше; измеряется в Unix-миллисекундах
* * `max-size` -- максимальный размер файла в байтах
* `dirs-to-memory` -- данная настройка определяет какие каталоги будут скопированы в ОЗУ при запуске SkunkyArt. Обязательна
* `download-proxy` -- адрес прокси для загрузки файлов
## Примеры reverse-прокси
Nginx:
```apache
server {
    listen 443 ssl;
    server_name skunky.example.com;
    
    location ((BASE URL)) { # если у вас отдельный поддомен для фронтенда, вместо '((BASE URL))' вставляйте '/'
        proxy_set_header Scheme $scheme;
        proxy_http_version 1.1;
        proxy_pass http://((IP)):((PORT));
    }
}
```
## Как добавить свой инстанс в список?
Чтобы это сделать, вы должны либо сделать PR, добавив в файл `instances.json` свой инстанс, либо сообщить о нём в комнате в Matrix. Думаю, он не нуждается в описании. Однако учтите, у этого списка есть пара правил:
1. Инстанс не должен использовать Cloudflare.
2. Если ваш инстанс имеет модифицированный исходный код, то вам нужно опубликовать его на любую свободную площадку. Например, Github и Gitlab таковыми не являются.
## Благодарности
* [Лис⚛](https://go.kde.org/matrix/#/@fox:matrix.org) -- помог разобраться в Go и много чего полезного посоветовал по этому языку.