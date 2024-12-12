FROM golang:1.23.4 AS builder

WORKDIR /bot

COPY go.mod go.sum ./
RUN go mod download 
COPY . .
COPY *.go ./

RUN CGO_ENABLED=0 go build -o /jonatas .

FROM alpine:3.21

RUN apk add --no-cache p7zip
COPY --from=builder /jonatas /jonatas

CMD ["/jonatas"]