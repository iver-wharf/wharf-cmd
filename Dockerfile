ARG REG=docker.io
FROM ${REG}/library/golang:1.18 AS build
WORKDIR /src
RUN go install github.com/swaggo/swag/cmd/swag@v1.8.0
COPY go.mod go.sum ./
RUN go mod download

COPY . .
ARG BUILD_VERSION="local docker"
ARG BUILD_GIT_COMMIT="HEAD"
ARG BUILD_REF="0"
ARG BUILD_DATE=""
RUN chmod +x scripts/update-version.sh  \
    && scripts/update-version.sh assets/version.yaml \
    && make swag check \
    && CGO_ENABLED=0 go build -o wharf ./cmd/wharf

ARG REG=docker.io
FROM ${REG}/library/alpine:3.15 AS final
RUN apk add --no-cache ca-certificates tzdata git
COPY --from=build /src/wharf /usr/local/bin/wharf
ENTRYPOINT ["/usr/local/bin/wharf"]

ARG BUILD_VERSION
ARG BUILD_GIT_COMMIT
ARG BUILD_REF
ARG BUILD_DATE
# The added labels are based on this: https://github.com/projectatomic/ContainerApplicationGenericLabels
LABEL name="iver-wharf/wharf-cmd" \
    url="https://github.com/iver-wharf/wharf-cmd" \
    release=${BUILD_REF} \
    build-date=${BUILD_DATE} \
    vendor="Iver" \
    version=${BUILD_VERSION} \
    vcs-type="git" \
    vcs-url="https://github.com/iver-wharf/wharf-cmd" \
    vcs-ref=${BUILD_GIT_COMMIT} \
    changelog-url="https://github.com/iver-wharf/wharf-cmd/blob/${BUILD_VERSION}/CHANGELOG.md" \
    authoritative-source-url="quay.io"
