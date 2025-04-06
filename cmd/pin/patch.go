package pin

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

func ApplyPatch(targetFilePath, patchFilePath string) error {
	parts := strings.Split(patchFilePath, "_")
	if len(parts) != 4 {
		return fmt.Errorf("invalid patch file name format")
	}

	operation := parts[1]
	offsetStr := strings.TrimPrefix(parts[2], "offset")
	lengthStr := strings.TrimPrefix(parts[3], "len")
	lengthStr = strings.TrimSuffix(lengthStr, ".bin")

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		return fmt.Errorf("invalid offset: %v", err)
	}

	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return fmt.Errorf("invalid length: %v", err)
	}

	patchData, err := ioutil.ReadFile(patchFilePath)
	if err != nil {
		return fmt.Errorf("failed to read patch file: %v", err)
	}

	targetFile, err := os.OpenFile(targetFilePath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open target file: %v", err)
	}
	defer targetFile.Close()

	switch operation {
	case "a":
		targetData, err := ioutil.ReadAll(targetFile)
		if err != nil {
			return fmt.Errorf("failed to read target file: %v", err)
		}

		newData := append(targetData[:offset], append(patchData, targetData[offset:]...)...)
		if err := ioutil.WriteFile(targetFilePath, newData, 0644); err != nil {
			return fmt.Errorf("failed to write patched file: %v", err)
		}

	case "s":
		targetData, err := ioutil.ReadAll(targetFile)
		if err != nil {
			return fmt.Errorf("failed to read target file: %v", err)
		}

		newData := append(targetData[:offset], targetData[offset+length:]...)
		if err := ioutil.WriteFile(targetFilePath, newData, 0644); err != nil {
			return fmt.Errorf("failed to write patched file: %v", err)
		}

	default:
		return fmt.Errorf("invalid operation: %s", operation)
	}

	return nil
}
