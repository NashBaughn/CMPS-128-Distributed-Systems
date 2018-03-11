FROM golang:latest

RUN mkdir /app
ENV GOPATH=/app/

RUN echo $GOPATH

ADD . /app/

WORKDIR /app

EXPOSE 8080

RUN go get github.com/gorilla/mux

RUN echo 'Dockerfile executing'

RUN ls /app/
CMD ["go", "run", "main.go"]
