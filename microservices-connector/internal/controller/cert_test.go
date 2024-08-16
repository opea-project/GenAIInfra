/*
* Copyright (C) 2024 Intel Corporation
* SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"reflect"
	"testing"
)

func parsePEM(pemBytes []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, nil
	}
	return x509.ParseCertificate(block.Bytes)
}

func isRootCA(caPEM *bytes.Buffer) bool {
	cert, err := parsePEM(caPEM.Bytes())
	if err != nil {
		return false
	}

	// Check if the certificate is a CA and if the subject equals the issuer
	if cert.IsCA && cert.Subject.String() == cert.Issuer.String() {
		return true
	} else {
		return false
	}
}

func verifyCert(rootPEM, certPEM []byte) error {
	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM(rootPEM)
	if !ok {
		return fmt.Errorf("failed to parse root certificate")
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		return fmt.Errorf("failed to parse certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %v", err.Error())
	}

	opts := x509.VerifyOptions{
		Roots: roots,
	}

	if _, err := cert.Verify(opts); err != nil {
		return fmt.Errorf("failed to verify certificate: %v", err.Error())
	}

	return nil
}

func Test_generateCert(t *testing.T) {
	commonName := "test-service.default.svc"
	dnsNames := []string{commonName}
	type args struct {
		orgs       []string
		dnsNames   []string
		commonName string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "no error",
			args: args{
				orgs:       []string{org},
				dnsNames:   dnsNames,
				commonName: commonName,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ca, cert, _, err := generateCert(tt.args.orgs, tt.args.dnsNames, tt.args.commonName)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateCert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !isRootCA(ca) {
				t.Errorf("CA is not a root CA: %v", ca)
				return
			}
			tlsCert, err := parsePEM(cert.Bytes())
			if err != nil {
				t.Errorf("generateCert() error = %v", err)
				return
			}
			if !reflect.DeepEqual(tlsCert.DNSNames, dnsNames) {
				t.Errorf("generateCert() = %v, want %v", tlsCert.DNSNames, dnsNames)
				return
			}
			if tlsCert.Subject.CommonName != commonName {
				t.Errorf("generateCert() = %v, want %v", tlsCert.Subject.CommonName, commonName)
				return
			}
			if err := verifyCert(ca.Bytes(), cert.Bytes()); err != nil {
				t.Errorf("generateCert() error = %v", err)
			}
		})
	}
}

func TestGenerateX509Cert(t *testing.T) {
	type args struct {
		webhookName      string
		webhookNamespace string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "no error",
			args: args{
				webhookName:      "test-service",
				webhookNamespace: "default",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := GenerateX509Cert(tt.args.webhookName, tt.args.webhookNamespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateX509Cert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

		})
	}
}

func Test_composeDNSNames(t *testing.T) {
	whName := "test-service"
	whNS := "default"
	type args struct {
		webhookName      string
		webhookNamespace string
	}
	tests := []struct {
		name    string
		args    args
		dns     []string
		cn      string
		wantErr bool
	}{
		{
			name: "no error",
			args: args{
				webhookName:      whName,
				webhookNamespace: whNS,
			},
			dns: []string{
				whName,
				fmt.Sprintf("%s.%s", whName, whNS),
				fmt.Sprintf("%s.%s.svc", whName, whNS),
			},
			cn:      fmt.Sprintf("%s.%s.svc", whName, whNS),
			wantErr: false,
		},
		{
			name: "namespace is empty",
			args: args{
				webhookName:      whName,
				webhookNamespace: "",
			},
			dns:     nil,
			cn:      "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dns, cn, err := composeDNSNames(tt.args.webhookName, tt.args.webhookNamespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("composeDNSNames() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(dns, tt.dns) {
				t.Errorf("composeDNSNames() got = %v, want %v", dns, tt.dns)
				return
			}
			if cn != tt.cn {
				t.Errorf("composeDNSNames() got1 = %v, want %v", cn, tt.cn)
			}
		})
	}
}
