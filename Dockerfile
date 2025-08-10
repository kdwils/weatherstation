FROM golang:1.24.5 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o weatherstation

FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=builder /app/weatherstation /app/
COPY --from=builder /app/static /app/static

EXPOSE 8080

ENTRYPOINT ["/app/weatherstation"]

CMD ["serve"]