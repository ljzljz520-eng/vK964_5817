package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"printshop-safety/assessment"
)

func main() {
	inputPath := flag.String("input", "examples/assessment.json", "评估输入 JSON 文件")
	flag.Parse()

	inputFile, err := os.Open(*inputPath)
	if err != nil {
		exitf("读取评估输入失败: %v", err)
	}
	defer inputFile.Close()

	var input assessment.AssessmentInput
	if err := json.NewDecoder(inputFile).Decode(&input); err != nil {
		exitf("解析评估输入失败: %v", err)
	}

	service := assessment.NewService(
		assessment.NewRiskEvidenceAggregator(assessment.FileEvidenceReader{}, assessment.HighRiskThreshold),
		assessment.JSONPublisher{Writer: os.Stdout},
		assessment.SystemClock{},
	)
	if _, err := service.Submit(input); err != nil {
		exitf("提交评估失败: %v", err)
	}
}

func exitf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
