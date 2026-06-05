package excel

import (
	"bytes"
	"fmt"

	"github.com/Wizzerin/immogucker-go/internal/models"
	"github.com/xuri/excelize/v2"
)

// GenerateResults creates an Excel file from the apartment list and returns it as a byte buffer.
func GenerateResults(apartments []models.Apartment) (*bytes.Buffer, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheet := "Sheet1"

	// Set headers
	f.SetCellValue(sheet, "A1", "Title")
	f.SetCellValue(sheet, "B1", "Price (€)")
	f.SetCellValue(sheet, "C1", "Link")

	// Set column widths for better readability
	f.SetColWidth(sheet, "A", "A", 50)
	f.SetColWidth(sheet, "B", "B", 15)
	f.SetColWidth(sheet, "C", "C", 70)

	for i, apt := range apartments {
		row := i + 2
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), apt.Title)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), apt.Price)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), "Open listing")

		err := f.SetCellHyperLink(sheet, fmt.Sprintf("C%d", row), apt.Link, "External")
		if err != nil {
			return nil, fmt.Errorf("failed to create hyperlink for cell C%d: %w", row, err)
		}
	}

	return f.WriteToBuffer()
}
