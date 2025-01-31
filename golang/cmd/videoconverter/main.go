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
	chunksFolder := "/tmp/videos/1"
	files := getAllFilesInDir(chunksFolder)
	fileNumbers := []int{}

	for _, file := range files {
		chunkNumber := extractNumber(file)
		// log.Default().Println(string(chunkNumber))
		fileNumbers = append(fileNumbers, chunkNumber)
	}

	// sortedFiles := bubbleSortFiles(fileNumbers)
	sortedFiles := mergeChunks(chunksFolder, "")

	fmt.Println(sortedFiles)

}

func bubbleSortFiles(numbers []int) []int {
	orderedNumbers := make([]int, len(numbers))
	copy(orderedNumbers, numbers)
	numbersLength := len(orderedNumbers)
	var swapped bool

	for currentNumber := 0; currentNumber < numbersLength-1; currentNumber++ {
		swapped = false
		for currentSwap := 0; currentSwap < numbersLength-currentNumber-1; currentSwap++ {
			nextNumber := orderedNumbers[currentSwap+1]
			previousNumber := orderedNumbers[currentSwap]
			if previousNumber > nextNumber {
				orderedNumbers[currentSwap] = nextNumber
				orderedNumbers[currentSwap+1] = previousNumber
				swapped = true
			}
		}
		if !swapped {
			break
		}

	}

	return orderedNumbers
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

func mergeChunks(inputDir, outputDir string) error {
	chunks, error := filepath.Glob(filepath.Join(inputDir, "*.chunk"))

	if error != nil {
		return fmt.Errorf("failed to find chunks: %v", error)
	}

	sort.Slice(chunks, func(i, j int) bool {
		return extractNumber(chunks[i]) < extractNumber(chunks[j])
	})

  output, error := os.Create(outputDir)

  if error != nil {
    return fmt.Errorf("failed to create output file: %v", error)
  }

  defer output.Close()

	return nil
}

func getAllFilesInDir(dirPath string) []string {
	files, err := filepath.Glob(dirPath + "/*")
	if err != nil {
		fmt.Println(err)
	}
	return files
}
