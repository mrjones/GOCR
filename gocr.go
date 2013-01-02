// $ go get code.google.com/p/goauth2/oauth
// $ go get code.google.com/p/google-api-go-client/drive/v2

package main

import (
	oauth "code.google.com/p/goauth2/oauth"
	drive "code.google.com/p/google-api-go-client/drive/v2"

	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
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

	transport := &oauth.Transport{Config: config}

	tokenCache := oauth.CacheFile("tokens.cache")

	token, err := tokenCache.Token()
	if err != nil {
		url := config.AuthCodeURL("")
		fmt.Println("Visit this URL to get a code, then type it in below:")
		fmt.Println(url)

		verificationCode := ""
		fmt.Scanln(&verificationCode)

		token, err := transport.Exchange(verificationCode)
		if err != nil {
			return nil, err
		}

		err = tokenCache.PutToken(token)
		if err != nil {
			log.Printf("Error: %s\n", err)
		}
	} else {
		transport.Token = token
	}

	return transport.Client(), nil
}

func uploadFile(service *drive.Service, localFileName string) error {
	log.Printf(" - file: %s\n", localFileName);
//	return nil

	localFile, err := os.Open(localFileName)
	if err != nil {
		return err
	}

	driveFile := &drive.File{Title: path.Base(localFileName)}
	// TODO(mrjones): Make directory configurable
  parent := &drive.ParentReference{Id: "0B1SxUBEP5_X2ZEdMaW45Qy1KcFk"}
  driveFile.Parents = []*drive.ParentReference{parent}
	_, err = service.Files.Insert(driveFile).Ocr(true).OcrLanguage("en").Media(localFile).Do();
	if err != nil {
		return err
	}

	return nil
}

type FilePredicate interface {
	Apply(info os.FileInfo) bool;
}

type AlwaysTrue struct { }

func (*AlwaysTrue) Apply(info os.FileInfo) bool {
	if strings.HasPrefix(info.Name(), ".") {
		return false;
	}
	return true;
}

func uploadDirectory(service *drive.Service, path string, shouldUpload FilePredicate) error {
	log.Printf("Uploading directory: %s\n", path);
	dir, err := os.Open(path)
	if err != nil {
		return err
	}

	info, err := dir.Stat()
	if err != nil {
		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory.", path)
	}

	files, err := dir.Readdir(-1)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			err = uploadDirectory(service, fmt.Sprintf("%s/%s", path, file.Name()), shouldUpload);
		} else {
			if shouldUpload.Apply(file) {
				err = uploadFile(service, fmt.Sprintf("%s/%s", path, file.Name()))
			}
		}
		if err != nil {
			return err
		}
	}

	return nil
}



func main() {
	var directory *string = flag.String(
		"directory",
		"/home/mrjones/scans",
		"Directory to (recursively) upload.")

	flag.Parse()

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

//	err = uploadFile(service, "/home/mrjones/gocrtest.jpg")
//	if err != nil {
//		log.Fatal(err)
//	}

	err = uploadDirectory(service, *directory, &AlwaysTrue{})
	if err != nil {
		log.Fatal(err)
	}
}
