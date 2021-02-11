FROM golang:1.15-alpine3.13 as builder
RUN apk --update --no-cache add ca-certificates
RUN addgroup -S loginsrv && adduser -S -g loginsrv loginsrv

WORKDIR /build

# Cache dependencies
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

# Copy code
COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-w -s" .

# ----------

FROM scratch
ENV LOGINSRV_HOST=0.0.0.0 LOGINSRV_PORT=8080
EXPOSE 8080
ENTRYPOINT ["./loginsrv"]

COPY --from=builder /etc/ssl/certs/ /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
USER loginsrv

COPY --from=builder /build/loginsrv /