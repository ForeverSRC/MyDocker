# MyDocker
DIY Docker

## build command
In order to build on macos, first install `x86_64-linux-musl-gcc`

```
arch -x86_64 brew install FiloSottile/musl-cross/musl-cross
```
Then, build with the following command:
```
CC=x86_64-linux-musl-gcc CGO_ENABLED=1 GOOS=linux GOARCH=amd64  go build -ldflags "-linkmode external -extldflags -static" -o my-docker main.go
```