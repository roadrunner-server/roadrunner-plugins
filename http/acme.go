package http

import (
	"context"
	"crypto"
	"crypto/tls"
	"time"

	"github.com/caddyserver/certmagic"
	"github.com/go-acme/lego/v4/registration"
	"go.uber.org/zap"
)

type challenge string

const (
	HTTP01    challenge = "http-01"
	TLSAlpn01 challenge = "tlsalpn-01"
)

// RRSslUser RR SSL user (needed to register the account)
type RRSslUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *RRSslUser) GetEmail() string {
	return u.Email
}
func (u RRSslUser) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u *RRSslUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

func ObtainCertificates(cacheDir, email, challengeType string, domains []string, useProduction bool, altHTTPPort, altTLSAlpnPort int) (*tls.Config, error) {
	z, _ := zap.NewProduction()
	cache := certmagic.NewCache(certmagic.CacheOptions{
		GetConfigForCert: func(c certmagic.Certificate) (*certmagic.Config, error) {
			return &certmagic.Config{
				RenewalWindowRatio: 0,
				MustStaple:         false,
				OCSP:               certmagic.OCSPConfig{},
				Storage:            &certmagic.FileStorage{Path: cacheDir},
				Logger:             z,
			}, nil
		},
		OCSPCheckInterval:  0,
		RenewCheckInterval: 0,
		Capacity:           0,
	})

	cfg := certmagic.New(cache, certmagic.Config{
		RenewalWindowRatio: 0,
		MustStaple:         false,
		OCSP:               certmagic.OCSPConfig{},
		Storage:            &certmagic.FileStorage{Path: cacheDir},
	})

	myAcme := certmagic.NewACMEManager(cfg, certmagic.ACMEManager{
		CA:                      certmagic.LetsEncryptProductionCA,
		TestCA:                  certmagic.LetsEncryptStagingCA,
		Email:                   email,
		Agreed:                  true,
		DisableHTTPChallenge:    false,
		DisableTLSALPNChallenge: false,
		ListenHost:              "127.0.0.1",
		AltHTTPPort:             altHTTPPort,
		AltTLSALPNPort:          altTLSAlpnPort,
		CertObtainTimeout:       time.Second * 240,
		PreferredChains:         certmagic.ChainPreference{},
	})

	if !useProduction {
		myAcme.CA = certmagic.LetsEncryptStagingCA
	}

	switch challenge(challengeType) {
	case HTTP01:
		myAcme.DisableTLSALPNChallenge = true
	case TLSAlpn01:
		myAcme.DisableHTTPChallenge = true
	default:
		// default - http
		myAcme.DisableTLSALPNChallenge = true
	}

	cfg.Issuers = append(cfg.Issuers, myAcme)

	err := cfg.ObtainCertAsync(context.Background(), email)
	if err != nil {
		return nil, err
	}

	err = cfg.ManageSync(context.Background(), domains)
	if err != nil {
		return nil, err
	}

	return cfg.TLSConfig(), nil
}
