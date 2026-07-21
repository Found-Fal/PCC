package ai

import (
	"context"
	"fmt"
	"sync"
)

type Router struct {
	mu sync.RWMutex

	providers map[string]Provider

	defaultProvider string

	fallback []string
}

func NewRouter(
	defaultProvider string,
) *Router {

	return &Router{
		providers:       make(map[string]Provider),
		defaultProvider: defaultProvider,
		fallback:        []string{},
	}
}

func (r *Router) Register(
	provider Provider,
) {

	r.mu.Lock()
	defer r.mu.Unlock()

	r.providers[provider.Name()] = provider
}

func (r *Router) SetFallback(
	providers []string,
) {

	r.mu.Lock()
	defer r.mu.Unlock()

	r.fallback = providers
}

func (r *Router) GetProvider(
	name string,
) (Provider, error) {

	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, exists := r.providers[name]

	if !exists {
		return nil, fmt.Errorf(
			"provider AI '%s' tidak ditemukan",
			name,
		)
	}

	return provider, nil
}

func (r *Router) Chat(
	ctx context.Context,
	messages []Message,
) (string, error) {

	r.mu.RLock()

	defaultProvider := r.defaultProvider
	fallback := append(
		[]string{},
		r.fallback...,
	)

	r.mu.RUnlock()

	providers := make(
		[]string,
		0,
		len(fallback)+1,
	)

	// Provider utama.
	if defaultProvider != "" {
		providers = append(
			providers,
			defaultProvider,
		)
	}

	// Provider fallback.
	for _, name := range fallback {

		if name == defaultProvider {
			continue
		}

		providers = append(
			providers,
			name,
		)
	}

	var lastErr error

	for _, providerName := range providers {

		provider, err := r.GetProvider(
			providerName,
		)

		if err != nil {
			lastErr = err
			continue
		}

		response, err := provider.Chat(
			ctx,
			messages,
		)

		if err == nil {
			return response, nil
		}

		lastErr = fmt.Errorf(
			"provider %s gagal: %w",
			providerName,
			err,
		)
	}

	if lastErr != nil {
		return "", fmt.Errorf(
			"semua provider AI gagal: %w",
			lastErr,
		)
	}

	return "", fmt.Errorf(
		"tidak ada provider AI yang tersedia",
	)
}
