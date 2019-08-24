# go-localstack
Lightweight library for integrating with localstack from go.

## Requirements
The only requirement, other than go of course, is docker. Localstack is most commonly run as a docker container, so you probably already have that!

## Getting Started
All you need to get started writing integration tests against localstack is a `localstack.Instance`. When creating a new instance, a localstack container gets spun up under the hood. Below is a minimal example.

```go
package test

import (
	"testing"
	"github.com/eriktate/go-localstack"
	"github.com/aws/aws-go-sdk-v2/service/s3"
)

func Test_AWS(t *testing.T) {
	instance, err := localstack.New()
	if err != nil {
		t.Fatal(err)
	}

	// wait for localstack container to spin up, but don't wait longer than 20 seconds
	if err := instance.Wait(20 * time.Second); err != nil {
		instance.Close()
		t.Fatal(err)
	}

	// you can generate an AWS config directly from the localstack Instance
	s3client := s3.New(instance.Config())

	// do your integration tets with the s3 client down here
}
```

## Tests
`go-localstack` uses go's built in testing capabilities, so you can run the full suite of tests with:

```bash
$ go test ./...
```

## Gotchas
`go-localstack` doesn't re-use containers and will leak them if they aren't cleaned up. If you don't want to end up with a million localstack container processes, you should call `instance.Close()` at every point your test might exit. Or just make sure it gets called in whatever clever cleanup magic you have.

## Contributing
Feel free to submit issues or PRs as you see fit. I can't promise I'll get to everything immediately, but I'll do my best to keep up with any issues or needs as they come up.

## TODO
- [x] Support v2 of AWS's go sdk
- [x] Spin up localstack container automagically
- [x] Provide a way to wait initially for the localstack container to spin up.
- [ ] Support v1 of AWS's go sdk.
- [ ] Write integration tests against each AWS service supported by localstack.
- [ ] More thorough unit test coverage.
- [ ] Write up some real world examples.
- [ ] Add repository badges.
- [ ] Detect old containers created by `go-localstack` and expose an option to clean those up.
