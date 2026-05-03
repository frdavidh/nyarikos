FROM golang:1.25-alpine AS builder

COPY go.mod go.sum ./
RUN go mod download

COPY . .


RUN mkdir -p /app/bin && \
    for cmd in cmd/*/; do \
        if [ -d "$cmd" ]; then \
			      binary=$(basename $cmd); \
			      echo "Building $binary"; \
			      CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/bin/$binary ./$cmd; \
		    fi \
	done


FROM alpine:latest


RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/bin/* ./


CMD ["./api"]