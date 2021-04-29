FROM golang:alpine
WORKDIR /app
COPY ./go.mod .
RUN go mod download
COPY . .
CMD ["cd", "src"]
CMD ["go", "run", "server.go"]