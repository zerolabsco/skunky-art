> [!NOTE]
> This is a fork of [SkunkyArt](https://git.macaw.me/skunky/SkunkyArt). The
> upstream repository has not been updated in over a year, contains a number of
> unresolved bugs, and all of its previously listed instances were found to be
> dead links. This fork exists to keep the project maintained.

<img src="static/images/logo.png" alt="SkunkyArt" title="SkunkyArt Logo" width="20%" loading="lazy"/>

Instances: [`INSTANCES.md`](INSTANCES.md)

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
