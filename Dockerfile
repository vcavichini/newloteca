# Estágio 1: Compilação
# Usamos uma imagem oficial do Go para compilar nossa aplicação.
FROM golang:1.21-alpine AS builder

# Define o diretório de trabalho dentro do contêiner
WORKDIR /app

# Copia os arquivos de módulo e baixa as dependências
COPY go.mod go.sum ./
RUN go mod download

# Copia todo o código-fonte
COPY . .

# Compila a aplicação para um executável estático
# CGO_ENABLED=0 é importante para criar um binário que não depende de bibliotecas do sistema.
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /newloteca .


# Estágio 2: Execução
# Usamos uma imagem "distroless", que é mínima e super segura, contendo apenas nossa aplicação e nada mais.
FROM gcr.io/distroless/static-debian11

# Define o diretório de trabalho
WORKDIR /

# Copia apenas o executável compilado do estágio anterior
COPY --from=builder /newloteca /newloteca

# Expõe a porta que nossa aplicação usa
EXPOSE 8080

# Define o comando para iniciar a aplicação quando o contêiner rodar
CMD ["/newloteca"]
