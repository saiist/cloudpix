package config

import "os"

type Config struct {
	S3BucketName      string
	TagsTableName     string
	MetadataTableName string
	AWSRegion         string
}

func NewConfig() *Config {
	return &Config{
		S3BucketName:      os.Getenv("S3_BUCKET_NAME"),
		TagsTableName:     os.Getenv("TAGS_TABLE_NAME"),
		MetadataTableName: os.Getenv("METADATA_TABLE_NAME"),
		AWSRegion:         os.Getenv("AWS_REGION"),
	}
}
