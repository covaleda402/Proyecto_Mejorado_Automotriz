# Build the backend as a static binary, then run it as a non-root user.
FROM golang:1.25-alpine AS build
WORKDIR /src
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
RUN CGO_ENABLED=0 go build -o /out/server ./cmd/server

FROM alpine:3.21
RUN adduser -D -u 10001 appuser
COPY --from=build /out/server /usr/local/bin/server
USER appuser
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/server"]
