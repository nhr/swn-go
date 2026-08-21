FROM registry.access.redhat.com/hi/go:1.25-builder AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o swn-go .

FROM registry.access.redhat.com/hi/static:latest
COPY --from=builder /build/swn-go /usr/local/bin/swn-go
EXPOSE 8080
CMD ["/usr/local/bin/swn-go"]
