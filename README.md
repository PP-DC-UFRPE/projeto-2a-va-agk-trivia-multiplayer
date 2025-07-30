# projeto-2a-va-agk
projeto-2a-va-agk created by GitHub Classroom

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