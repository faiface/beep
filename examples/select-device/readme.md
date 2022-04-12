this is an example using the malgo as unerlying subsystem instead of oto

to build locally, follow these steps:
```bash
go mod init select-device
go mod edit -replace github.com/faiface/beep=/path/to/local/src/beep
go mod tidy
go build -tags malgo
```
