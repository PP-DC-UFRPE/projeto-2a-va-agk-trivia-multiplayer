package server

import (
	"sort" //Funcionalidades de ordenação de dados
	"strings"
	"triviaMultiplayer/internal/models"
)

// CalcularPontos calcula a pontuação com base nas respostas
// O primeiro a acertar ganha 100, e cada subsequente ganha metade da pontuação anterior
func CalcularPontos(respostas []models.Resposta, opcaoCorreta string) []models.Pontuacao {
	var respostasCorretas []models.Resposta
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

	var listaPontos []models.Pontuacao
	pontosAtuais := 100 // Pontuação inicial

	for _, resp := range respostasCorretas {
		listaPontos = append(listaPontos, models.Pontuacao{Jogador: resp.Jogador, Pontos: pontosAtuais})
		pontosAtuais /= 2 // Metade para o próximo
	}

	return listaPontos
}
