# MyDocker
DIY Docker

## build command
In order to build on macos, first install `x86_64-linux-musl-gcc`

```
arch -x86_64 brew install FiloSottile/musl-cross/musl-cross
```
Then, build with the following command:
```shell
CC=x86_64-linux-musl-gcc CGO_ENABLED=1 GOOS=linux GOARCH=amd64  go build -ldflags "-linkmode external -extldflags -static" -o my-docker main.go
```

## Remote Debug
In order to debug my-docker in linux os env:

1. Install go on ubuntu linux
2. Install delve
```shell
go install github.com/go-delve/delve/cmd/dlv@latest
```
3. Compile with dlv:
```shell
CC=x86_64-linux-musl-gcc CGO_ENABLED=1 GOOS=linux GOARCH=amd64  go build -ldflags "-linkmode external -extldflags -static" -gcflags "all=-N -l" -o my-docker-dlv main.go

```
4. Start dlv on ubuntu
```shell
dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient exec ./my-docker-dlv [command] -- [-option of my-docker]
```
5. Use Go Remote in IDEA to start remote debug