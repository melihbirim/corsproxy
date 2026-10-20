FROM golang:1.21-alpine AS b
WORKDIR /a
COPY go.mod main.go ./
RUN CGO_ENABLED=0 go build -o up .
FROM alpine:3.20
COPY --from=b /a/up /up
CMD ["/up"]
