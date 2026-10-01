# ---- Build stage ----
FROM golang:1.27-alpine AS build
WORKDIR /src

# Cache deps first: copy go.mod/go.sum, download, THEN copy source.
# This layer is reused across builds unless deps change -> faster rebuilds.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=0 -> a fully static binary with no libc dependency,
# so it runs in a scratch/distroless image.
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/auth ./services/auth

# ---- Runtime stage ----
FROM alpine:3.20
RUN adduser -D -u 10001 appuser
COPY --from=build /out/auth /usr/local/bin/auth
USER appuser
EXPOSE 8080
ENTRYPOINT ["auth"]
