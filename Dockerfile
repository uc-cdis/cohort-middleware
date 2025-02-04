FROM quay.io/cdis/golang:1.23-bookworm AS base

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

RUN echo "nobody:x:65534:65534:Nobody:/:" > /etc_passwd

FROM scratch
COPY --from=builder /etc_passwd /etc/passwd
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /cohort-middleware /cohort-middleware
USER nobody
CMD ["/cohort-middleware"]
