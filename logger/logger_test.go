package logger_test

import (
	"fmt"
	"testing"

	"github.com/fast-response/fast-response/logger"
)

func BenchmarkFastResponseLogger(b *testing.B) {
	logger := logger.NewLogger(logger.DEBUG)
	for i := 0; i < b.N; i++ {
		logger.Info("test")
	}
}

func BenchmarkFmtPrint(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fmt.Print("test\n")
	}
}

func BenchmarkPrint(b *testing.B) {
	for i := 0; i < b.N; i++ {
		print("test\n")
	}
}

func BenchmarkFmtPrintf(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fmt.Printf("test\n")
	}
}

func BenchmarkFmtPrintln(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fmt.Println("test")
	}
}

func BenchmarkPrintln(b *testing.B) {
	for i := 0; i < b.N; i++ {
		println("test")
	}
}