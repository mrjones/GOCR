package main

import (
	oauth "code.google.com/p/goauth2/oauth"
	drive "code.google.com/p/google-api-go-client/drive/v2"

	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
)

type ConfigFile struct {
	ClientId string
	ClientSecret string
}

func main() {
	fmt.Println("Hello World!");

	bytes, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatal(err)
	}

	var configFile ConfigFile;
	err = json.Unmarshal(bytes, &configFile);
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(configFile.ClientId);

	// OAuth
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
  fmt.Println("Visit this URL to get a code, then run again with -code=YOUR_CODE\n")
  fmt.Println(url)

	verificationCode := ""
	fmt.Scanln(&verificationCode)


	transport := &oauth.Transport{Config: config}
	_, err = transport.Exchange(verificationCode)
	if err != nil {
		log.Fatal(err)
	}

	// Drive

	dc, err := drive.New(transport.Client())
	if err != nil {
		log.Fatal(err)
	}

	res, err := dc.Files.List().Do()

	if err != nil {
		log.Fatal(err)
	}

	for _, f := range res.Items {
		fmt.Printf("%s (%s)\n", f.Id, f.Title);
	}
}
