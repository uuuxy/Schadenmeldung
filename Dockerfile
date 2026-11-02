# 1. Oberfläche bauen
FROM node:26-alpine AS oberflaeche
WORKDIR /bau/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci --no-audit --no-fund
COPY frontend/ ./
RUN npm run build

# 2. Programm bauen: reines Go ohne CGO; Oberfläche und Migrationen werden eingebettet
FROM golang:1.27-alpine AS programm
WORKDIR /bau
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
COPY migrations/ ./migrations/
COPY --from=oberflaeche /bau/frontend/dist ./frontend/dist
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /schaden .

# 3. Laufzeit: nur das Programm, ohne Shell, als unprivilegierter Benutzer. Alle Daten liegen in
#    PostgreSQL; der Container selbst hat keine.
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=programm /schaden /schaden
ENV PORT=8090
EXPOSE 8090
USER nonroot
HEALTHCHECK --interval=30s --timeout=5s --retries=3 CMD ["/schaden", "gesund"]
ENTRYPOINT ["/schaden"]
