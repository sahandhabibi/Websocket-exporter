FROM golang:1.16-alpine 

WORKDIR /

COPY *.go ./
COPY go.* ./

RUN go get

RUN go build -o /websocket-exporter


FROM alpine:3.14 as production

ENV port 9143

COPY --from=0 websocket-exporter .

EXPOSE $port

CMD ./websocket-exporter
