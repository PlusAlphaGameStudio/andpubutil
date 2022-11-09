package main

import (
	"bufio"
	"context"
	"errors"
	"github.com/joho/godotenv"
	"google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	log.Println("andpubutil: hello~")
	// 리뷰 단계 스킵할 수 있는 것을 먼저 체크해서 시도하고,
	if err := upload(true); err != nil {
		log.Println("andpubutil: retrying with review flag on (i.e. skipReview=false)")
		// 실패하면 리뷰 단계 포함해서 시도한다.
		if err := upload(false); err != nil {
			panic(err)
		}
	}
	log.Println("andpubutil: Done successfully")
}

func upload(skipReview bool) error {

	var err error

	if len(os.Args) < 3 {
		_, programName := filepath.Split(os.Args[0])
		log.Printf("Usage: %v your.package.name uploadfile1 [uploadfile2 ...]\n", programName)
		os.Exit(1)
		return nil
	}

	log.Println("andpubutil: load .env file")
	_ = godotenv.Load(".env")

	ctx := context.Background()
	androidPublisherService, err := androidpublisher.NewService(ctx, option.WithCredentialsFile(os.Getenv("ANDPUBUTIL_ANDROID_PUBLISHER_KEY")))
	if err != nil {
		log.Printf("ANDPUBUTIL_ANDROID_PUBLISHER_KEY cred fail: %v", err)
		return err
	}

	packageName := os.Args[1]
	log.Printf("Package name: %v", packageName)

	appEdit, err := androidPublisherService.Edits.Insert(packageName, nil).Do()
	if err != nil {
		log.Printf("Edits.Insert fail: %v", err)
		return err
	}

	log.Printf("AppEdit ID: %v", appEdit.Id)

	var lastVersionCode int64
	for _, uploadPath := range os.Args[2:] {
		if strings.HasSuffix(uploadPath, ".apk") {
			lastVersionCode = uploadApkFile(androidPublisherService, packageName, appEdit, uploadPath)
		} else if strings.HasSuffix(uploadPath, ".aab") {
			lastVersionCode = uploadAabFile(androidPublisherService, packageName, appEdit, uploadPath)
		} else if strings.HasSuffix(uploadPath, ".obb") {
			uploadObbFile(androidPublisherService, packageName, appEdit, uploadPath, lastVersionCode)
		} else {
			log.Printf("unknown file extension to upload")
			return errors.New("unknown file extension to upload")
		}
	}

	commitCall := androidPublisherService.Edits.Commit(packageName, appEdit.Id)

	// 파일 업로드일 뿐이므로 리뷰 대상으로 만들 필요는 없다.
	commitCall.ChangesNotSentForReview(skipReview)

	commitResult, err := commitCall.Do()
	if err != nil {
		log.Printf("Commit result fail: %v", err)
		return err
	}

	log.Printf("Commit result status code: %v", commitResult.HTTPStatusCode)
	return nil
}

func uploadApkFile(androidPublisherService *androidpublisher.Service, packageName string, appEdit *androidpublisher.AppEdit, apkPath string) int64 {
	upload := androidPublisherService.Edits.Apks.Upload(packageName, appEdit.Id)
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
	return uploadResult.VersionCode
}

func uploadAabFile(androidPublisherService *androidpublisher.Service, packageName string, appEdit *androidpublisher.AppEdit, aabPath string) int64 {

	upload := androidPublisherService.Edits.Bundles.Upload(packageName, appEdit.Id)
	dat, err := os.Open(aabPath)
	if err != nil {
		log.Panicf("File open error: %v", err)
	}

	reader := bufio.NewReader(dat)
	upload.Media(reader, googleapi.ContentType("application/octet-stream"))

	log.Printf("Uploading %v...", aabPath)

	uploadResult, err := upload.Do()
	if err != nil {
		log.Panicf("Upload error: %v", err)
	}

	log.Printf("Upload result status code: %v", uploadResult.HTTPStatusCode)
	log.Printf("Uploaded version code: %v", uploadResult.VersionCode)
	return uploadResult.VersionCode
}

func uploadObbFile(androidPublisherService *androidpublisher.Service, packageName string, appEdit *androidpublisher.AppEdit, obbPath string, apkVersionCode int64) {
	upload := androidPublisherService.Edits.Expansionfiles.Upload(packageName, appEdit.Id, apkVersionCode, "main")
	dat, err := os.Open(obbPath)
	if err != nil {
		log.Panicf("File open error: %v", err)
	}

	reader := bufio.NewReader(dat)
	upload.Media(reader, googleapi.ContentType("application/octet-stream"))

	log.Printf("Uploading %v...", obbPath)

	uploadResult, err := upload.Do()
	if err != nil {
		log.Panicf("Upload error: %v", err)
	}

	log.Printf("Upload result status code: %v", uploadResult.HTTPStatusCode)
}