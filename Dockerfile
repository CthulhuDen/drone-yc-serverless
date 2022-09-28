FROM golang as build
COPY . /src
RUN cd /src && CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o /yc-serverless . && strip /yc-serverless

FROM debian:stable-slim
CMD ["/yc-serverless"]
COPY --from=build /yc-serverless /
