package main

import (
	"bufio"
	"context"
	"github.com/joho/godotenv"
	"google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"log"
	"os"
	"path/filepath"
)

func main() {
	log.Println("andpubutil: hello~")
	var err error

	if len(os.Args) < 3 {
		_, programName := filepath.Split(os.Args[0])
		log.Printf("Usage: %v your.package.name uploadfile1 [uploadfile2 ...]\n", programName)
		os.Exit(1)
		return
	}

	log.Println("andpubutil: load .env file")
	err = godotenv.Load(".env")
	if err != nil {
		log.Panicf("Environment file error: %v", err)
	}

	ctx := context.Background()
	androidpublisherService, err := androidpublisher.NewService(ctx, option.WithCredentialsFile(os.Getenv("ANDPUBUTIL_ANDROID_PUBLISHER_KEY")))
	if err != nil {
		log.Panicf("ANDPUBUTIL_ANDROID_PUBLISHER_KEY cred fail: %v", err)
	}

	packageName := os.Args[1]
	log.Printf("Package name: %v", packageName)

	appEdit, err := androidpublisherService.Edits.Insert(packageName, nil).Do()
	if err != nil {
		log.Panicf("Edits.Insert fail: %v", err)
	}

	log.Printf("AppEdit ID: %v", appEdit.Id)

	for _, uploadPath := range os.Args[2:] {
		uploadFile(androidpublisherService, packageName, appEdit, uploadPath)
	}

	commitResult, err := androidpublisherService.Edits.Commit(packageName, appEdit.Id).Do()
	if err != nil {
		log.Panicf("Commit result fail: %v", err)
	}

	log.Printf("Commit result status code: %v", commitResult.HTTPStatusCode)
}

func uploadFile(androidpublisherService *androidpublisher.Service, packageName string, appEdit *androidpublisher.AppEdit, apkPath string) {
	upload := androidpublisherService.Edits.Apks.Upload(packageName, appEdit.Id)
	dat, err := os.Open(apkPath)
	if err != nil {
		log.Panicf("File open error: %v", err)
	}

	reader := bufio.NewReader(dat)
	upload.Media(reader, googleapi.ContentType("application/octet-stream"))

	log.Printf("Uploading %v...", apkPath)

	uploadResult, err := upload.Do()
	if err != nil {
		log.Panicf("Upload error: %v", err)
	}

	log.Printf("Upload result status code: %v", uploadResult.HTTPStatusCode)
	log.Printf("Uploaded version code: %v", uploadResult.VersionCode)
}
