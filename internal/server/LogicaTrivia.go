package server

import (
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