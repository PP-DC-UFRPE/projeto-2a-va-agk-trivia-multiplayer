package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"
	"triviaMultiplayer/internal/server"
)

type Player struct {
	Nome  string
	Conn  net.Conn
	pontuacao int
}

type Pergunta struct {
	Tipo          string   `json:"tipo"`
	ID            int      `json:"id"`
	Texto        string   `json:"texto"`
	Opcoes       []string `json:"opcoes"`
	OpcaoCorreta string   `json:"-"` // Campo ignorado no JSON para o cliente
}

type Resposta struct {
	Tipo   string `json:"tipo"`
	ID     int    `json:"id"`
	Player string `json:"player"`
	Opcao  string `json:"opcao"`
	Tempo   time.Time
}

type Placar struct {
	Tipo       string         `json:"tipo"`
	Pontuacoes []server.Pontuacao `json:"pontuacoes"`
}

var players = make(map[net.Conn]*Player)

func handleCliente(conn net.Conn) {
	// Pede o nome do jogador
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

	fmt.Println("Aguardando 2 jogadores para começar...")
	for {
		if len(players) >= 2 {
			break
		}
		time.Sleep(1 * time.Second)
	}
	fmt.Println("O jogo vai começar!")
	time.Sleep(3 * time.Second)

	perguntas := []Pergunta{
		{Tipo: "pergunta", ID: 1, Texto: "Qual a capital da França?", Opcoes: []string{"A) Paris", "B) Roma", "C) Berlim", "D) Lisboa"}, OpcaoCorreta: "A"},
		{Tipo: "pergunta", ID: 2, Texto: "Quanto é 5 + 3?", Opcoes: []string{"A) 6", "B) 7", "C) 8", "D) 9"}, OpcaoCorreta: "C"},
	}

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
		time.Sleep(5 * time.Second) // Pausa entre as perguntas
	}

	fmt.Println("Fim de jogo!")
	// Envia placar final
	enviarPlacar()
}

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
					return // Apenas uma resposta por pergunta por jogador
				}
			}
		}(player)
	}

	for len(respostas) < len(players) && time.Now().Before(deadline) {
		select {
		case resp := <-canalResposta:
			respostas = append(respostas, resp)
		case <-time.After(100 * time.Millisecond):
			// continua esperando
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