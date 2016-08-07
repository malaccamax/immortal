package immortal

import (
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestHelperProcessSignals(*testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	select {
	case <-c:
		os.Exit(0)
	case <-time.After(30 * time.Second):
		os.Exit(1)
	}
}

func TestSignals(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	base := filepath.Base(os.Args[0]) // "exec.test"
	dir := filepath.Dir(os.Args[0])   // "/tmp/go-buildNNNN/os/exec/_test"
	if dir == "." {
		t.Skip("skipping; running test at root somehow")
	}
	parentDir := filepath.Dir(dir) // "/tmp/go-buildNNNN/os/exec"
	dirBase := filepath.Base(dir)  // "_test"
	if dirBase == "." {
		t.Skipf("skipping; unexpected shallow dir of %q", dir)
	}
	cfg := &Config{
		Env:     map[string]string{"GO_WANT_HELPER_PROCESS": "1"},
		command: []string{filepath.Join(dirBase, base), "-test.run=TestHelperProcessSignals"},
		Cwd:     parentDir,
		Pid: Pid{
			Parent: filepath.Join(parentDir, "parent.pid"),
			Child:  filepath.Join(parentDir, "child.pid"),
		},
	}
	c := make(chan os.Signal)
	wait := make(chan struct{})
	d := &Daemon{
		Config: cfg,
		Control: &Control{
			fifo:  make(chan Return),
			quit:  make(chan struct{}),
			state: make(chan error),
		},
		Forker: &myFork{},
		Logger: &LogWriter{
			logger: NewLogger(cfg),
		},
		process: &catchSignals{&os.Process{}, c, wait},
	}
	d.Run()
	sup := new(Sup)
	go Supervise(sup, d)

	// wait for process to startup
	select {
	case <-wait:
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for pid")
	}
	var testSignals = []struct {
		signal   string
		expected os.Signal
	}{
		{"p", syscall.SIGSTOP},
		{"pause", syscall.SIGSTOP},
		{"s", syscall.SIGSTOP},
		{"stop", syscall.SIGSTOP},
		{"c", syscall.SIGCONT},
		{"cont", syscall.SIGCONT},
		{"h", syscall.SIGHUP},
		{"hup", syscall.SIGHUP},
		{"a", syscall.SIGALRM},
		{"alrm", syscall.SIGALRM},
		{"i", syscall.SIGINT},
		{"int", syscall.SIGINT},
		{"q", syscall.SIGQUIT},
		{"quit", syscall.SIGQUIT},
		{"1", syscall.SIGUSR1},
		{"usr1", syscall.SIGUSR1},
		{"2", syscall.SIGUSR2},
		{"2", syscall.SIGUSR2},
		{"t", syscall.SIGTERM},
		{"term", syscall.SIGTERM},
		{"in", syscall.SIGTTIN},
		{"TTIN", syscall.SIGTTIN},
		{"ou", syscall.SIGTTOU},
		{"out", syscall.SIGTTOU},
		{"TTOU", syscall.SIGTTOU},
		{"w", syscall.SIGWINCH},
		{"winch", syscall.SIGWINCH},
	}
	for _, s := range testSignals {
		d.Control.fifo <- Return{err: nil, msg: s.signal}
		waitSig(t, c, s.expected)
	}

	// test kill process will restart
	old_pid := d.process.GetPid()
	d.Control.fifo <- Return{err: nil, msg: "k"}
	expect(t, d.count, uint32(1), "in 114")
	expect(t, d.count_defer, uint32(0), "in 115")

	// wait for process to came up and then send signal "once"
	d.Control.fifo <- Return{err: nil, msg: "o"}
	for d.count_defer != 1 {
	}
	expect(t, d.count, uint32(1), "in 121")
	expect(t, old_pid, d.process.GetPid())

	// kill for test once (should re-sestart)
	d.Control.fifo <- Return{err: nil, msg: "k"}
	// process shuld not start and pids remains the same
	expect(t, d.count, uint32(1), "in 127")
	expect(t, d.count_defer, uint32(1), "in 128")
	expect(t, old_pid, d.process.GetPid())

	for sup.IsRunning(d.process.GetPid()) {
		// wait for process to die
	}

	var testSignalsError = []struct {
		signal   string
		expected os.Signal
	}{
		{"p", syscall.SIGILL},
		{"pause", syscall.SIGILL},
		{"s", syscall.SIGILL},
		{"stop", syscall.SIGILL},
		{"c", syscall.SIGILL},
		{"cont", syscall.SIGILL},
		{"h", syscall.SIGILL},
		{"hup", syscall.SIGILL},
		{"a", syscall.SIGILL},
		{"alrm", syscall.SIGILL},
		{"i", syscall.SIGILL},
		{"int", syscall.SIGILL},
		{"q", syscall.SIGILL},
		{"quit", syscall.SIGILL},
		{"1", syscall.SIGILL},
		{"usr1", syscall.SIGILL},
		{"2", syscall.SIGILL},
		{"2", syscall.SIGILL},
		{"t", syscall.SIGILL},
		{"term", syscall.SIGILL},
		{"in", syscall.SIGILL},
		{"TTIN", syscall.SIGILL},
		{"ou", syscall.SIGILL},
		{"out", syscall.SIGILL},
		{"TTOU", syscall.SIGILL},
		{"w", syscall.SIGILL},
		{"winch", syscall.SIGILL},
	}
	for _, s := range testSignalsError {
		d.Control.fifo <- Return{err: nil, msg: s.signal}
		waitSig(t, c, s.expected)
	}

	// test u
	// bring up the service (new pid expected)
	d.Control.fifo <- Return{err: nil, msg: "u"}
	select {
	case <-wait:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for pid")
	}
	for old_pid == d.process.GetPid() {
	}

	// test down
	d.Control.fifo <- Return{err: nil, msg: "down"}
	for sup.IsRunning(d.process.GetPid()) {
		// waiting for process to exit
	}

	// test up
	// bring up the service (new pid expected)
	d.Control.fifo <- Return{err: nil, msg: "up"}
	for !sup.IsRunning(d.process.GetPid()) {
	}
	d.Control.fifo <- Return{err: nil, msg: "once"}
	for d.count_defer != 1 {
	}
	expect(t, d.count, uint32(1), "in 197")
	expect(t, d.count_defer, uint32(1), "in 198")

	// save old pid
	old_pid = d.process.GetPid()

	// send kill (should not start)
	d.Control.fifo <- Return{err: nil, msg: "k"}
	for sup.IsRunning(d.process.GetPid()) {
	}
	expect(t, old_pid, d.process.GetPid(), "in 207")
	expect(t, false, sup.IsRunning(d.process.GetPid()), "in 208")

	// test up
	// bring up the service (new pid expected)
	d.Control.fifo <- Return{err: nil, msg: "up"}
	for !sup.IsRunning(d.process.GetPid()) {
	}
	old_pid = d.process.GetPid()

	// send kill (should re-start, and get new pid)
	d.Control.fifo <- Return{err: nil, msg: "k"}
	for sup.IsRunning(d.process.GetPid()) {
	}
	select {
	case <-wait:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for pid")
	}
	for old_pid == d.process.GetPid() {
	}

	// should be running
	expect(t, true, sup.IsRunning(d.process.GetPid()))

	// quit
	d.Control.fifo <- Return{err: nil, msg: "k"}
	d.Control.fifo <- Return{err: nil, msg: "exit"}
}

func waitSig(t *testing.T, c <-chan os.Signal, sig os.Signal) {
	select {
	case s := <-c:
		if s != sig {
			t.Fatalf("signal was %v, want %v", s, sig)
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("timeout waiting for %v", sig)
	}
}