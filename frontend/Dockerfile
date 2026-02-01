# Build stage
FROM ghcr.io/cirruslabs/flutter:3.24.0 AS builder
WORKDIR /app

COPY pubspec.yaml pubspec.lock ./
RUN flutter pub get

# Copy the rest (no .dart_tool because of .dockerignore)
COPY . .

# Extra safety: ensure fresh tool state inside the container
RUN flutter clean

ARG API_BASE=
RUN if [ -n "$API_BASE" ]; then \
      flutter build web --release --dart-define=API_BASE="$API_BASE"; \
    else \
      flutter build web --release; \
    fi

# Serve stage
FROM nginx:alpine
COPY --from=builder /app/build/web /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]