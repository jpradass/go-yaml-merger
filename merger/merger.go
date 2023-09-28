package merger

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	HT = 0x09 // '\t' Horizontal Tab
	SP = 0x20 //      Space
	NL = 0x0a // '\n' New Line
)

func Merge(yamlpath string) ([]byte, error) {
	yaml, err := os.ReadFile(yamlpath)
	if err != nil {
		return nil, err
	}

	yaml, err = processFile(yaml, yamlpath)
	if err != nil {
		return nil, err
	}

	return yaml, nil
}

func processFile(yaml []byte, yamlpath string) ([]byte, error) {
	yamlfile, err := os.Stat(yamlpath)
	if err != nil {
		return nil, err
	}

	for 0 < strings.Count(string(yaml), "!include") {
		idx := strings.Index(string(yaml), "!include")

		include := readUntilNewLine(idx, yaml)
		space := estimateSpaceInFront(string(yaml), idx)

		includepath := strings.Split(include, " ")[1]
		_, err := os.Stat(
			filepath.Join(strings.ReplaceAll(yamlpath, yamlfile.Name(), ""), includepath),
		)
		if err != nil {
			return nil, err
		}

		yaml = append(yaml[:idx], yaml[idx+len(include)+1:]...)
		loweryaml := make([]byte, len(yaml[idx:]))
		copy(loweryaml, yaml[idx:])

		yaml, idx, err = includeFile(
			filepath.Join(strings.ReplaceAll(yamlpath, yamlfile.Name(), ""), includepath),
			idx,
			space,
			yaml[:idx],
			loweryaml,
		)
		if err != nil {
			fmt.Println("ERROR:", err.Error())
			return nil, err
		}
	}

	return yaml, nil
}

func readUntilNewLine(index int, yamlconf []byte) string {
	line := ""
	for i := index; i < len(yamlconf); i++ {
		char := yamlconf[i]
		if char == NL {
			return line
		}
		line = line + string(char)
	}

	return ""
}

func estimateSpaceInFront(yamlconf string, index int) []byte {
	var char byte = yamlconf[index-1]
	iteration := 1
	var ret []byte = make([]byte, 0)
	for char != NL {
		ret = append(ret, SP)
		char = yamlconf[index-iteration-1]
		iteration += 1
	}

	return ret
}

func includeFile(
	includepath string,
	index int,
	spaces, prependyaml, postpendyaml []byte,
) ([]byte, int, error) {
	includeFile, err := os.Open(includepath)
	if err != nil {
		return nil, 0, fmt.Errorf("error opening file %s. Error: %v", includepath, err)
	}

	// We defer the closing of the file
	defer includeFile.Close()
	firstline := true
	scanner := bufio.NewScanner(includeFile)

	for scanner.Scan() {
		line := scanner.Bytes()

		if !firstline {
			prependyaml = append(prependyaml, spaces...)
			index = index + len(spaces)
		} else {
			firstline = false
		}

		// Insert the line that we've read from the !include file
		prependyaml = append(prependyaml, line...)
		index = index + len(line)

		// Finally insert a new line char for the next line
		prependyaml = append(prependyaml, NL)
		index += 1
	}

	// We paste the lower part of the original configuration so nothing is lost
	return append(prependyaml, postpendyaml...), index, nil
}
