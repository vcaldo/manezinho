FROM golang:1.23.4 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download 
COPY . .
COPY *.go ./

RUN CGO_ENABLED=0 go build -o /garibaldi .

FROM alpine:3.21

RUN apk add --no-cache p7zip
COPY --from=builder /garibaldi /garibaldi

CMD ["/garibaldi"]