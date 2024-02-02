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
	Parent   *FileNode
}

var mutex sync.Mutex

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

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		addToTree(root, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return root, nil
}

func addToTree(root *FileNode, path string) {
	parts := strings.Split(path, "/")
	currentNode := root
	isLastComponentDir := path[len(path)-1] == '/'

	for i, part := range parts {
		if part == "" {
			if i < len(parts)-1 || !isLastComponentDir {
				continue
			}
		}

		mutex.Lock()
		if _, exists := currentNode.Children[part]; !exists {
			newNode := &FileNode{Name: part, IsDir: true, Children: make(map[string]*FileNode), Parent: currentNode}
			currentNode.Children[part] = newNode
		} else {
		}
		currentNode = currentNode.Children[part]
		mutex.Unlock()
		if i == len(parts)-1 {
			currentNode.IsDir = isLastComponentDir || part == ""
		}
	}
}

func compareDirectories(node1, node2 *FileNode, threshold float64) (float64, bool) {
	if node1 == nil || node2 == nil || isAncestor(node1, node2) || isAncestor(node2, node1) {
		return 0.0, false
	}
	if node1.Name == "" && node2.Name == "" {
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

func isAncestor(ancestor, descendant *FileNode) bool {
	for n := descendant; n != nil; n = n.Parent {
		if n == ancestor {
			return true
		}
	}
	return false
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
	children1 := make(map[string]struct{})
	for name := range node1.Children {
		children1[name] = struct{}{}
	}

	matches := 0
	total := len(children1)
	for name := range node2.Children {
		if _, exists := children1[name]; exists {
			matches++
		} else {
			total++
		}
	}

	if total == 0 {
		return 100.0
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

func collectDirectories(root *FileNode) []*FileNode {
	var directories []*FileNode
	var collect func(node *FileNode)
	collect = func(node *FileNode) {
		if node == nil {
			return
		}
		for _, child := range node.Children {
			collect(child)
		}
		if node.IsDir {
			directories = append(directories, node)
		}
	}
	collect(root)
	if len(directories) > 0 && directories[len(directories)-1] == root {
		directories = directories[:len(directories)-1]
	}
	return directories
}
