package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync" // Pacote para Mutex
	"triviaMultiplayer/internal/client"
)

// GameState armazena o estado compartilhado entre as goroutines
type GameState struct {
	mu              sync.Mutex // Protege o acesso ao ID da pergunta
	currentQuestionID int
}

// Mensagem genérica para identificar o tipo
type Mensagem struct {
	Tipo string `json:"tipo"`
}

type Pergunta struct {
	ID     int      `json:"id"`
	Texto  string   `json:"texto"`
	Opcoes []string `json:"opcoes"`
}

type Placar struct {
	Pontuacoes []struct {
		Player string `json:"player"`
		Pontos int    `json:"pontos"`
	} `json:"pontuacoes"`
}

// gameState é a nossa variável global para o estado compartilhado
var gameState GameState
var perguntaDisponivel = make(chan struct{})

// lerServidor lida com todas as mensagens recebidas do servidor
func lerServidor(conn net.Conn, done chan<- struct{}) {
	// Garante que o canal 'done' seja fechado quando a função retornar,
	// sinalizando para a main goroutine que a conexão foi perdida.
	defer close(done)
	reader := bufio.NewReader(conn)

	for {
		msg, err := reader.ReadBytes('\n')
		if err != nil {
			client.MostraMensagem("\nConexão perdida com o servidor. Pressione ENTER para sair.")
			return // Encerra a goroutine
		}

		var mensagemBase Mensagem
		if err := json.Unmarshal(msg, &mensagemBase); err != nil {
			continue
		}

		// Processa a mensagem com base no tipo
		switch mensagemBase.Tipo {
		case "nome_requisicao":
			// A outra goroutine vai lidar com o envio do nome

		case "pergunta":
			var pergunta Pergunta
			json.Unmarshal(msg, &pergunta)

			// Trava o mutex para atualizar o ID da pergunta de forma segura
			gameState.mu.Lock()
			gameState.currentQuestionID = pergunta.ID
			gameState.mu.Unlock()

			client.MostraMensagem(fmt.Sprintf("📢 Pergunta %d: %s\n", pergunta.ID, pergunta.Texto))
			for _, opt := range pergunta.Opcoes {
				client.MostraMensagem(opt)
			}
			perguntaDisponivel <- struct{}{}

		case "placar":
			var placar Placar
			json.Unmarshal(msg, &placar)
			client.MostraMensagem("\n\n--- PLACAR ---")
			for _, score := range placar.Pontuacoes {
				client.MostraMensagem(fmt.Sprintf("%s: %d pontos", score.Player, score.Pontos))
			}
			client.MostraMensagem("--------------")
			client.MostraMensagem("Aguardando próxima pergunta...")


		case "contagem_regressiva":
			var data struct {
				Valor int `json:"valor"`
			}
			if err := json.Unmarshal(msg, &data); err == nil {
				if data.Valor > 0 {
					client.MostraMensagem(fmt.Sprintf("\n⏳%d...", data.Valor))
				} else {
					client.MostraMensagem("\n🚦Vai!")
				}
			}
		
		case "inicio_jogo":
			client.MostraMensagem("\nO JOGO VAI COMEÇAR!")
		}
	}
}

// lerUsuario lida com todo o input do teclado do usuário
func lerUsuario(conn net.Conn) {
	isNameSent := false

	for {
		if !isNameSent {
			nome := client.PerguntaNome()
			conn.Write([]byte(nome + "\n"))
			client.MostraMensagem("Aguardando outros jogadores...")
			isNameSent = true
			continue
		}

		<- perguntaDisponivel
		// Após enviar o nome, todo input é uma resposta de pergunta
		resposta := client.PerguntaResposta()

		gameState.mu.Lock()
		idAtual := gameState.currentQuestionID
		gameState.mu.Unlock()

		msg := map[string]interface{}{
			"tipo":  "resposta",
			"id":    idAtual,
			"opcao": strings.ToUpper(resposta),
		}

		ansBytes, _ := json.Marshal(msg)
		conn.Write(append(ansBytes, '\n'))
	}
}

func main() {
    endereco := client.PerguntaIP()

	conn, err := net.Dial("tcp", endereco)
	if err != nil {
		panic(err)
	}

	defer conn.Close()
	client.MostraMensagem("Conectado ao servidor.")

	// Canal para sinalizar quando a goroutine de leitura da rede terminar
	done := make(chan struct{})

	// Inicia a goroutine para ler do servidor
	go lerServidor(conn, done)

	// Inicia a goroutine para ler do teclado do usuário
	go lerUsuario(conn)

	// A função main vai bloquear aqui. Ela só vai continuar (e o programa encerrar)
	// quando o canal 'done' for fechado pela goroutine lerServidor.
	<-done
}