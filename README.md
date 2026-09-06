#  Analise de Tendencias de Vendas solução apresentada

Este documento complementa o README da proposta e resume o que foi produzido em `main.go`, com blocos de validacao para conferir cada parte do desafio.

## O que o `main.go` faz

A solucao le arquivos de vendas em largura fixa, ignora o cabecalho ate a linha `-------`, acumula as quantidades por produto em dois periodos e gera o Top 20 por crescimento absoluto.

Periodos considerados:

- Periodo anterior: janeiro de 2024
- Periodo atual: fevereiro de 2024

Arquivos suportados por flag:

```text
small       -> data/sample/sales_small.dat
medium      -> data/challenge/sales_medium.dat
large       -> data/challenge/sales_large.dat
extra-large -> data/challenge/sales_extra_large.dat
```

Uso basico:

```powershell
go run . -size small
go run . -size medium
go run . -size large
go run . -size extra-large
```

Para processar todos:

```powershell
go run . -all
```

O programa tambem grava o resultado em Markdown dentro de `results/`, seguindo o modelo de `results/result-sample.md` que foi disponibilizado no repositório do desafio

## Estrategia

Em vez de carregar todas as linhas em memoria, o programa faz leitura sequencial com `bufio.Reader` e um buffer maior (`16 MB`). Para cada linha, ele extrai apenas os campos necessarios:

```go
product := parseProductNumber(line)
quantity := parseQuantity(line)

switch {
case isJanuary2024(line):
	quantities[product].previous += quantity
case isFebruary2024(line):
	quantities[product].current += quantity
}
```

O acumulador principal e um slice indexado pelo numero do produto:

```go
quantities := make([]totals, productLimit)
```

Isso evita mapas grandes no caminho quente da leitura e deixa a solucao previsivel para o arquivo `extra-large`.

## Regras de negocio implementadas

Produtos sem quantidade no periodo anterior sao descartados:

```go
if product == 0 || total.previous == 0 {
	continue
}
```

Crescimento absoluto e percentual:

```go
growth := total.current - total.previous
percent := ceilTwoDecimals((float64(growth) / float64(total.previous)) * 100)
```

Arredondamento para cima com duas casas:

```go
func ceilTwoDecimals(value float64) float64 {
	return math.Ceil(value*100) / 100
}
```

Ordenacao deterministica:

```go
sort.Slice(results, func(i, j int) bool {
	if results[i].growth != results[j].growth {
		return results[i].growth > results[j].growth
	}
	if results[i].percent != results[j].percent {
		return results[i].percent > results[j].percent
	}
	return results[i].productID < results[j].productID
})
```

Formato do produto na saida:

```go
func formatProductID(product int) string {
	return fmt.Sprintf("produto%02d", product)
}
```

## Metricas

Ao final de cada execucao, o programa imprime e grava:

```text
CPU
RAM
Tempo decorrido
```

Esses valores tambem sao escritos no Markdown de saida:

```go
fmt.Fprintf(writer, "| CPU | %s |\n", formatCPU(m))
fmt.Fprintf(writer, "| RAM | %s |\n", formatBytes(m.memoryBytes))
fmt.Fprintf(writer, "| Tempo decorrido | %s |\n", formatDuration(m.elapsed))
```

## Resultado esperado da entrega

Estrutura principal:

```text
.
|-- README.md
|-- README_SOLUCAO.md
|-- main.go
|-- go.mod
|-- cpu_other.go
|-- cpu_windows.go
|-- data/
`-- results/
    |-- small.md
    |-- medium.md
    |-- large.md
    `-- extra_large.md
```
