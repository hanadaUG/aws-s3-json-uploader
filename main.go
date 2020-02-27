package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"log"
	"time"

	"os"

	"github.com/gocarina/gocsv"
)

// authentication information
type AuthInfo struct {
	Id       string `json:"username"`
	Password string `json:"password"`
	LastUpdated string `json:"updatedAt"`
}

// aws credentials.csv
type Csv struct {
	UserName         string `csv:"User name"`
	Password         string `csv:"Password"`
	AccessKeyID      string `csv:"Access key ID"`
	SecretAccessKey  string `csv:"Secret access key"`
	ConsoleLoginLink string `csv:"Console login link"`
}

func main() {
	// 使用法を記載します
	flag.Usage = func() {
		usageTxt := `Usage example:
$ aws-s3-json-uploader <id> <password>`

		fmt.Fprintf(os.Stderr, "%s\n", usageTxt)
	}

	// コマンドライン引数をパースします
	flag.Parse()
	args := flag.Args()

	// コマンドライン引数の数が期待値ではなかったら終了します
	if len(args) != 2 {
		flag.Usage()
		return
	}

	// jsonデータを生成します
	t := time.Now()
	const layout = "20060102"
	data, _ := json.MarshalIndent(AuthInfo{Id: args[0], Password: args[1],LastUpdated: t.Format(layout)}, "", "\t")
	fmt.Println(string(data))

	// AWS S3にアクセスするためひ必要なアクセスキーとシークレットキーをcsvから読み込みます
	// IAMでユーザーを新規作成時にダウンロードできる「credentials.csv」を読み込みます
	file, err := os.Open("./credentials.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var csv []*Csv
	if err := gocsv.UnmarshalFile(file, &csv); err != nil {
		panic(err)
	}

	// S3と接続するためのインスタンスを生成します
	cred := credentials.NewStaticCredentials(csv[0].AccessKeyID, csv[0].SecretAccessKey, "")
	sess, err := session.NewSession(&aws.Config{
		Credentials: cred,
		Region:      aws.String(s3.BucketLocationConstraintApNortheast1)},
	)
	client := s3.New(sess)

	// jsonをS3にアップロードします
	_, err = client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String("xxx"), // TODO:
		Key:         aws.String("./test.json"),
		ContentType: aws.String("application/json"),
		Body:        bytes.NewReader(data),
	})
	if err != nil {
		log.Fatal(err)
	}
}
