package config

import (
	"field-service/common/util"
	"os"

	"github.com/sirupsen/logrus"
	_ "github.com/spf13/viper/remote"
)

var Config AppConfig

type AppConfig struct {
	Port                  int             `json:"port"`
	AppName               string          `json:"appName"`
	AppEnv                string          `json:"appEnv"`
	SignatureKey          string          `json:"signatureKey"`
	Database              Database        `json:"database"`
	RateLimiterMaxRequest float64         `json:"rateLimiterMaxRequest"`
	RateLimiterTimeSecond int             `json:"rateLimiterTimeSecond"`
	InternalService       InternalService `json:"internalService"`

	//S3 Config
	S3AccessKeyID     string `json:"s3AccessKeyID"`
	S3SecretAccessKey string `json:"s3SecretAccessKey"`
	S3Region          string `json:"s3Region"`
	S3BucketName      string `json:"s3BucketName"`

	// GCS Config
	GCSType                    string `json:"gcsType"`
	GCSProjectID               string `json:"gcsProjectID"`
	GCSPrivateKeyID            string `json:"gcsPrivateKeyID"`
	GCSPrivateKey              string `json:"gcsPrivateKey"`
	GCSClientEmail             string `json:"gcsClientEmail"`
	GCSClientID                string `json:"gcsClientID"`
	GCSAuthURI                 string `json:"gcsAuthURI"`
	GCSTokenURI                string `json:"gcsTokenURI"`
	GCSAuthProviderX509CertURL string `json:"gcsAuthProviderX509CertURL"`
	GCSClientX509CertURL       string `json:"gcsClientX509CertURL"`
	GCSUniverseDomain          string `json:"gcsUniverseDomain"`
	GCSBucketName              string `json:"gcsBucketName"`
}

type Database struct {
	Host                  string `json:"host"`
	Port                  int    `json:"port"`
	Name                  string `json:"name"`
	Username              string `json:"username"`
	Password              string `json:"password"`
	MaxOpenConnections    int    `json:"maxOpenConnections"`
	MaxLifeTimeConnection int    `json:"maxLifeTimeConnection"`
	MaxIdleConnections    int    `json:"maxIdleConnections"`
	MaxIdleTime           int    `json:"maxIdleTime"`
}

type User struct {
	Host         string `json:"host"`
	SignatureKey string `json:"signatureKey"`
}

type InternalService struct {
	User User `json:"user"`
}

func Init() {
	err := util.BindFromJson(&Config, "config.json", ".")
	if err != nil {
		logrus.Infof("failed to bind config: %v", err)
		err = util.BindFromConsul(&Config, os.Getenv("CONSUL_HTTP_URL"), os.Getenv("CONSUL_HTTP_PATH"))
		if err != nil {
			panic(err)
		}
	}
}
