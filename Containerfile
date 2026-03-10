FROM fedora:43 AS builder
RUN dnf install -y golang
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o swn-go .

FROM fedora:43
COPY --from=builder /build/swn-go /usr/local/bin/swn-go
EXPOSE 8080
CMD ["/usr/local/bin/swn-go"]
