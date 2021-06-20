FROM golang:1.16-alpine as builder

RUN apk update && \
    apk add --no-cache git

WORKDIR /parser

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o bin/parser *.go
EXPOSE 3000

FROM scratch
COPY --from=builder /parser/bin/parser .
ENTRYPOINT ["./parser"]
