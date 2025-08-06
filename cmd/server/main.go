package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"
	"triviaMultiplayer/internal/models"
	"triviaMultiplayer/internal/server"
)

// Constante para o número máximo de jogadores
const maxJogadores = 10

// Carrega perguntas do arquivo JSON
func carregarPerguntasDoArquivo(caminho string, limite int) ([]models.Pergunta, error) {
	//Lê o arquivo JSON
	arquivoBytes, err := os.ReadFile(caminho)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler o arquivo: %w", err)
	}

	// Decodifica o JSON para a struct
	var perguntasJSON []models.PerguntaJSON
	err = json.Unmarshal(arquivoBytes, &perguntasJSON)
	if err != nil {
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
	var perguntasJogo []models.Pergunta
	for i, pJSON := range perguntasJSON {
		perguntasJogo = append(perguntasJogo, models.Pergunta{
			Tipo:         "pergunta",
			ID:           i + 1, // ID sequencial
			Texto:        pJSON.Enunciado,
			Opcoes:       pJSON.Alternativas,
			OpcaoCorreta: pJSON.Resposta,
		})
	}

	return perguntasJogo, nil
}

// Faz contagem regressiva inicial
func contagemRegressiva(servidor *server.ServerJogo, valor int) {
	for i := valor; i > 0; i-- {
		servidor.TransmitirMsg([]byte(fmt.Sprintf("{\"tipo\":\"contagem_regressiva\",\"valor\":%d}\n", i)))
		fmt.Printf("Começando em %d...\n", i)
		time.Sleep(1 * time.Second)
	}
	servidor.TransmitirMsg([]byte("{\"tipo\":\"contagem_regressiva\",\"valor\":0}\n"))
}

// Envia placar
func enviarPlacar(servidor *server.ServerJogo, tipoMsg string) {
	jogadores := servidor.RetornarJogadores()
	var pontuacaoAtual []models.Pontuacao

	for _, jogador := range jogadores {
		pontuacaoAtual = append(pontuacaoAtual, models.Pontuacao{
			Jogador: jogador.Nome,
			Pontos:  jogador.Pontuacao,
		})
	}

	// Usa o tipo de mensagem que foi passado como argumento.
	placar := models.Placar{Tipo: tipoMsg, Pontuacoes: pontuacaoAtual}
	placarBytes, _ := json.Marshal(placar)
	servidor.TransmitirMsg(append(placarBytes, '\n'))
}

func executarJogo(servidor *server.ServerJogo, perguntas []models.Pergunta) {
	// Contagem regressiva
	contagemRegressiva(servidor, 3)

	// Executa cada pergunta
	for i, pergunta := range perguntas {
		perguntaBytes, _ := json.Marshal(pergunta)
		servidor.TransmitirMsg(append(perguntaBytes, '\n'))

		respostas := servidor.ColetarRespostas(10*time.Second, pergunta.OpcaoCorreta)

		pontos := server.CalcularPontos(respostas, pergunta.OpcaoCorreta)
		for _, ponto := range pontos {
			servidor.AtualizarPontos(ponto.Jogador, ponto.Pontos)
		}

		time.Sleep(5 * time.Second) // Pausa para a tela de Resultado da resposta

		//Verifica se não é a última pergunta antes de enviar o placar parcial
		ultimaPergunta := (i == len(perguntas)-1)
		if !ultimaPergunta {
			enviarPlacar(servidor, "placar")
			time.Sleep(5 * time.Second) // Pausa apenas se não for a última pergunta
		}
	}

	fmt.Println("Fim de jogo!")
	// Envia o placar final com o tipo "fim_de_jogo".
	enviarPlacar(servidor, "fim_de_jogo")

	time.Sleep(1 * time.Second)
}

func main() {
	servidor := server.NovoServer(maxJogadores)
	err := servidor.IniciarServer(":8080") //iniciando o servidor
	if err != nil {
		panic(err)
	}
	defer servidor.Parar()

	fmt.Printf("Servidor ouvindo em %s:8080\n", server.ObterIPlocal())

	for {
		fmt.Printf("\n----------------------------------\n")
		fmt.Printf("Aguardando jogadores... (%d/%d)\n", servidor.RetornarNumJogadores(), maxJogadores)
		fmt.Println("Pressione ENTER a qualquer momento para iniciar a partida com os jogadores conectados.")

		bufio.NewReader(os.Stdin).ReadBytes('\n')

		numPlayers := servidor.RetornarNumJogadores()
		if numPlayers == 0 {
			fmt.Println("Nenhum jogador conectado. Aguardando novamente...")
			continue // Volta para o início do ciclo.
		}

		fmt.Printf("\nO jogo vai começar com %d jogador(es)!\n", numPlayers)
		servidor.TransmitirMsg([]byte("{\"tipo\":\"inicio_jogo\"}\n"))

		perguntas, err := carregarPerguntasDoArquivo("perguntas.json", 5) //carrega 5 perguntas do perguntas.json
		if err != nil {
			fmt.Printf("Erro ao carregar perguntas: %v\n", err)
			return
		}

		executarJogo(servidor, perguntas)

		fmt.Println("Partida finalizada. O servidor está pronto para uma nova rodada.")
	}
}
