package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
	"triviaMultiplayer/internal/models"
)

// Jogador representa um jogador conectado
type Jogador struct {
	Nome      string
	Conn      net.Conn
	Pontuacao int
}

// ServerJogo gerencia as conexões e estado do jogo
type ServerJogo struct {
	jogadores      []*Jogador
	jogadoresMutex *sync.Mutex
	semaforo       chan struct{}
	maxJogadores   int
	sair           chan struct{}
	listener       net.Listener
}

// NovoServer cria uma nova instância do servidor
func NovoServer(maxJogadores int) *ServerJogo {
	return &ServerJogo{
		jogadores:      make([]*Jogador, 0),
		jogadoresMutex: &sync.Mutex{},
		semaforo:       make(chan struct{}, maxJogadores),
		maxJogadores:   maxJogadores,
		sair:           make(chan struct{}),
	}
}

// Iniciar inicia o servidor na porta especificada
func (server *ServerJogo) Iniciar(porta string) error {
	listener, err := net.Listen("tcp", porta)
	if err != nil {
		return fmt.Errorf("erro ao iniciar servidor: %w", err)
	}

	server.listener = listener
	fmt.Printf("Servidor ouvindo em %s\n", porta)

	// Goroutine para aceitar conexões
	go server.aceitarConnect()

	return nil
}

// aceitarConnect aceita novas conexões de clientes
func (server *ServerJogo) aceitarConnect() {
	for {
		select {
		case <-server.sair:
			return
		default:
			conn, err := server.listener.Accept()
			if err != nil {
				if strings.Contains(err.Error(), "conexão fechada") {
					return
				}
				fmt.Println("Erro ao aceitar conexão:", err)
				continue
			}

			// Controla o número máximo de jogadores
			select {
			case server.semaforo <- struct{}{}:
				go server.handleCliente(conn)
			default:
				// Servidor lotado
				conn.Write([]byte("{\"tipo\":\"servidor_lotado\"}\n"))
				conn.Close()
			}
		}
	}
}

// handleCliente gerencia a conexão de um cliente
func (server *ServerJogo) handleCliente(conn net.Conn) {
	jogador := &Jogador{Conn: conn}

	defer conn.Close()
	defer func() { <-server.semaforo }()
	defer server.removerJogador(jogador)

	// Solicita o nome do jogador
	if !server.pedirNome(jogador) {
		return
	}

	// Adiciona o jogador à lista
	server.addJogador(jogador)

	// Aguarda o fim do jogo
	<-server.sair
}

// pedirNome solicita e obtém o nome do jogador
func (server *ServerJogo) pedirNome(jogador *Jogador) bool {
	_, err := jogador.Conn.Write([]byte("{\"tipo\":\"nome_requisicao\"}\n"))
	if err != nil {
		return false
	}

	nome, err := bufio.NewReader(jogador.Conn).ReadString('\n')
	if err != nil {
		return false
	}

	jogador.Nome = strings.TrimSpace(nome)
	jogador.Pontuacao = 0
	return true
}

// addJogador adiciona um jogador à lista de forma thread-safe
func (server *ServerJogo) addJogador(jogador *Jogador) {
	server.jogadoresMutex.Lock()
	server.jogadores = append(server.jogadores, jogador)
	numPlayers := len(server.jogadores)
	server.jogadoresMutex.Unlock()

	fmt.Printf("%s conectou-se. (%d/%d jogadores conectados)\n",
		jogador.Nome, numPlayers, server.maxJogadores)
}

// removerJogador remove um jogador da lista de forma thread-safe
func (server *ServerJogo) removerJogador(jogador *Jogador) {
	server.jogadoresMutex.Lock()
	for i, p := range server.jogadores {
		if p == jogador {
			server.jogadores = append(server.jogadores[:i], server.jogadores[i+1:]...)
			break
		}
	}
	numPlayers := len(server.jogadores)
	server.jogadoresMutex.Unlock()

	if jogador.Nome != "" {
		fmt.Printf("%s desconectou-se. (%d/%d jogadores restantes)\n",
			jogador.Nome, numPlayers, server.maxJogadores)
	} else {
		fmt.Printf("Um jogador desconectou-se antes de se identificar. (%d/%d jogadores restantes)\n",
			numPlayers, server.maxJogadores)
	}
}

// Broadcast envia uma mensagem para todos os jogadores conectados
func (server *ServerJogo) Broadcast(mensagem []byte) {
	server.jogadoresMutex.Lock()
	defer server.jogadoresMutex.Unlock()

	for _, jogador := range server.jogadores {
		jogador.Conn.Write(mensagem)
	}
}

// GetJogadores retorna uma cópia dos jogadores atuais
func (server *ServerJogo) GetJogadores() []*Jogador {
	server.jogadoresMutex.Lock()
	defer server.jogadoresMutex.Unlock()

	players := make([]*Jogador, len(server.jogadores))
	copy(players, server.jogadores)
	return players
}

// GetNumJogadores retorna o número atual de jogadores
func (server *ServerJogo) GetNumJogadores() int {
	server.jogadoresMutex.Lock()
	defer server.jogadoresMutex.Unlock()
	return len(server.jogadores)
}

// AtualizarPontos atualiza a pontuação de um jogador
func (server *ServerJogo) AtualizarPontos(nomeJogador string, pontos int) {
	server.jogadoresMutex.Lock()
	defer server.jogadoresMutex.Unlock()

	for _, player := range server.jogadores {
		if player.Nome == nomeJogador {
			player.Pontuacao += pontos
			break
		}
	}
}

// ColetarRespostas coleta respostas dos jogadores por um tempo determinado
func (server *ServerJogo) ColetarRespostas(duration time.Duration) []models.Resposta {
	players := server.GetJogadores()

	var respostas []models.Resposta
	deadline := time.Now().Add(duration)
	canalResposta := make(chan models.Resposta, len(players))

	for _, player := range players {
		go server.lerResposta(player, deadline, canalResposta)
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

// lerResposta lê a resposta de um jogador específico
func (server *ServerJogo) lerResposta(jogador *Jogador, deadline time.Time, canal chan<- models.Resposta) {
	reader := bufio.NewReader(jogador.Conn)

	for {
		jogador.Conn.SetReadDeadline(time.Now().Add(1 * time.Second))
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

		var resp struct {
			Tipo   string `json:"tipo"`
			ID     int    `json:"id"`
			Player string `json:"player"`
			Opcao  string `json:"opcao"`
		}

		if err := json.Unmarshal(msg, &resp); err == nil && resp.Tipo == "resposta" {
			canal <- models.Resposta{
				Player: jogador.Nome,
				Opcao:  resp.Opcao,
				Tempo:  time.Now(),
			}
			return
		}
	}
}

// Parar para o servidor e desconecta todos os jogadores
func (server *ServerJogo) Parar() {
	close(server.sair)
	if server.listener != nil {
		server.listener.Close()
	}
}

// GetLocalIP busca o endereço IP local
func GetLocalIP() string {
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
