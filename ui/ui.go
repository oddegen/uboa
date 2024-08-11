package ui

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"runtime"
	"time"

	bam "github.com/oddegen/bam/internal/pkg"
)

func Render(data bam.ResultMetrics, filename string, skip bool) error {

	htmlTemplate, err := os.ReadFile("ui/result.tmpl")
	if err != nil {
		return err
	}

	var b bytes.Buffer
	tmpl := template.Must(template.New("metrics").Parse(string(htmlTemplate)))
	tmpl.Execute(&b, data)

	os.WriteFile(filename, b.Bytes(), 0644)

	if !skip {
		return preview(filename)
	}

	return nil

}

func preview(fname string) error {
	cName := ""
	cParams := []string{}

	switch runtime.GOOS {
	case "linux":
		cName = "xdg-open"
	case "windows":
		cName = "cmd.exe"
		cParams = []string{"/C", "start"}
	case "darwin":
		cName = "open"
	default:
		return fmt.Errorf("OS not supported")
	}

	cParams = append(cParams, fname)

	cPath, err := exec.LookPath(cName)
	if err != nil {
		return err
	}

	err = exec.Command(cPath, cParams...).Run()

	time.Sleep(2 * time.Second)
	return err
}
