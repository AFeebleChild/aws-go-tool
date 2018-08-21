package s3

import (
	"bytes"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type CopyObjectStruct struct {
	Bucket     string
	CopySource string
	Key        string
}

type GetObjectStruct struct {
	Bucket string
	Key    string
}

type PutObjectStruct struct {
	Body   []byte
	Bucket string
	Key    string
}

func GoCopyObject(sess *session.Session, input CopyObjectStruct) error {
	params := &s3.CopyObjectInput{
		Bucket:     aws.String(input.Bucket),
		CopySource: aws.String(input.CopySource),
		Key:        aws.String(input.Key),
	}

	_, err := s3.New(sess).CopyObject(params)
	if err != nil {
		return err
	}
	return nil
}

func GoGetObject(sess *session.Session, input GetObjectStruct) (io.ReadCloser, error) {
	params := &s3.GetObjectInput{
		Bucket: aws.String(input.Bucket),
		Key:    aws.String(input.Key),
	}

	resp, err := s3.New(sess).GetObject(params)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func GoPutObject(sess *session.Session, input PutObjectStruct) error {
	params := &s3.PutObjectInput{
		Body:   bytes.NewReader(input.Body),
		Bucket: aws.String(input.Bucket),
		Key:    aws.String(input.Key),
	}

	_, err := s3.New(sess).PutObject(params)
	if err != nil {
		return err
	}
	return nil
}
