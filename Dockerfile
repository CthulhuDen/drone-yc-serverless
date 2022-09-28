FROM golang as build
COPY . /src
RUN cd /src && CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o /yc-serverless .

FROM debian:stable-slim
RUN apt-get update && apt-get -y install ca-certificates && rm -rf /var/lib/apt/lists/* # Failed to connect without it
CMD ["/yc-serverless"]
COPY --from=build /yc-serverless /
