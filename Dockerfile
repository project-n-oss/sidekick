FROM golang:1.20-alpine

WORKDIR /go/src/github.com/project-n-oss/sidekick
COPY . .
RUN go mod tidy
RUN go generate ./...
RUN go build .

FROM golang:1.20-alpine

WORKDIR /usr/bin

COPY --from=0 /go/src/github.com/project-n-oss/sidekick/sidekick .
RUN ./sidekick --help > /dev/null

ENTRYPOINT ["/usr/bin/sidekick"]

