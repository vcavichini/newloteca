
package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// Estruturas para decodificar o JSON da API
type RateioPremio struct {
	DescricaoFaixa     string  `json:"descricaoFaixa"`
	NumeroDeGanhadores int     `json:"numeroDeGanhadores"`
	ValorPremio        float64 `json:"valorPremio"`
}

type LoteriaDados struct {
	Numero              int            `json:"numero"`
	DataApuracao        string         `json:"dataApuracao"`
	ListaDezenas        []string       `json:"listaDezenas"`
	ListaRateioPremio   []RateioPremio `json:"listaRateioPremio"`
	Acumulado           bool           `json:"acumulado"`
	DataProximoConcurso *string        `json:"dataProximoConcurso"`
	ValorEstimadoProximo float64        `json:"valorEstimadoProximoConcurso"`
}

// Meus jogos (hardcoded como no original)
var meusJogos = [][]string{
	{"03", "15", "18", "23", "40", "54"},
	{"01", "02", "12", "29", "46", "51"},
	{"04", "08", "22", "37", "41", "56"},
	{"02", "04", "07", "14", "25", "33"},
	{"04", "09", "13", "24", "30", "55"},
}

// Dados para passar ao template HTML
type TemplateData struct {
	DadosLoteria      LoteriaDados
	MeusJogos         [][]string
	MyNumbersSet      map[string]bool
	Erro              string
	NumeroAtual       int
	LatestContestNumber int
}

// Funções de ajuda para o template
var funcMap = template.FuncMap{
	"compareGames": func(sorteado []string, jogo []string) int {
		acertos := 0
		sorteadoMap := make(map[string]bool)
		for _, num := range sorteado {
			sorteadoMap[num] = true
		}
		for _, num := range jogo {
			if sorteadoMap[num] {
				acertos++
			}
		}
		return acertos
	},
	"isSorteado": func(sorteado []string, num string) bool {
		for _, n := range sorteado {
			if n == num {
				return true
			}
		}
		return false
	},
	"isInSet": func(set map[string]bool, key string) bool {
		return set[key]
	},
	"formatMoney": func(valor float64) string {
		p := message.NewPrinter(language.BrazilianPortuguese)
		return p.Sprintf("R$ %.2f", valor)
	},
	"add": func(a, b int) int {
		return a + b
	},
	"default": func(defaultValue, value interface{}) interface{} {
		if s, ok := value.(string); ok && s != "" {
			return s
		}
		if i, ok := value.(int); ok && i != 0 {
			return i
		}
		return defaultValue
	},
	"gt": func(a, b int) bool { // greater than
		return a > b
	},
	"lt": func(a, b int) bool { // less than
		return a < b
	},
}

// O template HTML principal
var tpl = template.Must(template.New("web").Funcs(funcMap).Parse(`
<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <title>Mega-Sena</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <style>
        body { background-color: #f8f9fa; }
        .badge { font-size: 1.1rem; }
        .card { box-shadow: 0 4px 8px rgba(0,0,0,0.1); }
        .dot {
          height: 8px;
          width: 8px;
          background-color: #198754; /* Bootstrap's success green */
          border-radius: 50%;
          display: block; /* Use block to position it below */
          margin: 4px auto 0; /* Center the dot */
        }
    </style>
</head>
<body>
    <div class="container mt-4">
        {{if .Erro}}
        <div class="alert alert-danger text-center" role="alert">
          <h2>🚨 Erro ao Carregar Dados da Loteria</h2>
          <p>{{.Erro}}</p>
          {{if ne .NumeroAtual 0}}
          <a href="/?concurso={{.NumeroAtual}}" class="btn btn-primary mt-3">Voltar para Sorteio Anterior</a>
          {{end}}
        </div>
        {{else}}
        <div class="card">
            <div class="card-body">
                <p class="text-center mb-2">
                    <strong>Concurso:</strong> {{.DadosLoteria.Numero}}
                </p>
                <p class="text-center mb-4">
                    {{.DadosLoteria.DataApuracao}}
                </p>

                <div class="mb-4">
                    <h3 class="text-center mb-3">Números Sorteados</h3>
                    <div class="d-flex justify-content-center gap-2 flex-wrap">
                        {{range .DadosLoteria.ListaDezenas}}
                        <div class="text-center">
                          <span class="badge bg-success p-2 fs-5">{{.}}</span>
                          {{if isInSet $.MyNumbersSet .}}
                            <span class="dot"></span>
                          {{else}}
                            <span class="dot" style="background-color: transparent;"></span>
                          {{end}}
                        </div>
                        {{end}}
                    </div>
                </div>

                <div class="mt-4">
                    <h3 class="text-center mb-3">Seus Jogos</h3>
                    {{range .MeusJogos}}
                        {{$acertos := compareGames $.DadosLoteria.ListaDezenas .}}
                        <div class="mb-3 p-2 rounded {{if ge $acertos 4}}bg-success bg-opacity-10{{end}}">
                            <div class="d-flex justify-content-center gap-2 flex-wrap">
                                {{range .}}
                                <span class="badge p-2 fs-5 {{if isSorteado $.DadosLoteria.ListaDezenas .}}bg-success{{else}}text-dark bg-secondary bg-opacity-10{{end}}">
                                    {{.}}
                                </span>
                                {{end}}
                            </div>
                        </div>
                    {{end}}
                </div>

                <div class="d-flex justify-content-center gap-3 mt-4">
                    {{if gt .DadosLoteria.Numero 1}}
                    <a href="/?concurso={{.DadosLoteria.Numero | add -1}}" class="btn btn-primary">{{.DadosLoteria.Numero | add -1}}</a>
                    {{end}}
                    <a href="/" class="btn btn-secondary">Mais Recente</a>
                    {{if lt .DadosLoteria.Numero .LatestContestNumber}}
                    <a href="/?concurso={{.DadosLoteria.Numero | add 1}}" class="btn btn-primary">{{.DadosLoteria.Numero | add 1}}</a>
                    {{end}}
                </div>
            </div>
        </div>

        {{if .DadosLoteria.ListaRateioPremio}}
        <div class="card mt-4">
            <div class="card-body">
                <h3 class="text-center mb-3">Distribuição de Prêmios</h3>
                <table class="table table-striped table-hover">
                    <thead class="table-dark">
                    <tr>
                        <th>Acertos</th>
                        <th>Ganhadores</th>
                        <th>Valor</th>
                    </tr>
                    </thead>
                    <tbody>
                    {{range .DadosLoteria.ListaRateioPremio}}
                    <tr>
                        <td>{{.DescricaoFaixa | default "N/A"}}</td>
                        <td>{{.NumeroDeGanhadores | default "N/A"}}</td>
                        <td>{{formatMoney .ValorPremio}}</td>
                    </tr>
                    {{end}}
                    </tbody>
                </table>
            </div>
        </div>
        {{end}}
        {{end}}
    </div>
    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/js/bootstrap.bundle.min.js"></script>
</body>
</html>
`))

// fetchContestData busca os dados de um concurso, com fallback.
// Se contestNumber for "", busca o último.
func fetchContestData(contestNumber string) (LoteriaDados, error) {
	var primaryURL, fallbackURL string

	if contestNumber == "" {
		primaryURL = "https://api.guidi.dev.br/loteria/megasena/ultimo"
		fallbackURL = "https://servicebus2.caixa.gov.br/portaldeloterias/api/megasena/"
	} else {
		primaryURL = "https://api.guidi.dev.br/loteria/megasena/" + contestNumber
		fallbackURL = "https://servicebus2.caixa.gov.br/portaldeloterias/api/megasena/" + contestNumber
	}

	client := http.Client{Timeout: 4 * time.Second}
	var dados LoteriaDados

	// Tenta a API primária
	log.Printf("Tentando API primária: %s", primaryURL)
	resp, err := client.Get(primaryURL)
	if err == nil {
		defer resp.Body.Close()
		if err := json.NewDecoder(resp.Body).Decode(&dados); err == nil && dados.Numero != 0 {
			log.Println("Sucesso na API primária.")
			return dados, nil
		}
	}

	// Se a primária falhar, tenta a de fallback
	log.Printf("API primária falhou (erro: %v). Tentando API de fallback: %s", err, fallbackURL)
	resp, err = client.Get(fallbackURL)
	if err != nil {
		return LoteriaDados{}, fmt.Errorf("ambas as APIs falharam. Erro final: %w", err)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&dados); err != nil || dados.Numero == 0 {
		return LoteriaDados{}, fmt.Errorf("API de fallback também falhou ou retornou dados inválidos")
	}

	log.Println("Sucesso na API de fallback.")
	return dados, nil
}


func getLatestContestNumber() (int, error) {
	dados, err := fetchContestData("")
	if err != nil {
		return 0, fmt.Errorf("erro ao buscar o número do último concurso de ambas as APIs: %w", err)
	}
	return dados.Numero, nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	// Primeiro, obtemos o número do concurso mais recente
	latestContestNum, err := getLatestContestNumber()
	if err != nil {
		log.Printf("Erro crítico ao obter o número do último concurso: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		tpl.Execute(w, TemplateData{Erro: "Não foi possível obter o número do concurso mais recente de nenhuma das fontes."})
		return
	}

	numeroStr := r.URL.Query().Get("concurso")
	
	dados, err := fetchContestData(numeroStr)
	if err != nil {
		log.Printf("Erro final ao obter dados do concurso '%s': %v", numeroStr, err)
		numero, _ := strconv.Atoi(numeroStr)
		if numero > 0 {
			numero--
		}
		tpl.Execute(w, TemplateData{Erro: fmt.Sprintf("Sorteio %s não encontrado ou ainda não realizado.", numeroStr), NumeroAtual: numero})
		return
	}

	// Cria um "set" com todos os números apostados para checagem rápida
	myNumbersSet := make(map[string]bool)
	for _, jogo := range meusJogos {
		for _, num := range jogo {
			myNumbersSet[num] = true
		}
	}

	data := TemplateData{
		DadosLoteria:      dados,
		MeusJogos:         meusJogos,
		MyNumbersSet:      myNumbersSet,
		LatestContestNumber: latestContestNum,
	}

	err = tpl.Execute(w, data)
	if err != nil {
		log.Printf("Erro ao executar template: %v", err)
		http.Error(w, "Erro ao renderizar a página.", http.StatusInternalServerError)
	}
}

func main() {
	http.HandleFunc("/", handler)
	fmt.Println("Servidor iniciado em http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
