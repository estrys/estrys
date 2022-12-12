FROM golang:1.19-alpine as builder
WORKDIR /go/src/app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 go build -o estrys ./cmd/estrys/
RUN CGO_ENABLED=0 go build -o worker ./cmd/worker/

FROM builder as dev
RUN make install-tools
ENTRYPOINT ["air"]

FROM dev as worker-dev
ENTRYPOINT ["air", "-c", ".air_worker.toml"]

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/app/.env /
COPY --from=builder /go/src/app/estrys /
COPY --from=builder /go/src/app/worker /
ENTRYPOINT ["/estrys"]