package localstack_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/eriktate/go-localstack"
)

func Test_S3(t *testing.T) {
	// SETUP
	ctx := context.TODO()
	bucket := "test-bucket"
	key := "testKey"
	content := "hello, world!"

	instance, err := localstack.New()
	if err != nil {
		t.Fatal(err)
	}

	if err := instance.Wait(20 * time.Second); err != nil {
		instance.Close()
		t.Fatal(err)
	}

	s3client := s3.New(instance.Config())
	s3client.ForcePathStyle = true // required for localhost testing

	bucketInput := s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	}

	putInput := s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader([]byte(content)),
	}

	getInput := s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	// RUN
	if _, err := s3client.CreateBucketRequest(&bucketInput).Send(ctx); err != nil {
		instance.Close()
		t.Fatalf("unexpected error creating bucket: %s", err)
	}

	if _, err := s3client.PutObjectRequest(&putInput).Send(ctx); err != nil {
		instance.Close()
		t.Fatalf("unexpected error creating object: %s", err)
	}

	object, err := s3client.GetObjectRequest(&getInput).Send(ctx)
	if err != nil {
		instance.Close()
		t.Fatalf("unexpected error retrieving object: %s", err)
	}

	data, err := ioutil.ReadAll(object.Body)
	if err != nil {
		instance.Close()
		t.Fatalf("unexpected error reading file: %s", err)
	}

	// ASSERT
	if string(data) != content {
		instance.Close()
		t.Fatalf("content fetched from localstack s3 should have matched the test content")
	}

	// CLEANUP
	instance.Close()
}

func Test_Dynamo(t *testing.T) {
	// SETUP
	ctx := context.TODO()
	tableName := "test_table"

	instance, err := localstack.New(localstack.WithServices("dynamodb"))
	if err != nil {
		t.Fatal(err)
	}

	if err := instance.Wait(20 * time.Second); err != nil {
		instance.Close()
		t.Fatal(err)
	}

	dynamoClient := dynamodb.New(instance.Config())

	keySchema := dynamodb.KeySchemaElement{
		KeyType:       dynamodb.KeyTypeHash,
		AttributeName: aws.String("id"),
	}

	idDef := dynamodb.AttributeDefinition{
		AttributeName: aws.String("id"),
		AttributeType: dynamodb.ScalarAttributeTypeS,
	}

	createTableInput := dynamodb.CreateTableInput{
		TableName:            aws.String(tableName),
		KeySchema:            []dynamodb.KeySchemaElement{keySchema},
		AttributeDefinitions: []dynamodb.AttributeDefinition{idDef},
		BillingMode:          dynamodb.BillingModePayPerRequest,
	}

	// RUN
	if _, err := dynamoClient.CreateTableRequest(&createTableInput).Send(ctx); err != nil {
		instance.Close()
		t.Fatalf("unexpected error creating dynamo table: %s", err)
	}

	// ASSERT

	// CLEANUP
	instance.Close()
}

func Test_SQS(t *testing.T) {
	// SETUP
	ctx := context.TODO()
	queueName := "test_queue"

	instance, err := localstack.New(localstack.WithServices("sqs"))
	if err != nil {
		t.Fatal(err)
	}

	if err := instance.Wait(20 * time.Second); err != nil {
		instance.Close()
		t.Fatal(err)
	}

	sqsClient := sqs.New(instance.Config())

	createQueueInput := sqs.CreateQueueInput{
		QueueName: aws.String(queueName),
	}

	// RUN
	if _, err := sqsClient.CreateQueueRequest(&createQueueInput).Send(ctx); err != nil {
		instance.Close()
		t.Fatalf("unexpected error creating sqs queue: %s", err)
	}

	// ASSERT

	// CLEANUP
	instance.Close()
}
