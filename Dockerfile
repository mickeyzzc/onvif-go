# Multi-stage build for the onvif-go helper CLIs
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /bin/onvif-quick ./cmd/onvif-quick \
 && CGO_ENABLED=0 GOOS=linux go build -trimpath -o /bin/onvif-diagnostics ./cmd/onvif-diagnostics \
 && CGO_ENABLED=0 GOOS=linux go build -trimpath -o /bin/onvif-server ./cmd/onvif-server

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
RUN addgroup -g 1001 -S onvif && \
    adduser -u 1001 -S onvif -G onvif

WORKDIR /app
COPY --from=builder /bin/onvif-quick /usr/local/bin/
COPY --from=builder /bin/onvif-diagnostics /usr/local/bin/
COPY --from=builder /bin/onvif-server /usr/local/bin/

RUN chown -R onvif:onvif /app
USER onvif

CMD ["onvif-quick"]

LABEL maintainer="mickeyzzc"
LABEL description="onvif-go helper CLIs (quick client, diagnostics, virtual camera server)"
