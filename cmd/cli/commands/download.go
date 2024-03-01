package commands

import (
	"archive/zip"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

const (
	RootArtifactURL = "https://minecraft.azureedge.net/bin-linux/bedrock-server-"
	ArtifactType    = ".zip"
)

var (
	artifactVersion string
	artifactOutput  string
)

var (
	downloadCmd = &cobra.Command{
		Use:   "download",
		Short: "Download a Minecraft server artifact",
		Long:  "Download a Minecraft server artifact",
		Run: func(cmd *cobra.Command, args []string) {
			// Do Stuff Here

			install := MinecraftInstall{}

			version := args[0]
			fmt.Printf("Downloading Minecraft server version %s\n", version)

			if version != "" || len(version) > 0 {
				install.Download(version, artifactOutput).Extract()
			}

		},
	}
)

func init() {
	downloadCmd.Flags().StringVarP(&artifactVersion, "version", "v", "", "Minecraft server version")
	downloadCmd.Flags().StringVarP(&artifactOutput, "output", "o", "", "Output directory")
}

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
