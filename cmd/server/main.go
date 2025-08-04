package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"
	"triviaMultiplayer/internal/server"
	"triviaMultiplayer/internal/models"
)

// Constante para o número máximo de jogadores
const maxJogadores = 10

// função para carregar perguntas do arquivo JSON
func carregarPerguntasDoArquivo(caminho string, limite int) ([]models.Pergunta, error) {
	//Lê o arquivo JSON
	arquivoBytes, err := os.ReadFile(caminho)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler o arquivo: %w", err)
	}

	// Decodifica o JSON para a struct
	var perguntasJSON []models.PerguntaJSON
	err = json.Unmarshal(arquivoBytes, &perguntasJSON)
	if err != nil {
		return nil, fmt.Errorf("erro ao decodificar o JSON: %w", err)
	}

	// Embaralha as perguntas
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(perguntasJSON), func(i, j int) {
		perguntasJSON[i], perguntasJSON[j] = perguntasJSON[j], perguntasJSON[i]
	})

	// Limita o número de perguntas
	if limite > 0 && len(perguntasJSON) > limite {
		perguntasJSON = perguntasJSON[:limite]
	}

	// Converte para o formato de Pergunta do jogo
	var perguntasJogo []models.Pergunta
	for i, pJSON := range perguntasJSON {
		perguntasJogo = append(perguntasJogo, models.Pergunta{
			Tipo:         "pergunta",
			ID:           i + 1, // ID sequencial
			Texto:        pJSON.Enunciado,
			Opcoes:       pJSON.Alternativas,
			OpcaoCorreta: pJSON.Resposta,
		})
	}

	return perguntasJogo, nil
}

func contagemRegressiva(gameServer *server.ServerJogo, valor int) {
	for i := valor; i > 0; i-- {
		gameServer.Broadcast([]byte(fmt.Sprintf("{\"tipo\":\"contagem_regressiva\",\"valor\":%d}\n", i)))
		fmt.Printf("Começando em %d...\n", i)
		time.Sleep(1 * time.Second)
	}
	gameServer.Broadcast([]byte("{\"tipo\":\"contagem_regressiva\",\"valor\":0}\n"))
}

func enviarPlacar(gameServer *server.ServerJogo) {
	players := gameServer.GetJogadores()
	var pontuacaoAtual []models.Pontuacao

	for _, player := range players {
		pontuacaoAtual = append(pontuacaoAtual, models.Pontuacao{
			Player: player.Nome,
			Pontos: player.Pontuacao,
		})
	}

	var pontuacaoModel []models.Pontuacao
	for _, p := range pontuacaoAtual {
		pontuacaoModel = append(pontuacaoModel, models.Pontuacao{
			Player: p.Player,
			Pontos: p.Pontos,
		})
	}

	placar := models.Placar{Tipo: "placar", Pontuacoes: pontuacaoModel}
	sbBytes, _ := json.Marshal(placar)
	gameServer.Broadcast(append(sbBytes, '\n'))
}

func executarJogo(serverJogo *server.ServerJogo, perguntas []models.Pergunta) {
	// Contagem regressiva
	contagemRegressiva(serverJogo, 3)

	// Executa cada pergunta
	for _, pergunta := range perguntas {
		qBytes, _ := json.Marshal(pergunta)
		serverJogo.Broadcast(append(qBytes, '\n'))

		respostas := serverJogo.ColetarRespostas(10 * time.Second)

		pontos := server.CalcularPontos(respostas, pergunta.OpcaoCorreta)
		for _, ponto := range pontos {
			serverJogo.AtualizarPontos(ponto.Player, ponto.Pontos)
		}

		enviarPlacar(serverJogo)
		time.Sleep(5 * time.Second)
	}

	fmt.Println("Fim de jogo!")
	enviarPlacar(serverJogo)
}

func main() {
	// Cria o servidor de jogo
	gameServer := server.NovoServer(maxJogadores)

	// Inicia o servidor
	err := gameServer.Iniciar(":8080")
	if err != nil {
		panic(err)
	}
	defer gameServer.Parar()

	fmt.Printf("Servidor ouvindo em %s:8080\n", server.GetLocalIP())
	fmt.Printf("Aguardando jogadores... (limite de %d)\n", maxJogadores)

	fmt.Println("\nO servidor está pronto para aceitar jogadores.")
	fmt.Println("Pressione ENTER a qualquer momento para iniciar a partida com os jogadores conectados.")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

	numPlayers := gameServer.GetNumJogadores()
	if numPlayers == 0 {
		fmt.Println("Nenhum jogador conectado. Encerrando o servidor.")
		return
	}

	fmt.Printf("\nO jogo vai começar com %d jogador(es)!\n", numPlayers)
	gameServer.Broadcast([]byte("{\"tipo\":\"inicio_jogo\"}\n"))

	// Carrega perguntas e inicia o jogo
	perguntas, err := carregarPerguntasDoArquivo("perguntas.json", 5)
	if err != nil {
		fmt.Printf("Erro fatal ao carregar perguntas: %v\n", err)
		return
	}

	// Inicia a lógica do jogo usando o gameServer
	executarJogo(gameServer, perguntas)
}