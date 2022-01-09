package collect

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/jyothri/hdd/db"
	"google.golang.org/api/iterator"
)

func CloudStorage(bucketName string) int {
	scanId := db.LogStartScan("google_storage")
	go startCloudStorage(scanId, bucketName)
	return scanId
}

func startCloudStorage(scanId int, bucketName string) {
	lock.Lock()
	defer lock.Unlock()
	ParseInfo = make([]db.FileData, 0)
	ctx := context.Background()

	// Create a client.
	client, err := storage.NewClient(ctx)
	checkError(err)
	defer client.Close()

	// Create a Bucket instance.
	bucket := client.Bucket(bucketName)

	query := &storage.Query{Prefix: ""}

	it := bucket.Objects(ctx, query)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		checkError(err)
		fd := db.FileData{
			FilePath:  attrs.MediaLink,
			IsDir:     false,
			ModTime:   attrs.Updated,
			FileCount: 1,
			Size:      uint(attrs.Size),
			Md5Hash:   fmt.Sprintf("%x", attrs.MD5),
		}
		fileName := getFileName(attrs.Name)
		fd.FileName = fileName
		ParseInfo = append(ParseInfo, fd)
	}
	db.SaveStatsToDb(scanId, &ParseInfo)
	db.LogCompleteScan(scanId)
}

func getFileName(objectPath string) string {
	fileParts := strings.Split(objectPath, "/")
	if len(fileParts) < 1 {
		panic("Does not have a valid filename. ObjectPath:" + objectPath)
	}
	return fileParts[len(fileParts)-1]
}
