# We'll choose the incredibly lightweight
# Go alpine image to work with
FROM golang:1.14-alpine3.11 AS builder

# We create an /app directory in which
# we'll put all of our project code
RUN mkdir /app
ADD . /app
WORKDIR /app
RUN apk update
RUN apk add gcc libstdc++ libc-dev
# We want to build our application's binary executable
RUN go build -a ./cmd/web/main.go

# the lightweight scratch image we'll
# run our application within
FROM alpine:3.11 AS production
# We have to copy the output from our
# builder stage to our production stage
COPY --from=builder /app/main .
COPY ./public ./public

RUN apk update
RUN apk upgrade
RUN apk add ffmpeg
RUN apk add tzdata
RUN cp /usr/share/zoneinfo/America/Los_Angeles /etc/localtime
RUN echo "America/Los_Angeles" >  /etc/timezone
RUN apk del tzdata

VOLUME /config
EXPOSE 8080
HEALTHCHECK --interval=5s --timeout=5s --retries=3 CMD wget localhost:8080/health -q -O - > /dev/null 2>&1

CMD ["./main"]
