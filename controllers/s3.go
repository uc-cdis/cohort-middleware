package controllers

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gin-gonic/gin"
	"github.com/uc-cdis/cohort-middleware/config"
	"github.com/uc-cdis/cohort-middleware/models"
)

type S3 struct {
	AWSConfig    *aws.Config
	S3BucketName string
}

func (s3 *S3) init(err error) {
	c := config.GetConfig()

	credsConfig := credentials.NewStaticCredentials(c.GetString("s3.userid"), c.GetString("s3.secret"), "")
	awsConfig := &aws.Config{
		Region:      aws.String(c.GetString("s3.region")),
		Credentials: credsConfig,
	}

	s3.AWSConfig = awsConfig
	s3.S3BucketName = c.GetString("s3.bucket")
}

func (s3 *S3) newS3Session() *session.Session {
	return session.Must(session.NewSession(s3.AWSConfig))
}

func (s3 *S3) retrieveFile(sourceName string, c *gin.Context) (file *bytes.Buffer, err error) {
	var cohortPhenotypeDataModel = new(models.CohortPhenotypeData)
	var f CohortPhenotypeData
	c.Bind(&f)

	if sourceName != "" {
		cohort, err := cohortPhenotypeDataModel.GetCohortDataPhenotype(sourceName)
		if err != nil {
			return nil, err
		}

		format := strings.ToLower(f.Format)

		if format == "tsv" || format == "csv" {
			b := GenerateCsv(format, cohort)
			return b, nil
		}
		return nil, fmt.Errorf("phenotype data is not in tsv or csv format, %s", err)
	}

	return nil, fmt.Errorf("sourceName was not specified")
}

func (s3 *S3) uploadFile(path string, sourceName string, context *gin.Context) (err error) {
	sess := s3.newS3Session()
	uploader := s3manager.NewUploader(sess)
	file, err := s3.retrieveFile(sourceName, context)
	if err != nil {
		return err
	}
	reader := bytes.NewReader(file.Bytes())

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s3.S3BucketName),
		Key:    aws.String(path),
		Body:   reader,
	})

	if err != nil {
		return err
	}

	return
}
