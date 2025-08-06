package models

import "time"

// PerguntaJSON representa a estrutura das perguntas no arquivo JSON
type PerguntaJSON struct {
	Enunciado    string   `json:"enunciado"`
	Alternativas []string `json:"alternativas"`
	Resposta     string   `json:"resposta_correta"`
}

// Pergunta representa uma pergunta do jogo
type Pergunta struct {
	Tipo         string   `json:"tipo"`
	ID           int      `json:"id"`
	Texto        string   `json:"texto"`
	Opcoes       []string `json:"opcoes"`
	OpcaoCorreta string   `json:"-"`
}

// Resposta representa a resposta de um jogador
type Resposta struct {
	Tipo    string    `json:"tipo"`
	ID      int       `json:"id"`
	Jogador string    `json:"jogador"`
	Opcao   string    `json:"opcao"`
	Tempo   time.Time `json:"tempo"`
}

// Pontuacao representa a pontuação de um jogador
type Pontuacao struct {
	Jogador string `json:"jogador"`
	Pontos  int    `json:"pontos"`
}

// Placar representa o placar do jogo
type Placar struct {
	Tipo       string      `json:"tipo"`
	Pontuacoes []Pontuacao `json:"pontuacoes"`
}

// Mensagem genérica para identificar tipos de mensagens
type Mensagem struct {
	Tipo string `json:"tipo"`
}

// ContagemRegressiva representa mensagens de contagem regressiva
type ContagemRegressiva struct {
	Tipo  string `json:"tipo"`
	Valor int    `json:"valor"`
}

// InicioJogo representa o sinal de início do jogo
type InicioJogo struct {
	Tipo string `json:"tipo"`
}

// ServidorLotado representa mensagem quando servidor está cheio
type ServidorLotado struct {
	Tipo string `json:"tipo"`
}

// NomeRequisicao representa solicitação de nome do jogador
type NomeRequisicao struct {
	Tipo string `json:"tipo"`
}
