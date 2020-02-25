# aws-s3-json-uploader
idとパスワードをコマンドライン引数で受け取り、jsonファイルに変換してS3にアップロードします<br>

## 構成
```
$ tree
.
├── README.md
├── aws-s3-json-uploader
├── credentials.csv
├── go.mod
├── go.sum
└── main.go
```

## AWSのアクセスキーとシックレットキー
`credentials.csv`<br>
AWS IAMで新規にユーザーを作成した時にだけダウンロードできるcsvファイルです<br>
`.gitignore`でgitの管理外にしています
```
User name,Password,Access key ID,Secret access key,Console login link
aws-s3-json-uploader,,hogehoge,fugafuga,https://xxx.signin.aws.amazon.com/console
```

## 使い方
```
# コマンドライン引数でidとパスワードを渡すとjsonファイルとしてS3に保存されます
$ go run main.go  <id> <password>

# 実行ファイル
$ ./aws-s3-json-uploader <id> <password>
```