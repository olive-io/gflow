/*
Copyright 2025 The gflow Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package clientgo

import (
	"context"
	"errors"
	"io"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func parseErr(err error) error {
	return err
}

func isUnavailable(err error) bool {
	return err == io.EOF ||
		errors.Is(err, context.Canceled) ||
		status.Code(err) == codes.Unavailable
}

func retryInterval(attempts int) time.Duration {
	if attempts <= 0 {
		return time.Millisecond * 100
	}
	if attempts > 5 {
		return time.Minute
	}
	return time.Duration(attempts*10) * time.Second
}
