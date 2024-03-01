package commands

import (
	"MinecraftServerController/internal"
	"bytes"
	"html/template"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var (
	generateCmd = &cobra.Command{
		Use:   "generate",
		Short: "Generate a Minecraft server.properties file",
		Long:  "Generate a Minecraft server.properties file",
		Run: func(cmd *cobra.Command, args []string) {
			if err := generateProperties(); err != nil {
				log.Fatal(err)
			}
		},
	}
)

func generateProperties() error {
	log.Println("Generating server.properties file")

	serverProperties := internal.NewServerProperties()
	serverProperties.AppName = ""
	serverProperties.AppVersion = ""

	tmplFile := "templates/1.20.20.tmpl"
	_, err := os.Stat(tmplFile)
	if err != nil {
		return err
	}

	template, err := template.ParseFiles(tmplFile)
	// Capture any error
	if err != nil {
		return err
	}

	var buffer = new(bytes.Buffer)
	// Print out the template to std
	template.Execute(buffer, serverProperties)

	// Write the buffer to a file
	file, err := os.Create("server.properties")
	if err != nil {
		return err
	}
	defer file.Close()
	file.Write(buffer.Bytes())
	log.Println("Generated server.properties file")
	return nil // Do Stuff Here
}
