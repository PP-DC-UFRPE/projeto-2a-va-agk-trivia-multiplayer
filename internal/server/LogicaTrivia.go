package server

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

type Pergunta struct {
	Enunciado       string   `json:"enunciado"`
	Alternativas    []string `json:"alternativas"`
	RespostaCorreta string   `json:"resposta_correta"`
}

type Resposta struct {
	Player string
	Opcao  string
	Tempo  time.Time
}
type Pontuacao struct {
	Player string
	Pontos int
}
type Jogador struct {
	Nome      string
	Pontuacao int
}

type Jogo struct {
	Perguntas   []Pergunta
	Jogadores   map[string]*Jogador
	RodadaAtual int
}

// funcao NovoJogo() *Jogo

func NovoJogo() *Jogo {
	return &Jogo{
		Jogadores:   make(map[string]*Jogador),
		RodadaAtual: -1,
	}
}

// CarregarPerguntas lê o arquivo JSON e popula a lista de perguntas do jogo.

func (j *Jogo) CarregarPerguntas(arquivo string) error {
	file, err := os.ReadFile(arquivo)
	if err != nil {
		return fmt.Errorf("erro ao ler o arquivo de perguntas: %w", err)
	}

	err = json.Unmarshal(file, &j.Perguntas)
	if err != nil {
		return fmt.Errorf("erro ao decodificar o JSON de perguntas: %w", err)
	}

	fmt.Println("Perguntas carregadas com sucesso!")
	return nil
}

// AdicionarJogador registra um novo jogador na partida.
func (j *Jogo) AdicionarJogador(nome string) {
	if _, existe := j.Jogadores[nome]; !existe {
		j.Jogadores[nome] = &Jogador{Nome: nome, Pontuacao: 0}
		fmt.Printf("Jogador %s entrou no jogo!\n", nome)
	}
}

// ObterPlacar retorna uma string formatada com o ranking dos jogadores.
func (j *Jogo) ObterPlacar() string {
	// 1. Copiar os jogadores para uma lista para poder ordenar
	listaJogadores := make([]*Jogador, 0, len(j.Jogadores))
	for _, jogador := range j.Jogadores {
		listaJogadores = append(listaJogadores, jogador)
	}

	// 2. Ordenar a lista por pontuação (do maior para o menor)
	sort.Slice(listaJogadores, func(i, j int) bool {
		return listaJogadores[i].Pontuacao > listaJogadores[j].Pontuacao
	})

	// 3. Montar a string do placar
	var placar strings.Builder
	placar.WriteString("\n--- PLACAR ATUAL ---\n")
	for i, jogador := range listaJogadores {
		placar.WriteString(fmt.Sprintf("%d. %s - %d pontos\n", i+1, jogador.Nome, jogador.Pontuacao))
	}
	placar.WriteString("--------------------\n")

	return placar.String()
}

// CalcularPontos calcula a pontuação com base nas respostas.
// A primeira resposta correta ganha 100 pontos, a segunda 50.
func CalcularPontos(respostas []Resposta, opcaoCorreta string) []Pontuacao {
	var respostaCorreta []Resposta
	for _, resp := range respostas {
		if resp.Opcao == opcaoCorreta {
			respostaCorreta = append(respostaCorreta, resp)
		}
	}

	// Ordena as respostas corretas pelo tempo de resposta
	sort.Slice(respostaCorreta, func(i, j int) bool {
		return respostaCorreta[i].Tempo.Before(respostaCorreta[j].Tempo)
	})

	pontos := make(map[string]int)
	if len(respostaCorreta) > 0 {
		pontos[respostaCorreta[0].Player] = 100 // Primeiro a responder
	}
	if len(respostaCorreta) > 1 {
		pontos[respostaCorreta[1].Player] = 50 // Segundo a responder
	}

	var listaPontos []Pontuacao
	for player, pontos := range pontos {
		listaPontos = append(listaPontos, Pontuacao{Player: player, Pontos: pontos})
	}

	return listaPontos
}

func (j *Jogo) ProcessarRespostasDaRodada(respostas []Resposta) {
	if j.RodadaAtual < 0 || j.RodadaAtual >= len(j.Perguntas) {
		return // Rodada inválida
	}
	perguntaAtual := j.Perguntas[j.RodadaAtual]
	pontosDaRodada := CalcularPontos(respostas, perguntaAtual.RespostaCorreta)

	for _, pontuacao := range pontosDaRodada {
		if jogador, existe := j.Jogadores[pontuacao.Player]; existe {
			jogador.Pontuacao += pontuacao.Pontos // Atualiza a pontuação TOTAL
		}
	}
}

func (j *Jogo) ProximaRodada() (Pergunta, bool) {
	j.RodadaAtual++
	if j.RodadaAtual >= len(j.Perguntas) {
		return Pergunta{}, false // Jogo acabou
	}
	return j.Perguntas[j.RodadaAtual], true // Jogo continua
}
