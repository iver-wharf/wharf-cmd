FROM golang:1.18 AS build
WORKDIR /src
ENV GO111MODULE=on
COPY . /src
RUN CGO_ENABLED=0 go build -o main && go test ./... -v

FROM ubuntu:20.04 AS final
WORKDIR /app
COPY --from=build /src/main /app/
ADD conf/config .kube/config
ADD conf/root.crt conf/int.crt /usr/local/share/ca-certificates/
RUN apt update && apt install -y ca-certificates && update-ca-certificates
ENTRYPOINT ["/app/main"]
