package client

import (
	"fmt"
)

/*
// PerguntaIP pergunta o IP do servidor ao usuário
func PerguntaIP() string {
    leitor := bufio.NewReader(os.Stdin)
    fmt.Print("IP do servidor (ex: 127.0.0.1:8080): ")
    endereco, _ := leitor.ReadString('\n')
    return strings.TrimSpace(endereco)
}

func PerguntaNome() string {
    leitor := bufio.NewReader(os.Stdin)
    fmt.Print("Seu nome: ")
    nome, _ := leitor.ReadString('\n')
    return strings.TrimSpace(nome)
}

func PerguntaResposta() string {
    leitor := bufio.NewReader(os.Stdin)
    fmt.Print("Sua resposta (A/B/C/D): ")
    resp, _ := leitor.ReadString('\n')
    return strings.TrimSpace(resp)
}
*/
// MostraMensagem imprime uma mensagem no terminal
func MostraMensagem(msg string) {
	fmt.Println(msg)
}
