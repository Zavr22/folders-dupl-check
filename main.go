package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
)

type FileNode struct {
	Name     string
	IsDir    bool
	Children map[string]*FileNode
	Mutex    sync.Mutex
	Parent   *FileNode
}

func main() {
	root, err := buildFileTree("input.txt")
	if err != nil {
		panic(err)
	}

	findAndPrintSimilarDirectories(root, 50)
}

func buildFileTree(filePath string) (*FileNode, error) {
	root := &FileNode{IsDir: true, Children: make(map[string]*FileNode)}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var wg sync.WaitGroup
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			addToTree(root, path)
		}(scanner.Text())
	}

	wg.Wait()

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return root, nil
}

func addToTree(root *FileNode, path string) {
	parts := strings.Split(path, "/")
	currentNode := root

	for _, part := range parts {
		if part == "" {
			continue
		}

		currentNode.Mutex.Lock()
		if _, exists := currentNode.Children[part]; !exists {
			newNode := &FileNode{Name: part, IsDir: true, Children: make(map[string]*FileNode), Parent: currentNode}
			currentNode.Children[part] = newNode
		}
		currentNode.Mutex.Unlock()

		currentNode = currentNode.Children[part]
	}

	currentNode.IsDir = false
}

func compareDirectories(node1, node2 *FileNode, threshold float64) (float64, bool) {
	if node1 == nil || node2 == nil {
		return 0.0, false
	}

	similarity := calculateSimilarity(node1, node2)
	if similarity >= threshold {
		if node1.Parent != nil && node2.Parent != nil {
			parentSimilarity, _ := compareDirectories(node1.Parent, node2.Parent, threshold)
			if parentSimilarity >= threshold {
				return parentSimilarity, false
			}
		}
		return similarity, true
	}

	return 0.0, false
}

func findAndPrintSimilarDirectories(root *FileNode, threshold float64) {
	directories := collectDirectories(root)

	for i := 0; i < len(directories); i++ {
		for j := i + 1; j < len(directories); j++ {
			similarity, isTopLevel := compareDirectories(directories[i], directories[j], threshold)
			if similarity >= threshold && isTopLevel {
				fmt.Printf("Схожие директории: %s и %s, схожесть: %.2f%%\n",
					getPath(directories[i]), getPath(directories[j]), similarity)
			}
		}
	}
}

func calculateSimilarity(node1, node2 *FileNode) float64 {
	totalUnique := make(map[string]bool)
	matches := 0

	// Учитываем имена дочерних элементов и файлов
	for name := range node1.Children {
		totalUnique[name] = true
	}
	for name := range node2.Children {
		if _, exists := totalUnique[name]; exists {
			matches++
		} else {
			totalUnique[name] = true
		}
	}

	total := len(totalUnique)
	if total == 0 {
		return 100
	}

	return float64(matches) / float64(total) * 100
}

func getPath(node *FileNode) string {
	var parts []string
	for node != nil && node.Name != "" {
		parts = append([]string{node.Name}, parts...)
		node = node.Parent
	}
	return "/" + strings.Join(parts, "/")
}

func collectDirectories(node *FileNode) []*FileNode {
	var directories []*FileNode
	var stack []*FileNode

	stack = append(stack, node)
	for len(stack) > 0 {
		n := len(stack) - 1
		current := stack[n]
		stack = stack[:n]

		if current.IsDir {
			directories = append(directories, current)
			for _, child := range current.Children {
				stack = append(stack, child)
			}
		}
	}

	return directories
}
