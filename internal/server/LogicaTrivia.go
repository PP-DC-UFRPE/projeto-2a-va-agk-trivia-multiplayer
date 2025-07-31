package server

import (
	"encoding/json"
	"fmt"
	"os" //Gerencia arquivos
	"sort" //Funcionalidades de ordenação de dados
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

func NovoJogo() *Jogo {
	return &Jogo{
		Jogadores:   make(map[string]*Jogador),
		RodadaAtual: -1,
	}
}

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

func (j *Jogo) AdicionarJogador(nome string) {
	if _, existe := j.Jogadores[nome]; !existe {
		j.Jogadores[nome] = &Jogador{Nome: nome, Pontuacao: 0}
		fmt.Printf("Jogador %s entrou no jogo!\n", nome)
	}
}

func (j *Jogo) ObterPlacar() string {
	listaJogadores := make([]*Jogador, 0, len(j.Jogadores))
	for _, jogador := range j.Jogadores {
		listaJogadores = append(listaJogadores, jogador)
	}

	sort.Slice(listaJogadores, func(i, j int) bool {
		return listaJogadores[i].Pontuacao > listaJogadores[j].Pontuacao
	})

	var placar strings.Builder
	placar.WriteString("\n--- PLACAR ATUAL ---\n")
	for i, jogador := range listaJogadores {
		placar.WriteString(fmt.Sprintf("%d. %s - %d pontos\n", i+1, jogador.Nome, jogador.Pontuacao))
	}
	placar.WriteString("--------------------\n")

	return placar.String()
}

// CalcularPontos calcula a pontuação com base nas respostas.
// O primeiro a acertar ganha 100, e cada subsequente ganha metade da pontuação anterior.
func CalcularPontos(respostas []Resposta, opcaoCorreta string) []Pontuacao {
	var respostasCorretas []Resposta
	for _, resp := range respostas {
		// A opção vem como "A) Paris", então pegamos apenas a letra.
		if strings.HasPrefix(resp.Opcao, opcaoCorreta) {
			respostasCorretas = append(respostasCorretas, resp)
		}
	}

	// Ordena as respostas corretas pelo tempo
	sort.Slice(respostasCorretas, func(i, j int) bool {
		return respostasCorretas[i].Tempo.Before(respostasCorretas[j].Tempo)
	})

	var listaPontos []Pontuacao
	pontosAtuais := 100 // Pontuação inicial

	for _, resp := range respostasCorretas {
		listaPontos = append(listaPontos, Pontuacao{Player: resp.Player, Pontos: pontosAtuais})
		pontosAtuais /= 2 // Metade para o próximo
	}

	return listaPontos
}

func (j *Jogo) ProcessarRespostasDaRodada(respostas []Resposta) {
	if j.RodadaAtual < 0 || j.RodadaAtual >= len(j.Perguntas) {
		return
	}
	perguntaAtual := j.Perguntas[j.RodadaAtual]
	pontosDaRodada := CalcularPontos(respostas, perguntaAtual.RespostaCorreta)

	for _, pontuacao := range pontosDaRodada {
		if jogador, existe := j.Jogadores[pontuacao.Player]; existe {
			jogador.Pontuacao += pontuacao.Pontos
		}
	}
}

func (j *Jogo) ProximaRodada() (Pergunta, bool) {
	j.RodadaAtual++
	if j.RodadaAtual >= len(j.Perguntas) {
		return Pergunta{}, false
	}
	return j.Perguntas[j.RodadaAtual], true
}