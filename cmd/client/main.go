package main

import (
	"bufio" //Otimiza a leitura dos dados
	"encoding/json"//Converte de JSON para Go e vice-versa
	"fmt" //Formata textos
	"net" //Permite criar clientes e servidores, lidar com endereços IP e usar protocolos como TCP
	"os" //Uso de I/O 
	"strings" //Funcionalidades para manipular Strings
	"time" //Dá acesso a funções que utilizam tempo
)

//Definição das structs
type Mensagem struct {
	Tipo string `json:"tipo"` //tag para utilizar o encoding/json
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
	leitura := bufio.NewReader(os.Stdin) //Constrói um leitor eficiente de dados com entrada para o teclado
	fmt.Print("IP do servidor (ex: 127.0.0.1:8080): ") 
	endereco, _ := leitura.ReadString('\n') //Lê o endereço digitado 
	endereco = strings.TrimSpace(endereco) //Elimina espaços e '\n' do endereço

	conn, err := net.Dial("tcp", endereco) //disca o endereço fornecido e conecta ao servidor
	if err != nil {
		panic(err)
	}
	defer conn.Close() //Programa encerramento da conexão
	fmt.Println("Conectado ao servidor.")

	var nome string

	for {
		msg, err := bufio.NewReader(conn).ReadBytes('\n') //Lê a mensagem enviada pela rede
		if err != nil {
			fmt.Println("Conexão perdida com o servidor.")
			break
		}

		var mensagemBase Mensagem
		if err := json.Unmarshal(msg, &mensagemBase); err != nil { //Transforma a 'msg' de JSON para Go e a coloca em 'mensagemBase'
			continue
		}

		switch mensagemBase.Tipo {

		//Digitar nome
		case "nome_requisicao":
			fmt.Print("Seu nome: ")
			nome, _ = leitura.ReadString('\n') //Lê o nome digitado pelo usuário
			nome = strings.TrimSpace(nome)
			conn.Write([]byte(nome + "\n")) //Concatena com '\n' e envia o nome para o servidor
			fmt.Println("Aguardando outros jogadores...")
		
		//Receber e responder pergunta
		case "pergunta":
			var pergunta Pergunta
			json.Unmarshal(msg, &pergunta) //Converte a 'msg' para Go e coloca em 'pergunta'

			fmt.Printf("\n📢 Pergunta %d: %s\n", pergunta.ID, pergunta.Texto) //Exibe a pergunta

			// Faz contagem regressiva antes de responder
			for i := 3; i > 0; i-- {
				fmt.Printf("%d, ", i)
				time.Sleep(1 * time.Second)
			}
			fmt.Println("\nAlternativas:")

			for _, opt := range pergunta.Opcoes { //Exibe todas as opções de resposta
				fmt.Println(opt)
			}

			fmt.Print("Sua resposta (A/B/C/D): ")
			resp, _ := leitura.ReadString('\n') //Lê resposta
			resp = strings.TrimSpace(resp)

			resposta := map[string]interface{}{ //Constrói um map para a resposta que pode receber múltiplos valores
				"tipo":   "resposta",
				"id":     pergunta.ID,
				"opcao":  strings.ToUpper(resp), //Coloca a resposta em letra maiúscula
			}
			ansBytes, _ := json.Marshal(resposta) //Transforma a resposta em JSON
			conn.Write(append(ansBytes, '\n')) //Concatena com '\n' e manda para o servidor
		
		//Mostrar placar
		case "placar":
			var placar Placar
			json.Unmarshal(msg, &placar) //Converte 'msg' para Go e coloca em 'placar'

			fmt.Println("\n--- PLACAR ---")
			for _, score := range placar.Pontuacoes { //Exibe cada jogador e sua pontuação
				fmt.Printf("%s: %d pontos\n", score.Player, score.Pontos)
			}
			fmt.Println("--------------")
		
		//Contagem regressiva
		case "contagem_regressiva":
			var data struct {  //struct local
                Valor int `json:"valor"`
            }

            if err := json.Unmarshal(msg, &data); err == nil { //Transforma 'msg' em Go e coloca em 'data'
                if data.Valor > 0 {
                    fmt.Printf("\n⏳%d...", data.Valor) //imprime de valor da contagem for maior que 0
                } else {
                    fmt.Println("\n🚦Vai!") //imprime "Vai" se a contagem chegou a 0
                }
            }
		}
	}
}
