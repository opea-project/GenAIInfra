/*
* Copyright (C) 2024 Intel Corporation
* SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"
)

const org = "opea.dev"

// generateCert generate a self-signed CA for given organization
// and sign certificate with the CA for given common name and dns names
// it resurns the CA in PEM format, certificate and private key in X509 format
func GenerateX509Cert(webhookName string, webhookNamespace string) (*tls.Certificate, *bytes.Buffer, error) {
	dnsNames := []string{
		webhookName,
		fmt.Sprintf("%s.%s", webhookName, webhookNamespace),
		fmt.Sprintf("%s.%s.svc", webhookName, webhookNamespace),
	}
	commonName := fmt.Sprintf("%s.%s.svc", webhookName, webhookNamespace)
	caPEM, certPEM, keyPEM, err := generateCert([]string{org}, dnsNames, commonName)
	if err != nil {
		return nil, nil, err
	}
	cert, err := tls.X509KeyPair(certPEM.Bytes(), keyPEM.Bytes())
	if err != nil {
		return nil, nil, err
	}

	return &cert, caPEM, nil
}

func generateCert(orgs, dnsNames []string, commonName string) (*bytes.Buffer, *bytes.Buffer, *bytes.Buffer, error) {
	// init CA config
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to generate serial number: %v", err)
	}
	ca := &x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               pkix.Name{Organization: orgs},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0), // expired in 1 year
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	// generate private key for CA
	caPrivateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, nil, err
	}

	// create the CA certificate
	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		return nil, nil, nil, err
	}

	// CA certificate with PEM encoded
	caPEM := new(bytes.Buffer)
	err = pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})
	if err != nil {
		return nil, nil, nil, err
	}

	// new certificate config
	serialNumber, err = rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to generate serial number: %v", err)
	}
	newCert := &x509.Certificate{
		DNSNames:     dnsNames,
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: orgs,
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(1, 0, 0), // expired in 1 year
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,
	}

	// generate new private key
	newPrivateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, nil, err
	}

	// sign the new certificate
	newCertBytes, err := x509.CreateCertificate(rand.Reader, newCert, ca, &newPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		return nil, nil, nil, err
	}

	// new certificate with PEM encoded
	newCertPEM := new(bytes.Buffer)
	err = pem.Encode(newCertPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: newCertBytes,
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to encode certificate: %v", err)
	}

	// new private key with PEM encoded
	newPrivateKeyPEM := new(bytes.Buffer)
	err = pem.Encode(newPrivateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(newPrivateKey),
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to encode key: %v", err)
	}

	return caPEM, newCertPEM, newPrivateKeyPEM, nil
}
