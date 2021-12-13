FROM golang:alpine AS builder

# 设置环境变量
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOPROXY="https://goproxy.io,https://gocenter.io,direct" \
    GOSUMDB=off

# 移动到工作目录
WORKDIR $GOPATH/src
# 下载依赖信息
COPY ./go.mod .
COPY ./go.sum .
RUN go mod download

COPY . .

# 编译
RUN go build -o my-docker .

FROM ubuntu:14.04
RUN apt-get update && apt-get install stress
COPY --from=builder go/src/my-docker .
CMD ["/bin/bash"]