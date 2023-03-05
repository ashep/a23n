FROM golang:1.20-alpine AS build

WORKDIR /build
RUN mkdir /build/out
COPY . /build
RUN go build -o out/a23n -ldflags "-s -w -X 'main.buildName=a23n' -X 'main.buildVer=0.0.1/$(uname -m)'" main.go

FROM alpine:latest
WORKDIR /app
COPY --from=build /build/out/a23n /app
ENTRYPOINT /app/a23n
