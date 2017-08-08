package ipam

import (
	"context"
	"fmt"
	"net"
	"strings"

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

	newService := &Service{
		// Dependencies.
		logger:  config.Logger,
		storage: config.Storage,

		// Settings.
		network: *config.Network,
	}

	return newService, nil
}

type Service struct {
	// Dependencies.
	logger  micrologger.Logger
	storage microstorage.Storage

	// Settings.
	network net.IPNet
}

// key returns a storage key for a given network.
func key(network net.IPNet) string {
	return fmt.Sprintf(
		IPAMSubnetStorageKeyFormat,
		strings.Replace(network.String(), "/", "-", -1),
	)
}

// listSubnets retrieves the stored subnets from storage and returns them.
func (s *Service) listSubnets(ctx context.Context) ([]net.IPNet, error) {
	s.logger.Log("info", "listing subnets")

	existingSubnetStrings, err := s.storage.List(ctx, IPAMSubnetStorageKey)
	if err != nil && !microstorage.IsNotFound(err) {
		return nil, microerror.Mask(err)
	}

	existingSubnets := []net.IPNet{}
	for _, existingSubnetString := range existingSubnetStrings {
		// TODO: memory storage seems to return the end of the key with List,
		// not the value. This reverts `key` to provide a valid CIDR string,
		// and is technically safe for other storages.
		// tl;dr - fix memory storage returning the wrong thing.
		existingSubnetString = strings.Replace(existingSubnetString, "-", "/", -1)

		_, existingSubnet, err := net.ParseCIDR(existingSubnetString)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		existingSubnets = append(existingSubnets, *existingSubnet)
	}

	return existingSubnets, nil
}

// NewSubnet returns the next available subnet, of the configured size,
// from the configured network.
func (s *Service) NewSubnet(mask net.IPMask) (net.IPNet, error) {
	s.logger.Log("info", "creating new subnet")

	ctx := context.Background()

	existingSubnets, err := s.listSubnets(ctx)
	if err != nil {
		return net.IPNet{}, microerror.Mask(err)
	}

	s.logger.Log("info", "computing next subnet")
	subnet, err := Free(s.network, mask, existingSubnets)
	if err != nil {
		return net.IPNet{}, microerror.Mask(err)
	}

	s.logger.Log("info", "putting subnet", "subnet", subnet.String())
	if err := s.storage.Put(ctx, key(subnet), subnet.String()); err != nil {
		return net.IPNet{}, microerror.Mask(err)
	}

	return subnet, nil
}

// DeleteSubnet deletes the given subnet from IPAM storage,
// meaning it can be given out again.
func (s *Service) DeleteSubnet(subnet net.IPNet) error {
	s.logger.Log("info", "deleting subnet", "subnet", subnet.String())

	ctx := context.Background()

	if err := s.storage.Delete(ctx, key(subnet)); err != nil {
		return microerror.Mask(err)
	}

	return nil
}
