package main

import (
	"bytes"
	"fmt"
	"strings"
)

const (
	tfSeparator = " "
)

type tableCellText struct {
	text             string
	themeComponentID ThemeComponentID
}

type tableCell struct {
	textEntries []tableCellText
}

// TableFormatter renders provided data in a tabular layout
type TableFormatter struct {
	config       Config
	maxColWidths []uint
	cells        [][]tableCell
}

// NewTableFormatter creates a new instance of the table formatter supporting the specified number of columns
func NewTableFormatter(cols uint, config Config) *TableFormatter {
	return &TableFormatter{
		maxColWidths: make([]uint, cols),
		config:       config,
	}
}

// Rows returns the number of rows in the table formatter
func (tableFormatter *TableFormatter) Rows() uint {
	return uint(len(tableFormatter.cells))
}

// Cols returns the number of cols in the table formatter
func (tableFormatter *TableFormatter) Cols() uint {
	if tableFormatter.Rows() > 0 {
		return uint(len(tableFormatter.cells[0]))
	}

	return 0
}

// Resize updates the number of rows the tableformatter can store
func (tableFormatter *TableFormatter) Resize(newRows uint) {
	rows := tableFormatter.Rows()

	if rows == newRows {
		return
	}

	cols := len(tableFormatter.maxColWidths)

	tableFormatter.cells = make([][]tableCell, newRows)
	for rowIndex := range tableFormatter.cells {
		tableFormatter.cells[rowIndex] = make([]tableCell, cols)
	}
}

// Clear text in all cells
func (tableFormatter *TableFormatter) Clear() {
	for rowIndex := range tableFormatter.cells {
		for colIndex := range tableFormatter.cells[rowIndex] {
			tableFormatter.cells[rowIndex][colIndex].textEntries = nil
		}
	}
}

// SetCell sets the text value of the cell at the specified coordinates
func (tableFormatter *TableFormatter) SetCell(rowIndex, colIndex uint, format string, args ...interface{}) (err error) {
	return tableFormatter.SetCellWithStyle(rowIndex, colIndex, CmpNone, format, args...)
}

// SetCellWithStyle sets text with style information value of the cell at the specified coordinates
func (tableFormatter *TableFormatter) SetCellWithStyle(rowIndex, colIndex uint, themeComponentID ThemeComponentID, format string, args ...interface{}) (err error) {
	if !(rowIndex < tableFormatter.Rows() && colIndex < tableFormatter.Cols()) {
		return fmt.Errorf("Invalid rowIndex (%v), colIndex (%v) for dimensions rows (%v), cols (%v)",
			rowIndex, colIndex, tableFormatter.Rows(), tableFormatter.Cols())
	}

	tableCell := &tableFormatter.cells[rowIndex][colIndex]

	tableCell.textEntries = []tableCellText{
		{
			text:             fmt.Sprintf(format, args...),
			themeComponentID: themeComponentID,
		},
	}

	return
}

// AppendToCell appends text to the specified cell
func (tableFormatter *TableFormatter) AppendToCell(rowIndex, colIndex uint, format string, args ...interface{}) (err error) {
	return tableFormatter.AppendToCellWithStyle(rowIndex, colIndex, CmpNone, format, args...)
}

// AppendToCellWithStyle appends text with style information to the specified cell
func (tableFormatter *TableFormatter) AppendToCellWithStyle(rowIndex, colIndex uint, themeComponentID ThemeComponentID, format string, args ...interface{}) (err error) {
	if !(rowIndex < tableFormatter.Rows() && colIndex < tableFormatter.Cols()) {
		return fmt.Errorf("Invalid rowIndex (%v), colIndex (%v) for dimensions rows (%v), cols (%v)",
			rowIndex, colIndex, tableFormatter.Rows(), tableFormatter.Cols())
	}

	tableCell := &tableFormatter.cells[rowIndex][colIndex]

	tableCell.textEntries = append(tableCell.textEntries, tableCellText{
		text:             fmt.Sprintf(format, args...),
		themeComponentID: themeComponentID,
	})

	return
}

// RowString returns the string representation of the row at the specified index
func (tableFormatter *TableFormatter) RowString(rowIndex uint) (rowString string, err error) {
	if rowIndex >= tableFormatter.Rows() {
		err = fmt.Errorf("Invalid rowIndex: %v, total rows %v", rowIndex, tableFormatter.Rows())
		return
	}

	var buf bytes.Buffer

	for colIndex := range tableFormatter.cells[rowIndex] {
		for _, textEntry := range tableFormatter.cells[rowIndex][colIndex].textEntries {
			buf.WriteString(textEntry.text)
		}

		buf.WriteString(tfSeparator)
	}

	rowString = buf.String()
	return
}

// Render pads the content of the table formatter and writes it to the provided window
func (tableFormatter *TableFormatter) Render(win RenderWindow, viewStartColumn uint, border bool) (err error) {
	if err = tableFormatter.PadCells(border); err != nil {
		return
	}

	var lineBuilder *LineBuilder

	for rowIndex := range tableFormatter.cells {
		adjustedRowIndex := uint(rowIndex)
		if border {
			adjustedRowIndex++
		}

		if lineBuilder, err = win.LineBuilder(adjustedRowIndex, viewStartColumn); err != nil {
			return
		}

		if border {
			lineBuilder.Append(" ")
		}

		for colIndex := range tableFormatter.cells[rowIndex] {
			tableCell := &tableFormatter.cells[rowIndex][colIndex]

			for _, textEntry := range tableCell.textEntries {
				lineBuilder.AppendWithStyle(textEntry.themeComponentID, "%v", textEntry.text)
			}

			lineBuilder.Append(tfSeparator)
		}
	}

	return
}

// PadCells pads each cell with whitespace so that the text in each column is of uniform width
func (tableFormatter *TableFormatter) PadCells(border bool) (err error) {
	tableFormatter.determineMaxColWidths(border)

	for rowIndex := range tableFormatter.cells {
		column := uint(1)

		if border {
			column++
		}

		for colIndex := range tableFormatter.cells[rowIndex] {
			width := tableFormatter.textWidth(rowIndex, colIndex, column)
			maxColWidth := tableFormatter.maxColWidths[colIndex]

			if width < maxColWidth {
				if err = tableFormatter.AppendToCell(uint(rowIndex), uint(colIndex), strings.Repeat(" ", int(maxColWidth-width))); err != nil {
					return
				}
			}

			column += maxColWidth + uint(len(tfSeparator))
		}
	}

	return
}

func (tableFormatter *TableFormatter) determineMaxColWidths(border bool) {
	for colIndex := 0; colIndex < len(tableFormatter.maxColWidths); colIndex++ {
		column := uint(1)

		if border {
			column++
		}

		for doneColIndex := 0; doneColIndex < colIndex; doneColIndex++ {
			column += tableFormatter.maxColWidths[doneColIndex]
			column += uint(len(tfSeparator))
		}

		for rowIndex := range tableFormatter.cells {
			width := tableFormatter.textWidth(rowIndex, colIndex, column)

			if width > tableFormatter.maxColWidths[colIndex] {
				tableFormatter.maxColWidths[colIndex] = width
			}
		}
	}

}

func (tableFormatter *TableFormatter) textWidth(rowIndex, colIndex int, column uint) (width uint) {
	textEntries := tableFormatter.cells[rowIndex][colIndex].textEntries

	for _, textEntry := range textEntries {
		for _, codePoint := range textEntry.text {
			renderedCodePoints := DetermineRenderedCodePoint(codePoint, column, tableFormatter.config)

			for _, renderedCodePoint := range renderedCodePoints {
				width += renderedCodePoint.width
				column += renderedCodePoint.width
			}
		}
	}

	return
}
