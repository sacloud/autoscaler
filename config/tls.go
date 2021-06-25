// Copyright 2021 The sacloud Authors
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
	"path/filepath"

	"google.golang.org/grpc/credentials"

	"github.com/goccy/go-yaml"
)

type TLSStruct struct {
	TLSCertPath string `yaml:"cert_file"`
	TLSKeyPath  string `yaml:"key_file"`
	ClientAuth  string `yaml:"client_auth_type"`
	ClientCAs   string `yaml:"client_ca_file"`
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
	conf.SetDirectory(filepath.Dir(configPath))

	return conf.TLSConfig()
}

func (t *TLSStruct) SetDirectory(dir string) {
	t.TLSCertPath = joinDir(dir, t.TLSCertPath)
	t.TLSKeyPath = joinDir(dir, t.TLSKeyPath)
	t.ClientCAs = joinDir(dir, t.ClientCAs)
}

func (t *TLSStruct) TLSConfig() (*tls.Config, error) {
	if t.TLSCertPath == "" && t.TLSKeyPath == "" && t.ClientAuth == "" && t.ClientCAs == "" {
		return nil, ErrNoTLSConfig
	}

	if t.TLSCertPath == "" {
		return nil, errors.New("missing cert_file")
	}

	if t.TLSKeyPath == "" {
		return nil, errors.New("missing key_file")
	}

	cert, err := tls.LoadX509KeyPair(t.TLSCertPath, t.TLSKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load X509KeyPair: %s", err)
	}

	cfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	if t.ClientCAs != "" {
		clientCAPool := x509.NewCertPool()
		clientCAFile, err := os.ReadFile(t.ClientCAs)
		if err != nil {
			return nil, err
		}
		clientCAPool.AppendCertsFromPEM(clientCAFile)
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

	if t.ClientCAs != "" && cfg.ClientAuth == tls.NoClientCert {
		return nil, errors.New("Client CA's have been configured without a Client Auth Policy")
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

func joinDir(dir, path string) string {
	if path == "" || filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(dir, path)
}
