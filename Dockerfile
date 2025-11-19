# Gunakan image Go 1.25
FROM golang:1.25-alpine

# Install Air untuk hot-reloading
RUN go install github.com/air-verse/air@latest

# Set working directory di dalam container
WORKDIR /app

# Salin file go.mod dan go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Port yang akan diekspos
EXPOSE 8080

# Perintah default untuk development dengan Air
CMD ["air", "-c", ".air.toml"]
