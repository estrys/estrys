FROM golang:1.19-alpine as builder
WORKDIR /go/src/app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 go build -o estrys ./cmd/estrys/
RUN CGO_ENABLED=0 go build -o worker ./cmd/worker/

FROM builder as dev
RUN apk add bash &&\
    go install github.com/cosmtrek/air@v1.40.4 &&\
    go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.15.2
ENTRYPOINT ["air"]

FROM dev as worker-dev
ENTRYPOINT ["air", "-c", ".air_worker.toml"]

FROM scratch
EXPOSE 8080
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/app/.env /
COPY --from=builder /go/src/app/estrys /
COPY --from=builder /go/src/app/worker /
ENTRYPOINT ["/estrys"]