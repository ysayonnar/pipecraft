FROM docker:stable-dind

WORKDIR /workspace

RUN apk add --no-cache git