# E um aviso
este repositório esta publico por obrigatoriedade do teste, os dados necessários para rodar esse script são privados. e não se encontram neste repositório

#  Analise de Tendencias de Vendas solução apresentada

Este documento segue de acordo o README da proposta e resume o que foi produzido em `main.go`, com blocos de validacao para conferir cada parte do desafio.

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
você também pode especificar o numero de nucleos que a cpu utilaza usando -cpus 'numero de cpus'

```powershell
go run . -size small -cpus -12
go run . -size medium -cpus -12
go run . -size large -cpus-12
go run . -size extra-large -21
```

Para processar todos:

```powershell
go run . -all
```

O programa tambem grava o resultado em Markdown dentro de `results/`, seguindo o modelo de `results/result-sample.md` que foi disponibilizado no repositório do desafio
caso quem estiver testando o código, existe 2 linhas comentadas que reportam tanto no console quanto no arquivo salvo o numero de linhas que foram percorridas

### Arquivos de CPU por sistema operacional

O projeto separa a medicao de CPU em dois arquivos porque o Go permite escolher arquivos diferentes conforme o sistema operacional usando build tags.

`cpu_windows.go` tem a tag `//go:build windows`, entao so entra na compilacao quando o programa roda no Windows. Ele usa `syscall.GetProcessTimes` para ler o tempo de CPU consumido pelo processo atual, somando tempo de kernel e tempo de usuario. O `main.go` compara essa medicao antes e depois do processamento para calcular o percentual de CPU exibido nos resultados.

`cpu_other.go` tem a tag `//go:build !windows`, entao entra na compilacao em sistemas que nao sejam Windows. Nesse caso a funcao existe apenas para manter o mesmo contrato do codigo, mas retorna `0, false`, indicando que a medicao automatica de CPU nao foi feita naquele sistema.

## Estrategia

Em vez de carregar todas as linhas em memoria, o programa faz leitura sequencial com `bufio.Reader` e um buffer maior (`16 MB`). Para cada linha, ele extrai apenas os campos necessarios:

### Como o arquivo `.dat` e tratado

O arquivo `.dat` e lido como um arquivo de largura fixa. Primeiro o programa ignora o cabecalho ate encontrar a linha separadora `-------`; a partir dali, cada linha passa a representar uma venda.

Como os campos sempre ocupam as mesmas posicoes, o programa nao precisa quebrar a linha por delimitadores. Ele acessa diretamente os bytes onde ficam `product_id`, `timestamp` e `quantity`. Por exemplo, o produto e lido nas posicoes do identificador `PR00000`, a quantidade fica nas posicoes `[45:47]`, e o mes do `timestamp` e verificado para decidir se a venda pertence a janeiro ou fevereiro de 2024.

Durante essa leitura, o programa acumula apenas as quantidades por produto:

- vendas de janeiro entram em `previous`;
- vendas de fevereiro entram em `current`;
- os outros campos, como loja e preco unitario, sao ignorados porque nao entram no calculo pedido.

Esse tratamento evita guardar todas as vendas em memoria. No fim da leitura, o programa percorre os totais acumulados por produto, calcula crescimento absoluto e percentual, ordena os resultados e grava o Top 20.

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
