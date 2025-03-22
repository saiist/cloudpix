package config

import (
	"os"
	"strconv"
)

type Config struct {
	S3BucketName       string
	TagsTableName      string
	MetadataTableName  string
	AWSRegion          string
	UserPoolID         string
	ClientID           string
	Environment        string
	FunctionName       string
	EnableMetrics      bool
	EnableXRay         bool
	ImageRetentionDays int
}

func NewConfig() *Config {
	enableMetrics := os.Getenv("ENABLE_METRICS") == "true"
	enableXRay := os.Getenv("ENABLE_XRAY") == "true"

	// 保持日数の取得
	retentionDays := 90 // デフォルト値
	if retentionDaysStr := os.Getenv("RETENTION_DAYS"); retentionDaysStr != "" {
		if days, err := strconv.Atoi(retentionDaysStr); err == nil {
			retentionDays = days
		}
	}

	return &Config{
		S3BucketName:       os.Getenv("S3_BUCKET_NAME"),
		TagsTableName:      os.Getenv("TAGS_TABLE_NAME"),
		MetadataTableName:  os.Getenv("METADATA_TABLE_NAME"),
		AWSRegion:          os.Getenv("AWS_REGION"),
		UserPoolID:         os.Getenv("USER_POOL_ID"),
		ClientID:           os.Getenv("USER_POOL_CLIENT_ID"),
		Environment:        os.Getenv("ENVIRONMENT"),
		FunctionName:       os.Getenv("AWS_LAMBDA_FUNCTION_NAME"),
		EnableMetrics:      enableMetrics,
		EnableXRay:         enableXRay,
		ImageRetentionDays: retentionDays,
	}
}
