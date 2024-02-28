package main

import (
	"MinecraftServerController/internal"
	"archive/zip"
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const (
	AppName    = "CraftCube"
	AppVersion = "1.0.0-alpha"

	RootArtifactURL = "https://minecraft.azureedge.net/bin-linux/bedrock-server-"
	ArtifactType    = ".zip"
)

type Color string

const (
	ColorBlack  Color = "\u001b[30m"
	ColorRed          = "\u001b[31m"
	ColorGreen        = "\u001b[32m"
	ColorYellow       = "\u001b[33m"
	ColorBlue         = "\u001b[34m"
	ColorReset        = "\u001b[0m"
)

func colorize(color Color, message string) {
	fmt.Println(string(color), message, string(ColorReset))
}

var skipExtract bool = false
var skipDownload bool = false

type MinecraftInstall struct {
	Version                string
	ArtifactURL            string
	Filepath               string
	TargetPath             string
	GeneratePropertiesFile bool
}

func (ma *MinecraftInstall) Download(version string, target string) *MinecraftInstall {
	ma.Version = version
	ma.TargetPath = target

	ma.Filepath = ma.TargetPath + "/bedrock-server-" + ma.Version + ".zip"
	fmt.Println("Downloading Minecraft server version ", ma.Version)
	fmt.Println("Target path: ", ma.TargetPath)

	if err := ma.download(); err != nil {
		log.Fatal(err)
	}

	return ma
}

func (ma *MinecraftInstall) Extract() *MinecraftInstall {
	if err := unzipFile(ma.Filepath, ma.TargetPath); err != nil {
		log.Fatal(err)
	}
	return ma
}

func (ma *MinecraftInstall) GenerateConfig() *MinecraftInstall {
	if ma.GeneratePropertiesFile {
		err := generateServerProperties()
		if err != nil {
			log.Fatal(err)
		}
	}
	return nil
}

func (ma MinecraftInstall) artifactExists() bool {
	if _, err := os.Stat(ma.Filepath); err == nil {
		log.Printf("File %s already exists. Skipping download...\n", ma.Filepath)
		return true
	}
	return false
}

func (ma *MinecraftInstall) download() error {
	log.Println("Downloading Minecraft server artifact")

	fullURLFile := RootArtifactURL + ma.Version + ArtifactType

	// Build fileName from fullPath
	fileURL, err := url.Parse(fullURLFile)
	if err != nil {
		return fmt.Errorf("Error parsing URL %s", fullURLFile)
	}
	path := fileURL.Path
	segments := strings.Split(path, "/")

	fileName := segments[len(segments)-1]
	fileName = ma.TargetPath + "/" + fileName

	if _, err := os.Stat(ma.TargetPath); os.IsNotExist(err) {
		os.MkdirAll(ma.TargetPath, os.ModePerm)
	}

	log.Println("-----> ", fileName)

	if artifactExists(fileName) {
		log.Println("**************** File already exists. Skipping download... ****************")
		return nil
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
		Transport: tr,
	}

	log.Printf("Downloading %s from %s\n", fileName, fullURLFile)
	// Put content on file
	resp, err := client.Get(fullURLFile)
	if err != nil {
		return fmt.Errorf("Error downloading file %s", fileName)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Error: status code received: %d", resp.StatusCode)
	} else {
		log.Printf("File found. Starting download...")
	}

	// Create blank file
	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("Error creating file %s", fileName)
	}

	size, err := io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("Error writing to file %s", fileName)
	}

	defer file.Close()

	log.Printf("Downloaded %s (size:  %d KB)", fileName, size)

	return nil

}

// https://minecraft.azureedge.net/bin-linux/bedrock-server-1.20.62.02.zip
func main() {
	//useColor := flag.Bool("color", false, "display colorized output")
	artifactVersion := flag.String("version", "", "specify the version of the Minecraft server to download. E.g. 1.20.62.02")
	//generateProperties := flag.Bool("generate", false, "generate a server.properties file")
	artifactOutput := flag.String("output", "", "specify the output file directory for the server artifact")
	flag.Parse()

	log.Println("Kubecraft CLI v0.1.0")

	install := MinecraftInstall{}

	if *artifactVersion != "" {
		install.Download(*artifactVersion, *artifactOutput).Extract().GenerateConfig()
	}

}

func generateServerProperties() error {
	log.Println("Generating server.properties file")

	serverProperties := internal.NewServerProperties()
	serverProperties.AppName = AppName
	serverProperties.AppVersion = AppVersion

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
	return nil
}

func downloadServerArtifact(version string, output string) {

}

// url:
// https://minecraft.azureedge.net/bin-linux/bedrock-server-1.20.62.02.zip

func artifactExists(filepath string) bool {
	_, err := os.Stat(filepath)
	if err == nil {
		log.Printf("File %s already exists. Skipping download...\n", filepath)
		return true
	}
	log.Printf("File %s does not exist in cache.\n", filepath)
	return false

}

func unzipFile(zipfilepath string, dst string) error {
	log.Println("Unzipping file ", zipfilepath)

	archive, err := zip.OpenReader(zipfilepath)
	if err != nil {
		return err
	}
	defer archive.Close()

	if _, err := os.Stat(dst); os.IsNotExist(err) {
		os.MkdirAll(dst, os.ModePerm)
	}

	for _, f := range archive.File {
		filePath := filepath.Join(dst, f.Name)
		//fmt.Println("-- unzipping file ", filePath)

		if !strings.HasPrefix(filePath, filepath.Clean(dst)+string(os.PathSeparator)) {
			return fmt.Errorf("%s: illegal file path", filePath)
		}
		if f.FileInfo().IsDir() {
			//fmt.Println("creating directory...")
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return err
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		fileInArchive, err := f.Open()
		if err != nil {
			return err
		}

		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			return err
		}

		dstFile.Close()
		fileInArchive.Close()
	}

	log.Printf("Extracted %d files\n", len(archive.File))

	return nil
}

/*
func rcon() {
	// command: /opt/minecraft/tools/mcrcon/mcrcon -H 127.0.0.1 -P <PORT> -p <PASSWORD> "$1"
}
*/
