package sort

import (
	"log"
	"math/big"
	"os"
	"sort"
	"strings"
	"unicode"
)

type Compare func(file1, file2 os.DirEntry) bool

func (cmp Compare) Sort(files []os.DirEntry) {
	fileSort := &FileSorter{
		files: files,
		cmp:   cmp,
	}
	sort.Sort(fileSort)
}

type FileSorter struct {
	files []os.DirEntry
	cmp   func(file1, file2 os.DirEntry) bool
}

func CompareStringNumber(file1, file2 os.DirEntry) bool {
	return extractNumberFromString(file1.Name(), 0).Cmp(extractNumberFromString(file2.Name(), 0)) == -1
}

func extractNumberFromString(str string, size int) *big.Int {
	strSlice := make([]string, 0)
	for _, v := range str {
		if unicode.IsDigit(v) {
			strSlice = append(strSlice, string(v))
		}
	}

	numStr := strings.Join(strSlice, "") // default use all numbers
	if size != 0 {
		numStr = strSlice[size-1]
	}

	num, ok := new(big.Int).SetString(numStr, 10)
	if !ok {
		log.Fatal("Fail to parse string into number ", numStr)
	}
	return num
}

func (f *FileSorter) Len() int { return len(f.files) }

func (f *FileSorter) Swap(i, j int) { f.files[i], f.files[j] = f.files[j], f.files[i] }

func (f *FileSorter) Less(i, j int) bool { return f.cmp(f.files[i], f.files[j]) }
