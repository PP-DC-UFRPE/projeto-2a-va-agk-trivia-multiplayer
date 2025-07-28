//  conexões TCP e protocolo JSON

package client

import (
    "bufio"
    "fmt"
    "net"
    "strings"
    "time"
)

type Client struct {
    conn   net.Conn
    reader *bufio.Reader
}

func NewClient(address string) (*Client, error) {
    conn, err := net.Dial("tcp", address)
    if err != nil {
        return nil, fmt.Errorf("erro ao conectar: %v", err)
    }
    
    return &Client{
        conn:   conn,
        reader: bufio.NewReader(conn),
    }, nil
}

func (c *Client) Close() error {
    return c.conn.Close()
}

func (c *Client) SendMessage(message string) error {
    _, err := c.conn.Write([]byte(message + "\n"))
    return err
}

func (c *Client) ReadMessage() (string, error) {
    message, err := c.reader.ReadString('\n')
    if err != nil {
        return "", err
    }
    return strings.TrimSpace(message), nil
}

func (c *Client) PingPong() error {
    // Envia ping
    if err := c.SendMessage("ping"); err != nil {
        return fmt.Errorf("erro ao enviar ping: %v", err)
    }
    
    // Lê resposta
    response, err := c.ReadMessage()
    if err != nil {
        return fmt.Errorf("erro ao ler resposta: %v", err)
    }
    
    if response != "pong" {
        return fmt.Errorf("resposta inesperada: %s", response)
    }
    
    return nil
}

func (c *Client) StartPingPongLoop(interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    
    for range ticker.C {
        if err := c.PingPong(); err != nil {
            fmt.Printf("Erro no ping/pong: %v\n", err)
            break
        }
        fmt.Println("Ping/Pong bem-sucedido")
    }
}