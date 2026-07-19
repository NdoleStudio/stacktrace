// Copyright 2016 Palantir Technologies
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package stacktrace_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/NdoleStudio/stacktrace"
)

type wrappedTestError struct {
	err error
}

func (e *wrappedTestError) Error() string {
	return "wrapped test error"
}

func (e *wrappedTestError) Unwrap() error {
	return e.err
}

func TestUnwrap(t *testing.T) {
	root := errors.New("root")
	typed := &wrappedTestError{err: root}
	inner := stacktrace.Propagate(typed, "inner")
	outer := stacktrace.Propagate(inner, "outer")

	assert.Same(t, inner, errors.Unwrap(outer))
	assert.True(t, errors.Is(outer, root))

	var target *wrappedTestError
	assert.True(t, errors.As(outer, &target))
	assert.Same(t, typed, target)
}

func TestMessage(t *testing.T) {
	useFixturePaths(t)

	err := startDoing()
	err = PublicObj{}.DoPublic(err)
	err = PublicObj{}.doPrivate(err)
	err = privateObj{}.DoPublic(err)
	err = privateObj{}.doPrivate(err)
	err = (&ptrObj{}).doPtr(err)
	err = doClosure(err)

	expected := strings.Join([]string{
		"so closed",
		fmt.Sprintf(" --- at %s:51 (doClosure.func1) ---", fixturePath("functions_for_test.go")),
		"Caused by: pointedly",
		fmt.Sprintf(" --- at %s:46 (ptrObj.doPtr) ---", fixturePath("functions_for_test.go")),
		fmt.Sprintf(" --- at %s:42 (privateObj.doPrivate) ---", fixturePath("functions_for_test.go")),
		fmt.Sprintf(" --- at %s:38 (privateObj.DoPublic) ---", fixturePath("functions_for_test.go")),
		fmt.Sprintf(" --- at %s:34 (PublicObj.doPrivate) ---", fixturePath("functions_for_test.go")),
		fmt.Sprintf(" --- at %s:30 (PublicObj.DoPublic) ---", fixturePath("functions_for_test.go")),
		"Caused by: failed to start doing",
		fmt.Sprintf(" --- at %s:26 (startDoing) ---", fixturePath("functions_for_test.go")),
	}, "\n")
	stacktrace.DefaultFormat = stacktrace.FormatFull
	assert.Equal(t, expected, err.Error())
	assert.Equal(t, expected, fmt.Sprint(err))
}

func TestGetCode(t *testing.T) {
	for _, test := range []struct {
		originalError error
		originalCode  stacktrace.ErrorCode
	}{
		{
			originalError: errors.New("err"),
			originalCode:  stacktrace.NoCode,
		},
		{
			originalError: stacktrace.NewError("err"),
			originalCode:  stacktrace.NoCode,
		},
		{
			originalError: stacktrace.NewErrorWithCode(EcodeInvalidVillain, "err"),
			originalCode:  EcodeInvalidVillain,
		},
		{
			originalError: stacktrace.NewMessageWithCode(EcodeNoSuchPseudo, "err"),
			originalCode:  EcodeNoSuchPseudo,
		},
	} {
		err := test.originalError
		assert.Equal(t, test.originalCode, stacktrace.GetCode(err))

		err = stacktrace.Propagate(err, "")
		assert.Equal(t, test.originalCode, stacktrace.GetCode(err))

		err = stacktrace.PropagateWithCode(err, EcodeNotFastEnough, "")
		assert.Equal(t, EcodeNotFastEnough, stacktrace.GetCode(err))

		err = stacktrace.PropagateWithCode(err, EcodeTimeIsIllusion, "")
		assert.Equal(t, EcodeTimeIsIllusion, stacktrace.GetCode(err))
	}
}

func TestPropagateNil(t *testing.T) {
	var err error

	err = stacktrace.Propagate(err, "")
	assert.Nil(t, err)

	err = stacktrace.PropagateWithCode(err, EcodeNotImplemented, "")
	assert.Nil(t, err)

	assert.Equal(t, stacktrace.NoCode, stacktrace.GetCode(err))
}

func TestExitCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{
			name:     "no code",
			err:      stacktrace.NewError("err"),
			expected: 1,
		},
		{
			name:     "explicit code",
			err:      stacktrace.NewErrorWithCode(EcodeNotImplemented, "err"),
			expected: int(EcodeNotImplemented),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			exitCoder, ok := test.err.(interface{ ExitCode() int })
			assert.True(t, ok)
			assert.Equal(t, test.expected, exitCoder.ExitCode())
		})
	}
}

func TestFormattingHelpers(t *testing.T) {
	root := errors.New("root")

	tests := []struct {
		name     string
		err      error
		expected string
		code     stacktrace.ErrorCode
	}{
		{
			name:     "new error",
			err:      stacktrace.NewErrorf("new %d", 7),
			expected: "new 7",
			code:     stacktrace.NoCode,
		},
		{
			name:     "propagate",
			err:      stacktrace.Propagatef(root, "propagate %d", 7),
			expected: "propagate 7: root",
			code:     stacktrace.NoCode,
		},
		{
			name:     "new error with code",
			err:      stacktrace.NewErrorWithCodef(EcodeInvalidVillain, "coded %d", 7),
			expected: "coded 7",
			code:     EcodeInvalidVillain,
		},
		{
			name:     "propagate with code",
			err:      stacktrace.PropagateWithCodef(root, EcodeNotFastEnough, "coded propagate %d", 7),
			expected: "coded propagate 7: root",
			code:     EcodeNotFastEnough,
		},
		{
			name:     "new message with code",
			err:      stacktrace.NewMessageWithCodef(EcodeInvalidVillain, "message %d", 7),
			expected: "message 7",
			code:     EcodeInvalidVillain,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, fmt.Sprintf("%#s", test.err))
			assert.Equal(t, test.code, stacktrace.GetCode(test.err))
		})
	}
}

func TestFormattingHelpersPreserveCode(t *testing.T) {
	err := stacktrace.NewErrorWithCodef(EcodeInvalidVillain, "inner %d", 7)
	err = stacktrace.Propagatef(err, "outer %d", 8)

	assert.Equal(t, EcodeInvalidVillain, stacktrace.GetCode(err))
	assert.Equal(t, "outer 8: inner 7", fmt.Sprintf("%#s", err))
}

func TestFormattingPropagationNil(t *testing.T) {
	assert.Nil(t, stacktrace.Propagatef(nil, "propagate %d", 7))
	assert.Nil(t, stacktrace.PropagateWithCodef(nil, EcodeInvalidVillain, "propagate %d", 7))
}

func TestFormattingHelpersCaptureCaller(t *testing.T) {
	useFixturePaths(t)
	root := errors.New("root")

	tests := []struct {
		name     string
		err      error
		function string
	}{
		{name: "new error", err: newErrorfAtCallSite(), function: "newErrorfAtCallSite"},
		{name: "propagate", err: propagatefAtCallSite(root), function: "propagatefAtCallSite"},
		{name: "new error with code", err: newErrorWithCodefAtCallSite(), function: "newErrorWithCodefAtCallSite"},
		{name: "propagate with code", err: propagateWithCodefAtCallSite(root), function: "propagateWithCodefAtCallSite"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			trace := fmt.Sprintf("%+s", test.err)
			assert.Contains(t, trace, fixturePath("functions_for_test.go"))
			assert.Contains(t, trace, "("+test.function+")")
		})
	}
}
