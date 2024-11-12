ARG AZLINUX_BASE_VERSION=master

FROM 707767160287.dkr.ecr.us-east-1.amazonaws.com/gen3/golang-build-base:${AZLINUX_BASE_VERSION} AS base
# FROM --platform=$BUILDPLATFORM quay.io/cdis/golang-build-base:${AZLINUX_BASE_VERSION} AS base

ARG TARGETOS
ARG TARGETARCH

ENV appname=cohort-middleware

ENV CGO_ENABLED=0
ENV GOOS=${TARGETOS}
ENV GOARCH=${TARGETARCH}
ENV GOPATH=/root/go

FROM base AS builder
WORKDIR $GOPATH/src/github.com/uc-cdis/cohort-middleware/

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN GITCOMMIT=$(git rev-parse HEAD) \
    GITVERSION=$(git describe --always --tags) \
    && go build \
    -ldflags="-X 'github.com/uc-cdis/cohort-middleware/version.GitCommit=${GITCOMMIT}' -X 'github.com/uc-cdis/cohort-middleware/version.GitVersion=${GITVERSION}'" \
    -o /cohort-middleware

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /cohort-middleware /cohort-middleware
CMD ["/cohort-middleware"]
