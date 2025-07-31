package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync" // Pacote para Mutex
	"time"
	"triviaMultiplayer/internal/server"
)

// Constante para o número máximo de jogadores
const maxPlayers = 10

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

// Usando uma fatia e um Mutex para os jogadores
var players []*Player
var playersMutex = &sync.Mutex{}

// Semáforo para limitar o número de jogadores
var semaphore = make(chan struct{}, maxPlayers)

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

// ALTERAÇÃO 1: A função agora recebe um canal 'quit' para saber quando terminar.
func handleCliente(conn net.Conn, quit <-chan struct{}) {
	player := &Player{Conn: conn}

	// Defer executa na ordem inversa (LIFO - Last In, First Out)
	// 1. A conexão é fechada.
	// 2. O lugar no semáforo é liberado.
	// 3. O jogador é removido da lista global.
	defer conn.Close()
	defer func() { <-semaphore }()
	defer func() {
		playersMutex.Lock()
		// Encontra e remove o jogador da fatia
		for i, p := range players {
			if p == player {
				players = append(players[:i], players[i+1:]...)
				break
			}
		}
		numPlayers := len(players)
		playersMutex.Unlock()
		if player.Nome != "" {
			fmt.Printf("%s desconectou-se. (%d/%d jogadores restantes)\n", player.Nome, numPlayers, maxPlayers)
		} else {
			fmt.Printf("Um jogador desconectou-se antes de se identificar. (%d/%d jogadores restantes)\n", numPlayers, maxPlayers)
		}

	}()

	_, err := conn.Write([]byte("{\"tipo\":\"nome_requisicao\"}\n"))
	if err != nil {
		return // Se não conseguir escrever, encerra a goroutine
	}

	nome, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return // Se não conseguir ler o nome, encerra a goroutine
	}

	player.Nome = strings.TrimSpace(nome)
	player.pontuacao = 0

	// Adiciona o jogador à lista de forma segura
	playersMutex.Lock()
	players = append(players, player)
	numPlayers := len(players)
	playersMutex.Unlock()

	fmt.Printf("%s conectou-se. (%d/%d jogadores conectados)\n", player.Nome, numPlayers, maxPlayers)

	// ALTERAÇÃO 2: Removemos o loop de leitura. A goroutine agora espera passivamente
	// até que o canal 'quit' seja fechado na função main, indicando o fim do jogo.
	<-quit
}

func broadcast(mensagem []byte) {
	for _, player := range players {
		player.Conn.Write(mensagem)
	}
}

//Busca o endereço IP do ADM
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

//Realiza a contagem regressiva para os jogadores 
func contagemRegressiva(valor int) {
	for i := valor; i > 0; i-- {
		broadcast([]byte(fmt.Sprintf("{\"tipo\":\"contagem_regressiva\",\"valor\":%d}\n", i)))
		fmt.Printf("Começando em %d...\n", i)
		time.Sleep(1 * time.Second)
	}
	broadcast([]byte("{\"tipo\":\"contagem_regressiva\",\"valor\":0}\n"))
}

//Coleta as respostas dos jogadores
func coletarRespostas(duration time.Duration) []server.Resposta {
	playersMutex.Lock()
	jogadoresAtuais := make([]*Player, len(players))
	copy(jogadoresAtuais, players)
	playersMutex.Unlock()

	var respostas []server.Resposta
	deadline := time.Now().Add(duration)
	canalResposta := make(chan server.Resposta, len(jogadoresAtuais))

	for _, player := range jogadoresAtuais {
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

	for len(respostas) < len(jogadoresAtuais) && time.Now().Before(deadline) {
		select {
		case resp := <-canalResposta:
			respostas = append(respostas, resp)
		case <-time.After(100 * time.Millisecond):
		}
	}

	return respostas
}

//Envia placar aos jogadores
func enviarPlacar() {
	playersMutex.Lock()
	defer playersMutex.Unlock()

	var pontuacaoAtual []server.Pontuacao
	for _, player := range players {
		pontuacaoAtual = append(pontuacaoAtual, server.Pontuacao{Player: player.Nome, Pontos: player.pontuacao})
	}

	placar := Placar{Tipo: "placar", Pontuacoes: pontuacaoAtual}
	sbBytes, _ := json.Marshal(placar)
	broadcast(append(sbBytes, '\n'))
}

//Inicializa o servidor
func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

//Programa seu encerramento 
	defer listener.Close()
	fmt.Printf("Servidor ouvindo em %s:8080\n", getIpLocal())
	fmt.Printf("Aguardando jogadores... (limite de %d)\n", maxPlayers)

	// ALTERAÇÃO 3: Canal para sinalizar o fim do jogo para as goroutines.
	quit := make(chan struct{})

//Gerencia a entrada de usuários
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				// Se o listener foi fechado, encerra a goroutine
				if strings.Contains(err.Error(), "use of closed network connection") {
					return
				}
				fmt.Println("Erro ao aceitar conexão:", err)
				continue
			}

			semaphore <- struct{}{}
			// ALTERAÇÃO 4: Passa o canal 'quit' para o handler.
			go handleCliente(conn, quit)
		}
	}()

	fmt.Println("\nO servidor está pronto para aceitar jogadores.")
	fmt.Println("Pressione ENTER a qualquer momento para iniciar a partida com os jogadores conectados.")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

	playersMutex.Lock()
	numPlayers := len(players)
	playersMutex.Unlock()

	if numPlayers == 0 {
		fmt.Println("Nenhum jogador conectado. Encerrando o servidor.")
		return
	}

	fmt.Printf("\nO jogo vai começar com %d jogador(es)!\n", numPlayers)
	broadcast([]byte("{\"tipo\":\"inicio_jogo\"}\n"))
	time.Sleep(1 * time.Second)

	perguntas, err := carregarPerguntasDoArquivo("perguntas.json", 5)
	if err != nil {
		fmt.Printf("Erro fatal ao carregar perguntas: %v\n", err)
		return
	}
	fmt.Printf("Jogo iniciado com %d perguntas aleatórias.\n", len(perguntas))
	time.Sleep(2 * time.Second)

	// contagem regressiva antes de iniciar o jogo
	contagemRegressiva(3)

	for _, pergunta := range perguntas {
		qBytes, _ := json.Marshal(pergunta)
		broadcast(append(qBytes, '\n'))

		respostas := coletarRespostas(10 * time.Second)

		pontos := server.CalcularPontos(respostas, pergunta.OpcaoCorreta)
		playersMutex.Lock()
		fmt.Printf("Pergunta %d respondida. Distribuindo pontos...\n", pergunta.ID)
		for _, ponto := range pontos {
			for _, player := range players {
				if player.Nome == ponto.Player {
					player.pontuacao += ponto.Pontos
				}
			}
		}
		playersMutex.Unlock()

		enviarPlacar()
		fmt.Printf("placar enviado. Aguardando 5 segundos...\n")
		time.Sleep(5 * time.Second)
	}

	fmt.Println("Fim de jogo!")
	enviarPlacar()
	time.Sleep(1 * time.Second)

	// ALTERAÇÃO 5: Fecha o canal 'quit', sinalizando para todas as goroutines 'handleCliente'
	// que elas podem encerrar, fechar suas conexões e limpar os recursos.
	close(quit)

	time.Sleep(1 * time.Second)
}
