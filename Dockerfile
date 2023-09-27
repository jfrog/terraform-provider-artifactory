FROM alpine AS base
RUN apk add --no-cache git terraform wget make && \
    wget -q https://github.com/goreleaser/goreleaser/releases/download/v1.21.2/goreleaser_1.21.2_x86_64.apk  && \
    wget -q https://go.dev/dl/go1.21.1.linux-amd64.tar.gz && \
    rm -rf /usr/local/go && \
    tar -C /usr/local -xzf go1.21.1.linux-amd64.tar.gz && \
	apk add --allow-untrusted goreleaser_1.21.2_x86_64.apk && \
    mkdir -p /src/terraform-provider-artifactory
ENV PATH=$PATH:/usr/local/go/bin
WORKDIR /root

FROM base as builder
COPY . .
RUN go mod download
RUN make
WORKDIR /root/v5-v6-migrator
RUN make

FROM hashicorp/terraform as plugin
RUN adduser -S jfrog
WORKDIR /home/jfrog
COPY --from=builder /src/terraform-provider-artifactory/terraform-provider-artifactory /home/jfrog/.terraform.d/plugins/

FROM alpine as migrator
RUN adduser -S jfrog
WORKDIR /home/jfrog
COPY --from=builder /src/terraform-provider-artifactory/v5-v6-migrator/tf-v5-migrator /home/jfrog/tf-v5-migrator
WORKDIR /home/jfrog
