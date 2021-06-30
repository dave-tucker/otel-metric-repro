FROM golang:1.16
WORKDIR /src
COPY . .
CMD [ "go", "run", "main.go"]