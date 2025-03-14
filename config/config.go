package config

import "os"

type Config struct {
	S3BucketName      string
	DynamoDBTableName string
	TagsTableName     string
	MetaTableName     string
	AWSRegion         string
}

func NewConfig() *Config {
	return &Config{
		S3BucketName:      os.Getenv("S3_BUCKET_NAME"),
		DynamoDBTableName: os.Getenv("DYNAMODB_TABLE_NAME"),
		TagsTableName:     os.Getenv("TAGS_TABLE_NAME"),
		MetaTableName:     os.Getenv("METADATA_TABLE_NAME"),
		AWSRegion:         os.Getenv("AWS_REGION"),
	}
}
