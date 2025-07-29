# syntax=docker/dockerfile:1

FROM golang:1.23.1
WORKDIR /app
COPY ./backend ./
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o ./main ./main.go
EXPOSE 8888
CMD ["/app/main"]