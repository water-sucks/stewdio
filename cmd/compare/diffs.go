package compare

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"

	"github.com/go-audio/wav"
)

func GenerateDiffs(oldFile, newFile *os.File, outputPath string) error {
	oldDecoder := wav.NewDecoder(oldFile)
	oldAudioBuf, err := oldDecoder.FullPCMBuffer()
	if err != nil {
		return err
	}

	newDecoder := wav.NewDecoder(newFile)
	newAudioBuf, err := newDecoder.FullPCMBuffer()
	if err != nil {
		return err
	}

	if oldDecoder.BitDepth != newDecoder.BitDepth {
		return fmt.Errorf("bit depth mismatch: old file is %d-bit, new file is %d-bit", oldDecoder.BitDepth, newDecoder.BitDepth)
	}

	additions, subtractions, offsets := calculateDiffs(oldAudioBuf.Data, newAudioBuf.Data, int(oldDecoder.BitDepth), getFormat(oldDecoder))

	additionsLength := len(additions) / bitDepthToBytes(int(oldDecoder.BitDepth))
	subtractionsLength := len(subtractions) / bitDepthToBytes(int(oldDecoder.BitDepth))

	oldFileName := filepath.Base(oldFile.Name())

	if additionsLength > 0 {
		additionsFileName := fmt.Sprintf("%s/%s_a_offset%d_len%d.bin", outputPath, oldFileName, offsets["additions"], len(additions))
		err = os.WriteFile(additionsFileName, additions, 0644)
		if err != nil {
			return err
		}
	}

	if subtractionsLength > 0 {
		subtractionsFileName := fmt.Sprintf("%s/%s_s_offset%d_len%d.bin", outputPath, oldFileName, offsets["subtractions"], len(subtractions))
		err = os.WriteFile(subtractionsFileName, subtractions, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

func getFormat(decoder *wav.Decoder) string {
	if decoder.BitDepth == 32 && decoder.WavAudioFormat == 3 {
		return "float"
	}
	return "pcm"
}

func calculateDiffs(oldData, newData []int, bitDepth int, format string) ([]byte, []byte, map[string]int) {
	additions := make([]byte, 0)
	subtractions := make([]byte, 0)
	offsets := map[string]int{"additions": -1, "subtractions": -1}

	oldLen := len(oldData)
	newLen := len(newData)
	maxLen := max(newLen, oldLen)

	for i := 0; i < maxLen; i++ {
		if i < oldLen && i < newLen {
			if oldData[i] != newData[i] {
				if offsets["additions"] == -1 {
					offsets["additions"] = i * bitDepthToBytes(bitDepth)
				}
				if offsets["subtractions"] == -1 {
					offsets["subtractions"] = i * bitDepthToBytes(bitDepth)
				}
				additions = append(additions, intToBytes(newData[i], bitDepth, format)...)
				subtractions = append(subtractions, intToBytes(oldData[i], bitDepth, format)...)
			}
		} else if i < oldLen {
			if offsets["subtractions"] == -1 {
				offsets["subtractions"] = i * bitDepthToBytes(bitDepth)
			}
			subtractions = append(subtractions, intToBytes(oldData[i], bitDepth, format)...)
		} else {
			if offsets["additions"] == -1 {
				offsets["additions"] = i * bitDepthToBytes(bitDepth)
			}
			additions = append(additions, intToBytes(newData[i], bitDepth, format)...)
		}
	}

	return additions, subtractions, offsets
}

func intToBytes(value int, bitDepth int, format string) []byte {
	switch bitDepth {
	case 16:
		buf := make([]byte, 2)
		binary.LittleEndian.PutUint16(buf, uint16(value))
		return buf
	case 32:
		buf := make([]byte, 4)
		if format == "float" {
			binary.LittleEndian.PutUint32(buf, math.Float32bits(float32(value)))
		} else {
			binary.LittleEndian.PutUint32(buf, uint32(value))
		}
		return buf
	default:
		return nil
	}
}

func bitDepthToBytes(bitDepth int) int {
	switch bitDepth {
	case 16:
		return 2
	case 32:
		return 4
	default:
		return 0
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
