package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
)

func main() {
  chunksDir := "../videos/5"
  mergeChunks(chunksDir, "merged.mp4")
}

func extractNumber(fileName string) int {
	regex := regexp.MustCompile(`\d+`)
	numberStr := regex.FindString(filepath.Base(fileName))
	chunkNumber, error := strconv.Atoi(numberStr)

	if error != nil {
		log.Fatal(error)
		return -1
	}

	return chunkNumber
}

func mergeChunks(inputDir, outputFile string) error {
	chunks, err := filepath.Glob(filepath.Join(inputDir, "*.chunk"))

	if err != nil {
		return fmt.Errorf("failed to find chunks: %v", err)
	}

	sort.Slice(chunks, func(i, j int) bool {
		return extractNumber(chunks[i]) < extractNumber(chunks[j])
	})

	output, err := os.Create(outputFile)

	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}

	defer output.Close()

	for _, chunk := range chunks {
		input, err := os.Open(chunk)

		if err != nil {
			return fmt.Errorf("failed to open chunk: %v", err)
		}

		_, err = output.ReadFrom(input)

		if err != nil {
			return fmt.Errorf("failed to write chunk %s to output: %v", chunk, err)
		}

		input.Close()
	}

	return nil
}

// func getAllFilesInDir(dirPath string) []string {
// 	files, err := filepath.Glob(dirPath + "/*")
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	return files
// }

// func bubbleSortFiles(numbers []int) []int {
// 	orderedNumbers := make([]int, len(numbers))
// 	copy(orderedNumbers, numbers)
// 	numbersLength := len(orderedNumbers)
// 	var swapped bool

// 	for currentNumber := 0; currentNumber < numbersLength-1; currentNumber++ {
// 		swapped = false
// 		for currentSwap := 0; currentSwap < numbersLength-currentNumber-1; currentSwap++ {
// 			nextNumber := orderedNumbers[currentSwap+1]
// 			previousNumber := orderedNumbers[currentSwap]
// 			if previousNumber > nextNumber {
// 				orderedNumbers[currentSwap] = nextNumber
// 				orderedNumbers[currentSwap+1] = previousNumber
// 				swapped = true
// 			}
// 		}
// 		if !swapped {
// 			break
// 		}

// 	}

// 	return orderedNumbers
// }
