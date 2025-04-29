package io

import (
	"fmt"
	"io"

	"github.com/go-logr/logr"
)

func CloseOrLog(c io.Closer, name string, log logr.Logger) {
	if err := c.Close(); err != nil {
		log.Error(err, fmt.Sprintf("unable to close resource %q", name))
	}
}
