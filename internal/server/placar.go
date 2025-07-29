package server

import (
	"sort"
	"time"
)

type Resposta struct {
	Player string
	Opcao string
	Tempo   time.Time
}

type Pontuacao struct {
	Player string
	Pontos int
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