package http

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"os"
	"path"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge/http01"
	"github.com/go-acme/lego/v4/challenge/tlsalpn01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner-plugins/v2/logger"
	"github.com/spiral/roadrunner/v2/utils"
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

func ObtainCertificates(log logger.Logger, cacheDir, keyName, certName, email, challengeType, challengePort, challengeIface string, domains []string, useProduction bool) error {
	const op = errors.Op("letsencrypt_obtain_certificates")
	// Create a user. New accounts need an email and private key to start.
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return errors.E(op, err)
	}

	_, err = os.Stat(cacheDir)
	if err != nil {
		if os.IsNotExist(err) {
			errMk := os.MkdirAll(cacheDir, 0600)
			if errMk != nil {
				return errors.E(op, errMk)
			}
		}
	}

	rrUser := &RRSslUser{
		Email: email,
		key:   privateKey,
	}

	config := lego.NewConfig(rrUser)

	if useProduction {
		config.CADirURL = lego.LEDirectoryProduction
	} else {
		config.CADirURL = lego.LEDirectoryStaging
	}

	config.Certificate.KeyType = certcrypto.RSA4096

	// A client facilitates communication with the CA server.
	client, err := lego.NewClient(config)
	if err != nil {
		return errors.E(op, err)
	}

	switch challenge(challengeType) {
	case HTTP01:
		err = client.Challenge.SetHTTP01Provider(http01.NewProviderServer(challengeIface, challengePort))
		if err != nil {
			return errors.E(op, err)
		}
	case TLSAlpn01:
		err = client.Challenge.SetTLSALPN01Provider(tlsalpn01.NewProviderServer(challengeIface, challengePort))
		if err != nil {
			return errors.E(op, err)
		}
	default:
		err = client.Challenge.SetHTTP01Provider(http01.NewProviderServer(challengeIface, challengePort))
		if err != nil {
			return errors.E(op, err)
		}
	}

	// New users will need to register
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return errors.E(op, err)
	}

	rrUser.Registration = reg

	request := certificate.ObtainRequest{
		Domains:    domains,
		Bundle:     true,
		MustStaple: false,
	}
	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		return errors.E(op, err)
	}

	err = os.WriteFile(path.Join(cacheDir, keyName), certificates.PrivateKey, 0600)
	if err != nil {
		return errors.E(op, err)
	}
	log.Debug("private key saved", "dir", path.Join(cacheDir, keyName))

	err = os.WriteFile(path.Join(cacheDir, certName), certificates.Certificate, 0644) //nolint:gosec
	if err != nil {
		return errors.E(op, err)
	}
	log.Debug("certificate saved", "dir", path.Join(cacheDir, certName))

	err = os.WriteFile(path.Join(cacheDir, "cert_url.txt"), utils.AsBytes(certificates.CertURL), 0644) //nolint:gosec
	if err != nil {
		return errors.E(op, err)
	}
	log.Debug("certificate url saved", "dir", path.Join(cacheDir, "cert_url.txt"))

	return nil
}
