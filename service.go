package ipam

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/microstorage"
	"github.com/giantswarm/microstorage/memory"
)

// Config represents the configuration used to create a new ipam service.
type Config struct {
	// Dependencies.
	Logger  micrologger.Logger
	Storage microstorage.Storage
}

// DefaultConfig provides a default configuration to create a new ipam service
// by best effort.
func DefaultConfig() Config {
	var err error

	var newLogger micrologger.Logger
	{
		config := micrologger.DefaultConfig()
		newLogger, err = micrologger.New(config)
		if err != nil {
			panic(err)
		}
	}

	var newStorage microstorage.Storage
	{
		config := memory.DefaultConfig()
		newStorage, err = memory.New(config)
		if err != nil {
			panic(err)
		}
	}

	return Config{
		// Dependencies.
		Logger:  newLogger,
		Storage: newStorage,
	}
}

// New creates a new configured ipam service.
func New(config Config) (*Service, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "logger must not be empty")
	}
	if config.Storage == nil {
		return nil, microerror.Maskf(invalidConfigError, "storage must not be empty")
	}

	newService := &Service{
		// Dependencies.
		logger:  config.Logger,
		storage: config.Storage,
	}

	return newService, nil
}

type Service struct {
	// Dependencies.
	logger  micrologger.Logger
	storage microstorage.Storage
}
