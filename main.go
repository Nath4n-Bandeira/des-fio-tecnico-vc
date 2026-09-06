package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

const (
	productLimit = 100_000
	readBuffer   = 16 * 1024 * 1024
)

var datasetPaths = map[string]string{
	"small":       filepath.Join("data", "sample", "sales_small.dat"),
	"medium":      filepath.Join("data", "challenge", "sales_medium.dat"),
	"large":       filepath.Join("data", "challenge", "sales_large.dat"),
	"extra-large": filepath.Join("data", "challenge", "sales_extra_large.dat"),
}

type totals struct {
	previous int
	current  int
}

type productGrowth struct {
	productID string
	previous  int
	current   int
	growth    int
	percent   float64
}

type metrics struct {
	elapsed     time.Duration
	cpuMeasured bool
	cpuTotal    time.Duration
	cpuPercent  float64
	memoryBytes uint64
	logicalCPUs int
}

func main() {
	size := flag.String("size", "small", "Dataset: small, medium, large ou extra-large")
	all := flag.Bool("all", false, "Processa todos os datasets conhecidos")
	inputPath := flag.String("input", "", "Caminho do arquivo .dat de entrada")
	outputPath := flag.String("output", "", "Caminho do arquivo Markdown de saída")
	cpus := flag.Int("cpus", runtime.NumCPU(), "Número máximo de núcleos lógicos usados pelo Go")
	flag.Parse()

	if *cpus <= 0 {
		logFatal(fmt.Errorf("o valor de -cpus precisa ser maior que zero"))
	}
	runtime.GOMAXPROCS(*cpus)

	if *all {
		for _, name := range []string{"small", "medium", "large", "extra-large"} {
			if err := processDataset(name, datasetPaths[name], defaultOutputPath(name)); err != nil {
				logFatal(err)
			}
		}
		return
	}

	name := *size
	input := *inputPath
	if input == "" {
		var ok bool
		input, ok = datasetPaths[name]
		if !ok {
			logFatal(fmt.Errorf("dataset desconhecido: %s", name))
		}
	}

	output := *outputPath
	if output == "" {
		output = defaultOutputPath(name)
	}

	if err := processDataset(name, input, output); err != nil {
		logFatal(err)
	}
}

func processDataset(name, inputPath, outputPath string) error {
	startWall := time.Now()
	startCPU, cpuMeasured := currentProcessCPU()

	rows, top, err := analyzeFile(inputPath)
	if err != nil {
		return err
	}

	elapsed := time.Since(startWall)
	endCPU, endCPUMeasured := currentProcessCPU()
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	resultMetrics := metrics{
		elapsed:     elapsed,
		cpuMeasured: cpuMeasured && endCPUMeasured,
		memoryBytes: mem.Sys,
		logicalCPUs: runtime.GOMAXPROCS(0),
	}
	if resultMetrics.cpuMeasured {
		resultMetrics.cpuTotal = endCPU - startCPU
		if elapsed > 0 && resultMetrics.logicalCPUs > 0 {
			resultMetrics.cpuPercent = (float64(resultMetrics.cpuTotal) / float64(elapsed) / float64(resultMetrics.logicalCPUs)) * 100
		}
	}

	printConsole(name, rows, top, resultMetrics)
	return writeMarkdown(outputPath, name, rows, top, resultMetrics)
}

func analyzeFile(path string) (int, []productGrowth, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, nil, fmt.Errorf("abrir entrada %q: %w", path, err)
	}
	defer file.Close()

	reader := bufio.NewReaderSize(file, readBuffer)
	if err := skipHeader(reader); err != nil {
		return 0, nil, err
	}

	quantities := make([]totals, productLimit)
	rows := 0

	for {
		line, err := reader.ReadSlice('\n')
		if len(line) > 0 {
			if len(line) < 54 {
				return rows, nil, fmt.Errorf("linha %d inválida: esperados ao menos 54 caracteres", rows+1)
			}

			product := parseProductNumber(line)
			quantity := parseQuantity(line)

			switch {
			case isJanuary2024(line):
				quantities[product].previous += quantity
			case isFebruary2024(line):
				quantities[product].current += quantity
			}
			rows++
		}

		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return rows, nil, fmt.Errorf("ler linha %d: %w", rows+1, err)
		}
	}

	results := make([]productGrowth, 0, 20_000)
	for product, total := range quantities {
		if product == 0 || total.previous == 0 {
			continue
		}

		growth := total.current - total.previous
		percent := ceilTwoDecimals((float64(growth) / float64(total.previous)) * 100)
		results = append(results, productGrowth{
			productID: formatProductID(product),
			previous:  total.previous,
			current:   total.current,
			growth:    growth,
			percent:   percent,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].growth != results[j].growth {
			return results[i].growth > results[j].growth
		}
		if results[i].percent != results[j].percent {
			return results[i].percent > results[j].percent
		}
		return results[i].productID < results[j].productID
	})

	if len(results) > 20 {
		results = results[:20]
	}

	return rows, results, nil
}

func skipHeader(reader *bufio.Reader) error {
	for {
		line, err := reader.ReadString('\n')
		if strings.TrimSpace(line) == "-------" {
			return nil
		}
		if errors.Is(err, io.EOF) {
			return errors.New("cabeçalho inválido: separador ------- não encontrado")
		}
		if err != nil {
			return fmt.Errorf("ler cabeçalho: %w", err)
		}
	}
}

func parseProductNumber(line []byte) int {
	return int(line[20]-'0')*10_000 +
		int(line[21]-'0')*1_000 +
		int(line[22]-'0')*100 +
		int(line[23]-'0')*10 +
		int(line[24]-'0')
}

func parseQuantity(line []byte) int {
	return int(line[45]-'0')*10 + int(line[46]-'0')
}

func isJanuary2024(line []byte) bool {
	return line[25] == '2' &&
		line[26] == '0' &&
		line[27] == '2' &&
		line[28] == '4' &&
		line[30] == '0' &&
		line[31] == '1'
}

func isFebruary2024(line []byte) bool {
	return line[25] == '2' &&
		line[26] == '0' &&
		line[27] == '2' &&
		line[28] == '4' &&
		line[30] == '0' &&
		line[31] == '2'
}

func ceilTwoDecimals(value float64) float64 {
	return math.Ceil(value*100) / 100
}

func formatProductID(product int) string {
	return fmt.Sprintf("produto%02d", product)
}

func defaultOutputPath(size string) string {
	fileName := strings.ReplaceAll(size, "-", "_") + ".md"
	return filepath.Join("results", fileName)
}

func printConsole(name string, rows int, top []productGrowth, m metrics) {
	fmt.Printf("\nResultado - %s\n", name)
	//fmt.Printf("Linhas processadas: %d\n\n", rows)
	fmt.Printf("%-10s %24s %20s %12s %17s\n", "Produto", "Qtd. Período Anterior", "Qtd. Período Atual", "Crescimento", "Crescimento (%)")
	for _, item := range top {
		fmt.Printf("%-10s %24d %20d %12d %16.2f%%\n", item.productID, item.previous, item.current, item.growth, item.percent)
	}
	fmt.Printf("\nCPU: %s\n", formatCPU(m))
	fmt.Printf("RAM: %s\n", formatBytes(m.memoryBytes))
	fmt.Printf("Tempo decorrido: %s\n", formatDuration(m.elapsed))
}

func writeMarkdown(path, name string, rows int, top []productGrowth, m metrics) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("criar diretório de saída: %w", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("criar saída %q: %w", path, err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	fmt.Fprintf(writer, "# Resultado - %s\n\n", name)
	//fmt.Fprintf(writer, "Linhas processadas: %d\n\n", rows)
	fmt.Fprintln(writer, "## Top 20")
	fmt.Fprintln(writer)
	fmt.Fprintln(writer, "| Produto | Qtd. Período Anterior | Qtd. Período Atual | Crescimento | Crescimento (%) |")
	fmt.Fprintln(writer, "|---|---:|---:|---:|---:|")
	for _, item := range top {
		fmt.Fprintf(writer, "| %s | %d | %d | %d | %.2f%% |\n", item.productID, item.previous, item.current, item.growth, item.percent)
	}

	fmt.Fprintln(writer)
	fmt.Fprintln(writer, "## Recursos")
	fmt.Fprintln(writer)
	fmt.Fprintln(writer, "| Métrica | Valor |")
	fmt.Fprintln(writer, "|---|---|")
	fmt.Fprintf(writer, "| CPU | %s |\n", formatCPU(m))
	fmt.Fprintf(writer, "| RAM | %s |\n", formatBytes(m.memoryBytes))
	fmt.Fprintf(writer, "| Tempo decorrido | %s |\n", formatDuration(m.elapsed))

	return nil
}

func formatCPU(m metrics) string {
	cores := "núcleos"
	if m.logicalCPUs == 1 {
		cores = "núcleo"
	}

	if !m.cpuMeasured {
		return fmt.Sprintf("não medido automaticamente (%d %s)", m.logicalCPUs, cores)
	}
	return fmt.Sprintf("%.0f%% (%d %s)", m.cpuPercent, m.logicalCPUs, cores)
}

func formatBytes(bytes uint64) string {
	const mb = 1024 * 1024
	return fmt.Sprintf("%.2f MB", float64(bytes)/mb)
}

func formatDuration(duration time.Duration) string {
	return strings.ReplaceAll(fmt.Sprintf("%.2f s", duration.Seconds()), ".", ",")
}

func logFatal(err error) {
	fmt.Fprintln(os.Stderr, "erro:", err)
	os.Exit(1)
}
