FROM golang:1.23.4 AS builder

WORKDIR /bot

COPY go.mod go.sum ./
RUN go mod download 
COPY . .
COPY *.go ./

RUN CGO_ENABLED=0 go build -o /manezinho_bot .

FROM alpine:3.21

COPY --from=builder /manezinho_bot /manezinho_bot

CMD ["/manezinho_bot"]