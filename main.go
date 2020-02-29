package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gocarina/gocsv"
)

// authentication information
type AuthInfo struct {
	UserName    string `json:"username"`
	Password    string `json:"password"`
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
	const layout = "2006/01/02"
	data, err := json.MarshalIndent(AuthInfo{UserName: args[0], Password: args[1], LastUpdated: t.Format(layout)}, "", "\t")
	if err != nil {
		panic(err)
	}
	fmt.Printf("json  : %s (%d byte)\n", string(data), len(data))
	fmt.Printf("json  : %x (%d byte)\n", data, len(data))

	// 暗号化
	key := []byte("00000000000000000000000000000000") // 32 bytes => AES-256
	fmt.Printf("key   : %s (%d bytes)\n", key, len(key))

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	iv := []byte("0000000000000000")
	fmt.Printf("iv    : %s (%d bytes)\n", iv, len(iv))

	cbc := cipher.NewCBCEncrypter(block, iv)
	padding := PKCS5Padding(data)
	fmt.Printf("PKCS#5: %x (%d bytes)\n", padding, len(padding))

	encrypt := make([]byte, len(padding))
	cbc.CryptBlocks(encrypt, padding)
	fmt.Printf("暗号化: %x (%d bytes)\n", encrypt, len(encrypt))

	enc := base64.StdEncoding.EncodeToString(encrypt)
	fmt.Printf("符号化: %s (%d byte)\n", enc, len(enc))

	//dec, err := base64.StdEncoding.DecodeString(enc)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Printf("decode: %x\n", dec)
	//plainText := make([]byte, len(dec))
	//cbc = cipher.NewCBCDecrypter(block, iv)
	//cbc.CryptBlocks(plainText, dec)
	//fmt.Printf("復号化: %s (%d bytes)\n", PKCS5Trimming(plainText), len(PKCS5Trimming(plainText)))

	// ローカルに暗号化したバイナリーデータを保存します
	//ioutil.WriteFile("encrypt", []byte(enc), 0644)

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
		Bucket: aws.String("xxx"), // TODO:
		Key:    aws.String("./sample"),
		Body:   bytes.NewReader([]byte(enc)),
		ACL:    aws.String(s3.BucketCannedACLPublicRead),
	})
	if err != nil {
		log.Fatal(err)
	}
}

// PKCS#5 Padding
// ブロック長に満たないサイズ（＝埋めるサイズ）の値を表すバイト値で足りない分を埋めます
// (例) ブロック長が16バイトでデータサイズが72バイトなら0x08で足りない8バイト分埋めます
// ブロック長ぴったりな場合は1ブロック分丸ごとパディングされます
func PKCS5Padding(data []byte) []byte {
	//fmt.Printf("data: %d bytes\n", len(data))
	padding := aes.BlockSize - len(data)%aes.BlockSize
	//fmt.Printf("padding: %xbytes\n", padding)
	pad := bytes.Repeat([]byte{byte(padding)}, padding)
	//fmt.Printf("pad: %x\n", pad)
	return append(data, pad...)
}

func PKCS5Trimming(encrypt []byte) []byte {
	padding := encrypt[len(encrypt)-1]
	return encrypt[:len(encrypt)-int(padding)]
}
