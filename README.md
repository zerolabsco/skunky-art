> [!NOTE]
> This is a fork of [SkunkyArt](https://git.macaw.me/skunky/SkunkyArt). The
> upstream repository has not been updated in over a year, contains a number of
> unresolved bugs, and all of its previously listed instances were found to be
> dead links. This fork exists to keep the project maintained.

<img src="static/images/logo.png" alt="SkunkyArt" title="SkunkyArt Logo" width="20%" loading="lazy"/>

Instances: [`INSTANCES.md`](INSTANCES.md)

# EN 🇺🇸
## Description
SkunkyArt 🦨 — an alternative frontend for DeviantArt that works entirely
without JavaScript.

## Build
It is recommended to build with the 'embed' tag because it embeds the presets in
the binary. If you plan to modify the templates, then do not use this tag. You
can also add the `-ldflags "-w -s"` argument (GCCGO has a different name for it
— `gccgoflags`) to reduce the size of the output file. Here is an example:

`go build -tags embed -ldflags "-w -s"`

## Setup
The sample config is in the `config.example.json` file. For custom config, use
the `--config` option.

See the [`SETUP.md`](SETUP.md) file for more info about directives.

## Adding instance to the list
To do this, you must either make a PR by adding your instance to the
`instances.json` and `INSTANCES.md` files (you can use `--add-instance`
cli-argument to automatically add the instance to these files) or create an
Issue.

## Acknowledgements
* [vlnst](https://git.bloat.cat/vlnst) — wrote a Docker file.
* [Лис⚛](https://go.kde.org/matrix/#/@fox:matrix.org) — helped me understand Go
  and gave me a lot of useful advice on this language.
* [meoww](https://codeberg.org/meoww) — translated some sentences into English
  and wrote a service for openrc

# RU 🇷🇺
## Описание
SkunkyArt 🦨 — альтернативный фронтенд к DeviantArt, который полностью работает
без JS (JavaScript).

## Сборка
Рекомендуется производить сборку с тегом 'embed', поскольку он встраивает
заготовки в бинарный файл. Если вы планируете изменять заготовки, то не
используйте этот тег. Также вы можете добавить аргумент `-ldflags "-w -s"` (у
GCCGO он называется по-другому — `gccgoflags`) для уменьшения размера выходного
файла. Вот пример:

`go build -tags embed -ldflags "-w -s"`

## Настройка
Пример конфига находится в файле `config.example.json`. Чтобы указать свой
конфиг, используйте cli-аргумент `--config`.

См. [`SETUP-RU.md`](SETUP-RU.md) для информации о настройки фронтенда.

## Добавление инстанса в список
Чтобы это сделать, вы должны либо сделать PR, добавив в файлы `instances.json` и
`INSTANCES.md` свой инстанс (можете воспользоваться cli-аргументом
`--add-instance`, который автоматически это сделает), либо создать Issue.

## Благодарности
* [vlnst](https://git.bloat.cat/vlnst) — написал Docker-файл.
* [Лис⚛](https://go.kde.org/matrix/#/@fox:matrix.org) — помог разобраться в Go и
  много чего полезного посоветовал по этому языку.
* [meoww](https://codeberg.org/meoww) — перевела некоторые предложения на
  английский язык и написала сервис для openrc
