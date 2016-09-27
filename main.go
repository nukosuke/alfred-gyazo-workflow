package main

import (
	"errors"
	"fmt"
	"github.com/Tomohiro/go-gyazo/gyazo"
	"github.com/nukosuke/go-alfred"
	"github.com/spf13/viper"
	"github.com/urfave/cli"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

const (
	MARKDOWN_FORMAT = "[![%s](%s)](%s)"
	HTML_FORMAT     = `<a href="%s"><img src="%s" alt="%s"/></a>`
)

func loadConfig() error {
	viper.SetConfigName("config")
	viper.AddConfigPath(os.Getenv("HOME") + "/.gyazo")
	viper.SetConfigType("json")

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "alfred-gyazo-workflow"

	app.Commands = []cli.Command{
		{
			Name: "fetch",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "format",
					Value: "markdown",
				},
			},
			Action: func(c *cli.Context) error {
				err := loadConfig()
				if err != nil {
					return err
				}

				access_token := viper.GetString("access_token")
				if access_token == "" {
					return errors.New("access_token was not found")
				}

				client, err := gyazo.NewClient(access_token)
				if err != nil {
					panic(err)
				}

				res, _ := client.List(&gyazo.ListOptions{Page: 1, PerPage: 25})

				HOME := os.Getenv("HOME")
				os.Mkdir(HOME+"/.gyazo", 0755)

				list := alfred.NewResponse()

				for _, item := range *res.Images {
					url := item.ThumbURL
					response, err := http.Get(url)
					if err != nil {
						panic(err)
					}
					defer response.Body.Close()

					file, err := os.Create(HOME + "/.gyazo/" + item.ID + ".png")
					if err != nil {
						panic(err)
					}

					_, err = io.Copy(file, response.Body)
					if err != nil {
						panic(err)
					}
					file.Close()

					var arg string
					switch c.String("format") {
					case "markdown":
						arg = fmt.Sprintf(MARKDOWN_FORMAT, item.PermalinkURL, item.URL, item.PermalinkURL)
					case "html":
						arg = fmt.Sprintf(HTML_FORMAT, item.PermalinkURL, item.URL, item.PermalinkURL)
					case "direct":
						arg = item.URL
					default:
						arg = item.URL
					}

					list.AddItem(&alfred.AlfredResponseItem{
						Valid:    true,
						Uid:      item.ID,
						Title:    item.CreatedAt,
						Arg:      arg,
						Subtitle: item.PermalinkURL,
						Icon: alfred.Icon{
							Type: "png",
							Path: HOME + "/.gyazo/" + item.ID + ".png",
						},
					})
				}

				list.PrintJSON()
				return nil
			},
		},
		{
			Name: "token",
			Action: func(c *cli.Context) error {
				HOME := os.Getenv("HOME")
				os.Mkdir(HOME+"/.gyazo", 0755)

				token := c.Args().Get(0)
				if token == "" {
					return errors.New("argument is empty")
				}

				ioutil.WriteFile(HOME+"/.gyazo/"+"config.json", []byte(`{"access_token":"`+token+`"}`), os.ModePerm)
				return nil
			},
		},
	}

	app.Run(os.Args)
}
