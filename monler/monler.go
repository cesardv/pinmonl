package monler

import (
	"errors"

	"github.com/pinmonl/pinmonl/monler/provider"
	"github.com/pinmonl/pinmonl/pkgs/pkguri"
)

var (
	providers = make(map[string]provider.Provider)
)

// Errors.
var (
	ErrUnknownProvider = errors.New("monler: unknown provider")
)

func Register(name string, provider provider.Provider) {
	providers[name] = provider
}

func Providers() []string {
	list := make([]string, 0)
	for name := range providers {
		list = append(list, name)
	}
	return list
}

func Open(providerName, url string) (provider.Repo, error) {
	pvd, ok := providers[providerName]
	if !ok {
		return nil, ErrUnknownProvider
	}
	return pvd.Open(url)
}

func Parse(uri string) (provider.Repo, error) {
	pu, err := pkguri.Parse(uri)
	if err != nil {
		return nil, err
	}
	pvd, ok := providers[pu.Provider]
	if !ok {
		return nil, ErrUnknownProvider
	}
	return pvd.Parse(uri)
}

func Ping(providerName, url string) error {
	pvd, ok := providers[providerName]
	if !ok {
		return ErrUnknownProvider
	}
	return pvd.Ping(url)
}

func Guess(url string) ([]provider.Repo, error) {
	repos := make([]provider.Repo, 0)
	for _, pvd := range providers {
		if err := pvd.Ping(url); err != nil {
			continue
		}
		repo, err := pvd.Open(url)
		if err != nil {
			return nil, err
		}
		repos = append(repos, repo)
	}
	return repos, nil
}
