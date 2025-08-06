//  conexões TCP e protocolo JSON

package client

import (
	"bufio"
	"encoding/json"
	"net"
)

// Encapsula a conexão e o writer/reader
type ConexCliente struct {
	Conex  net.Conn
	Leitor *bufio.Reader
}

// Cria uma nova conexão e retorna a struct
func NovaConexCliente(endereco string) (*ConexCliente, error) {
	conex, err := net.Dial("tcp", endereco)
	if err != nil {
		return nil, err
	}
	return &ConexCliente{
		Conex:  conex,
		Leitor: bufio.NewReader(conex),
	}, nil
}

// Envia um objeto como JSON para o servidor
func (c *ConexCliente) EnviarJSON(msg interface{}) error {
	bytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = c.Conex.Write(append(bytes, '\n'))
	return err
}

// Lê uma linha e faz o unmarshal para o objeto passado
func (c *ConexCliente) ReceberJSON(msg interface{}) error { //recebe msg de forma genérica
	linha, err := c.Leitor.ReadBytes('\n')
	if err != nil {
		return err
	}
	return json.Unmarshal(linha, msg)
}

// Fecha a conexão
func (c *ConexCliente) Fechar() error {
	return c.Conex.Close()
}
