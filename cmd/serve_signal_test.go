// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/sentrie-sh/sentrie/api"
)

const signalTestPack = `[schema]
version = 1

[pack]
name = "signal-pack"
version = "1.0.0"

[engine]
sentrie = ">=0.1.0 <2.0.0"
`

const signalTestPolicy = `namespace sh/sentrie/signal

shape User {
  role: string
}

policy access {
  fact user: User as user

  rule allow_admin = {
    yield user.role == "admin"
  }

  export decision of allow_admin
}
`

const signalTestBody = `{"facts":{"user":{"role":"admin"}}}`

// lockedBuffer is a bytes.Buffer safe for a process's stderr copier and the
// test to use at once.
type lockedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// startedServe is a running `sentrie serve` process.
type startedServe struct {
	cmd    *exec.Cmd
	addr   string
	logs   *lockedBuffer
	exited chan struct{} // closed once the process has been waited for
	err    error         // Wait's result, valid after exited is closed
}

var (
	signalBinOnce sync.Once
	signalBinDir  string
	signalBinPath string
	signalBinErr  error
)

// signalTestBinary builds the sentrie binary once per test process, into a
// directory TearDownSuite removes.
func (s *CmdTestSuite) signalTestBinary() string {
	s.T().Helper()
	signalBinOnce.Do(func() {
		repoRoot, err := filepath.Abs("..")
		if err != nil {
			signalBinErr = err
			return
		}
		signalBinDir, err = os.MkdirTemp("", "sentrie-signal-test-")
		if err != nil {
			signalBinErr = err
			return
		}
		signalBinPath = filepath.Join(signalBinDir, "sentrie")
		build := exec.Command("go", "build", "-o", signalBinPath, ".")
		build.Dir = repoRoot
		build.Env = append(os.Environ(), "GOWORK=off")
		if out, err := build.CombinedOutput(); err != nil {
			signalBinErr = fmt.Errorf("go build: %w\n%s", err, out)
		}
	})
	s.Require().NoError(signalBinErr)
	return signalBinPath
}

// TearDownSuite removes the binary signalTestBinary built and resets the
// build, so a repeated suite run (go test -count=N) builds again.
func (s *CmdTestSuite) TearDownSuite() {
	if signalBinDir != "" {
		_ = os.RemoveAll(signalBinDir)
	}
	signalBinOnce = sync.Once{}
	signalBinDir, signalBinPath, signalBinErr = "", "", nil
}

// startServe serves a minimal pack with the built binary on a free loopback
// port and returns once /health answers.
func (s *CmdTestSuite) startServe() *startedServe {
	s.T().Helper()
	if runtime.GOOS == "windows" {
		s.T().Skip("POSIX signals are not available on Windows")
	}

	bin := s.signalTestBinary()
	tmp := s.T().TempDir()

	packDir := filepath.Join(tmp, "pack")
	s.Require().NoError(os.Mkdir(packDir, 0o755))
	s.Require().NoError(os.WriteFile(filepath.Join(packDir, "sentrie.pack.toml"), []byte(signalTestPack), 0o600))
	s.Require().NoError(os.WriteFile(filepath.Join(packDir, "signal.sentrie"), []byte(signalTestPolicy), 0o600))

	// Reserve a free port, release it for the child to bind.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	s.Require().NoError(err)
	port := ln.Addr().(*net.TCPAddr).Port
	s.Require().NoError(ln.Close())

	logs := &lockedBuffer{}
	proc := exec.Command(bin, "serve",
		"--pack-location", packDir,
		"--http-port", strconv.Itoa(port),
		"--http-listen", "127.0.0.1")
	proc.Stdout = logs
	proc.Stderr = logs
	s.Require().NoError(proc.Start())

	sp := &startedServe{
		cmd:    proc,
		addr:   net.JoinHostPort("127.0.0.1", strconv.Itoa(port)),
		logs:   logs,
		exited: make(chan struct{}),
	}
	go func() {
		sp.err = proc.Wait()
		close(sp.exited)
	}()
	s.T().Cleanup(func() {
		select {
		case <-sp.exited:
		default:
			_ = proc.Process.Kill()
			<-sp.exited
		}
	})

	deadline := time.Now().Add(60 * time.Second)
	for {
		select {
		case <-sp.exited:
			s.FailNow("serve exited before it was ready", "err: %v\nlogs:\n%s", sp.err, logs.String())
		default:
		}
		resp, err := http.Get("http://" + sp.addr + "/health")
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return sp
			}
		}
		s.Require().True(time.Now().Before(deadline), "serve never became healthy\nlogs:\n%s", logs.String())
		time.Sleep(20 * time.Millisecond)
	}
}

// openPartialDecision opens a connection and sends a decision request with
// "Expect: 100-continue" and no body. Go writes the 100 Continue when the
// handler first reads the body, so reading it here proves the request is in
// flight and the handler is blocked waiting for the body. finish sends the
// body and returns the rest of the response.
func (s *CmdTestSuite) openPartialDecision(addr string) (finish func() string, abandon func()) {
	s.T().Helper()
	conn, err := net.Dial("tcp", addr)
	s.Require().NoError(err)
	s.T().Cleanup(func() { _ = conn.Close() })

	req := "POST /decision/sh/sentrie/signal/access/allow_admin HTTP/1.1\r\n" +
		"Host: " + addr + "\r\n" +
		"Content-Type: application/json\r\n" +
		"Content-Length: " + strconv.Itoa(len(signalTestBody)) + "\r\n" +
		"Expect: 100-continue\r\n" +
		"Connection: close\r\n\r\n"
	_, err = io.WriteString(conn, req)
	s.Require().NoError(err)

	s.Require().NoError(conn.SetReadDeadline(time.Now().Add(30 * time.Second)))
	br := bufio.NewReader(conn)
	status, err := br.ReadString('\n')
	s.Require().NoError(err)
	s.Require().Equal("HTTP/1.1 100 Continue\r\n", status)
	blank, err := br.ReadString('\n')
	s.Require().NoError(err)
	s.Require().Equal("\r\n", blank)

	finish = func() string {
		_, err := io.WriteString(conn, signalTestBody)
		s.Require().NoError(err)
		s.Require().NoError(conn.SetReadDeadline(time.Now().Add(30 * time.Second)))
		resp, err := io.ReadAll(br)
		s.Require().NoError(err)
		return string(resp)
	}
	return finish, func() { _ = conn.Close() }
}

// waitRefusing blocks until nothing accepts connections at addr: the server
// has begun shutting down.
func (s *CmdTestSuite) waitRefusing(sp *startedServe) {
	s.T().Helper()
	deadline := time.Now().Add(30 * time.Second)
	for {
		c, err := net.DialTimeout("tcp", sp.addr, time.Second)
		if err != nil {
			return
		}
		_ = c.Close()
		select {
		case <-sp.exited:
			s.FailNow("serve exited without refusing connections first", "err: %v", sp.err)
		default:
		}
		s.Require().True(time.Now().Before(deadline), "serve kept accepting connections after the signal\nlogs:\n%s", sp.logs.String())
		time.Sleep(5 * time.Millisecond)
	}
}

// TestServeFinishesAnInFlightRequestOnSIGTERM: the process stops accepting
// connections on SIGTERM, answers the request that was already in flight with
// its decision, and then exits 0 (#117).
func (s *CmdTestSuite) TestServeFinishesAnInFlightRequestOnSIGTERM() {
	sp := s.startServe()
	finish, _ := s.openPartialDecision(sp.addr)

	s.Require().NoError(sp.cmd.Process.Signal(syscall.SIGTERM))
	s.waitRefusing(sp)

	resp := finish()
	s.True(strings.HasPrefix(resp, "HTTP/1.1 200"), "response: %s\nlogs:\n%s", resp, sp.logs.String())
	s.Contains(resp, `"decisions"`)

	select {
	case <-sp.exited:
		s.NoError(sp.err, "serve should exit 0 after a clean drain\nlogs:\n%s", sp.logs.String())
	case <-time.After(30 * time.Second):
		s.FailNow("serve did not exit after draining", "logs:\n%s", sp.logs.String())
	}
}

// TestServeExitsOnSIGTERMWhenIdle: with nothing in flight the process exits 0
// promptly on SIGTERM, rather than waiting out the grace period.
func (s *CmdTestSuite) TestServeExitsOnSIGTERMWhenIdle() {
	sp := s.startServe()

	s.Require().NoError(sp.cmd.Process.Signal(syscall.SIGTERM))

	select {
	case <-sp.exited:
		s.NoError(sp.err, "logs:\n%s", sp.logs.String())
	case <-time.After(api.DefaultShutdownGracePeriod):
		s.FailNow("an idle serve outlived the grace period", "logs:\n%s", sp.logs.String())
	}
}

// TestServeSecondSignalTerminatesAtOnce: while a stuck request holds the drain
// open, a second SIGTERM kills the process instead of waiting for the grace
// period to run out.
func (s *CmdTestSuite) TestServeSecondSignalTerminatesAtOnce() {
	sp := s.startServe()
	_, abandon := s.openPartialDecision(sp.addr)
	defer abandon()

	s.Require().NoError(sp.cmd.Process.Signal(syscall.SIGTERM))
	s.waitRefusing(sp)

	// Resend until the process dies: the second signal only kills once main
	// has restored the default action, which happens right after the first.
	// The request never completes, so a surviving process would sit in the
	// drain for the whole grace period.
	deadline := time.After(api.DefaultShutdownGracePeriod / 2)
	tick := time.NewTicker(20 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-sp.exited:
			var ee *exec.ExitError
			s.Require().ErrorAs(sp.err, &ee, "logs:\n%s", sp.logs.String())
			ws, ok := ee.Sys().(syscall.WaitStatus)
			s.Require().True(ok)
			s.True(ws.Signaled(), "expected death by signal, got %v", ee)
			s.Equal(syscall.SIGTERM, ws.Signal())
			return
		case <-deadline:
			s.FailNow("a second SIGTERM did not terminate serve during the drain", "logs:\n%s", sp.logs.String())
		case <-tick.C:
			_ = sp.cmd.Process.Signal(syscall.SIGTERM)
		}
	}
}
