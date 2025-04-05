package compare

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"

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

	additions, subtractions := calculateDiffs(oldAudioBuf.Data, newAudioBuf.Data, int(oldDecoder.BitDepth), getFormat(oldDecoder))

	err = os.WriteFile(fmt.Sprintf("%s/additions.bin", outputPath), additions, 0644)
	if err != nil {
		return err
	}

	err = os.WriteFile(fmt.Sprintf("%s/subtractions.bin", outputPath), subtractions, 0644)
	if err != nil {
		return err
	}

	return nil
}

func getFormat(decoder *wav.Decoder) string {
	if decoder.BitDepth == 32 && decoder.WavAudioFormat == 3 {
		return "float"
	}
	return "pcm"
}

func calculateDiffs(oldData, newData []int, bitDepth int, format string) ([]byte, []byte) {
	additions := make([]byte, 0)
	subtractions := make([]byte, 0)

	oldLen := len(oldData)
	newLen := len(newData)
	maxLen := max(newLen, oldLen)

	for i := 0; i < maxLen; i++ {
		if i < oldLen && i < newLen {
			if oldData[i] != newData[i] {
				additions = append(additions, intToBytes(newData[i], bitDepth, format)...)
				subtractions = append(subtractions, intToBytes(oldData[i], bitDepth, format)...)
			}
		} else if i < oldLen {
			subtractions = append(subtractions, intToBytes(oldData[i], bitDepth, format)...)
		} else {
			additions = append(additions, intToBytes(newData[i], bitDepth, format)...)
		}
	}

	return additions, subtractions
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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
