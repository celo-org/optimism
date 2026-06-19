package environment

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/ethereum/go-ethereum/common"
)

// LineReader is an interface that abstracts out the ability to read a whole
// line from a source.
//
// The definition is extracted from bufio.Reader, but is here for convenience
// of reference and implementation
type LineReader interface {
	ReadLine() (line []byte, isPrefix bool, err error)
}

// ansiEscapeCodeLineReader is a LineReader that removes ANSI escape codes
// from the line it reads. This is useful for cleaning up log lines that
// contain ANSI escape codes for coloring or formatting. The reader wraps
// another LineReader and processes the lines it reads to remove the escape
// codes before returning them.
type ansiEscapeCodeLineReader struct {
	r LineReader
}

// NewAnsiEscapeCodeLineReader creates a new LineReader from the LineReader
// passed in.  It removes any ANSI escape code for color formatting from
// line entries encountered.
func NewAnsiEscapeCodeLineReader(r LineReader) LineReader {
	return &ansiEscapeCodeLineReader{
		r: r,
	}
}

// ReadLine implements LineReader.  It reads a line from the underlying
// LineReader and removes any ANSI escape codes from the line.
//
// This avoids extra allocation by replacing contents from the line returned
// by the underlying LineReader with the contents of the line without
// escape codes.
func (a *ansiEscapeCodeLineReader) ReadLine() (line []byte, isPrefix bool, err error) {
	line, isPrefix, err = a.r.ReadLine()
	if err != nil {
		return line, isPrefix, err
	}

	// Go through the Escape sequence codes, and remove them from the
	// line

	i := 0
	o := 0
	for l := len(line); i < l; i, o = i+1, o+1 {
		line[o] = line[i]

		if line[i] != 0x1b {
			// this is not the escape character
			continue
		}

		if (i+1 < l) && line[i+1] != '[' {
			// We'll ignore this case for now
			continue
		}

		// We want to ignore this escape sequence
		o--

		// We have already read the ESC rune and '['
		i++
		i++

		for i < l && line[i] != 'm' {
			i++
		}
	}

	// truncate the line to the new length
	return line[:o], isPrefix, err
}

// EspressoDevNodeLogEntry represents a simple log entry from the
// Espresso Dev Node.
//
// It contains the timestamp, the logging level, the file location, and the
// rest of the message for quick reference.
//
// The Format of the log lines is anticipated to be of the following form:
// <timestamp> <level> <location>: <message>
type EspressoDevNodeLogEntry struct {
	Time     time.Time
	Level    string
	Location string
	Message  string
}

// EspressoDevNodeLogEntryReader is an interface that abstracts out the
// ability to read a log entry from a source.
type EspressoDevNodeLogEntryReader interface {
	ReadLogLine() (EspressoDevNodeLogEntry, error)
}

// espressoDevNodeLogEntryReader is a struct that will implement the
// EspressoDevNodeLogEntryReader interface.
type espressoDevNodeLogEntryReader struct {
	r LineReader
}

// NewEspressoDevNodeLogReader creates a new EspressoDevNodeLogEntryReader
// from the LineReader passed in.
func NewEspressoDevNodeLogReader(r LineReader) EspressoDevNodeLogEntryReader {
	return &espressoDevNodeLogEntryReader{
		r: r,
	}
}

func readOffsetOfCondition(line []byte, offset int, cond func(r rune) bool) int {
	i, l := offset, len(line)

	for i < l {
		r, size := utf8.DecodeRune(line[i:])
		if cond(r) {
			// We have our entry
			return i
		}

		i += size
	}

	return i

}

// isNotSpace is a helper function that will return true if the rune
// is not a space character.  This is used to skip over whitespace in the
// log line.
func isNotSpace(r rune) bool {
	return !unicode.IsSpace(r)
}

// isColon is a helper function that will return true if the rune
// is a colon character.
func isColon(r rune) bool {
	return r == ':'
}

func skipWhitespace(line []byte, offset int) int {
	// Skip the whitespace in between
	whiteSpaceBytesEnd := readOffsetOfCondition(line, offset, isNotSpace)
	return whiteSpaceBytesEnd
}

// ReadLogLine implements EspressoDevNodeLogEntryReader.
//
// It will read lines from the LineReader until it encounters a line that
// matches the expected format.  Once the expected format is encountered, and
// the values are able to be parsed into an `EspressoDevNodeLogEntry`, it will
// return the entry.
//
// If there is an error in the underlying LineReader, it will return that
// error instead of returning an entry.
func (e *espressoDevNodeLogEntryReader) ReadLogLine() (EspressoDevNodeLogEntry, error) {
	for {
		line, _, err := e.r.ReadLine()
		if err != nil {
			return EspressoDevNodeLogEntry{}, err
		}

		// Trim the spaces from the line
		line = bytes.TrimSpace(line)

		offset := 0
		var tsField, infoField, locField, messageField []byte
		{
			// Read the first field
			tsFieldStartOffset := offset
			tsFieldEndOffset := readOffsetOfCondition(line, offset, unicode.IsSpace)
			tsField = line[tsFieldStartOffset:tsFieldEndOffset]

			// Ignore the white space in between
			offset = skipWhitespace(line, tsFieldEndOffset)
		}

		// Read the second field
		{
			levelFieldStartOffset := offset
			levelFieldEndOffset := readOffsetOfCondition(line, offset, unicode.IsSpace)
			infoField = line[levelFieldStartOffset:levelFieldEndOffset]

			// Ignore the white space in between
			offset = skipWhitespace(line, levelFieldEndOffset)
		}

		{
			// Read the third field
			locFieldStartOffset := offset
			locFieldEndOffset := readOffsetOfCondition(line, offset, unicode.IsSpace)
			locField = line[locFieldStartOffset:locFieldEndOffset]

			// Ignore the white space in between
			offset = skipWhitespace(line, locFieldEndOffset)
		}

		{
			// Message field
			messageField = line[offset:]
		}

		ts, err := time.Parse(time.RFC3339Nano, string(tsField))
		if err != nil {
			// This isn't a log entry we're expecting or wanting, skip
			continue
		}

		lvl := string(infoField)
		loc := string(bytes.TrimRightFunc(locField, isColon))
		msg := string(messageField)

		entry := EspressoDevNodeLogEntry{
			Time:     ts,
			Level:    lvl,
			Location: loc,
			Message:  msg,
		}

		return entry, nil
	}
}

// There are two types of log entries we are interested in.  Deployed contract
// lines, and Server Listening Lines.

// EspressoDeployedContractLogEntry represents a log entry for a
// deployed contract.
// It contains the original log entry, and for convenience it also contains
// the Name of the contract and the address of the contract for easy access.
type EspressoDeployedContractLogEntry struct {
	Entry   EspressoDevNodeLogEntry
	Name    string
	Address common.Address
}

// handleDeployedLogEntry is a helper function that will take a log entry
// and parse it into an `EspressoDeployedContractLogEntry`.
//
// It is expected to be of the form:
// <timestamp> <level> <location>: deployed <CONTRACT_ENV_NAME> at <address>
func handleDeployedLogEntry(entry EspressoDevNodeLogEntry, fields []string) (EspressoDeployedContractLogEntry, error) {
	// deployed <CONTRACT_ENV_NAME> at <address>
	if len(fields) != 4 {
		return EspressoDeployedContractLogEntry{}, fmt.Errorf("invalid deployed entry: %s", entry.Message)
	}

	name := fields[1]
	address := common.HexToAddress(fields[3])

	return EspressoDeployedContractLogEntry{
		Entry:   entry,
		Name:    name,
		Address: address,
	}, nil
}

// EspressoStartListeningLogEntry represents a log entry for a
// server listening entry.
// It contains the original log entry, and for convenience it also contains
// the URL of the server for easy access.
type EspressoStartListeningLogEntry struct {
	Entry EspressoDevNodeLogEntry
	Url   url.URL
}

// handleServerListeningLogEntry is a helper function that will take a log entry
// and parse it into an `EspressoStartListeningLogEntry`.
//
// It is expected to be of the form:
// <timestamp> <level> <location>: Server listening on <url>
func handleServerListeningLogEntry(entry EspressoDevNodeLogEntry, fields []string) (EspressoStartListeningLogEntry, error) {
	// Server listening on <url>
	if len(fields) != 4 {
		return EspressoStartListeningLogEntry{}, fmt.Errorf("invalid server listening entry: %s", entry.Message)
	}

	rawUrl := fields[3]
	u, err := url.Parse(rawUrl)
	if err != nil {
		return EspressoStartListeningLogEntry{}, fmt.Errorf("invalid url: %s", rawUrl)
	}

	return EspressoStartListeningLogEntry{
		Entry: entry,
		Url:   *u,
	}, nil
}

// EspressoDeployedContractLogEntryReader is an interface that abstracts out
// the ability to read a log entry from a source.
type EspressoDeployedContractLogEntryReader interface {
	ReadDeployedContractLogEntry() (EspressoDeployedContractLogEntry, error)
}

// espressoDeployedContractDevNodeLogEntryReader is a struct that will
// implement the EspressoDeployedContractLogEntryReader interface.
type espressoDeployedContractDevNodeLogEntryReader struct {
	r                        EspressoDevNodeLogEntryReader
	numStartListeningEntries int
}

// NewEspressoDeployedContractLogEntryReader creates a new
// EspressoDeployedContractLogEntryReader from the
// EspressoDevNodeLogEntryReader passed in.
func NewEspressoDeployedContractLogEntryReader(r EspressoDevNodeLogEntryReader) EspressoDeployedContractLogEntryReader {
	return &espressoDeployedContractDevNodeLogEntryReader{
		r: r,
	}
}

// MAX_NUM_STARTING_ENTIRES represents the maximum number of starting entries
// we expect to encounter before we consider ourselves "started"
const MAX_NUM_STARTING_ENTRIES = 4

// ReadDeployedContractLogEntry implements EspressoDeployedContractLogEntryReader.
//
// It will read entries from the EspressoDevNodeLogEntryReader until it
// encounters an entry that matches the expected deployed contract entry.
// Once the expected format is encountered, and the values are able to
// be parsed into an `EspressoDeployedContractLogEntry`, it will return the
// entry.
//
// If an error is encountered from the underlying EspressoDevNodeLogEntryReader,
// it will return that error instead of returning an entry.
//
// If more than 4 server listening entries are encountered, it will return
// an EOF error. This is an anticipated condition for the Espresso Dev Node
// being started.
func (e *espressoDeployedContractDevNodeLogEntryReader) ReadDeployedContractLogEntry() (EspressoDeployedContractLogEntry, error) {
	for {
		if e.numStartListeningEntries >= MAX_NUM_STARTING_ENTRIES {
			return EspressoDeployedContractLogEntry{}, io.EOF
		}

		entry, err := e.r.ReadLogLine()
		if err != nil {
			return EspressoDeployedContractLogEntry{}, err
		}

		fields := strings.Fields(entry.Message)
		// <message> = deployed <CONTRACT_ENV_NAME> at <address>
		// <message> = Server listening on <url>

		if len(fields) < 4 {
			// This isn't a log entry we're considering at the moment, skip it
			continue
		}

		switch {
		default:
			// This isn't a log entry we're considering at the moment, skip it
			continue

		case strings.HasPrefix(strings.ToLower(entry.Message), "deployed"):
			// We have a deployed contract entry
			return handleDeployedLogEntry(entry, fields)

		case strings.HasPrefix(strings.ToLower(entry.Message), "server listening on"):
			// We have a server listening entry
			_, err := handleServerListeningLogEntry(entry, fields)
			if err == nil {
				e.numStartListeningEntries++
			}
			// Skip this entry
			continue
		}
	}
}
