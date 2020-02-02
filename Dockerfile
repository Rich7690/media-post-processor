# We'll choose the incredibly lightweight
# Go alpine image to work with
FROM golang:1-alpine3.11 AS builder

# We create an /app directory in which
# we'll put all of our project code
RUN mkdir /app
ADD . /app
WORKDIR /app
RUN apk update
RUN apk upgrade
RUN apk add gcc libstdc++ libc-dev
# We want to build our application's binary executable
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags "-linkmode external -extldflags -static" -a ./cmd/web/main.go

# the lightweight scratch image we'll
# run our application within
FROM alpine:3.11 AS production
# We have to copy the output from our
# builder stage to our production stage
COPY --from=builder /app/main .
COPY ./public ./public
# we can then kick off our newly compiled
# binary exectuable!!

RUN apk update
RUN apk upgrade
RUN apk add ffmpeg tini

VOLUME /config
EXPOSE 8080
ENV GIN_MODE=release
CMD ["/sbin/tini","--","./main"]
