package counter

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"time"
	"os"
)

func TestNewCounter(t *testing.T) {
	cntr := NewCounter(5, "Go")
	assert.NotNil(t, cntr)
}

func TestCounter_ScanFile(t *testing.T) {

	cntr := NewCounter(5, "Go")
	assert.NotNil(t, cntr)
	input, output, err := os.Pipe()
	if err != nil {
		t.Error(err)
	}
	testCntNum := 5
	defer output.Close()
	for i := 0; i < testCntNum; i++ {
		output.WriteString("https://golang.org\n")
	}

	cntr.ScanFile(input)
	time.Sleep(time.Second)

	assert.Equal(t, testCntNum, cntr.GetURLsChanLen())
}

func TestCounter_Count(t *testing.T) {
	cntr := NewCounter(5, "Go")
	assert.NotNil(t, cntr)
	input, output, err := os.Pipe()
	if err != nil {
		t.Error(err)
	}
	testCntNum := 5
	defer output.Close()
	for i := 0; i < testCntNum; i++ {
		output.WriteString("https://golang.org\n")
	}

	cntr.ScanFile(input)
	time.Sleep(time.Second)

	assert.Equal(t, testCntNum, cntr.GetURLsChanLen())

	cntr.Count()
	assert.Equal(t, 9 * 5, cntr.total)
}
