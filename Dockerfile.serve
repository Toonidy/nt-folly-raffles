# Build Go API Server
FROM golang:1.16-alpine AS go_builder
RUN go version
ARG CAPROVER_GIT_COMMIT_SHA
ADD . /app
WORKDIR /app
RUN go build -ldflags="-X 'follyteam/build.Version=1.0.0' -X 'follyteam/build.BuildHash=$CAPROVER_GIT_COMMIT_SHA'" -o /main ./cmd/serve/main.go

# Final stage build, this will be the container
# that we will deploy to production
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=go_builder /main ./

# Execute Main Server
RUN adduser -D follyteam
USER follyteam
CMD ./main service
