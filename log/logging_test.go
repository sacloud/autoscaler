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

package log

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLogger_Error(t *testing.T) {
	buf := bytes.NewBufferString("")
	logger := NewLogger(&LoggerOption{
		Writer:    buf,
		JSON:      false,
		TimeStamp: false,
		Caller:    false,
		Level:     LevelError,
	})

	logger.Debug("msg", "this value will never be displayed") //nolint
	logger.Error("msg", "this value will be displayed")       //nolint

	require.Equal(t, `level=error msg="this value will be displayed"`+"\n", buf.String())
}

func TestLogger_With(t *testing.T) {
	buf := bytes.NewBufferString("")

	logger := NewLogger(&LoggerOption{
		Writer:    buf,
		JSON:      false,
		TimeStamp: false,
		Caller:    false,
		Level:     LevelError,
	})

	logger.Error("msg", "message without prefix") //nolint
	logger = logger.With("prefix", "foobar")
	logger.Error("msg", "message with prefix") //nolint

	expected := []string{
		`level=error msg="message without prefix"`,
		`level=error prefix=foobar msg="message with prefix"`,
		"",
	}

	require.Equal(t, expected, strings.Split(buf.String(), "\n"))
}
