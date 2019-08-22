package localstack_test

import (
	"testing"

	"github.com/eriktate/go-localstack"
)

func Test_Localstack(t *testing.T) {
	instance, err := localstack.New()
	if err != nil {
		instance.Close()
		t.Fatal(err)
	}

	instance.Close()
}
