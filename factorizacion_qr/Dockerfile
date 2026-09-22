FROM golang:1.27-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/factqr .

FROM alpine:3.21
COPY --from=builder /out/factqr /usr/local/bin/factqr
EXPOSE 3000
CMD ["factqr"]