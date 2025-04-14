ARG PORT=8080
ARG GO_VERSION=1.23
ARG BASE_OS=alpine3.21
ARG BASE_IMAGE=golang:${GO_VERSION}-${BASE_OS}
ARG APP_PATH="/app"
ARG APP_NAME="pipo-dispatcher"

FROM $BASE_IMAGE AS base

FROM base AS builder-base
ARG APP_PATH

# copy project requirement files to ensure they are cached
WORKDIR $APP_PATH
ARG PROGRAM_VERSION
COPY go.mod go.sum ./

ENV GIN_MODE=release

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=bind,source=go.mod,target=go.mod \
    go mod download

COPY . .

# build application
RUN --mount=type=cache,target="/root/.cache/go-build" \
    CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-w -s -X main.version=${PROGRAM_VERSION}" \
    -o ./build/app \
    ./cmd

# test application
FROM builder-base AS test
RUN go test -v ./...

# runtime
FROM gcr.io/distroless/static:nonroot AS production
ARG APP_PATH
ARG APP_NAME

COPY --from=builder-base ${APP_PATH}/build/app /usr/bin
COPY ./config /etc/${APP_NAME}

EXPOSE $PORT
ENTRYPOINT ["app"]
