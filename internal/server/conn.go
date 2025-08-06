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

// Inicia o servidor na porta especificada
func (server *ServerJogo) IniciarServer(porta string) error {
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
				go server.controlaCliente(conn)
			default:
				// Servidor lotado
				conn.Write([]byte("{\"tipo\":\"servidor_lotado\"}\n"))
				conn.Close()
			}
		}
	}
}

// Gerencia a conexão de um cliente
func (server *ServerJogo) controlaCliente(conn net.Conn) {
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

// Adiciona um jogador à lista de forma thread-safe
func (server *ServerJogo) addJogador(jogador *Jogador) {
	server.jogadoresMutex.Lock()
	server.jogadores = append(server.jogadores, jogador)
	numPlayers := len(server.jogadores)
	server.jogadoresMutex.Unlock()

	fmt.Printf("%s conectou-se. (%d/%d jogadores conectados)\n",
		jogador.Nome, numPlayers, server.maxJogadores)
}

// Remove um jogador da lista de forma thread-safe
func (server *ServerJogo) removerJogador(jogador *Jogador) {
	server.jogadoresMutex.Lock()
	for i, p := range server.jogadores {
		if p == jogador {
			server.jogadores = append(server.jogadores[:i], server.jogadores[i+1:]...)
			break
		}
	}
	numJogadores := len(server.jogadores)
	server.jogadoresMutex.Unlock()

	if jogador.Nome != "" {
		fmt.Printf("%s desconectou-se. (%d/%d jogadores restantes)\n",
			jogador.Nome, numJogadores, server.maxJogadores)
	} else {
		fmt.Printf("Um jogador desconectou-se antes de se identificar. (%d/%d jogadores restantes)\n",
			numJogadores, server.maxJogadores)
	}
}

// Envia uma mensagem para todos os jogadores conectados
func (server *ServerJogo) TransmitirMsg(msg []byte) {
	server.jogadoresMutex.Lock()
	defer server.jogadoresMutex.Unlock()

	for _, jogador := range server.jogadores {
		jogador.Conn.Write(msg)
	}
}

// Retorna uma cópia dos jogadores atuais
func (server *ServerJogo) RetornarJogadores() []*Jogador {
	server.jogadoresMutex.Lock()
	defer server.jogadoresMutex.Unlock()

	players := make([]*Jogador, len(server.jogadores))
	copy(players, server.jogadores)
	return players
}

// Retorna o número atual de jogadores
func (server *ServerJogo) RetornarNumJogadores() int {
	server.jogadoresMutex.Lock()
	defer server.jogadoresMutex.Unlock()
	return len(server.jogadores)
}

// AtualizarPontos atualiza a pontuação de um jogador
func (server *ServerJogo) AtualizarPontos(nomeJogador string, pontos int) {
	server.jogadoresMutex.Lock()
	defer server.jogadoresMutex.Unlock()

	for _, jogador := range server.jogadores {
		if jogador.Nome == nomeJogador {
			jogador.Pontuacao += pontos
			break
		}
	}
}

// Coleta respostas e dá feddback
func (server *ServerJogo) ColetarRespostas(tempo_duracao time.Duration, opcaoCorreta string) []models.Resposta {
	jogadores := server.RetornarJogadores()
	var respostas []models.Resposta
	tempo_limite := time.Now().Add(tempo_duracao)
	canalResposta := make(chan models.Resposta, len(jogadores))

	for _, jogador := range jogadores {
		// Passa a resposta correta para a goroutine que lê a resposta do jogador
		go server.lerResposta(jogador, tempo_limite, canalResposta, opcaoCorreta)
	}

	// Continua a coletar respostas até que o tempo se esgote ou todos respondam
	for i := 0; i < len(jogadores); i++ {
		select {
		case resp := <-canalResposta:
			respostas = append(respostas, resp)
		case <-time.After(tempo_limite.Sub(time.Now())):
			return respostas // O tempo acabou
		}
	}
	return respostas
}

// Verifica a resposta e envia feedback imediato
func (server *ServerJogo) lerResposta(jogador *Jogador, tempo_limite time.Time, canal chan<- models.Resposta, opcaoCorreta string) {
	leitor := bufio.NewReader(jogador.Conn)
	jogador.Conn.SetReadDeadline(tempo_limite)

	msg, err := leitor.ReadBytes('\n')
	if err != nil {
		return // O jogador não respondeu a tempo
	}

	var resp struct {
		Opcao string `json:"opcao"`
	}

	if err := json.Unmarshal(msg, &resp); err == nil {
		// Lógica de feedback imediato
		correta := (resp.Opcao == opcaoCorreta)
		feedbackMsg := fmt.Sprintf("{\"tipo\":\"resultado_resposta\",\"correta\":%t}\n", correta)
		_, err := jogador.Conn.Write([]byte(feedbackMsg))
		if err != nil {
			fmt.Printf("Erro ao enviar feedback para %s: %v\n", jogador.Nome, err)
		}

		// Envia a resposta para o canal principal para ser usada no cálculo de pontos
		canal <- models.Resposta{
			Jogador: jogador.Nome,
			Opcao:   resp.Opcao,
			Tempo:   time.Now(),
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
func ObterIPlocal() string {
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
