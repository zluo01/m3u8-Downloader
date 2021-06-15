package sort

import (
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

type Compare func(file1, file2 os.FileInfo) bool

func (cmp Compare) Sort(files []os.FileInfo) {
	fileSort := &FileSorter{
		files: files,
		cmp:   cmp,
	}
	sort.Sort(fileSort)
}

type FileSorter struct {
	files []os.FileInfo
	cmp   func(file1, file2 os.FileInfo) bool
}

func CompareStringNumber(file1, file2 os.FileInfo) bool {
	return extractNumberFromString(file1.Name(), 0) < extractNumberFromString(file2.Name(), 0)
}

func extractNumberFromString(str string, size int) int {
	strSlice := make([]string, 0)
	for _, v := range str {
		if unicode.IsDigit(v) {
			strSlice = append(strSlice, string(v))
		}
	}

	if size == 0 { // default
		num, err := strconv.Atoi(strings.Join(strSlice, ""))
		if err != nil {
			log.Fatal(err)
		}
		return num
	}
	num, err := strconv.Atoi(strSlice[size-1])
	if err != nil {
		log.Fatal(err)
	}
	return num
}

func (f *FileSorter) Len() int { return len(f.files) }

func (f *FileSorter) Swap(i, j int) { f.files[i], f.files[j] = f.files[j], f.files[i] }

func (f *FileSorter) Less(i, j int) bool { return f.cmp(f.files[i], f.files[j]) }
