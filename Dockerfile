FROM golang:1.23-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY internal/ ./internal/
COPY cmd/ ./cmd/

# Build the application with CGO disabled for a static binary suitable for a FROM scratch image
# ldflags -s -w for an as small as possible binary
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /knx-mqtt ./cmd

# Build from scratch for added leanness and security
FROM scratch AS knx-mqtt-minimal

WORKDIR /app

COPY --from=build /knx-mqtt .

# Use non-root user for security
USER 1337:1337

CMD ["/app/knx-mqtt"]


FROM alpine:latest AS knx-mqtt

WORKDIR /app

COPY --from=build /knx-mqtt .

CMD ["/app/knx-mqtt"]
