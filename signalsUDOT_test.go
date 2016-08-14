package immortal

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"testing"
	"time"
)

func TestHelperProcessSignalsUDOT(*testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	select {
	case <-c:
		os.Exit(1)
	case <-time.After(10 * time.Second):
		os.Exit(0)
	}
}

func TestSignalsUDOT(t *testing.T) {
	//log.SetOutput(ioutil.Discard)
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
		command: []string{filepath.Join(dirBase, base), "-test.run=TestHelperProcessSignalsUDOT"},
		Cwd:     parentDir,
		Pid: Pid{
			Parent: filepath.Join(parentDir, "parent.pid"),
			Child:  filepath.Join(parentDir, "child.pid"),
		},
	}
	d, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	d.Run()
	sup := new(Sup)
	go Supervise(sup, d)

	fmt.Printf("d.process.Pid XXXXXX= %+v\n", d.process.Pid)

	for !d.IsRunning() {
	}
	fmt.Printf("d.process.Pid = %+v\n", d.process.Pid)
	sup.HandleSignals("k", d)
	fmt.Printf("d.process.Pid = %+v\n", d.process.Pid)

	for {

	}

	// # ----------------------------------------------------------------------------

	// test "k", process should restart and get a new pid
	//d.Control.fifo <- Return{err: nil, msg: "k"}
	//sup.HandleSignals("k", d)
	//	expect(t, d.lock, uint32(1))
	//	expect(t, d.lock_defer, uint32(0))
	for {

	}
	// want it down
	//	fmt.Printf("d.process.Pid = %+v\n", d.process.Pid)

	//for d.process.Pid == 0 {
	//// wait for process to come  up
	//}
	//expect(t, true, sup.IsRunning(d.process.Pid))

	//// just to track using: watch -n 0.1 "pgrep -fl run=TestSignals | awk '{print $1}' | xargs -n1 pstree -p "
	//time.Sleep(500 * time.Millisecond)

	//// test "once", process should not restart after going down
	//d.Control.fifo <- Return{err: nil, msg: "o"}
	//d.Control.fifo <- Return{err: nil, msg: "k"}
	//// process shuld not start
	//for d.process.Pid != 0 {
	//// wait for process to restart and came up
	//}
	//expect(t, false, sup.IsRunning(d.process.Pid))

	//// test "u" bring up the service (new pid expected)
	//d.Control.fifo <- Return{err: nil, msg: "u"}
	//for d.process.Pid == 0 {
	//// wait for new pid
	//}
	//expect(t, true, sup.IsRunning(d.process.Pid))

	//time.Sleep(500 * time.Millisecond)

	//// test "down"
	//d.Control.fifo <- Return{err: nil, msg: "down"}
	//for d.process.Pid != 0 {
	//// wait for new pid
	//}
	//expect(t, false, sup.IsRunning(d.process.Pid))

	//// test "up" bring up the service
	//d.Control.fifo <- Return{err: nil, msg: "up"}
	//for d.process.Pid == 0 {
	//// wait for new pid
	//}
	//expect(t, true, sup.IsRunning(d.process.Pid))

	//// run only one command at a time
	//d.Run()

	//d.Control.fifo <- Return{err: nil, msg: "t"}
	//for d.process.Pid != 0 {
	//// wait for process to stop
	//}

	//expect(t, false, sup.IsRunning(d.process.Pid))
	//d.Control.fifo <- Return{err: nil, msg: "exit"}
}