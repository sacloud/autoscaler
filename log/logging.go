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
	"log/slog"
	"os"
)

// LoggerOption ログ出力のオプション
type LoggerOption struct {
	Writer    io.Writer  // 出力先(デフォルトはos.Stderr)
	JSON      bool       // JSON出力するか(falseの場合はlogfmt)
	TimeStamp bool       // タイムスタンプを含めるか
	Caller    bool       // caller(呼び出し箇所)を含めるか
	Level     slog.Level // 出力するログのレベル
}

func (opt *LoggerOption) handlerOption() *slog.HandlerOptions {
	replaceFunc := func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey {
			return slog.Attr{}
		}
		return a
	}
	if opt.TimeStamp {
		replaceFunc = nil
	}
	return &slog.HandlerOptions{
		AddSource:   opt.Caller,
		Level:       opt.Level,
		ReplaceAttr: replaceFunc,
	}
}

// NewLogger 指定のオプションで新しいロガーを生成して返す
//
// optは省略(nil)でも可、デフォルトでは標準エラーに出力される
func NewLogger(opt *LoggerOption) *slog.Logger {
	if opt == nil {
		opt = &LoggerOption{}
	}
	if opt.Writer == nil {
		opt.Writer = os.Stderr
	}

	switch {
	case opt.JSON:
		return slog.New(slog.NewJSONHandler(opt.Writer, opt.handlerOption()))
	default:
		return slog.New(slog.NewTextHandler(opt.Writer, opt.handlerOption()))
	}
}
