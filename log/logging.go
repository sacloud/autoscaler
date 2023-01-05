// Copyright 2021-2023 The sacloud/autoscaler Authors
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

package log

import (
	"io"
	"os"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

// Logger ログ出力
type Logger struct {
	internal log.Logger
	opt      *LoggerOption
}

// Level ログレベル
type Level string

const (
	LevelError = Level("error")
	LevelWarn  = Level("warn")
	LevelInfo  = Level("info")
	LevelDebug = Level("debug")
)

// LoggerOption ログ出力のオプション
type LoggerOption struct {
	Writer    io.Writer // 出力先(デフォルトはos.Stderr)
	JSON      bool      // JSON出力するか(falseの場合はlogfmt)
	TimeStamp bool      // タイムスタンプを含めるか
	Caller    bool      // caller(呼び出し箇所)を含めるか
	Level     Level     // 出力するログのレベル
}

// NewLogger 指定のオプションで新しいロガーを生成して返す
//
// optは省略(nil)でも可、デフォルトでは標準エラーに出力される
func NewLogger(opt *LoggerOption) *Logger {
	if opt == nil {
		opt = &LoggerOption{}
	}
	if opt.Writer == nil {
		opt.Writer = os.Stderr
	}

	logger := &Logger{}
	logger.initLogger(opt)
	return logger
}

func (l *Logger) initLogger(opt *LoggerOption) {
	w := log.NewSyncWriter(opt.Writer)
	var logger log.Logger
	switch {
	case opt.JSON:
		logger = log.NewJSONLogger(w)
	default:
		logger = log.NewLogfmtLogger(w)
	}

	if opt.TimeStamp {
		logger = log.With(logger, "timestamp", log.TimestampFormat(time.Now, time.RFC3339))
	}
	if opt.Caller {
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	switch opt.Level {
	case LevelError:
		logger = level.NewFilter(logger, level.AllowError())
	case LevelWarn:
		logger = level.NewFilter(logger, level.AllowWarn())
	case LevelInfo:
		logger = level.NewFilter(logger, level.AllowInfo())
	case LevelDebug:
		logger = level.NewFilter(logger, level.AllowDebug())
	}

	l.internal = logger
	l.opt = opt
}

func (l *Logger) Log(keyValues ...interface{}) error {
	return l.internal.Log(keyValues...)
}

// Reset 現在のLoggerOptionを元にロガーをリセット
//
//	Withxxxの影響を元に戻したい時などに利用する
func (l *Logger) Reset() {
	l.initLogger(l.opt)
}

// With 指定されたkey-valuesを持つコンテキストロガーを返す
//
// see: https://pkg.go.dev/github.com/go-kit/log
func (l *Logger) With(keyvals ...interface{}) *Logger {
	logger := NewLogger(l.opt)
	logger.internal = log.With(l.internal, keyvals...)
	return logger
}

// WithPrefix 指定されたkey-valuesを持つコンテキストロガーを返す
//
// see: https://pkg.go.dev/github.com/go-kit/log
func (l *Logger) WithPrefix(keyvals ...interface{}) *Logger {
	logger := NewLogger(l.opt)
	logger.internal = log.WithPrefix(l.internal, keyvals...)
	return logger
}

// WithSuffix 指定されたkey-valuesを持つコンテキストロガーを返す
//
// see: https://pkg.go.dev/github.com/go-kit/log
func (l *Logger) WithSuffix(keyvals ...interface{}) *Logger {
	logger := NewLogger(l.opt)
	logger.internal = log.WithSuffix(l.internal, keyvals...)
	return logger
}

func (l *Logger) Fatal(keyvals ...interface{}) {
	l.Error(keyvals...) //nolint
	os.Exit(1)
}

// Error レベルErrorでログ出力
func (l *Logger) Error(keyvals ...interface{}) error {
	return level.Error(l.internal).Log(keyvals...)
}

// Warn レベルWarnでログ出力
func (l *Logger) Warn(keyvals ...interface{}) error {
	return level.Warn(l.internal).Log(keyvals...)
}

// Info レベルInfoでログ出力
func (l *Logger) Info(keyvals ...interface{}) error {
	return level.Info(l.internal).Log(keyvals...)
}

// Debug レベルDebugでログ出力
func (l *Logger) Debug(keyvals ...interface{}) error {
	return level.Debug(l.internal).Log(keyvals...)
}
