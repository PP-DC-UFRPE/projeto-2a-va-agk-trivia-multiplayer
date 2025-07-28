// entry point do servidor TCP
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"
)

type Question struct {
	Type    string   `json:"type"`
	ID      int      `json:"id"`
	Text    string   `json:"text"`
	Options []string `json:"options"`
}

type Answer struct {
	Type   string `json:"type"`
	ID     int    `json:"id"`
	Player string `json:"player"`
	Option string `json:"option"`
	Time   time.Time
}

var clients []net.Conn

func handleClient(conn net.Conn) {
	clients = append(clients, conn)
	fmt.Printf("Aguardando jogadores... (%d jogadores conectados)\n", len(clients))
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "localhost"
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.IsLoopback() {
			continue
		}
		if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
			return ipnet.IP.String()
		}
	}
	return "localhost"
}

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Servidor ouvindo na porta em %s:8080\n", getLocalIP())

	go func() {
		for {
			conn, err := ln.Accept()
			if err == nil {
				fmt.Println("Cliente conectado:", conn.RemoteAddr())
				go handleClient(conn)
			}
		}
	}()

	for {
		if len(clients) >= 2 {
			fmt.Println("Jogadores conectados:", len(clients))
			break
		}
		time.Sleep(3 * time.Second)
	}

	questions := []Question{
		{Type: "question", ID: 1, Text: "Qual a capital da França?", Options: []string{"A) Paris", "B) Roma", "C) Berlim", "D) Lisboa"}},
		{Type: "question", ID: 2, Text: "Quanto é 5 + 3?", Options: []string{"A) 6", "B) 7", "C) 8", "D) 9"}},
	}

	reader := bufio.NewReader(nil)
	for _, question := range questions {
		fmt.Println("Enviando pergunta:", question.Text)

		qBytes, _ := json.Marshal(question)
		for _, client := range clients {
			client.Write(append(qBytes, '\n'))
		}

		var respostas []Answer
		deadline := time.Now().Add(10 * time.Second)

		for time.Now().Before(deadline) {
			for _, client := range clients {
				client.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
				reader = bufio.NewReader(client)
				line, err := reader.ReadBytes('\n')
				if err == nil && len(strings.TrimSpace(string(line))) > 0 {
					var a Answer
					if err := json.Unmarshal(line, &a); err == nil && a.Type == "answer" {
						a.Time = time.Now()
						respostas = append(respostas, a)
					}
				}
			}
		}

		fmt.Println("Respostas:")
		for _, r := range respostas {
			fmt.Printf("- %s respondeu %s\n", r.Player, r.Option)
		}
	}
}
