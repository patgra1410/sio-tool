package judge

import (
	"os"
	"strings"

	"fmt"
	"github.com/fatih/color"
)

func Judge(inPath, ansPath, sampleID, command string, oiejqOptions *OiejqOptions) Verdict {
	input, err := os.Open(inPath)
	if err != nil {
		return Verdict{INT, 0, 0, "", err}
	}
	defer input.Close()

	var processInfo ProcessInfo
	if oiejqOptions != nil {
		processInfo, err = RunProcessWithOiejq(command, input, oiejqOptions)
	} else {
		processInfo, err = RunProcess(command, input, nil)
	}
	if err != nil || processInfo.Status != OK {
		return Verdict{processInfo.Status, processInfo.TimeInSeconds, processInfo.MemoryInMegabytes, "", err}
	}
	if err == nil && len(processInfo.Stderr) > 0 { // Check if stderr was captured
		fmt.Println(color.New(color.FgCyan).Sprint("-----Stderr-----"))
    	fmt.Fprintln(os.Stderr, "Captured Stderr Output:\n"+string(processInfo.Stderr))
	}


	b, err := os.ReadFile(ansPath)
	if err != nil {
		return Verdict{INT, 0, 0, "", err}
	}
	return GenerateVerdict(sampleID, Plain(b), processInfo)
}

func ExtractTaskName(file string) (task string) {
	task, _, _ = strings.Cut(file, "-")
	return
}
