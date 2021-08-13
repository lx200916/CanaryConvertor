package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "hcyconverter",
		Usage: "Convert HttpCanary zip achieve to Postman Collection",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Value:   "./",
				Usage:   "Output Path",
			},
		},
		Action: func(c *cli.Context) error {
			var name string
			if c.NArg() > 0 {
				name = c.Args().Get(0)
			} else {
				log.Fatal("Input File not specified")
			}
			if !strings.HasSuffix(name, ".zip") {
				log.Fatal("Only HttpCanary zip achieve supported.")

			}
			outputPath := c.String("output")
			_, err := os.Stat(name)
			if err != nil {
				log.Fatal("Invalid Input File")
			}

			file, err := os.Stat(outputPath)
			if err != nil {
				if os.IsNotExist(err) {
					err := os.MkdirAll(outputPath, 0755)
					if err != nil {
						fmt.Println(err)
						log.Fatal("Error Creating output directory")
					}
				} else {
					log.Fatal("Invalid output directory")

				}
			}
			if !file.IsDir() {
				log.Fatal("Invalid output directory")
			}

			return toPostman(name, outputPath)
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
