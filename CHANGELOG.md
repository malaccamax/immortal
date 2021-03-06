## 0.20.0

* Added the `require_cmd` option that prevents starting a service based on the output of command (exit 0), thanks @luetge

## 0.19.0

* Added option `-n` no-daemon mode, stays in the foreground [#40](https://github.com/immortal/immortal/pull/40), thanks @loafoe
* Use service name derived from config when using `immortal -c service.yml` [#39](https://github.com/immortal/immortal/pull/30), thanks @loafoe

## 0.18.0

* Added option `retries`, `-r` to specify the maximum number of tries before exiting the program
* Environment `IMMORTAL_EXIT` used to exit when running immortal with a config
file, helps to avoid a race condition (start/stop) when using immortaldir
* `immortalctl` prints now process that are about to start with a defined `wait` value
* Renamed option `-s` to `-w` to be more consistent with the config file option `wait`
* Signals are only sent to process then this is up and running

## 0.17.0

* Cleaned tests (Dockerfile for linux)
* Created a Supervisor struct to handle better the supervise logic
* Give priority to environment `$HOME` instead of HomeDir from `user.Current()`
* Improved lint
* Print cmd name (not just PID) in the log when the process terminates [#29](https://github.com/immortal/immortal/pull/29), thanks @marko-turk
* Removed info.go (signal.Notify) from supervise.go
* Replaced lock/map with sync.Map in scandir.go
* Updated HandleSignal to use `GetParam` from violetear
