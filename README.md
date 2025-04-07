> [!NOTE]
> Currently, due to school, I cannot actively develop this project :(
> However, this does not mean that development has stopped. Just wait for the summer. For questions, write either to the Matrix room or to me in DM.

<img src="static/images/logo.png" alt="SkunkyArt" title="SkunkyArt Logo" width="20%" loading="lazy"/>

[![Matrix room](https://img.shields.io/badge/matrix-000000?style=for-the-badge&logo=Matrix&logoColor=white)](https://go.kde.org/matrix/#/#skunkyart:gnulinux.club)

Instances: [`INSTANCES.md`](/skunky/SkunkyArt/src/branch/master/INSTANCES.md)

# EN 🇺🇸
## Description
SkunkyArt 🦨 — alternative frontend for DevianArt, which works without JS.
## Build (translated via DeepL)
It is recommended to build with the 'embed' tag because it embeds the presets in the binary. If you plan to modify the templates, then do not use this tag. You can also add the `-ldflags "-w -s"` argument (GCCGO has a different name for it — `gccgoflags`) to reduce the size of the output file. Here is an example:

`go build -tags embed -ldflags "-w -s"`

Pre-compiled binaries can be found in the [Releases](https://git.macaw.me/skunky/skunkyart/releases) tab.
## Setup
The sample config is in the `config.example.json` file. For custom config, use `--config` option.
See the [`SETUP.md`](/skunky/SkunkyArt/src/branch/master/SETUP.md) file for more info about directives. 
## Adding instance to the list
To do this, you must either make a PR by adding your instance to the `instances.json` and `INSTANCES.md` files (you can use `--add-instance` cli-argument to automatically add the instance to these files), or create an Issue, or report it to the room in Matrix. Keep in mind that your instance must comply with the following rules:
1. the Instance must not use Cloudflare.
2. If your instance has modified source code, you need to publish it to any free platform. For example, Github and Gitlab are not.
## Acknowledgements
* [vlnst](https://git.bloat.cat/vlnst) — wrote a Docker file.
* [Лис⚛](https://go.kde.org/matrix/#/@fox:matrix.org) — helped me understand Go and gave me a lot of useful advice on this language.
* [meoww](https://codeberg.org/meoww) — translated some sentences into English and wrote a service for openrc

# RU 🇷🇺
## Описание
SkunkyArt 🦨 — альтернативный фронтенд к DeviantArt, который полностью работает без JS (JavaScript).
## Сборка
Рекомендуется производить сборку с тегом 'embed', поскольку он встраивает заготовки в бинарный файл. Если вы планируете изменять заготовки, то не используйте этот тег. Также вы можете добавить аргумент `-ldflags "-w -s"` (у GCCGO он называется по-другому — `gccgoflags`) для уменьшения размера выходного файла. Вот пример:

`go build -tags embed -ldflags "-w -s"`

Готовые бинари находятся во вкладке [Releases](https://git.macaw.me/skunky/skunkyart/releases).
## Настройка
Пример конфига находится в файле `config.example.json`. Чтобы указать свой конфиг, используйте cli-аргумент `--config`.
См. [`SETUP-RU.md`](/skunky/SkunkyArt/src/branch/master/SETUP-RU.md) для информации о настройки фронтенда. 
## Добавление инстанса в список
Чтобы это сделать, вы должны либо сделать PR, добавив в файлы `instances.json` и `INSTANCES.md` свой инстанс (можете воспользоваться cli-аргументом `--add-instance`, который автоматически это сделает), либо создать Issue, или сообщить о нём в комнате в Matrix. Учтите, что ваш инстанс должен соблюсти следущие правила:
1. Инстанс не должен использовать Cloudflare итп.
2. Если ваш инстанс имеет модифицированный исходный код, то вам нужно опубликовать его на любую свободную площадку. Например, Github и Gitlab таковыми не являются.
## Благодарности
* [vlnst](https://git.bloat.cat/vlnst) — написал Docker-файл.
* [Лис⚛](https://go.kde.org/matrix/#/@fox:matrix.org) — помог разобраться в Go и много чего полезного посоветовал по этому языку.
* [meoww](https://codeberg.org/meoww) — перевела некоторые предложения на английский язык и написала сервис для openrc