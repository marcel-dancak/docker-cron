FROM golang:alpine as go-builder
RUN apk add --no-cache git
ENV GOPATH=
COPY ./go.mod /go/
COPY ./go.sum /go/
RUN go mod download
COPY ./cmd /go/cmd
COPY *.go /go/
RUN go build -ldflags="-s -w" -o /go/bin/dcron cmd/dcron/main.go


FROM node:12-alpine AS webapp
WORKDIR /webapp/
COPY web-app/package*.json ./
RUN npm install
COPY web-app/ .
RUN npm run build


FROM alpine:latest
ENV DCRON_WEB_ROOT /var/www
ENV DCRON_WEB_PORT 8090
COPY --from=go-builder /go/bin/dcron /usr/local/bin/dcron
COPY --from=webapp /webapp/dist/ $DCRON_WEB_ROOT
CMD ["dcron"]
