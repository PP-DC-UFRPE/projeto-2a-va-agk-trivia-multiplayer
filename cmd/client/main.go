package main

import (
	"bufio"
	"encoding/json"//
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

type Mensagem struct {
	Tipo string `json:"tipo"`
}

type Pergunta struct {
	ID      int      `json:"id"`
	Texto   string   `json:"texto"`
	Opcoes  []string `json:"opcoes"`
}

type Placar struct {
	Pontuacoes []struct {
		Player string `json:"player"`
		Pontos  int    `json:"pontos"`
	} `json:"pontuacoes"`
}

func main() {
	leitura := bufio.NewReader(os.Stdin)
	fmt.Print("IP do servidor (ex: 127.0.0.1:8080): ")
	endereco, _ := leitura.ReadString('\n')
	endereco = strings.TrimSpace(endereco)

	conn, err := net.Dial("tcp", endereco)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	fmt.Println("Conectado ao servidor.")

	var nome string

	for {
		msg, err := bufio.NewReader(conn).ReadBytes('\n')
		if err != nil {
			fmt.Println("Conexão perdida com o servidor.")
			break
		}

		var mensagemBase Mensagem
		if err := json.Unmarshal(msg, &mensagemBase); err != nil {
			continue
		}

		switch mensagemBase.Tipo {
		case "nome_requisicao":
			fmt.Print("Seu nome: ")
			nome, _ = leitura.ReadString('\n')
			nome = strings.TrimSpace(nome)
			conn.Write([]byte(nome + "\n"))
			fmt.Println("Aguardando outros jogadores...")

		case "pergunta":
			var pergunta Pergunta
			json.Unmarshal(msg, &pergunta)

			fmt.Printf("\n📢 Pergunta %d: %s\n", pergunta.ID, pergunta.Texto)

			// contagem regressiva antes de responder
			for i := 3; i > 0; i-- {
				fmt.Printf("%d, ", i)
				time.Sleep(1 * time.Second)
			}
			fmt.Println("\nAlternativas:")

			for _, opt := range pergunta.Opcoes {
				fmt.Println(opt)
			}

			fmt.Print("Sua resposta (A/B/C/D): ")
			resp, _ := leitura.ReadString('\n')
			resp = strings.TrimSpace(resp)

			resposta := map[string]interface{}{
				"tipo":   "resposta",
				"id":     pergunta.ID,
				"opcao":  strings.ToUpper(resp),
			}
			ansBytes, _ := json.Marshal(resposta)
			conn.Write(append(ansBytes, '\n'))

		case "placar":
			var placar Placar
			json.Unmarshal(msg, &placar)

			fmt.Println("\n--- PLACAR ---")
			for _, score := range placar.Pontuacoes {
				fmt.Printf("%s: %d pontos\n", score.Player, score.Pontos)
			}
			fmt.Println("--------------")

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
		}
	}
}
