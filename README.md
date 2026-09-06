# Solução - Análise de Tendências de Vendas no Varejo

## Aviso Sobre Os Dados

Este repositório está público por obrigatoriedade do teste. Os datasets de desafio são privados e não devem ser versionados. Por isso, apenas o arquivo `data/sample/sales_small.dat` fica no repositório; os arquivos de `data/challenge/` devem ser baixados ou gerados localmente conforme o enunciado.

## O Que A Solução Faz

O programa, escrito em Go, lê arquivos `.dat` de largura fixa e encontra os 20 produtos com maior crescimento entre:

- período anterior: janeiro de 2024;
- período atual: fevereiro de 2024.

Para cada produto, ele soma a quantidade vendida em cada período, calcula o crescimento absoluto e percentual, ordena o resultado e grava uma tabela em Markdown dentro de `results/`.

## Como Executar

```powershell
go run . -size small
go run . -size medium
go run . -size large
go run . -size extra-large
```

Também é possível limitar o número de núcleos lógicos usados pelo Go:

```powershell
go run . -size small -cpus 12
```

Para processar todos os datasets em sequência:

```powershell
go run . -all
```

O comando `-all` exige que todos os arquivos `.dat` estejam presentes localmente.

## Como O `.dat` É Tratado

O arquivo é lido sequencialmente com `bufio.Reader`. Primeiro, o programa ignora o cabeçalho até encontrar a linha `-------`. Depois disso, cada linha é tratada como uma venda.

Como o layout é de largura fixa, o programa acessa os campos diretamente pelas posições dos bytes, sem usar split ou parser de CSV:

- `product_id`: usado para identificar o produto;
- `timestamp`: usado para separar janeiro e fevereiro de 2024;
- `quantity`: usado para acumular a quantidade vendida.

Campos como loja e preço unitário são ignorados porque não entram no cálculo pedido.

## Estratégia

As quantidades são acumuladas em um slice indexado pelo número do produto, o que evita guardar todas as vendas em memória e reduz o custo de busca durante a leitura.

Depois da leitura, o programa:

1. descarta produtos sem vendas no período anterior;
2. calcula `growth = current - previous`;
3. calcula o percentual de crescimento;
4. arredonda o percentual para cima com duas casas decimais;
5. ordena por maior crescimento, maior percentual em empate e `product_id` crescente no empate final;
6. grava o Top 20.

## Métricas

Cada execução imprime e grava:

- CPU;
- RAM;
- tempo decorrido.

No Windows, `cpu_windows.go` usa `syscall.GetProcessTimes` para medir o tempo de CPU do processo. Em outros sistemas, `cpu_other.go` mantém o mesmo contrato de código e informa quando a medição automática de CPU não está disponível.

## Resultados

Os arquivos de saída seguem o modelo de `results/result-sample.md`:

```text
results/
  small.md
  medium.md
  large.md
  extra_large.md
```

## Validação

```powershell
go test ./...
```

Os testes cobrem cálculo, descarte de produto sem período anterior, produto descontinuado, arredondamento e ordenação.

Mais detalhes da implementação estão em `README_SOLUCAO.md`.
