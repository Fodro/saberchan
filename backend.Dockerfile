# syntax=docker/dockerfile:1

FROM golang:1.23.1
WORKDIR /app
COPY ./backend ./
RUN ls
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o ./main ./main.go
EXPOSE 8080
CMD ["/app/main"]