FROM golang:1.25 AS build-server

# Required for go-sqlite3
ENV CGO_ENABLED=1

WORKDIR /go/src/server
COPY ./server .

WORKDIR /go/src/server/cmd/middlewarr

RUN go mod download
RUN go vet -v
RUN go test -v

RUN GOOS=linux go build -o /go/bin/middlewarr

FROM gcr.io/distroless/base

COPY --from=build-server /go/bin/middlewarr /server/middlewarr
COPY --from=build-web /web/dist /web/dist

EXPOSE 80

VOLUME /data

CMD ["/server/middlewarr"]
