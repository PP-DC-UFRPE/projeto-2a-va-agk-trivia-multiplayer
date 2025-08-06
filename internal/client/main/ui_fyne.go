// Este ficheiro deve estar em: internal/client/ui_fyne.go
package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"strconv"
	"sync" // Importado para usar o sync.Once
	"time"
	"triviaMultiplayer/internal/client"
	"triviaMultiplayer/internal/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// AppUI gere o estado da nossa interface
type AppUI struct {
	janela  fyne.Window
	conexao *client.ConexCliente
}

// A função main é o ponto de entrada da nossa aplicação gráfica
func main() {
	a := app.New()
	w := a.NewWindow("Trivia Multiplayer")
	w.Resize(fyne.NewSize(400, 350))

	ui := &AppUI{
		janela: w,
	}

	w.SetContent(telaInicial(ui))
	w.ShowAndRun()
}

// É a primeira tela que o jogador vê
func telaInicial(ui *AppUI) fyne.CanvasObject {
	label0 := canvas.NewText("💡 ", color.Gray{})
	label0.TextSize = 20
	label1 := canvas.NewText("TRIVIA MULTIPLAYER", color.RGBA{R: 0, G: 119, B: 190, A: 255})
	label1.TextSize = 20
	label1.TextStyle = fyne.TextStyle{
		Bold: true}
	label2 := canvas.NewText("💡 ", color.Gray{})
	label2.TextSize = 20

	label3 := canvas.NewText("Digite o IP do servidor:", color.RGBA{R: 0, G: 119, B: 190, A: 255})
	entryIP := widget.NewEntry()
	entryIP.SetPlaceHolder("Ex: 127.0.0.1:8080")

	label4 := canvas.NewText("Digite seu nome de jogador:", color.RGBA{R: 0, G: 119, B: 190, A: 255})
	entryNome := widget.NewEntry()
	entryNome.SetPlaceHolder("Ex: Eistein")

	button := widget.NewButton("Entrar", func() {
		ip := entryIP.Text
		nomeJogador := entryNome.Text

		// Inicia a conexão numa goroutine
		go func() {
			conex, err := client.NovaConexCliente(ip)
			if err != nil {
				dialog.ShowError(err, ui.janela)
				return
			}
			ui.conexao = conex

			err = ui.conexao.EnviarJSON(nomeJogador)
			if err != nil {
				dialog.ShowError(err, ui.janela)
				return
			}

			ui.janela.SetContent(telaAguardoInicial(ui)) //se a conexão der certo vai para a tela de espera inicial
			go lerServidorEAtualizarUI(ui)               //inicia a escuta de todas as mensagens do servidor
		}()
	})

	return container.NewCenter(container.NewVBox(
		(container.NewHBox(label0, label1, label2)),
		label3,
		entryIP,
		label4,
		entryNome,
		button,
	))
}

func telaAguardoInicial(ui *AppUI) fyne.CanvasObject {
	label1 := canvas.NewText("O jogo começará em breve", color.RGBA{R: 0, G: 119, B: 190, A: 255})
	label1.TextSize = 20
	label1.TextStyle = fyne.TextStyle{Bold: true}

	label2 := widget.NewLabel("Aguardando os outros jogadores...")
	barraProgresso := widget.NewProgressBarInfinite()

	return container.NewCenter(container.NewVBox(
		label1,
		label2,
		barraProgresso,
	))
}

func telaContagem(ui *AppUI, tempo int) fyne.CanvasObject {
	texto := fmt.Sprintf("O jogo começa em:")
	label := canvas.NewText(texto, color.RGBA{R: 0, G: 119, B: 190, A: 255})
	label.Alignment = fyne.TextAlignCenter
	label.TextSize = 20
	label.TextStyle.Bold = true

	texto2 := fmt.Sprintf("%d", tempo)
	if tempo <= 0 {
		texto2 = "VAI!"
		time.Sleep(1 * time.Second)
	}
	labelCont := canvas.NewText(texto2, color.RGBA{R: 173, G: 216, B: 230, A: 255})
	labelCont.TextStyle.Bold = true
	labelCont.Alignment = fyne.TextAlignCenter
	labelCont.TextSize = 70

	return container.NewCenter(container.NewVBox(label, labelCont))
}

func telaPerguntas(ui *AppUI, pergunta models.Pergunta) fyne.CanvasObject {
	label1 := canvas.NewText(pergunta.Texto, color.Black)

	timerLabel := canvas.NewText("", color.RGBA{R: 0, G: 119, B: 190, A: 255})
	timerLabel.TextStyle = fyne.TextStyle{Bold: true}
	timerLabel.Alignment = fyne.TextAlignCenter

	var botoes []*widget.Button
	var once sync.Once // Garante que a ação de resposta, clique ou tempoEsgotado, só acontece uma vez

	// Ação a ser executada quando um botão de resposta é clicado ou o tempo esgota
	acaoResposta := func(opcao string, tempoEsgotado bool) {
		once.Do(func() {
			for _, b := range botoes {
				b.Disable()
			}
			if tempoEsgotado {
				timerLabel.Text = "Tempo esgotado!"
				// Se o tempo esgotou, o jogador não enviou resposta, então mostramos "Incorreta"
				ui.janela.SetContent(telaResultadoResposta(ui, false))
			} else {
				enviarResposta(ui, pergunta.ID, opcao)
			}
		})
	}

	buttonA := widget.NewButton(pergunta.Opcoes[0], func() { acaoResposta("A", false) })
	buttonB := widget.NewButton(pergunta.Opcoes[1], func() { acaoResposta("B", false) })
	buttonC := widget.NewButton(pergunta.Opcoes[2], func() { acaoResposta("C", false) })
	buttonD := widget.NewButton(pergunta.Opcoes[3], func() { acaoResposta("D", false) })

	botoes = []*widget.Button{buttonA, buttonB, buttonC, buttonD}

	go func(tempoRestante int) {
		for i := tempoRestante; i >= 0; i-- {
			timerLabel.Text = "Tempo restante: " + strconv.Itoa(i) //strconv converte o i (int) em string
			time.Sleep(1 * time.Second)                            //garante duração de tempo correta da função
		}
		// Quando o ciclo termina, o tempo esgotou
		acaoResposta("", true)
	}(10) //o parâmetro de tempo da goroutine é 10 segundos

	return container.NewVBox(
		label1,
		widget.NewSeparator(),
		timerLabel,
		widget.NewSeparator(),
		buttonA,
		buttonB,
		buttonC,
		buttonD,
	)
}

// Mostra o resultado da resposta do jogador
func telaResultadoResposta(ui *AppUI, correta bool) fyne.CanvasObject {
	var textoResultado string
	var corTexto color.Color

	textoResultado = "Resposta Correta!"
	corTexto = color.NRGBA{R: 0, G: 180, B: 0, A: 255}

	if !correta {
		textoResultado = "Resposta Incorreta!"
		corTexto = color.NRGBA{R: 200, G: 0, B: 0, A: 255}
	}

	labelResultado := canvas.NewText(textoResultado, corTexto)
	labelResultado.TextSize = 24
	labelResultado.TextStyle = fyne.TextStyle{Bold: true}
	labelResultado.Alignment = fyne.TextAlignCenter

	label2 := widget.NewLabel("Esperando os outros jogadores responderem...")
	label2.Alignment = fyne.TextAlignCenter

	return container.NewCenter(container.NewVBox(
		labelResultado,
		label2))
}

func telaPlacar(ui *AppUI, placar models.Placar, fim_de_jogo bool) fyne.CanvasObject {
	titulo := canvas.NewText("Placar Parcial", color.RGBA{R: 0, G: 119, B: 190, A: 255})
	titulo.Alignment = fyne.TextAlignCenter
	titulo.TextStyle.Bold = true
	titulo.TextSize = 30

	if fim_de_jogo {
		titulo.Text = "Placar Final"
	}

	listaDePontos := container.NewVBox()
	listaDePontos.Add(titulo)

	for _, p := range placar.Pontuacoes {
		nomeJogador := canvas.NewText(fmt.Sprintf("%s: ", p.Jogador), color.NRGBA{R: 128, G: 0, B: 128, A: 255})     // roxo
		pontosJogador := canvas.NewText(fmt.Sprintf("%d pontos", p.Pontos), color.NRGBA{R: 0, G: 200, B: 0, A: 255}) // Verde
		linhaDePonto := container.NewHBox(nomeJogador, pontosJogador)
		listaDePontos.Add(linhaDePonto)
	}

	if !fim_de_jogo {
		barraProgresso := widget.NewProgressBar()
		go func() {
			duracaoTotal := 5.0
			for i := 0.0; i <= duracaoTotal; i += 0.1 {

				valor := i / duracaoTotal
				barraProgresso.SetValue(valor)
				time.Sleep(100 * time.Millisecond)
			}
		}()

		conteudoInferior := container.NewVBox(
			widget.NewLabel("Carregando próxima pergunta..."),
			barraProgresso,
		)

		return container.NewBorder(nil, conteudoInferior, nil, nil, listaDePontos)

	} else {
		botaoJogarDenovo := widget.NewButton("Jogar Novamente", func() {
			ui.janela.SetContent(telaAguardoInicial(ui))

			// Reiniciamos a goroutine que ouve o servidor para a nova partida.
			go lerServidorEAtualizarUI(ui)
		})

		botaoSair := widget.NewButton("Sair do Jogo", func() {
			// Fecha a aplicação do cliente
			ui.janela.Close()
		})

		botoesFinais := container.NewHBox(botaoJogarDenovo, botaoSair)

		listaDePontos.Add(widget.NewSeparator())
		listaDePontos.Add(container.NewCenter(botoesFinais))
		return listaDePontos
	}
}

// Envia a resposta do jogador para o servidor
func enviarResposta(ui *AppUI, idPergunta int, opcao string) {
	msg := map[string]interface{}{
		"tipo":  "resposta",
		"id":    idPergunta,
		"opcao": opcao,
	}
	err := ui.conexao.EnviarJSON(msg)
	if err != nil {
		dialog.ShowError(err, ui.janela)
	}
}

// É a função principal que ouve o servidor e muda as telas
func lerServidorEAtualizarUI(ui *AppUI) {
	for {
		var rawMsg map[string]interface{}
		err := ui.conexao.ReceberJSON(&rawMsg) //recebe a msg de forma genérica em um map
		if err != nil {
			dialog.ShowError(fmt.Errorf("Ligação perdida: %w", err), ui.janela)
			ui.janela.SetContent(telaInicial(ui))
			return
		}

		tipo, _ := rawMsg["tipo"].(string)
		switch tipo {
		case "pergunta":
			var pergunta models.Pergunta
			bytes, _ := json.Marshal(rawMsg)     //Não dá para usar a msg genérica, então colocamos a msg em formato de JSON novamente
			_ = json.Unmarshal(bytes, &pergunta) //pegamos a msg em JSON e convertemos para o formato específico de pergunta que é uma struct
			ui.janela.SetContent(telaPerguntas(ui, pergunta))

		case "resultado_resposta":
			var data struct {
				Correta bool `json:"correta"`
			}
			bytes, _ := json.Marshal(rawMsg)
			_ = json.Unmarshal(bytes, &data)
			ui.janela.SetContent(telaResultadoResposta(ui, data.Correta))

		case "placar":
			var placar models.Placar
			bytes, _ := json.Marshal(rawMsg)
			_ = json.Unmarshal(bytes, &placar)
			ui.janela.SetContent(telaPlacar(ui, placar, false))
		case "contagem_regressiva":
			var data struct {
				Valor int `json:"valor"`
			}
			bytes, _ := json.Marshal(rawMsg)
			_ = json.Unmarshal(bytes, &data)
			ui.janela.SetContent(telaContagem(ui, data.Valor))
		case "fim_de_jogo":
			var placar models.Placar
			bytes, _ := json.Marshal(rawMsg)
			_ = json.Unmarshal(bytes, &placar)
			ui.janela.SetContent(telaPlacar(ui, placar, true))
			return
		}
	}
}
