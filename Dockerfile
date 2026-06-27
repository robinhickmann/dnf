FROM golang:1.26.4-alpine AS build

WORKDIR /go/src/app

RUN apk add --no-cache make openssl

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN --mount=type=cache,target=/root/.cache \
    --mount=type=cache,target=/go/pkg/mod \
    make tls && \
    GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
    make build BIN=/go/bin/app

FROM gcr.io/distroless/static-debian13:nonroot

WORKDIR /usr/src/app

COPY --from=build /go/bin/app ./dnf
COPY --from=build --chown=nonroot \
    /go/src/app/cert.pem /go/src/app/key.pem ./

ENTRYPOINT ["./dnf"]
