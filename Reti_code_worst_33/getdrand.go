package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type drandResult struct {
	Randomness []byte `json:"randomness"`
	Signature  []byte `json:"signature"`
}

const drandURL = "https://api.drand.sh/public/latest"

func getDrandResult() ([]byte, error) {
	resp, err := http.Get(drandURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get drand result: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get drand result: status code %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read drand result: %v", err)
	}

	var result drandResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal drand result: %v", err)
	}

	// Save to  drand.json
	err = os.MkdirAll("./data", os.ModePerm)
	err = ioutil.WriteFile("./data/drand.json", body, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to save drand result to file: %v", err)
	}

	return result.Randomness, nil
}

func generateNodes(x int) ([]*Node, error) {
	err := os.MkdirAll("./data", os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("failed to create data directory: %v", err)
	}

	// 生成并保存节点信息到文本文件
	err = generateAndSaveNodes(x)
	if err != nil {
		return nil, err
	}

	// 从文本文件中读取节点信息并存入nodes切片
	nodes, err := readNodesFromFile()
	if err != nil {
		return nil, err
	}

	return nodes, nil
}

// 生成节点并保存到文本文件
func generateAndSaveNodes(x int) error {
	filePath := "./data/nodelist_org.txt"

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create node list file: %v", err)
	}
	defer file.Close()

	for i := 0; i < x; i++ {
		nodeID := fmt.Sprintf("node%d", i+1)
		nodeAddr := fmt.Sprintf("localhost:%d", 9000+i+1)
		node := &Node{ID: nodeID, Addr: nodeAddr}
		_, err := fmt.Fprintf(file, "ID: %s, Addr: %s\n", node.ID, node.Addr)
		if err != nil {
			return fmt.Errorf("failed to write node to file: %v", err)
		}
	}

	return nil
}

// 从文本文件中读取节点信息并存入nodes切片
func readNodesFromFile() ([]*Node, error) {
	filePath := "./data/nodelist_org.txt"

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open node list file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	nodes := []*Node{}

	for scanner.Scan() {
		line := scanner.Text()
		nodeID, nodeAddr, err := parseNodeInfo(line)
		if err != nil {
			return nil, err
		}

		node := &Node{ID: nodeID, Addr: nodeAddr}
		nodes = append(nodes, node)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read node list file: %v", err)
	}

	return nodes, nil
}

// 解析节点信息
func parseNodeInfo(line string) (string, string, error) {
	// 假设节点信息的格式为 "ID: nodeX, Addr: localhost:port"
	idPrefix := "ID: "
	addrPrefix := ", Addr: "

	idIndex := strings.Index(line, idPrefix)
	addrIndex := strings.Index(line, addrPrefix)

	if idIndex == -1 || addrIndex == -1 {
		return "", "", fmt.Errorf("failed to parse node info: %s", line)
	}

	nodeID := line[idIndex+len(idPrefix) : addrIndex]
	nodeAddr := line[addrIndex+len(addrPrefix):]

	return nodeID, nodeAddr, nil
}
