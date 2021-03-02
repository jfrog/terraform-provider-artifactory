# Fetch the dependencies
FROM golang:1.15-alpine AS builder

RUN apk add --update ca-certificates git gcc g++ libc-dev
WORKDIR /src/

ENV GO111MODULE=on

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY pkg/ /src/pkg/
COPY main.go /src/

RUN CGO_ENABLED=0 GOOS=linux go build


# Build the final image
FROM hashicorp/terraform:0.13

COPY --from=builder /src/terraform-provider-artifactory /root/.terraform.d/plugins/
