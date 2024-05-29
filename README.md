# slidingo
A sliding window algorithm-based request counter server implemented in go.

# Features

* Counts received requests within an interval, based on `window` argument (default 60 seconds), with millisecond accuracy
* Autosaves to a file using an interval based on `autosave` argument (default 30 seconds)

# How to run

```bash
go run ./main.go
```

# Configuration

```
  -addr string
        address the server listens (default ":3000")
  -autosave string
        autosave interval (default "30s")
  -save string
        file path the state will be saved to (default "/tmp/requestcounter.json")
  -window string
        window of the request counter (default "60s")
```
