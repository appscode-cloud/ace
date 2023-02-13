package config

import (
	"os"

	ace "go.bytebuilders.dev/client"
)

type Factory struct {
	Client    func() (*ace.Client, error)
	Canceller func() chan os.Signal
}
