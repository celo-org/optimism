package environment_test

import (
	"bufio"
	"io"
	"strings"
	"testing"
	"time"

	env "github.com/ethereum-optimism/optimism/espresso/environment"
	"github.com/ethereum/go-ethereum/common"
)

type SingleLineReader struct {
	Line string
}

func (e *SingleLineReader) ReadLine() (line []byte, isPrefix bool, err error) {
	return []byte(e.Line), false, nil
}

// TestAnsiEscapeCodeLineReader tests the AnsiEscapeCodeLineReader in order
// to ensure that it removes ANSI escape codes from the line it reads.
// correctly.
func TestAnsiEscapeCodeLineReader(t *testing.T) {
	{
		// Correctly prunes the lien
		example := "  \x1b[2m2025-04-03T17:13:35.610901Z\x1b[0m \x1b[32m INFO\x1b[0m \x1b[1;32msequencer::context\x1b[0m\x1b[32m: \x1b[32mproposal missing from storage; fetching from network: proposal ViewNumber(59) not available, \x1b[1;32mview\x1b[0m\x1b[32m: ViewNumber(59), \x1b[1;32mleaf\x1b[0m\x1b[32m: COMMIT~5j6HMckzhEdAh_l02alyHAjD72NsLGozO7Um7cYsSsSa\x1b[0m"
		line, _, _ := env.NewAnsiEscapeCodeLineReader(&SingleLineReader{example}).ReadLine()

		if have, want := string(line), "  2025-04-03T17:13:35.610901Z  INFO sequencer::context: proposal missing from storage; fetching from network: proposal ViewNumber(59) not available, view: ViewNumber(59), leaf: COMMIT~5j6HMckzhEdAh_l02alyHAjD72NsLGozO7Um7cYsSsSa"; have != want {
			t.Errorf("failed to remove ANSI escape codes:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
		}
	}

	{
		// Correctly does not alter the line
		example := "hello\n\tworld"
		line, _, _ := env.NewAnsiEscapeCodeLineReader(&SingleLineReader{example}).ReadLine()

		if have, want := string(line), example; have != want {
			t.Errorf("failed to not alter the example line:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
		}
	}

	{
		// Correctly ignores escape sequences that do not conform to the expected
		// color escape sequence
		example := "hello\x1bworld"
		line, _, _ := env.NewAnsiEscapeCodeLineReader(&SingleLineReader{example}).ReadLine()

		if have, want := string(line), example; have != want {
			t.Errorf("failed to not alter the example line:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
		}
	}

	{
		// Correctly detects the escape sequence, and returns the truncated line
		// without a runtime error.
		example := "Hello World\x1b[5"
		line, _, _ := env.NewAnsiEscapeCodeLineReader(&SingleLineReader{example}).ReadLine()

		if have, want := string(line), "Hello World"; have != want {
			t.Errorf("failed to not error reading a truncated escape sequence:\nhave:\n\t\"%v\"\nwant:\n\t\"%v\"\n", have, want)
		}
	}
}

func TestEspressoDevNodeLogEntry(t *testing.T) {
	exampleLogs := `
	foo
	bar
		2025-04-01T17:13:35.610901Z INFO espresso_dev_node_logs_test.go: foo bar baz
	other
	garbage
	`

	reader := env.NewEspressoDevNodeLogReader(env.NewAnsiEscapeCodeLineReader(bufio.NewReader(strings.NewReader(exampleLogs))))

	{
		// Read the First Line Entry
		entry, err := reader.ReadLogLine()
		if have, want := err, error(nil); have != want {
			t.Errorf("failed to read log line:\nhave:\n\t%v\nwant:\n\t%v\n", have, want)
		}

		if have, want := entry.Time, time.Date(2025, 4, 1, 17, 13, 35, 610901000, time.UTC); have != want {
			t.Errorf("failed to parse log line time:\nhave:\n\t%v\nwant:\n\t%v\n", have, want)
		}

		if have, want := entry.Level, "INFO"; have != want {
			t.Errorf("failed to parse log line level:\nhave:\n\t%v\nwant:\n\t%v\n", have, want)
		}

		if have, want := entry.Location, "espresso_dev_node_logs_test.go"; have != want {
			t.Errorf("failed to parse log line location:\nhave:\n\t%v\nwant:\n\t%v\n", have, want)
		}

		if have, want := entry.Message, "foo bar baz"; have != want {
			t.Errorf("failed to parse log line message:\nhave:\n\t%v\nwant:\n\t%v\n", have, want)
		}
	}

	_, err := reader.ReadLogLine()
	if have, want := err, io.EOF; have != want {
		t.Errorf("expecting next ReadLogLine to return EOF:\nhave:\n\t%v\nwant:\n\t%v\n", have, want)
	}
}

func TestEspressoDeployedContractLogEntryReader(t *testing.T) {
	exampleLogs := `
	2025-04-01T17:13:35.610901Z  INFO some::rust::library.rs: Server listening on http://0.0.0.0:1234


	2025-04-01T17:13:35.610901Z  INFO some::rust::library.rs: deployed SOME_WICKED_CONTRACT at 0x1234567890abcdef1234567890abcdef12345678

	2025-04-01T17:13:35.610901Z  INFO some::rust::library.rs: Server listening on http://0.0.0.0:1234
	2025-04-01T17:13:35.610901Z  INFO some::rust::library.rs: Server listening on http://0.0.0.0:1234
	2025-04-01T17:13:35.610901Z  INFO some::rust::library.rs: deployed SOME_WICKED_CONTRACT_2 at 0x1234567890abcdef1234567890abcdef12345679
	2025-04-01T17:13:35.610901Z  INFO some::rust::library.rs: Server listening on http://0.0.0.0:1234

	2025-04-01T17:13:35.610901Z  INFO some::rust::library.rs: deployed SOME_WICKED_CONTRACT_3 at 0x1234567890abcdef1234567890abcdef1234567a
	`

	reader := env.NewEspressoDeployedContractLogEntryReader(env.NewEspressoDevNodeLogReader(env.NewAnsiEscapeCodeLineReader(bufio.NewReader(strings.NewReader(exampleLogs)))))

	{
		// Read the First Deployed Contract Entry
		entry, err := reader.ReadDeployedContractLogEntry()
		if have, want := err, error(nil); have != want {
			t.Errorf("failed to read log line:\nhave:\n\t%v\nwant:\n\t%v\n", have, want)
		}

		if have, want := entry.Name, "SOME_WICKED_CONTRACT"; have != want {
			t.Errorf("failed to parse log line name:\nhave:\n\t%v\nwant:\n\t%v\n", have, want)
		}

		if have, want := entry.Address, common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678"); have.Cmp(want) != 0 {
			t.Errorf("failed to parse log line address:\nhave:\n\t%s\nwant:\n\t%s\n", have, want)
		}
	}

	{
		// Read the Second Deployed Contract Entry
		entry, err := reader.ReadDeployedContractLogEntry()
		if have, want := err, error(nil); have != want {
			t.Errorf("failed to read log line:\nhave:\n\t%v\nwant:\n\t%v\n", have, want)
		}

		if have, want := entry.Name, "SOME_WICKED_CONTRACT_2"; have != want {
			t.Errorf("failed to parse deployed contract entry, name:\nhave:\n\t%v\nwant:\n\t%v\n", have, want)
		}

		if have, want := entry.Address, common.HexToAddress("0x1234567890abcdef1234567890abcdef12345679"); have.Cmp(want) != 0 {
			t.Errorf("failed to parse deployed contract entry address:\nhave:\n\t%s\nwant:\n\t%s\n", have, want)
		}
	}

	{
		// Read the Second Deployed Contract Entry
		_, err := reader.ReadDeployedContractLogEntry()
		if have, want := err, io.EOF; have != want {
			t.Errorf("should have received EOF, instead received:\nhave:\n\t%v\nwant:\n\t%v\n", have, want)
		}
	}

}
