package output

import (
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
)

var (
	// Success prints text in bright green
	Success = color.New(color.FgHiGreen, color.Bold).SprintFunc()
	// Error prints text in bright red
	Error = color.New(color.FgHiRed, color.Bold).SprintFunc()
	// Warning prints text in bright yellow
	Warning = color.New(color.FgHiYellow).SprintFunc()
	// Info prints text in bright blue
	Info = color.New(color.FgHiBlue).SprintFunc()
	// Highlight prints text in bright cyan
	Highlight = color.New(color.FgHiCyan).SprintFunc()
	// Bold prints text in bold
	Bold = color.New(color.Bold).SprintFunc()
	// Underline prints text with underline
	Underline = color.New(color.Underline).SprintFunc()
	// Header prints text in magenta and bold
	Header = color.New(color.FgHiMagenta, color.Bold).SprintFunc()
)

// ColorWriter is a wrapper around io.Writer that supports color output
type ColorWriter struct {
	Writer io.Writer
}

// NewColorWriter creates a new ColorWriter
func NewColorWriter(w io.Writer) *ColorWriter {
	return &ColorWriter{Writer: w}
}

// Printf prints a formatted string to the writer
func (w *ColorWriter) Printf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(w.Writer, format, a...)
}

// Println prints a line to the writer
func (w *ColorWriter) Println(a ...interface{}) (n int, err error) {
	return fmt.Fprintln(w.Writer, a...)
}

// SuccessPrintf prints a success message
func (w *ColorWriter) SuccessPrintf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(w.Writer, "%s", Success(fmt.Sprintf(format, a...)))
}

// SuccessPrintln prints a success message with a newline
func (w *ColorWriter) SuccessPrintln(a ...interface{}) (n int, err error) {
	return fmt.Fprintln(w.Writer, Success(fmt.Sprint(a...)))
}

// ErrorPrintf prints an error message
func (w *ColorWriter) ErrorPrintf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(w.Writer, "%s", Error(fmt.Sprintf(format, a...)))
}

// ErrorPrintln prints an error message with a newline
func (w *ColorWriter) ErrorPrintln(a ...interface{}) (n int, err error) {
	return fmt.Fprintln(w.Writer, Error(fmt.Sprint(a...)))
}

// WarningPrintf prints a warning message
func (w *ColorWriter) WarningPrintf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(w.Writer, "%s", Warning(fmt.Sprintf(format, a...)))
}

// WarningPrintln prints a warning message with a newline
func (w *ColorWriter) WarningPrintln(a ...interface{}) (n int, err error) {
	return fmt.Fprintln(w.Writer, Warning(fmt.Sprint(a...)))
}

// InfoPrintf prints an info message
func (w *ColorWriter) InfoPrintf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(w.Writer, "%s", Info(fmt.Sprintf(format, a...)))
}

// InfoPrintln prints an info message with a newline
func (w *ColorWriter) InfoPrintln(a ...interface{}) (n int, err error) {
	return fmt.Fprintln(w.Writer, Info(fmt.Sprint(a...)))
}

// HighlightPrintf prints a highlighted message
func (w *ColorWriter) HighlightPrintf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(w.Writer, "%s", Highlight(fmt.Sprintf(format, a...)))
}

// HighlightPrintln prints a highlighted message with a newline
func (w *ColorWriter) HighlightPrintln(a ...interface{}) (n int, err error) {
	return fmt.Fprintln(w.Writer, Highlight(fmt.Sprint(a...)))
}

// BoldPrintf prints a bold message
func (w *ColorWriter) BoldPrintf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(w.Writer, "%s", Bold(fmt.Sprintf(format, a...)))
}

// BoldPrintln prints a bold message with a newline
func (w *ColorWriter) BoldPrintln(a ...interface{}) (n int, err error) {
	return fmt.Fprintln(w.Writer, Bold(fmt.Sprint(a...)))
}

// HeaderPrintf prints a header message
func (w *ColorWriter) HeaderPrintf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(w.Writer, "%s", Header(fmt.Sprintf(format, a...)))
}

// HeaderPrintln prints a header message with a newline
func (w *ColorWriter) HeaderPrintln(a ...interface{}) (n int, err error) {
	return fmt.Fprintln(w.Writer, Header(fmt.Sprint(a...)))
}

// Default is a ColorWriter that writes to os.Stdout
var Default = NewColorWriter(os.Stdout)

// DisableColors disables all colors
func DisableColors() {
	color.NoColor = true
}

// EnableColors enables all colors
func EnableColors() {
	color.NoColor = false
}

// MakeBold formats a string in bold
func MakeBold(text string) string {
	return Bold(text)
}

// MakeUnderline formats a string with underline
func MakeUnderline(text string) string {
	return Underline(text)
}

// MakeHeader formats a string as a header
func MakeHeader(text string) string {
	return Header(text)
}
