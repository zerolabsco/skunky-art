package app

import (
	"bufio"
	"bytes"
	"encoding/json"
	"html/template"
	"os"
	"time"
)

func ExecuteCommandLineArguments() {
	var helpmsg = `SkunkyArt v{{.Version}} [{{.Description}}]
Usage:
	- [-c|--config] 	| path to config
	- [-a|--add-instance]	| generates 'instances.json' and 'INSTANCES.md' files with ur instance
	- [-h|--help]		| returns this message
Example:
	./skunkyart -c config.json
Copyright lost+skunk, X11. https://git.macaw.me/skunky/skunkyart/src/tag/v{{.Version}}`

	a := os.Args[1:]
	for n, x := range a {
		switch x {
		case "-c", "--config":
			if len(a) >= 2 {
				CFG.cfg = a[n+1]
			} else {
				exit("Not enought arguments", 1)
			}
		case "-h", "--help":
			var buf bytes.Buffer
			t := template.New("help")
			t.Parse(helpmsg)
			t.Execute(&buf, &Release)
			exit(buf.String(), 0)
		case "-a", "--add-instance":
			addInstance()
		}
	}
}

type settingsUrls struct {
	I2P      string `json:"i2p,omitempty"`
	Ygg      string `json:"ygg,omitempty"`
	Tor      string `json:"tor,omitempty"`
	Clearnet string `json:"clearnet,omitempty"`
}

type settingsParams struct {
	Nsfw  bool `json:"nsfw"`
	Proxy bool `json:"proxy"`
}

type settings struct {
	Title       string         `json:"title"`
	Country     string         `json:"country"`
	ModifiedSrc string         `json:"modified-src,omitempty"`
	Urls        settingsUrls   `json:"urls"`
	Settings    settingsParams `json:"settings"`
}

func addInstance() {
	prompt := func(txt string, necessary bool) string {
		input := bufio.NewScanner(os.Stdin)
		for {
			print(txt)
			print(": ")
			input.Scan()

			if i := input.Text(); necessary && i == "" {
				println("Please specify the", txt)
			} else {
				return i
			}
		}
	}

	var settingsVar struct {
		Instances []settings `json:"instances"`
	}
	instancesJson, err := os.OpenFile("instances.json", os.O_CREATE|os.O_WRONLY, 0644)
	try(err)
	defer instancesJson.Close()

	instancesFile, err := os.OpenFile("INSTANCES.md", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	try(err)
	defer instancesFile.Close()

	for {
		if string(instances) == "" {
			print("\rDownloading instance list...")
		} else {
			println("\r\033[2KDownloaded!")
			try(json.Unmarshal(instances, &settingsVar))

			settingsVar.Instances = append(settingsVar.Instances, settings{
				Title:       prompt("Title", true),
				Country:     prompt("Country", true),
				ModifiedSrc: prompt("Link to modified sources", false),
				Settings: settingsParams{
					Nsfw:  CFG.Nsfw,
					Proxy: CFG.Proxy,
				},
				Urls: settingsUrls{
					Clearnet: prompt("Clearnet link", false),
					Ygg:      prompt("Yggdrasil link", false),
					Tor:      prompt("Onion link", false),
					I2P:      prompt("I2P link", false),
				},
			})

			j, err := json.MarshalIndent(&settingsVar, "", "    ")
			try(err)

			instancesJson.Write(j)

			settingsVar := &settingsVar.Instances[len(settingsVar.Instances)-1]
			var mdstr bytes.Buffer

			mdbuilder := func(yes bool, link string, title string) {
				switch {
				case yes && (title != "" && link != ""):
					mdstr.WriteString("[")
					mdstr.WriteString(title)
					mdstr.WriteString("](")
					mdstr.WriteString(link)
					mdstr.WriteString(")")
				case yes && link != "":
					mdstr.WriteString("[Yes](")
					mdstr.WriteString(link)
					mdstr.WriteString(")")
				case yes:
					mdstr.WriteString("Yes")
				default:
					mdstr.WriteString("No")
				}
				mdstr.WriteString("|")
			}

			mdstr.WriteString("\n|")
			mdbuilder(settingsVar.Urls.Clearnet != "", settingsVar.Urls.Clearnet, settingsVar.Title)

			urls := []string{settingsVar.Urls.Ygg, settingsVar.Urls.I2P, settingsVar.Urls.Tor}
			for i, l := 0, len(urls); i < l; i++ {
				url := urls[i]
				mdbuilder(url != "", url, "")
			}

			settings := []bool{settingsVar.Settings.Nsfw, settingsVar.Settings.Proxy}
			for i, l := 0, len(settings); i < l; i++ {
				mdbuilder(settings[i], "", "")
			}

			mdbuilder(settingsVar.ModifiedSrc != "", settingsVar.ModifiedSrc, "")

			mdstr.WriteString(settingsVar.Country)
			mdstr.WriteString("|")

			instancesFile.Write(mdstr.Bytes())
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	exit("Done! Now add the files 'instances.json' and 'INSTANCES.md' to the 'master' branch in the repository https://git.macaw.me/skunky/SkunkyArt", 0)
}
