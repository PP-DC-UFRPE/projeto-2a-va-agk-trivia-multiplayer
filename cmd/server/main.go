package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"  // Pacote para embaralhar
	"net"
	"os"
	"strings"
	"time"
	"triviaMultiplayer/internal/server"
)

// Estrutura para carregar as perguntas do arquivo perguntas.json
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
	Tipo       string           `json:"tipo"`
	Pontuacoes []server.Pontuacao `json:"pontuacoes"`
}

var players = make(map[net.Conn]*Player)

//funcao para carregar perguntas do arquivo JSON
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

func handleCliente(conn net.Conn) {
	conn.Write([]byte("{\"tipo\":\"nome_requisicao\"}\n"))
	nome, _ := bufio.NewReader(conn).ReadString('\n')
	nome = strings.TrimSpace(nome)

	players[conn] = &Player{Nome: nome, Conn: conn, pontuacao: 0}
	fmt.Printf("%s conectou-se. (%d jogadores conectados)\n", nome, len(players))
}

func broadcast(mensagem []byte) {
	for _, player := range players {
		player.Conn.Write(mensagem)
	}
}

func getIpLocal() string {
	enderecos, err := net.InterfaceAddrs()
	if err != nil {
		return "localhost"
	}
	for _, endereco := range enderecos {
		if ipnet, ok := endereco.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "localhost"
}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	fmt.Printf("Servidor ouvindo em %s:8080\n", getIpLocal())

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				fmt.Println("Erro ao aceitar conexão:", err)
				continue
			}
			go handleCliente(conn)
		}
	}()

	fmt.Println("\nO servidor está pronto para aceitar jogadores.")
	fmt.Println("Pressione ENTER a qualquer momento para iniciar a partida com os jogadores conectados.")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

	if len(players) == 0 {
		fmt.Println("Nenhum jogador conectado. Encerrando o servidor.")
		return
	}

	fmt.Printf("\nO jogo vai começar com %d jogador(es)!\n", len(players))
	broadcast([]byte("{\"tipo\":\"inicio_jogo\"}\n"))
	time.Sleep(1 * time.Second)

	// --- CARREGANDO AS PERGUNTAS DO ARQUIVO ---
	perguntas, err := carregarPerguntasDoArquivo("perguntas.json", 5) // Limite de 5 perguntas
	if err != nil {
		fmt.Printf("Erro fatal ao carregar perguntas: %v\n", err)
		return
	}
	fmt.Printf("Jogo iniciado com %d perguntas aleatórias.\n", len(perguntas))
	time.Sleep(2 * time.Second)

	for _, pergunta := range perguntas {
		qBytes, _ := json.Marshal(pergunta)
		broadcast(append(qBytes, '\n'))

		respostas := coletarRespostas(10 * time.Second)

		pontos := server.CalcularPontos(respostas, pergunta.OpcaoCorreta)
		for _, ponto := range pontos {
			for _, player := range players {
				if player.Nome == ponto.Player {
					player.pontuacao += ponto.Pontos
				}
			}
		}

		enviarPlacar()
		time.Sleep(5 * time.Second)
	}

	fmt.Println("Fim de jogo!")
	enviarPlacar()
}

// coletarRespostas e enviarPlacar permanecem os mesmos
func coletarRespostas(duration time.Duration) []server.Resposta {
	var respostas []server.Resposta
	deadline := time.Now().Add(duration)
	canalResposta := make(chan server.Resposta, len(players))

	for _, player := range players {
		go func(p *Player) {
			reader := bufio.NewReader(p.Conn)
			for {
				p.Conn.SetReadDeadline(time.Now().Add(1 * time.Second))
				msg, err := reader.ReadBytes('\n')
				if err != nil {
					if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
						if time.Now().After(deadline) {
							return
						}
						continue
					}
					return
				}

				var resp Resposta
				if err := json.Unmarshal(msg, &resp); err == nil && resp.Tipo == "resposta" {
					resp.Tempo = time.Now()
					resp.Player = p.Nome
					canalResposta <- server.Resposta{Player: resp.Player, Opcao: resp.Opcao, Tempo: resp.Tempo}
					return
				}
			}
		}(player)
	}

	for len(respostas) < len(players) && time.Now().Before(deadline) {
		select {
		case resp := <-canalResposta:
			respostas = append(respostas, resp)
		case <-time.After(100 * time.Millisecond):
		}
	}

	return respostas
}

func enviarPlacar() {
	var pontuacaoAtual []server.Pontuacao
	for _, player := range players {
		pontuacaoAtual = append(pontuacaoAtual, server.Pontuacao{Player: player.Nome, Pontos: player.pontuacao})
	}

	placar := Placar{Tipo: "placar", Pontuacoes: pontuacaoAtual}
	sbBytes, _ := json.Marshal(placar)
	broadcast(append(sbBytes, '\n'))
}