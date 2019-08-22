package localstack

import (
	"bytes"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/defaults"
	"github.com/aws/aws-sdk-go-v2/aws/endpoints"
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

	resource, err := pool.Run("localstack/localstack", "", []string{fmt.Sprintf("SERVICES=%s", makeCsv(instance.services))})
	if err != nil {
		return nil, err
	}

	withDefaults(instance)
	instance.resolver = instance.makeResolver()
	instance.pool = pool
	instance.resource = resource

	return instance, nil
}

type InstanceOpt func(instance *Instance) error

func WithHost(host string) InstanceOpt {
	return func(i *Instance) error {
		i.host = host
		return nil
	}
}

func WithCredentials(key, secret, session string) InstanceOpt {
	return func(i *Instance) error {
		i.key = key
		i.secret = secret
		i.session = session
		return nil
	}
}

func WithRegion(region string) InstanceOpt {
	return func(i *Instance) error {
		i.region = region
		return nil
	}
}

func WithServices(services ...string) InstanceOpt {
	return func(i *Instance) error {
		i.services = services
		return nil
	}
}

func (i *Instance) Close() error {
	return i.pool.Purge(i.resource)
}

func withDefaults(i *Instance) {
	if i.host == "" {
		i.host = "localhost"
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

// Config gives an AWS client configuration for talking to localstack.
func (i *Instance) Config() aws.Config {
	return aws.Config{
		Credentials:                    aws.NewStaticCredentialsProvider(i.key, i.secret, i.session),
		Region:                         i.region,
		DisableRestProtocolURICleaning: true,
		DisableEndpointHostPrefix:      true,
		HTTPClient:                     defaults.HTTPClient(),
		Handlers:                       defaults.Handlers(),
		Logger:                         defaults.Logger(),
		EndpointResolver:               aws.EndpointResolverFunc(i.resolver),
	}
}

func makeCsv(strings []string) string {
	buffer := bytes.NewBufferString("")
	for idx, str := range strings {
		buffer.WriteString(str)
		if idx != len(strings) {
			buffer.WriteString(",")
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
		case "dynamo":
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
		default:
			return defaultResolver.ResolveEndpoint(service, region)
		}
	}
}
