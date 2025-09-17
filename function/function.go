package function

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"golang.org/x/sync/errgroup"
)

type kv struct {
	Word  string
	Count int
}

func DirectoryCheck() (string, error) {
	if len(os.Args) == 1 {
		return "", fmt.Errorf("directory not provided")
	} else if len(os.Args) > 2 {
		return "", fmt.Errorf("argument must be 1")
	}

	home, _ := os.UserHomeDir()
	fullPath := filepath.Join(home, os.Args[1])

	_, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("directory %s not found", fullPath)
		} else {
			fmt.Println()
			return "", fmt.Errorf("error accessing: %v", err)
		}
	}

	return fullPath, nil
}

func Producer(directory string) (<-chan string, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("error reading directory: %v", err)
	}

	fileCh := make(chan string)

	go func() {
		defer close(fileCh)
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			fullPath := filepath.Join(directory, entry.Name())
			fileCh <- fullPath
		}
	}()

	return fileCh, nil
}

func countWordsInFile(path string, ctx context.Context) (map[string]int, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		result := make(map[string]int)

		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanWords)

		for scanner.Scan() {
			word := scanner.Text()

			word = strings.TrimFunc(word, func(r rune) bool {
				return !unicode.IsLetter(r) && !unicode.IsDigit(r)
			})
			word = strings.ToLower(word)
			if word != "" {
				result[word]++
			}
		}

		if err := scanner.Err(); err != nil {
			return nil, err
		}

		return result, nil
	}
}

func ReadFile(fileCh <-chan string, n int) (<-chan map[string]int, <-chan error) {
	resCh := make(chan map[string]int)
	errCh := make(chan error)

	g, ctx := errgroup.WithContext(context.Background())

	for range n {
		g.Go(func() error {
			for file := range fileCh {
				m, err := countWordsInFile(file, ctx)
				if err != nil {
					return err
				}
				select {
				case resCh <- m:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			return nil
		})
	}

	go func() {
		defer close(resCh)
		defer close(errCh)

		if err := g.Wait(); err != nil {
			errCh <- err
		}
	}()

	return resCh, errCh
}

func Statistics(words map[string]int) []kv {

	var pairs []kv
	for k, v := range words {
		pairs = append(pairs, kv{k, v})
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Count > pairs[j].Count
	})

	return pairs
}
