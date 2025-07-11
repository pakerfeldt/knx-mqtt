FROM golang:1.23-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY internal/ ./internal/
COPY cmd/ ./cmd/

# CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -o /knx-mqtt ./cmd

FROM alpine:latest

WORKDIR /app

COPY --from=build /knx-mqtt .

CMD ["/app/knx-mqtt"]
