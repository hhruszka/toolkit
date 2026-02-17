package reports

import (
	"bptvnftester/testengine"
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/fatih/color"
	"github.com/xuri/excelize/v2"
)

var setStyleOnce sync.Once
var styleWrappedId int
var styleNotWrappedId int

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

// setSheetName sets the name of the first or a new sheet in the Excel file, truncating the name if it exceeds 31 characters.
// It returns the updated sheet name or an error if the operation fails.
func setSheetName(xlsxFile *excelize.File, sheetName string) (string, error) {
	var err error

	if len(sheetName) > 31 {
		sheetName = sheetName[:excelize.MaxSheetNameLength]
	}
	if xlsxFile.SheetCount == 1 && xlsxFile.GetSheetName(0) == "Sheet1" {
		err = xlsxFile.SetSheetName(xlsxFile.GetSheetName(0), sheetName)
		if err != nil {
			// this is the first and the only tab. We cannot change its name to 31 char name
			return "", err
		}
	} else if _, err = xlsxFile.NewSheet(sheetName); err != nil {
		return "", err
	}
	return sheetName, nil
}

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
			if err := sw.SetColWidth(col+i, col+i, float64(width)*0.9+2); err != nil {
				return err
			}
		}
	}
	return nil
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
	return style.Font.Size, nil
}

var _ = getFontSize

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
func setColStyle(xlsxFile *excelize.File, sheetName string, startCol int, endCol int, _ int, _ int) error {
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

func resultToString(res bool) string {
	return map[bool]string{true: color.GreenString("PASSED"), false: color.RedString("FAILED")}[res]
}

// testStatus determines the test resultToString based on execution status or boolean value, returning "TIMEOUT", "PASSED", or "FAILED".
func testResult(test *testengine.TestResult, colorFlg bool) string {
	var (
		passedFailed map[bool]string = map[bool]string{true: "PASSED", false: "FAILED"}
		timeout      string          = "TIMEOUT"
	)

	if colorFlg {
		passedFailed[false] = color.RedString(passedFailed[false])
		passedFailed[true] = color.GreenString(passedFailed[true])
		timeout = color.YellowString("TIMEOUT")
	}

	if test.TimedOut() {
		return timeout
	} else {
		return passedFailed[test.Result]
	}
}

func wrap(text []string, n int) string {
	var buffer bytes.Buffer

	for _, str := range text {
		wrapLen := n - 1
		strLen := len(str) - 1
		for idx, c := range str {
			buffer.WriteRune(c)
			if idx%n == wrapLen && idx != strLen {
				buffer.WriteRune('\n')
			}
		}
		buffer.WriteRune('\n')
	}
	return buffer.String()
}

// sanitizeFilePath ensures a file path has the correct extension and is cleaned; defaults apply if the path is empty or invalid.
func sanitizeFilePath(filePath, defaultFilePath, defaultExtension string) string {
	if filePath == "" || defaultExtension == "" {
		filePath = defaultFilePath
	}

	if defaultExtension[0] != '.' {
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
