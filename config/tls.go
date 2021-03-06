// Copyright 2021-2022 The sacloud/autoscaler Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/goccy/go-yaml"
	"google.golang.org/grpc/credentials"
)

type TLSStruct struct {
	TLSCert    StringOrFilePath `yaml:"cert_file"`
	TLSKey     StringOrFilePath `yaml:"key_file"`
	ClientAuth string           `yaml:"client_auth_type"` // NoClientCert | RequestClientCert | RequireAnyClientCert | VerifyClientCertIfGiven | RequireAndVerifyClientCert
	ClientCAs  StringOrFilePath `yaml:"client_ca_file"`
	RootCAs    StringOrFilePath `yaml:"root_ca_file"`
}

var ErrNoTLSConfig = errors.New("TLS config is not present")

func LoadTLSConfig(configPath string) (*tls.Config, error) {
	if configPath == "" {
		return nil, ErrNoTLSConfig
	}
	reader, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return LoadTLSConfigFromReader(configPath, reader)
}

func LoadTLSConfigFromReader(configPath string, reader io.Reader) (*tls.Config, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	conf := &TLSStruct{}
	if err := yaml.UnmarshalWithOptions(data, conf, yaml.Strict()); err != nil {
		return nil, err
	}
	return conf.TLSConfig()
}

func (t *TLSStruct) TLSConfig() (*tls.Config, error) {
	if t.TLSCert.Empty() && t.TLSKey.Empty() && t.ClientAuth == "" && t.ClientCAs.Empty() && t.RootCAs.Empty() {
		return nil, ErrNoTLSConfig
	}

	cfg := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	if !t.TLSCert.Empty() && !t.TLSKey.Empty() {
		cert, err := tls.X509KeyPair(t.TLSCert.Bytes(), t.TLSKey.Bytes())
		if err != nil {
			return nil, fmt.Errorf("failed to load X509KeyPair: %s", err)
		}
		cfg.Certificates = []tls.Certificate{cert}
	}

	if !t.ClientCAs.Empty() {
		clientCAPool := x509.NewCertPool()
		clientCAPool.AppendCertsFromPEM(t.ClientCAs.Bytes())
		cfg.ClientCAs = clientCAPool
	}

	switch t.ClientAuth {
	case "RequestClientCert":
		cfg.ClientAuth = tls.RequestClientCert
	case "RequireClientCert":
		cfg.ClientAuth = tls.RequireAnyClientCert
	case "VerifyClientCertIfGiven":
		cfg.ClientAuth = tls.VerifyClientCertIfGiven
	case "RequireAndVerifyClientCert":
		cfg.ClientAuth = tls.RequireAndVerifyClientCert
	case "", "NoClientCert":
		cfg.ClientAuth = tls.NoClientCert
	default:
		return nil, errors.New("Invalid ClientAuth: " + t.ClientAuth)
	}

	if !t.ClientCAs.Empty() && cfg.ClientAuth == tls.NoClientCert {
		return nil, errors.New("Client CA's have been configured without a Client Auth Policy")
	}

	if !t.RootCAs.Empty() {
		rootCAPool := x509.NewCertPool()
		rootCAPool.AppendCertsFromPEM(t.RootCAs.Bytes())
		cfg.RootCAs = rootCAPool
	}

	return cfg, nil
}

func (t *TLSStruct) TransportCredentials() (credentials.TransportCredentials, error) {
	tlsConfig, err := t.TLSConfig()
	if err != nil {
		return nil, err
	}
	return credentials.NewTLS(tlsConfig), nil
}
