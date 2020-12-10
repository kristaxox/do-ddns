# build stage
FROM golang:alpine AS build-env
RUN apk --no-cache add git
ADD . /src
RUN cd /src && go build -o do-ddns ./...

# final stage
FROM alpine
WORKDIR /app
COPY --from=build-env /src/do-ddns /app/
ENTRYPOINT ./do-ddns
