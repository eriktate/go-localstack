package localstack

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/defaults"
	"github.com/aws/aws-sdk-go-v2/aws/endpoints"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ory/dockertest"
)

type serviceResolver func(service, region string) (aws.Endpoint, error)

// An Instance keeps track of the localstack container state.
type Instance struct {
	host     string
	key      string
	secret   string
	session  string
	region   string
	services []string

	pool     *dockertest.Pool
	resource *dockertest.Resource
	resolver serviceResolver
}

// New spins up a new localstack container and returns an Instance tracking it.
func New(opts ...InstanceOpt) (*Instance, error) {
	instance := &Instance{}

	for _, opt := range opts {
		if err := opt(instance); err != nil {
			return nil, err
		}
	}

	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, err
	}

	resource, err := pool.Run("localstack/localstack", "", []string{instance.serviceString()})
	if err != nil {
		return nil, err
	}

	withDefaults(instance)
	instance.resolver = instance.makeResolver()
	instance.pool = pool
	instance.resource = resource

	return instance, nil
}

// An InstanceOpt is a configuration option for the New constructor.
type InstanceOpt func(instance *Instance) error

// WithHost sets the Instance host value.
func WithHost(host string) InstanceOpt {
	return func(i *Instance) error {
		i.host = host
		return nil
	}
}

// WithCredentials sets the Instance key, secret, and session values.
func WithCredentials(key, secret, session string) InstanceOpt {
	return func(i *Instance) error {
		i.key = key
		i.secret = secret
		i.session = session
		return nil
	}
}

// WithRegion sets the AWS region for the Instance.
func WithRegion(region string) InstanceOpt {
	return func(i *Instance) error {
		i.region = region
		return nil
	}
}

// WithServices configures the Instance to only spin up the listed services.
func WithServices(services ...string) InstanceOpt {
	return func(i *Instance) error {
		i.services = services
		return nil
	}
}

// Wait for localstack to be ready.
func (i *Instance) Wait(max time.Duration) error {
	s3Client := s3.New(i.Config())
	start := time.Now()
	input := s3.ListBucketsInput{}
	for {
		if _, err := s3Client.ListBucketsRequest(&input).Send(context.TODO()); err != nil {
			if time.Now().Add(-1 * max).After(start) {
				return errors.New("localstack failed to respond in time")
			}

			time.Sleep(500 * time.Millisecond)
			continue
		}

		return nil
	}
}

// Close the Instance and clean up docker artifacts.
func (i *Instance) Close() error {
	return i.pool.Purge(i.resource)
}

func withDefaults(i *Instance) {
	if i.host == "" {
		i.host = "http://localhost"
	}

	if i.region == "" {
		i.region = "us-east-1"
	}

	if i.key == "" {
		i.key = "key"
	}

	if i.secret == "" {
		i.secret = "secret"
	}

	if i.session == "" {
		i.session = "session"
	}
}

func (i *Instance) serviceString() string {
	foundS3 := false
	for _, service := range i.services {
		if service == "s3" {
			foundS3 = true
		}
	}

	// s3 always has to be available in order for Wait() to work.
	if !foundS3 {
		i.services = append(i.services, "s3")
	}

	return fmt.Sprintf("SERVICES=%s", makeCsv(i.services))
}

// Config gives an AWS client configuration for talking to localstack.
func (i *Instance) Config() aws.Config {
	return aws.Config{
		Credentials: aws.NewStaticCredentialsProvider(i.key, i.secret, i.session),
		Region:      i.region,
		// DisableRestProtocolURICleaning: true,
		DisableEndpointHostPrefix: true,
		HTTPClient:                defaults.HTTPClient(),
		Handlers:                  defaults.Handlers(),
		Logger:                    defaults.Logger(),
		EndpointResolver:          aws.EndpointResolverFunc(i.resolver),
	}
}

func makeCsv(strings []string) string {
	buffer := bytes.NewBufferString("")
	for idx, str := range strings {
		_, _ = buffer.WriteString(str)
		if idx != len(strings) {
			_, _ = buffer.WriteString(",")
		}
	}

	return buffer.String()
}

func (i *Instance) makeResolver() serviceResolver {
	defaultResolver := endpoints.NewDefaultResolver()
	return func(service, region string) (aws.Endpoint, error) {
		switch service {
		case "apigateway":
			return aws.Endpoint{
				URL:           fmt.Sprintf("%s:%s", i.host, i.resource.GetPort("4567/tcp")),
				SigningRegion: "test-siging-region",
			}, nil
		case "kinesis":
			return aws.Endpoint{
				URL:           fmt.Sprintf("%s:%s", i.host, i.resource.GetPort("4568/tcp")),
				SigningRegion: "test-siging-region",
			}, nil
		case "dynamodb":
			return aws.Endpoint{
				URL:           fmt.Sprintf("%s:%s", i.host, i.resource.GetPort("4569/tcp")),
				SigningRegion: "test-siging-region",
			}, nil
		case "streams.dynamodb":
			return aws.Endpoint{
				URL:           fmt.Sprintf("%s:%s", i.host, i.resource.GetPort("4570/tcp")),
				SigningRegion: "test-siging-region",
			}, nil
		case "elasticsearch":
			return aws.Endpoint{
				URL:           fmt.Sprintf("%s:%s", i.host, i.resource.GetPort("4571/tcp")),
				SigningRegion: "test-siging-region",
			}, nil
		case "s3":
			return aws.Endpoint{
				URL:           fmt.Sprintf("%s:%s", i.host, i.resource.GetPort("4572/tcp")),
				SigningRegion: "test-siging-region",
			}, nil
		case "firehose":
			return aws.Endpoint{
				URL:           fmt.Sprintf("%s:%s", i.host, i.resource.GetPort("4573/tcp")),
				SigningRegion: "test-siging-region",
			}, nil
		case "lambda":
			return aws.Endpoint{
				URL:           fmt.Sprintf("%s:%s", i.host, i.resource.GetPort("4574/tcp")),
				SigningRegion: "test-siging-region",
			}, nil
		case "sns":
			return aws.Endpoint{
				URL:           fmt.Sprintf("%s:%s", i.host, i.resource.GetPort("4575/tcp")),
				SigningRegion: "test-siging-region",
			}, nil
		case "sqs":
			return aws.Endpoint{
				URL:           fmt.Sprintf("%s:%s", i.host, i.resource.GetPort("4576/tcp")),
				SigningRegion: "test-siging-region",
			}, nil
		case "redshift":
			return aws.Endpoint{
				URL:           fmt.Sprintf("%s:%s", i.host, i.resource.GetPort("4577/tcp")),
				SigningRegion: "test-siging-region",
			}, nil
		case "es":
			return aws.Endpoint{
				URL:           fmt.Sprintf("%s:%s", i.host, i.resource.GetPort("4578/tcp")),
				SigningRegion: "test-siging-region",
			}, nil
		case "ses":
			return aws.Endpoint{
				URL:           fmt.Sprintf("%s:%s", i.host, i.resource.GetPort("4579/tcp")),
				SigningRegion: "test-siging-region",
			}, nil
		case "route53":
			return aws.Endpoint{
				URL:           fmt.Sprintf("%s:%s", i.host, i.resource.GetPort("4580/tcp")),
				SigningRegion: "test-siging-region",
			}, nil
		case "cloudformation":
			return aws.Endpoint{
				URL:           fmt.Sprintf("%s:%s", i.host, i.resource.GetPort("4581/tcp")),
				SigningRegion: "test-siging-region",
			}, nil
		case "cloudwatch":
			return aws.Endpoint{
				URL:           fmt.Sprintf("%s:%s", i.host, i.resource.GetPort("4582/tcp")),
				SigningRegion: "test-siging-region",
			}, nil
		case "ssm":
			return aws.Endpoint{
				URL:           fmt.Sprintf("%s:%s", i.host, i.resource.GetPort("4583/tcp")),
				SigningRegion: "test-siging-region",
			}, nil
		case "secretsmanager":
			return aws.Endpoint{
				URL:           fmt.Sprintf("%s:%s", i.host, i.resource.GetPort("4584/tcp")),
				SigningRegion: "test-siging-region",
			}, nil
		// case "stepfunctions":
		// 	return aws.Endpoint{
		// 		URL:           fmt.Sprintf("%s:%s", i.host, i.resource.GetPort("4585/tcp")),
		// 		SigningRegion: "test-siging-region",
		// 	}, nil
		case "logs":
			return aws.Endpoint{
				URL:           fmt.Sprintf("%s:%s", i.host, i.resource.GetPort("4586/tcp")),
				SigningRegion: "test-siging-region",
			}, nil
		case "events":
			return aws.Endpoint{
				URL:           fmt.Sprintf("%s:%s", i.host, i.resource.GetPort("4587/tcp")),
				SigningRegion: "test-siging-region",
			}, nil
		case "sts":
			return aws.Endpoint{
				URL:           fmt.Sprintf("%s:%s", i.host, i.resource.GetPort("4592/tcp")),
				SigningRegion: "test-siging-region",
			}, nil
		case "iam":
			return aws.Endpoint{
				URL:           fmt.Sprintf("%s:%s", i.host, i.resource.GetPort("4593/tcp")),
				SigningRegion: "test-siging-region",
			}, nil
		case "ec2":
			return aws.Endpoint{
				URL:           fmt.Sprintf("%s:%s", i.host, i.resource.GetPort("4597/tcp")),
				SigningRegion: "test-siging-region",
			}, nil
		default:
			return defaultResolver.ResolveEndpoint(service, region)
		}
	}
}
