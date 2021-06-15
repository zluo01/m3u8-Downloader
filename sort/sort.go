package sort

import (
	"log"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

type Compare func(str1, str2 string) bool

func (cmp Compare) Sort(strs []string) {
	strSort := &strSorter{
		strs: strs,
		cmp:  cmp,
	}
	sort.Sort(strSort)
}

type strSorter struct {
	strs []string
	cmp  func(str1, str2 string) bool
}

func ExtractNumberFromString(str string, size int) (num int) {
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
	} else {

		num, err := strconv.Atoi(strSlice[size-1])
		if err != nil {
			log.Fatal(err)
		}

		return num
	}
}

func (s *strSorter) Len() int { return len(s.strs) }

func (s *strSorter) Swap(i, j int) { s.strs[i], s.strs[j] = s.strs[j], s.strs[i] }

func (s *strSorter) Less(i, j int) bool { return s.cmp(s.strs[i], s.strs[j]) }
