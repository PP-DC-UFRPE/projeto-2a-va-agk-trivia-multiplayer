//  conexões TCP e protocolo JSON

package client

import (
    "bufio"
    "encoding/json"
    "net"
)

// ConnClient encapsula a conexão e o writer/reader
type ConnClient struct {
    Conn   net.Conn
    Reader *bufio.Reader
}

// NovaConnClient cria uma nova conexão e retorna a struct
func NovaConnClient(endereco string) (*ConnClient, error) {
    conn, err := net.Dial("tcp", endereco)
    if err != nil {
        return nil, err
    }
    return &ConnClient{
        Conn:   conn,
        Reader: bufio.NewReader(conn),
    }, nil
}

// EnviaJSON envia um objeto como JSON para o servidor
func (c *ConnClient) EnviarJSON(msg interface{}) error {
    bytes, err := json.Marshal(msg)
    if err != nil {
        return err
    }
    _, err = c.Conn.Write(append(bytes, '\n'))
    return err
}

// RecebeJSON lê uma linha e faz o unmarshal para o objeto passado
func (c *ConnClient) ReceberJSON(dest interface{}) error {
    line, err := c.Reader.ReadBytes('\n')
    if err != nil {
        return err
    }
    return json.Unmarshal(line, dest)
}

// Fecha a conexão
func (c *ConnClient) Fechar() error {
    return c.Conn.Close()
}


