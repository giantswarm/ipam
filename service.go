package ipam

import (
	"context"
	"fmt"
	"net"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/microstorage"
	"github.com/giantswarm/microstorage/memory"
)

const (
	IPAMSubnetStorageKey       = "/ipam/subnet"
	IPAMSubnetStorageKeyFormat = "/ipam/subnet/%s"
)

// Config represents the configuration used to create a new ipam service.
type Config struct {
	// Dependencies.
	Logger  micrologger.Logger
	Storage microstorage.Storage

	// Settings.
	// Network is the network in which all returned subnets should exist.
	Network *net.IPNet
	// Mask is the mask for all returned subnets.
	Mask *net.IPMask
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

	// Settings.
	if config.Network == nil {
		return nil, microerror.Maskf(invalidConfigError, "network must not be empty")
	}
	if config.Mask == nil {
		return nil, microerror.Maskf(invalidConfigError, "mask must not be empty")
	}

	newService := &Service{
		// Dependencies.
		logger:  config.Logger,
		storage: config.Storage,

		// Settings.
		network: *config.Network,
		mask:    *config.Mask,
	}

	return newService, nil
}

type Service struct {
	// Dependencies.
	logger  micrologger.Logger
	storage microstorage.Storage

	// Settings.
	network net.IPNet
	mask    net.IPMask
}

// NewSubnet returns the next available subnet, of the configured size,
// from the configured network.
func (s *Service) NewSubnet() (net.IPNet, error) {
	s.logger.Log("info", "creating new subnet")

	ctx := context.Background()

	s.logger.Log("info", "fetching existing subnets")

	// Fetch existing subnets from storage.
	existingSubnetStrings, err := s.storage.List(ctx, IPAMSubnetStorageKey)
	if err != nil && !microstorage.IsNotFound(err) {
		return net.IPNet{}, microerror.Mask(err)
	}

	existingSubnets := []net.IPNet{}
	for _, existingSubnetString := range existingSubnetStrings {
		_, existingSubnet, err := net.ParseCIDR(existingSubnetString)
		if err != nil {
			return net.IPNet{}, microerror.Mask(err)
		}
		existingSubnets = append(existingSubnets, *existingSubnet)
	}

	s.logger.Log("info", "computing next subnet")

	// Compute the next subnet.
	subnet, err := Free(s.network, s.mask, existingSubnets)
	if err != nil {
		return net.IPNet{}, microerror.Mask(err)
	}

	s.logger.Log("info", "storing computed subnet")

	// Store the next subnet.
	if err := s.storage.Put(
		ctx,
		fmt.Sprintf(IPAMSubnetStorageKeyFormat, subnet.String()),
		subnet.String(),
	); err != nil {
		return net.IPNet{}, microerror.Mask(err)
	}

	return subnet, nil
}
