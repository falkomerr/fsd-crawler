package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"fsd-crawler/pkg/analyzer"
	"fsd-crawler/pkg/config"
	"fsd-crawler/pkg/exporter"
	"fsd-crawler/pkg/model"
)

func main() {
	startTime := time.Now()
	
	cfg, err := config.FindAndLoadConfig()
	if err != nil {
		fmt.Printf("Предупреждение: не удалось загрузить конфигурацию: %v\n", err)
		fmt.Println("Используются значения по умолчанию.")
		cfg = &config.DefaultConfig
	}

	if cfg.SrcDir == "." {
		if _, err := os.Stat("src"); err == nil {
			cfg.SrcDir = "src"
		}
	}

	if len(cfg.Aliases) == 0 {
		cfg.Aliases = map[string]string{
			"@": "src",
			"~": "src",
		}
	}

	model.UpdateFromConfig(cfg)

	structure := analyzer.AnalyzeProject(cfg)

	outputFormats := cfg.OutputFormats
	if len(outputFormats) == 0 {
		outputFormats = []string{"html"}
	}

	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		fmt.Printf("Ошибка при создании директории для вывода: %v\n", err)
		return
	}

	htmlPath := ""
	
	for _, format := range outputFormats {
		switch format {
		case "html":
			if err := exporter.GenerateHTML(structure, cfg); err != nil {
				fmt.Printf("Ошибка при генерации HTML: %v\n", err)
			} else {
				htmlPath = filepath.Join(cfg.OutputDir, "fsd_structure.html")
			}
		case "json":
			if err := exporter.ExportJSON(structure, cfg); err != nil {
				fmt.Printf("Ошибка при экспорте в JSON: %v\n", err)
			}
		default:
			fmt.Printf("Неподдерживаемый формат вывода: %s\n", format)
		}
	}
	
	if cfg.ServeHTML && htmlPath != "" {
		port := cfg.Port
		if port == 0 {
			port = 3123
		}
		
		fs := http.FileServer(http.Dir(cfg.OutputDir))
		http.Handle("/", fs)
		
		url := fmt.Sprintf("http://localhost:%d/fsd_structure.html", port)
		
		go func() {
			err := http.ListenAndServe(":"+strconv.Itoa(port), nil)
			if err != nil {
				fmt.Printf("Ошибка при запуске веб-сервера: %v\n", err)
			}
		}()
		
		time.Sleep(100 * time.Millisecond)
		
		openBrowser(url)
		
		elapsed := time.Since(startTime).Milliseconds()
		
		clearConsole()
		
		fmt.Println()
		fmt.Printf("  FSD ANALYZER  ready in %d ms\n\n", elapsed)
		fmt.Printf("  ➜  Local:   %s\n", url)
		
		select {}
	}
}

func clearConsole() {
	cmd := exec.Command("clear")
	if _, err := os.Stat("/usr/bin/clear"); os.IsNotExist(err) {
		cmd = exec.Command("cmd", "/c", "cls")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func openBrowser(url string) {
	var err error
	
	switch {
	case commandExists("xdg-open"):
		err = runCommand("xdg-open", url)
	case commandExists("open"):
		err = runCommand("open", url)
	case commandExists("start"):
		err = runCommand("start", url)
	default:
		fmt.Printf("Не удалось автоматически открыть браузер. Пожалуйста, откройте %s вручную.\n", url)
		return
	}
	
	if err != nil {
		fmt.Printf("Ошибка при открытии браузера: %v\n", err)
	}
}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func runCommand(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	return cmd.Start()
} 