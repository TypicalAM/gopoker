############################
# STEP 1 build executable binary
############################
FROM golang:alpine AS builder

# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache 'git=~2'

# Copy dependencies
ENV GO111MODULE=on
WORKDIR $GOPATH/src/gopoker/
COPY . .

# Fetch dependencies.
RUN go mod download

# Build the binary.
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /go/main cmd/gopoker/gopoker.go

############################
# STEP 2 build a small image
############################
FROM alpine:3

WORKDIR /

# Copy our static executable.
COPY --from=builder /go/main /go/main
COPY ./entrypoint.sh /go/entrypoint.sh
RUN chmod +x go/entrypoint.sh

ENV PORT 8080
ENV GIN_MODE release
EXPOSE 8080

WORKDIR /go

# Run the Go Gin binary.
ENTRYPOINT ["/go/entrypoint.sh"]
