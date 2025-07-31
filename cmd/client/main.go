package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"sync" // Pacote para Mutex
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

// lerServidor lida com todas as mensagens recebidas do servidor
func lerServidor(conn net.Conn, done chan<- struct{}) {
	// Garante que o canal 'done' seja fechado quando a função retornar,
	// sinalizando para a main goroutine que a conexão foi perdida.
	defer close(done)
	reader := bufio.NewReader(conn)

	for {
		msg, err := reader.ReadBytes('\n')
		if err != nil {
			fmt.Println("\nConexão perdida com o servidor. Pressione ENTER para sair.")
			return // Encerra a goroutine
		}

		var mensagemBase Mensagem
		if err := json.Unmarshal(msg, &mensagemBase); err != nil {
			continue
		}

		// Processa a mensagem com base no tipo
		switch mensagemBase.Tipo {
		case "nome_requisicao":
			fmt.Print("Seu nome: ")
			// A outra goroutine vai lidar com o envio do nome

		case "pergunta":
			var pergunta Pergunta
			json.Unmarshal(msg, &pergunta)

			// Trava o mutex para atualizar o ID da pergunta de forma segura
			gameState.mu.Lock()
			gameState.currentQuestionID = pergunta.ID
			gameState.mu.Unlock()

			fmt.Printf("\n\n📢 Pergunta %d: %s\n", pergunta.ID, pergunta.Texto)
			for _, opt := range pergunta.Opcoes {
				fmt.Println(opt)
			}
			fmt.Print("Sua resposta (A/B/C/D): ")

		case "placar":
			var placar Placar
			json.Unmarshal(msg, &placar)
			fmt.Println("\n\n--- PLACAR ---")
			for _, score := range placar.Pontuacoes {
				fmt.Printf("%s: %d pontos\n", score.Player, score.Pontos)
			}
			fmt.Println("--------------")
			fmt.Print("Aguardando próxima pergunta...")


		case "contagem_regressiva":
			var data struct {
				Valor int `json:"valor"`
			}
			if err := json.Unmarshal(msg, &data); err == nil {
				if data.Valor > 0 {
					fmt.Printf("\n⏳%d...", data.Valor)
				} else {
					fmt.Println("\n🚦Vai!")
				}
			}
		
		case "inicio_jogo":
			fmt.Println("\nO JOGO VAI COMEÇAR!")
		}
	}
}

// lerUsuario lida com todo o input do teclado do usuário
func lerUsuario(conn net.Conn) {
	reader := bufio.NewReader(os.Stdin)
	isNameSent := false

	for {
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if !isNameSent {
			// A primeira coisa que o usuário digita é o nome
			conn.Write([]byte(input + "\n"))
			fmt.Println("Aguardando outros jogadores...")
			isNameSent = true
			continue
		}
		
		// Após enviar o nome, todo input é uma resposta de pergunta
		gameState.mu.Lock()
		currentID := gameState.currentQuestionID
		gameState.mu.Unlock()

		resposta := map[string]interface{}{
			"tipo":  "resposta",
			"id":    currentID,
			"opcao": strings.ToUpper(input),
		}
		ansBytes, _ := json.Marshal(resposta)
		conn.Write(append(ansBytes, '\n'))
	}
}

func main() {
	leitor := bufio.NewReader(os.Stdin)
	fmt.Print("IP do servidor (ex: 127.0.0.1:8080): ")
	endereco, _ := leitor.ReadString('\n')
	endereco = strings.TrimSpace(endereco)

	conn, err := net.Dial("tcp", endereco)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	fmt.Println("Conectado ao servidor.")

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