# AGGK Trivia Multiplayer

## Membros da Equipe:
* Álvaro Ribeiro;
* Gabriel Felipe;
* Guilherme Oliveira;
* Kleber Barbosa.

## Descrição

Este projeto é um jogo de perguntas e respostas, onde dois ou mais jogadores conectados recebem simultaneamente perguntas acompanhadas de quatro alternativas de resposta, e precisam responder corretamente antes do fm da contagem regressiva para receber uma pontuação, que varia de acordo com quão rápida foi a resposta em relação aos outros jogadores, e apresenta um placar ordenado com todos os jogadores e suas pontuações. O programa apresenta uma interface CLI simples para cada jogador, e utiliza goroutines e canais(concorrência visível).

# como executar
1. abirir o bin/trivia-server.exe
2. abrir o bin/trivia-cliente.exe
3. cliente digitar o ip do server (caso seja no mesmo pc ip será localhost:8080)
4. esperar pelo menos 2 jogadores para começar o jogo

# como gerar novamente executável em caso de atualização e executar
go build -o bin/trivia-server.exe ./cmd/server \n
go build -o bin/trivia-client.exe ./cmd/client \n
go run ./cmd/server \n
go run ./cmd/client \n
