/*
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

// glloq implements advisory locks on various backends.
//
// Checkout https://github.com/gilbsgilbs/glloq for a quick overview.
package glloq

import (
	"context"
	"errors"
	"log"
	"time"
)

// Options lock options.
type Options struct {
	// DSN is the connection string to the database.
	DSN string

	// Key is a lock key.
	Key string

	// Timeout defines how long to wait for the backend to be up.
	Timeout time.Duration

	// Params are beckend-specific options.
	Params map[string]string
}

// UseLocker waits for a locker to hold the lock then calls fn().
func UseLocker(locker Locker, opts *Options, fn func() error) error {
	if !locker.SupportsDSN(opts.DSN) {
		return ErrUnsupportedDSN
	}

	timeout := opts.Timeout
	if timeout == time.Duration(0) {
		timeout = 60 * time.Second
	}

	ctx, cancel := context.WithDeadline(
		context.Background(),
		time.Now().Add(timeout),
	)
	defer cancel()

	for {
		err := locker.Open(ctx, opts.DSN)
		if err == nil {
			defer locker.Close()
			break
		}

		log.Printf("glloq: couldn't open locker (%s). Retrying...\n", err)
		select {
		case <-ctx.Done():
			return ErrTimeout
		case <-time.After(1 * time.Second):
		}
	}

	err := locker.WithLock(ctx, opts, fn)
	if errors.Is(err, context.DeadlineExceeded) {
		return ErrTimeout
	}
	return err
}
