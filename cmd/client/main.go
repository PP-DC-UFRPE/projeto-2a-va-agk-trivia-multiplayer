// entry point do cliente CLI/GUI
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
)

type Question struct {  
	Type    string   `json:"type"`
	ID      int      `json:"id"`
	Text    string   `json:"text"`
	Options []string `json:"options"`
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Seu nome: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	fmt.Print("IP do servidor (ex: 127.0.0.1:8080): ")
	addr, _ := reader.ReadString('\n')
	addr = strings.TrimSpace(addr)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	fmt.Println("Conectado ao servidor.")

	for {
		msg, err := bufio.NewReader(conn).ReadBytes('\n')
		if err != nil {
			break
		}

		var question Question
		if err := json.Unmarshal(msg, &question); err != nil {
			continue
		}

		fmt.Printf("\n📢 Pergunta %d: %s\n", question.ID, question.Text)
		for _, opt := range question.Options {
			fmt.Println(opt)
		}

		fmt.Print("Sua resposta (A/B/C/D): ")
		resp, _ := reader.ReadString('\n')
		resp = strings.TrimSpace(resp)

		answer := map[string]interface{}{
			"type":   "answer",
			"id":     question.ID,
			"player": name,
			"option": strings.ToUpper(resp),
		}

		ansBytes, _ := json.Marshal(answer)
		conn.Write(append(ansBytes, '\n'))
	}
}
