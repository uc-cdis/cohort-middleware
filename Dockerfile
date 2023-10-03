FROM public.ecr.aws/amazonlinux/amazonlinux:2023-minimal as base

ARG TARGETOS
ARG TARGETARCH

ENV appname=cohort-middleware

ENV CGO_ENABLED=0
ENV GOOS=${TARGETOS}
ENV GOARCH=${TARGETARCH}

FROM base as builder
RUN dnf install -y go \
    && dnf clean all \
    && rm -rf /var/cache/yum/

WORKDIR $GOPATH/src/github.com/uc-cdis/cohort-middleware/

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN GITCOMMIT=$(git rev-parse HEAD) \
    GITVERSION=$(git describe --always --tags) \
    && go build \
    -ldflags="-X 'github.com/uc-cdis/cohort-middleware/version.GitCommit=${GITCOMMIT}' -X 'github.com/uc-cdis/cohort-middleware/version.GitVersion=${GITVERSION}'" \
    -o /cohort-middleware

FROM scratch
COPY --from=builder /etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /cohort-middleware /cohort-middleware
CMD ["/cohort-middleware"]
