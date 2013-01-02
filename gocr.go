package main

import (
	oauth "code.google.com/p/goauth2/oauth"
	drive "code.google.com/p/google-api-go-client/drive/v2"

	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type ConfigFile struct {
	ClientId string
	ClientSecret string
}

func parseConfigFile(filename string) (*ConfigFile, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var configFile ConfigFile;
	err = json.Unmarshal(bytes, &configFile);
	if err != nil {
		return nil, err
	}

	return &configFile, nil
}

func authorize(configFile *ConfigFile) (*http.Client, error) {
	config := &oauth.Config{
		ClientId: configFile.ClientId,
		ClientSecret: configFile.ClientSecret,
//		Scope: "https://www.googleapis.com/auth/drive.file",
		Scope: "https://www.googleapis.com/auth/drive",
		AuthURL: "https://accounts.google.com/o/oauth2/auth",
		TokenURL: "https://accounts.google.com/o/oauth2/token",
		RedirectURL: "oob",
	}

	url := config.AuthCodeURL("")
  fmt.Println("Visit this URL to get a code, then type it in below:")
  fmt.Println(url)

	verificationCode := ""
	fmt.Scanln(&verificationCode)


	transport := &oauth.Transport{Config: config}
	_, err := transport.Exchange(verificationCode)
	if err != nil {
		return nil, err
	}

	return transport.Client(), nil
}

func upload(service *drive.Service, localFileName string) error {
	localFile, err := os.Open(localFileName)
	if err != nil {
		return err
	}

	driveFile := &drive.File{Title: "GOCR File"}
	_, err = service.Files.Insert(driveFile).Ocr(true).OcrLanguage("en").Media(localFile).Do();
	if err != nil {
		return err
	}

	return nil
}

func main() {
	configFile, err := parseConfigFile("config.json");
	if err != nil {
		log.Fatal(err)
	}

	httpClient, err := authorize(configFile);
	if err != nil {
		log.Fatal(err)
	}


	service, err := drive.New(httpClient)
	if err != nil {
		log.Fatal(err)
	}

// Code to list files, just to test that things work
//	res, err := service.Files.List().Do()
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	for _, f := range res.Items {
//		fmt.Printf("%s (%s)\n", f.Id, f.Title);
//	}

	err = upload(service, "/home/mrjones/gocrtest.jpg")
	if err != nil {
		log.Fatal(err)
	}

}
