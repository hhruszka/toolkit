package bptcommon

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/xuri/excelize/v2"
)

// _col converts a column number to its corresponding Excel column letter. It panics if an error occurs during conversion.
func _col(col int) string {
	colStr, err := excelize.ColumnNumberToName(col)
	if err != nil {
		panic(err)
	}
	return colStr
}

// _cell converts column and row numbers into an Excel cell reference string. It panics if the conversion fails.
func _cell(col int, row int) string {
	cell, err := excelize.CoordinatesToCellName(col, row)
	if err != nil {
		panic(err)
	}
	return cell
}

// cellLen calculates and returns the width of the widest line in a cell's value, adding a margin of 2 characters.
func cellLen(cellValue string) float64 {
	lines := strings.Split(cellValue, "\n")
	var width float64

	for _, line := range lines {
		width = max(width, float64(utf8.RuneCountInString(line)+2))
	}
	return width
}

// getFontSize retrieves the font size of a specified cell in a given Excel sheet.
// It returns the font size as a float64 or an error if the operation fails.
func getFontSize(xlsxFile *excelize.File, sheetName string, cell string) (float64, error) {
	var styleId int
	var style *excelize.Style
	var err error

	styleId, err = xlsxFile.GetCellStyle(sheetName, cell)
	if err != nil {
		return 0, err
	}
	style, err = xlsxFile.GetStyle(styleId)
	if err != nil {
		return 0, err
	}

	if style.Font == nil {
		return 11, nil // Return default size if none is set
	}
	return style.Font.Size, nil
}

// setColWidth adjusts the width of specified columns in an Excel sheet based on the longest content in the given range.
// It calculates the optimal width for each column and sets it, ensuring the width does not exceed the maximum allowable.
// Parameters include the Excel file, sheet name, start and end columns, and start and end rows for the range.
// Returns an error if any issues occur during the adjustment of column widths.
func setColWidth(xlsxFile *excelize.File, sheetName string, startCol int, endCol int, startRow int, endRow int) error {
	var maxColWidth = make([]float64, endCol+1)
	var cellValue string
	var err error

	for col := startCol; col <= endCol; col++ {
		for row := startRow; row <= endRow; row++ {
			cellValue, err = xlsxFile.GetCellValue(sheetName, _cell(col, row))
			if err != nil {
				return err
			}

			maxColWidth[col] = max(maxColWidth[col], cellLen(cellValue))
		}
	}

	for col, width := range maxColWidth {
		if width == 0 {
			continue
		}

		if width > excelize.MaxColumnWidth {
			width = excelize.MaxColumnWidth
		}

		err = xlsxFile.SetColWidth(sheetName, _col(col), _col(col), width)
		if err != nil {
			return err
		}
	}
	return err
}

// setColStyle applies a predefined style to a range of columns in the specified sheet of an Excel file.
// xlsxFile is the Excel file object.
// sheetName is the name of the worksheet where the style will be applied.
// startCol and endCol define the range of columns to style.
// startRow and endRow are not currently used but reserved for potential future functionality.
// Returns an error if the style creation or application fails.
func setColStyle(xlsxFile *excelize.File, sheetName string, startCol int, endCol int, startRow int, endRow int) error {
	style := &excelize.Style{
		Border:        nil,
		Fill:          excelize.Fill{},
		Font:          nil,
		Alignment:     &excelize.Alignment{Horizontal: "left", Vertical: "top", WrapText: true},
		Protection:    nil,
		NumFmt:        0,
		DecimalPlaces: nil,
		CustomNumFmt:  nil,
		NegRed:        false,
	}

	styleId, err := xlsxFile.NewStyle(style)
	if err != nil {
		return err
	}

	err = xlsxFile.SetColStyle(sheetName, fmt.Sprintf("%s:%s", _col(startCol), _col(endCol)), styleId)
	return err
}

const StyleDoesNotExist = -1 // if style does not exist excelize returns -1

var setStyleOnce sync.Once
var styleWrappedId int = StyleDoesNotExist
var styleNotWrappedId int = StyleDoesNotExist

func SetDefaultStyles(xlsxFile *excelize.File) {
	setStyleOnce.Do(func() {
		font := excelize.Font{
			Size:   11,
			Family: "Calibri",
		}
		styleWrapped := &excelize.Style{
			Border:        nil,
			Fill:          excelize.Fill{},
			Font:          &font,
			Alignment:     &excelize.Alignment{Horizontal: "left", Vertical: "top", WrapText: true},
			Protection:    nil,
			NumFmt:        0,
			DecimalPlaces: nil,
			CustomNumFmt:  nil,
			NegRed:        false,
		}

		styleNotWrapped := &excelize.Style{
			Border:        nil,
			Fill:          excelize.Fill{},
			Font:          &font,
			Alignment:     &excelize.Alignment{Horizontal: "left", Vertical: "top", WrapText: false},
			Protection:    nil,
			NumFmt:        0,
			DecimalPlaces: nil,
			CustomNumFmt:  nil,
			NegRed:        false,
		}

		styleNotWrappedId, _ = xlsxFile.NewStyle(styleNotWrapped)
		styleWrappedId, _ = xlsxFile.NewStyle(styleWrapped)
	})
}

func _Wrapped(v any) any {
	if styleWrappedId == StyleDoesNotExist {
		panic("SetDefaultStyles must be called before using _Wrapped")
	}
	return excelize.Cell{StyleID: styleWrappedId, Value: v}
}

func _notWrapped(v any) any {
	if styleNotWrappedId == StyleDoesNotExist {
		panic("SetDefaultStyles must be called before using _notWrapped")
	}
	return excelize.Cell{StyleID: styleNotWrappedId, Value: v}
}

// formatCols applies column styles and adjusts column widths for a specified range in an Excel sheet.
// It takes an `xlsxFile`, a `sheetName`, starting and ending column indices, and starting and ending row indices as parameters.
// Returns an error if setting styles or widths fails.
func formatCols(xlsxFile *excelize.File, sheetName string, startCol int, endCol int, startRow int, endRow int) error {
	err := setColStyle(xlsxFile, sheetName, startCol, endCol, startRow, endRow)
	if err != nil {
		return err
	}
	err = setColWidth(xlsxFile, sheetName, startCol, endCol, startRow, endRow)
	return err
}

func calculateWidth(s string, fontSize float64) float64 {
	if fontSize == 0 {
		fontSize = 11 // Default Excel font size
	}

	// 1. Calculate base pixels for Calibri 11 (as discussed before)
	basePixels := 0.0
	for _, r := range s {
		switch {
		case r >= 'A' && r <= 'Z':
			basePixels += 9 // average upper
		case r >= 'a' && r <= 'z':
			basePixels += 7 // average lower
		case r >= '0' && r <= '9':
			basePixels += 7
		default:
			basePixels += 8
		}
	}

	// 2. Scale based on font size ratio
	// (CurrentSize / 11.0)
	scaleFactor := fontSize / 11.0
	scaledPixels := basePixels * scaleFactor

	// 3. Add padding and convert to Excel Width Units
	// (Pixels + 5px padding) / 7px_per_unit
	return (scaledPixels + 5.0) / 7.0
}

// setColWidthWithStreamWriter sets the widths of specified columns using the provided StreamWriter.
// sw is the Excel StreamWriter used for writing operations.
// colWidths is a slice of integers representing the widths to be applied to corresponding columns.
// Returns an error if the operation fails or if no column widths are specified.
func setColWidthWithStreamWriter(sw *excelize.StreamWriter, col int, colWidths []int) error {
	if len(colWidths) == 0 {
		return errors.New("no column widths specified")
	}
	for i, width := range colWidths {
		if width > 0 {
			if width > excelize.MaxColumnWidth {
				width = excelize.MaxColumnWidth
			}
			if err := sw.SetColWidth(col+i, col+i, float64(width)); err != nil {
				return err
			}
		}
	}
	return nil
}

// sanitizeFilePath ensures a file path has a default extension and returns a cleaned version of the path.
func sanitizeFilePath(filePath, defaultFilePath, defaultExtension string) string {
	if len(defaultExtension) == 0 {
		panic("default extension cannot be empty")
	}

	if len(defaultFilePath) == 0 {
		panic("default file path cannot be empty")
	}

	if filePath == "" {
		filePath = defaultFilePath
	}

	if len(defaultExtension) > 0 && defaultExtension[0] != '.' {
		defaultExtension = "." + defaultExtension
	}

	fileExt := filepath.Ext(filePath)

	if fileExt == "" {
		filePath = filePath + defaultExtension
		fileExt = filepath.Ext(filePath)
	}

	if fileExt != defaultExtension {
		filePath = strings.TrimSuffix(filePath, fileExt) + defaultExtension
	}

	return filepath.Clean(filePath)
}
