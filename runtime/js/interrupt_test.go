// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

import (
	"context"
	"errors"
	"time"

	"github.com/dop251/goja"
)

func (s *JSTestSuite) TestInterruptOnDoneInterruptsARunningVM() {
	ctx, cancel := context.WithTimeout(s.T().Context(), 20*time.Millisecond)
	defer cancel()
	vm := goja.New()

	stop := InterruptOnDone(ctx, vm)
	_, err := vm.RunString(`while (true) {}`)
	stop()

	var interrupted *goja.InterruptedError
	s.Require().ErrorAs(err, &interrupted)
	s.ErrorIs(interrupted.Value().(error), context.DeadlineExceeded)
}

// TestInterruptOnDoneDoesNotFireAfterStop is the #105 watchdog race: once stop
// returns, a later cancellation must not interrupt whatever runs next on the VM.
func (s *JSTestSuite) TestInterruptOnDoneDoesNotFireAfterStop() {
	vm := goja.New()
	for range 200 {
		ctx, cancel := context.WithCancel(s.T().Context())
		stop := InterruptOnDone(ctx, vm)
		_, err := vm.RunString(`1 + 1`)
		s.Require().NoError(err)
		stop()
		cancel()

		_, err = vm.RunString(`2 + 2`)
		s.Require().NoError(err, "a cancellation after stop interrupted the VM")
	}
}

// TestInterruptOnDoneStopWaitsForAnInFlightInterrupt races cancellation against
// stop. Under -race this proves stop synchronises with the interrupt goroutine.
func (s *JSTestSuite) TestInterruptOnDoneStopWaitsForAnInFlightInterrupt() {
	for range 200 {
		vm := goja.New()
		ctx, cancel := context.WithCancel(s.T().Context())
		stop := InterruptOnDone(ctx, vm)
		go cancel()
		_, err := vm.RunString(`for (let i = 0; i < 1000; i++) {}`)
		stop()
		if err != nil {
			var interrupted *goja.InterruptedError
			s.Require().True(errors.As(err, &interrupted), "unexpected error %v", err)
		}
		cancel()
	}
}

func (s *JSTestSuite) TestInterruptOnDoneWithNilContextIsANoOp() {
	//nolint:staticcheck // a nil context is the documented no-op input.
	stop := InterruptOnDone(nil, goja.New())
	stop()
}
