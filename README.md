# Análise de Tendências de Vendas no Varejo

## Contexto

A **Varejo Consolidado** opera uma rede de lojas em diferentes regiões. Cada venda é registrada e enviada para um data warehouse central. A equipe de inteligência comercial precisa identificar, entre milhares de produtos, quais mais cresceram ao comparar dois períodos consecutivos.

Essa análise orienta reposição de estoque, campanhas e negociação com fornecedores. O volume de vendas é grande e continua crescendo — a solução precisa ser correta com poucos registros e continuar viável com dezenas de milhões.

## Problema

A partir de um arquivo de vendas, encontre os **20 produtos** que mais cresceram entre o período **anterior** (janeiro de 2024) e o período **atual** (fevereiro de 2024).

Linguagem livre. O problema — entrada, regras e saída — precisa ser compatível com qualquer linguagem escolhida.

## Entrada

Arquivo de largura fixa. O cabeçalho lista cada campo como `nome:início,tamanho`. Depois da linha `-------` começam os registros — um por linha, sem delimitador. IDs têm sempre o mesmo tamanho em todos os datasets.

```
sale_id:0,11
store_id:11,7
product_id:18,7
timestamp:25,20
quantity:45,2
unit_price:47,7
-------
S0000000090ST00010PR000022024-01-23T16:45:22Z06   8.58
S0000000091ST00003PR000012024-02-01T00:00:00Z12  12.87
```

| Campo | Tipo | Slice | Descrição |
|---|---|---|---|
| `sale_id` | string | `[0:11]` | `S` + 10 dígitos. Identificador único da venda. |
| `store_id` | string | `[11:18]` | `ST` + 5 dígitos. Identificador da loja. Não entra no cálculo. |
| `product_id` | string | `[18:25]` | `PR` + 5 dígitos. Identificador do produto. |
| `timestamp` | string | `[25:45]` | Data e hora em UTC, ISO 8601 (`YYYY-MM-DDTHH:MM:SSZ`). |
| `quantity` | int | `[45:47]` | Quantidade vendida, sempre `>= 1`. |
| `unit_price` | float | `[47:54]` | Preço unitário, sempre `> 0`. Não entra no cálculo. |

**Períodos.** Instante de corte: `2024-02-01T00:00:00Z`.

- **Anterior:** `[2024-01-01T00:00:00Z, 2024-02-01T00:00:00Z)`
- **Atual:** `[2024-02-01T00:00:00Z, 2024-03-01T00:00:00Z)`

Fevereiro de 2024 inclui o dia 29. Os dados são sempre válidos.

## Saída

Imprima no console, em português, uma tabela com os 20 produtos: Produto, Qtd. Período Anterior, Qtd. Período Atual, Crescimento e Crescimento (%).

Só entram produtos com quantidade **maior que zero** no período anterior. Produto sem venda em janeiro fica de fora (não há base de comparação).

Exemplo:

```
Produto      Qtd. Período Anterior   Qtd. Período Atual   Crescimento   Crescimento (%)
produto01    80                      320                  240           300.00%
produto02    110                     340                  230           209.09%
```

Salve o mesmo resultado em markdown, um arquivo por dataset — veja a árvore do projeto em Entrega. O formato esperado está em `results/result-sample.md`.

## Regras de negócio

### Cálculo

```
growth = current_period_quantity - previous_period_quantity
growth_percentage = (growth / previous_period_quantity) * 100
```

Arredonde `growth_percentage` para cima, com duas casas decimais.

### Casos especiais

1. **Sem quantidade no período anterior** (`previous == 0`): o produto **não entra** no resultado — inclusive produto novo, vendido só em fevereiro.
2. **Produto descontinuado** (`previous > 0` e `current == 0`): `growth = -previous_period_quantity`; `growth_percentage = -100.0`.
3. **Sem vendas nos dois períodos**: produto não aparece no resultado.
4. Os dados de entrada são sempre válidos.

### Ordenação determinística

1. Maior `growth` primeiro.
2. Empate em `growth` → maior `growth_percentage` primeiro.
3. Empate persistente → `product_id` em ordem crescente.

## Dataset

```
data/
  sample/
    sales_small.dat
  challenge/
    sales_medium.dat
    sales_large.dat
    sales_extra_large.dat
```

| Dataset | Lojas | Produtos | Linhas (aprox.) | Tamanho (aprox.) | Uso |
|---|---|---|---|---|---|
| `small` | 5 | 30 | ~36 mil | ~2 MB | Desenvolvimento |
| `medium` | 50 | 500 | ~360 mil | ~20 MB | Validação |
| `large` | 500 | 5.000 | ~3,6 milhões | ~200 MB | Escalabilidade |
| `extra-large` | 5.000 | 20.000 | ~36 milhões | ~2 GB | Escalabilidade |

Somente `small` está no repositório. Para `medium`, `large` e `extra-large`:

1. **Download:** [pasta compartilhada](https://varejoconsolidado868-my.sharepoint.com/:f:/g/personal/kaua_rob_varejoconsolidado_com_br/IgCR4yjTPNo0T7neKygGNJgCAZVQh_w3TfGsM_DR_cVh3f8?e=uSEwJ7) — a senha será enviada ao candidato por e-mail, junto com o teste.
2. **Geração local:** `python scripts/generate_dataset.py --size <tamanho>` (determinístico, seed `20240101`).

## Performance

Meça e reporte, para cada dataset: **uso de CPU**, **uso de memória (RAM)** e **tempo decorrido total**. Performance é critério de avaliação — inclua esses três números em `results/[size].md`.

## Entrega

Repositório GitHub público, estruturado assim:

```
projeto/
├── README.md
├── (seu código-fonte)
└── results/
    ├── small.md
    ├── medium.md
    ├── large.md
    └── extra_large.md
```

- **README.md** — explique sua solução.
- **results/[size].md** — tabela dos 20 produtos daquele dataset + CPU, RAM e tempo decorrido. Use `results/result-sample.md` como modelo.

## Avaliação

1. **Correção** — regras de negócio, casos especiais, saída no formato pedido.
2. **Performance** — CPU, RAM e tempo decorrido nos quatro datasets.
3. **Escalabilidade** — funciona nos quatro datasets.
4. **Clareza** do código.
5. **Robustez** — casos especiais tratados corretamente.
6. **Entrega** — repositório, README, resultados documentados.# des-fio-tecnico-vc
