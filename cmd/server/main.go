package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"os"
	"time"
	"triviaMultiplayer/internal/server"
)

// Constante para o número máximo de jogadores
const maxJogadores = 10

// struct para carregar as perguntas do arquivo perguntas.json
type PerguntaJSON struct {
	Enunciado    string   `json:"enunciado"`
	Alternativas []string `json:"alternativas"`
	Resposta     string   `json:"resposta_correta"`
}

type Player struct {
	Nome      string
	Conn      net.Conn
	pontuacao int
}

type Pergunta struct {
	Tipo         string   `json:"tipo"`
	ID           int      `json:"id"`
	Texto        string   `json:"texto"`
	Opcoes       []string `json:"opcoes"`
	OpcaoCorreta string   `json:"-"`
}

type Resposta struct {
	Tipo   string `json:"tipo"`
	ID     int    `json:"id"`
	Player string `json:"player"`
	Opcao  string `json:"opcao"`
	Tempo  time.Time
}

type Placar struct {
	Tipo       string             `json:"tipo"`
	Pontuacoes []server.Pontuacao `json:"pontuacoes"`
}

// funcao para carregar perguntas do arquivo JSON
func carregarPerguntasDoArquivo(caminho string, limite int) ([]Pergunta, error) {
	//Lê o arquivo JSON
	arquivoBytes, err := os.ReadFile(caminho)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler o arquivo: %w", err)
	}

	// Decodifica o JSON para a struct
	var perguntasJSON []PerguntaJSON
	if err := json.Unmarshal(arquivoBytes, &perguntasJSON); err != nil {
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
	var perguntasJogo []Pergunta
	for i, pJSON := range perguntasJSON {
		perguntasJogo = append(perguntasJogo, Pergunta{
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
	var pontuacaoAtual []server.Pontuacao

	for _, player := range players {
		pontuacaoAtual = append(pontuacaoAtual, server.Pontuacao{
			Player: player.Nome,
			Pontos: player.Pontuacao,
		})
	}

	placar := Placar{Tipo: "placar", Pontuacoes: pontuacaoAtual}
	sbBytes, _ := json.Marshal(placar)
	gameServer.Broadcast(append(sbBytes, '\n'))
}

// Inicializa o servidor
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

func executarJogo(gameServer *server.ServerJogo, perguntas []Pergunta) {
	// Contagem regressiva
	contagemRegressiva(gameServer, 3)

	// Executa cada pergunta
	for _, pergunta := range perguntas {
		qBytes, _ := json.Marshal(pergunta)
		gameServer.Broadcast(append(qBytes, '\n'))

		respostas := gameServer.ColetarRespostas(10 * time.Second)

		pontos := server.CalcularPontos(respostas, pergunta.OpcaoCorreta)
		for _, ponto := range pontos {
			gameServer.AtualizarPontos(ponto.Player, ponto.Pontos)
		}

		enviarPlacar(gameServer)
		time.Sleep(5 * time.Second)
	}

	fmt.Println("Fim de jogo!")
	enviarPlacar(gameServer)
}