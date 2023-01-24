package config

import (
	ace "go.bytebuilders.dev/client"
)

type Factory struct {
	Client func() (*ace.Client, error)
}
