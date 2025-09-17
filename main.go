package main

import (
	"fmt"

	"file/function"
)

func main() {
	fullPath, err := function.DirectoryCheck()
	if err != nil {
		fmt.Println(err)
		return
	}

	words := make(map[string]int)

	fileCh, err := function.Producer(fullPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	mapCh, errCh := function.ReadFile(fileCh, 4)

	for mapCh != nil || errCh != nil {
		select {
		case m, ok := <-mapCh:
			if !ok {
				mapCh = nil
				continue
			}
			for w, c := range m {
				words[w] += c
			}

		case err, ok := <-errCh:
			if !ok {
				errCh = nil
				continue
			}
			fmt.Println("ошибка:", err)
			return
		}
	}

	pairs := function.Statistics(words)

	if len(pairs) > 0 {
		fmt.Printf("most used\n%s: %d\n", pairs[0].Word, pairs[0].Count)
	} else {
		fmt.Println("no words")
	}
}
