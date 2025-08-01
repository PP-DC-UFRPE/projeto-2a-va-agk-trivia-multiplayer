package main

import (
	"encoding/json"
	"fmt"
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
func lerServidor(conn *client.ConnClient, done chan<- struct{}) {
    defer close(done)

    for {
        // Lê a mensagem bruta
        var rawMsg map[string]interface{}
        err := conn.ReceberJSON(&rawMsg)
        if err != nil {
            client.MostraMensagem("\nConexão perdida com o servidor. Pressione ENTER para sair.")
            return
        }

        tipo, _ := rawMsg["tipo"].(string)
        switch tipo {
        case "nome_requisicao":
            // A outra goroutine vai lidar com o envio do nome

        case "pergunta":
            // Converte o rawMsg para JSON e depois para Pergunta
            bytes, _ := json.Marshal(rawMsg)
            var pergunta Pergunta
            _ = json.Unmarshal(bytes, &pergunta)

            gameState.mu.Lock()
            gameState.currentQuestionID = pergunta.ID
            gameState.mu.Unlock()

            client.MostraMensagem(fmt.Sprintf("📢 Pergunta %d: %s\n", pergunta.ID, pergunta.Texto))
            for _, opt := range pergunta.Opcoes {
                client.MostraMensagem(opt)
            }
            perguntaDisponivel <- struct{}{}

        case "placar":
            bytes, _ := json.Marshal(rawMsg)
            var placar Placar
            _ = json.Unmarshal(bytes, &placar)
            client.MostraMensagem("\n\n--- PLACAR ---")
            for _, score := range placar.Pontuacoes {
                client.MostraMensagem(fmt.Sprintf("%s: %d pontos", score.Player, score.Pontos))
            }
            client.MostraMensagem("--------------")
            client.MostraMensagem("Aguardando próxima pergunta...")

        case "contagem_regressiva":
            bytes, _ := json.Marshal(rawMsg)
            var data struct {
                Valor int `json:"valor"`
            }
            _ = json.Unmarshal(bytes, &data)
            if data.Valor > 0 {
                client.MostraMensagem(fmt.Sprintf("\n⏳%d...", data.Valor))
            } else {
                client.MostraMensagem("\n🚦Vai!")
            }

        case "inicio_jogo":
            client.MostraMensagem("\nO JOGO VAI COMEÇAR!")
        }
    }
}

// lerUsuario lida com todo o input do teclado do usuário
func lerUsuario(conn *client.ConnClient) {
	isNameSent := false

	for {
		if !isNameSent {
			nome := client.PerguntaNome()
			conn.EnviarJSON(nome)
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

		conn.EnviarJSON(msg)
	}
}

func main() {
    endereco := client.PerguntaIP()

	connClient, err := client.NovaConnClient(endereco)
	if err != nil {
		panic(err)
	}

	defer connClient.Fechar()
	client.MostraMensagem("Conectado ao servidor.")

	// Canal para sinalizar quando a goroutine de leitura da rede terminar
	done := make(chan struct{})

	// Inicia a goroutine para ler do servidor
	go lerServidor(connClient, done)

	// Inicia a goroutine para ler do teclado do usuário
	go lerUsuario(connClient)

	// A função main vai bloquear aqui. Ela só vai continuar (e o programa encerrar)
	// quando o canal 'done' for fechado pela goroutine lerServidor.
	<-done
}