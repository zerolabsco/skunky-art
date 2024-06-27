package app

import (
	"encoding/json"
	"strings"

	"git.macaw.me/skunky/devianter"
)

type text struct {
	TXT  string
	from int
	to   int
}

func ParseDescription(dscr devianter.Text) string {
	var parseddescription strings.Builder
	TagBuilder := func(tag string, content string) string {
		if tag != "" {
			var htm strings.Builder
			htm.WriteString("<")
			htm.WriteString(tag)
			htm.WriteString(">")

			htm.WriteString(content)

			htm.WriteString("</")
			htm.WriteString(tag)
			htm.WriteString(">")
			return htm.String()
		}
		return content
	}

	if description, dl := dscr.Html.Markup, len(dscr.Html.Markup); dl != 0 &&
		description[0] == '{' &&
		description[dl-1] == '}' {
		var descr struct {
			Blocks []struct {
				Key, Text, Type   string
				InlineStyleRanges []struct {
					Offset, Length int
					Style          string
				}
			}
		}
		e := json.Unmarshal([]byte(description), &descr)
		err(e)

		for _, x := range descr.Blocks {
			ranges := make(map[int]text)
			for i, rngs := range x.InlineStyleRanges {
				var tag string

				switch rngs.Style {
				case "BOLD":
					tag = "b"
				case "UNDERLINE":
					tag = "u"
				case "ITALIC":
					tag = "i"
				}

				fromto := rngs.Offset + rngs.Length
				ranges[i] = text{
					TXT:  TagBuilder(tag, x.Text[rngs.Offset:fromto]),
					from: rngs.Offset,
					to:   fromto,
				}
			}

			for _, r := range ranges {
				var tag string
				switch x.Type {
				case "header-two":
					tag = "h2"
				case "unstyled":
					tag = "p"
				}
				parseddescription.WriteString(r.TXT)
				parseddescription.WriteString(TagBuilder(tag, x.Text[r.to:]))
			}
		}
	} else if dl != 0 {
		parseddescription.WriteString(description)
	}

	return parseddescription.String()
}
