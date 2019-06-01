package s3

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

func getNewSession() (*session.Session, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2")},
	)
	return sess, err
}

func ListBucketItems(bucketName string) {
	sess, err := getNewSession()

	// Create S3 service client
	svc := s3.New(sess)
	resp, err := svc.ListObjects(&s3.ListObjectsInput{Bucket: aws.String(bucketName)})
	if err != nil {
		exitErrorf("Unable to list items in bucket %q, %v", bucketName, err)
	}

	for _, item := range resp.Contents {
		fmt.Println("Name:         ", *item.Key)
		fmt.Println("Last modified:", *item.LastModified)
		fmt.Println("Size:         ", *item.Size)
		fmt.Println("Storage class:", *item.StorageClass)
		fmt.Println("")
	}
}

func UploadFile(bucketName string, filename string) (bool, error) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Unable to open file %q, %v\n", filename, err)
		return false, err
	}

	defer file.Close()

	sess, err := getNewSession()
	uploader := s3manager.NewUploader(sess)

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(filename),
		Body:   file,
	})
	if err != nil {
		// Print the error and exit.
		fmt.Printf("Unable to upload %q to %q, %v\n", filename, bucketName, err)
		return false, err
	}

	fmt.Printf("Successfully uploaded %q to %q\n", filename, bucketName)
	return true, nil
}

func DownloadFile(bucketName string, filename string) (bool, error) {
	// Remove file locally if already exists
	_, err := os.Stat(filename)
	if err == nil {
		os.Remove(filename)
	}

	sess, err := getNewSession()
	downloader := s3manager.NewDownloader(sess)

	// Create a file to write the S3 Object contents to.
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error in creating s3 file: ", filename, err)
		return false, err
	}

	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(filename),
		})
	if err != nil {
		fmt.Println("Error in fetching s3 file: ", filename, err)
		return false, err
	}

	fmt.Println("Downloaded", file.Name(), numBytes, "bytes")
	return true, nil
}
