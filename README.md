# AGGK Trivia Multiplayer

## Membros da Equipe:
* Álvaro Ribeiro;
* Gabriel Felipe;
* Guilherme Oliveira;
* Kleber Barbosa.

## Descrição

Este projeto é um jogo de perguntas e respostas, onde dois ou mais jogadores conectados recebem simultaneamente perguntas acompanhadas de quatro alternativas de resposta, e precisam responder corretamente antes do fm da contagem regressiva para receber uma pontuação, que varia de acordo com quão rápida foi a resposta em relação aos outros jogadores, e apresenta um placar ordenado com todos os jogadores e suas pontuações. O programa apresenta uma interface CLI simples para cada jogador, e utiliza goroutines e canais(concorrência visível).

##Funcionamento

### Conexão
* O servidor é iniciado e aguarda conexões de jogadores.
* Jogadores se conectam com o endereço do servidor e escolhem um nome.

### Início
* O administrador do console do servidor pressiona 'ENTER' para iniciar a partida com os jogadores conectados.
* O servidor carrega perguntas do arquivo 'perguntas.json', embaralha e seleciona a quantidade desejada.

### Rodadas
* O jogo começa com uma contagem regressiva, visualizada por todos os jogadores.
* O servidor envia uma perginta e suas alternativas para todos os jogadores simultaneamente.
* O cliente tem um tempo limitado para responder.
* O jogador digita 'A', 'B', 'C' ou 'D' e envia a resposta ao servidor.

### Pontuação
* Os pontos são calculados com base na ordem de chegada das respostas corretas.
* O primeiro jogador que acertou recebe 100 pontos, e os seguintes recebem metade da pontuação do anteerior.
* O placar atualizado e transmitido ao fim de cada rodada.

### Fim de jogo
* Depois que todas as perguntas selecionadas são feitas, o placar final é exibido e o jogo é encerrado.

# Recursos compartilhados
* 'players'
* 'net.Conn'
* 'canalResposta'
* 'perguntas.json'

# Sincronização

O servidor utiliza Goroutines para lidar com a concorrência. que são iniciadas sempre que um jogador se conecta, sendo responsável por solicitar o nome do jogador sem impedir que novos jogadores se conectem, além de serem criadas uma para cada jogador durante a coleta de respostas, que envia as respostas atraves de um canal, permitindo a chegada de forma paralela.

# Parametros
* Número de perguntas.
* Tempo para responder.

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
