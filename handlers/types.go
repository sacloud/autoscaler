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

package handlers

import (
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
	"github.com/sacloud/autoscaler/config"
	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/autoscaler/log"
)

// CustomHandler カスタムハンドラーが実装すべきインターフェース
type CustomHandler interface {
	HandlerMeta
	Listener
}

// HandlerMeta ハンドラーのメタ情報
//
// ビルトイン/カスタム問わず全てのハンドラーが実装すべきインターフェース
type HandlerMeta interface {
	// Name プレフィックスなしの短い名前を返す
	Name() string
	// Version バージョン情報を返す
	Version() string

	Logger
}

// Logger ログ出力のためのインターフェース
type Logger interface {
	GetLogger() *log.Logger
	SetLogger(logger *log.Logger)
}

// Listener gRPCサーバとしてリッスンするためのインターフェース
type Listener interface {
	ListenAddress() string
	TLSConfigPath() string
}

// Handler CoreからのHandleリクエストを処理するためのインターフェース
type Handler interface {
	Handle(*handler.HandleRequest, ResponseSender) error
}

// PreHandler CoreからのPreHandleリクエストを処理するためのインターフェース
type PreHandler interface {
	PreHandle(*handler.PreHandleRequest, ResponseSender) error
}

// PostHandler CoreからのPostHandleリクエストを処理するためのインターフェース
type PostHandler interface {
	PostHandle(*handler.PostHandleRequest, ResponseSender) error
}

// ResponseSender gRPCのサーバストリームのレスポンスをラップするインターフェース
type ResponseSender interface {
	Send(*handler.HandleResponse) error
}

// TLSConfig .
type TLSConfig struct {
	HandlerTLSConfig *config.TLSStruct `yaml:"tls_config"`
}

// LoadTLSConfig ファイルパスからTLSConfigを読み込む
func LoadTLSConfig(configPath string) (*TLSConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	conf := &TLSConfig{}
	if err := yaml.UnmarshalWithOptions(data, conf, yaml.Strict()); err != nil {
		return nil, err
	}

	if conf.HandlerTLSConfig != nil {
		conf.HandlerTLSConfig.SetDirectory(filepath.Dir(configPath))
	}
	return conf, nil
}
