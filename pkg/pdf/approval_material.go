package pdf

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/claudio-nehemia/interior_backend/internal/dto"
	"github.com/jung-kurt/gofpdf/v2"
)

// GenerateApprovalMaterialPDF creates a professional PDF sheet for Material Approval.
func GenerateApprovalMaterialPDF(am *dto.ApprovalMaterialResponse, cp CompanyProfile, uploadDir string) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(10, 10, 10)
	pdf.SetTextColor(50, 50, 50)
	pdf.SetAutoPageBreak(false, 10)

	// Add Font if necessary, otherwise standard Arial is fine
	pdf.AddPage()

	// 1. Kop Surat / Letterhead (Dynamic from Company Profile)
	drawKopSurat(pdf, cp, uploadDir)

	// 2. Order Metadata Header Grid
	drawMetadataHeader(pdf, am)

	// 3. Materials Table Grouped by Category
	// Group items by category: bahan_baku, finishing, aksesoris
	var bahanBakuItems []dto.ApprovalMaterialItemResponse
	var finishingItems []dto.ApprovalMaterialItemResponse
	var aksesorisItems []dto.ApprovalMaterialItemResponse

	for _, item := range am.Items {
		cat := strings.ToLower(item.Category)
		if cat == "bahan_baku" {
			bahanBakuItems = append(bahanBakuItems, item)
		} else if cat == "finishing_dalam" || cat == "finishing_luar" || cat == "finishing" {
			finishingItems = append(finishingItems, item)
		} else if cat == "aksesoris" || cat == "aksesori" {
			aksesorisItems = append(aksesorisItems, item)
		} else {
			// fallback
			bahanBakuItems = append(bahanBakuItems, item)
		}
	}

	// Render Tables
	err := renderCategoryTable(pdf, "A. BAHAN BAKU", bahanBakuItems, uploadDir)
	if err != nil {
		return nil, err
	}

	err = renderCategoryTable(pdf, "B. FINISHING", finishingItems, uploadDir)
	if err != nil {
		return nil, err
	}

	err = renderCategoryTable(pdf, "C. AKSESORIS", aksesorisItems, uploadDir)
	if err != nil {
		return nil, err
	}

	// Signatures/Footer
	checkPageLimit(pdf, 40)
	pdf.Ln(8)
	ySign := pdf.GetY()
	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(80, 80, 80)
	pdf.SetXY(15, ySign)
	pdf.CellFormat(80, 5, "Diajukan Oleh,", "", 0, "C", false, 0, "")
	pdf.SetXY(115, ySign)
	pdf.CellFormat(80, 5, "Disetujui Oleh,", "", 1, "C", false, 0, "")

	pdf.Ln(15)

	yName := pdf.GetY()
	pdf.SetFont("Arial", "B", 9)
	pdf.SetTextColor(50, 50, 50)
	pdf.SetXY(15, yName)
	designerName := "Designer / Drafter"
	if am.ResponseBy != "" {
		designerName = am.ResponseBy
	}
	pdf.CellFormat(80, 5, designerName, "", 0, "C", false, 0, "")

	pdf.SetXY(115, yName)
	clientName := "Klien / Pelanggan"
	if am.Order != nil {
		clientName = am.Order.NamaCustomer
	}
	pdf.CellFormat(80, 5, clientName, "", 1, "C", false, 0, "")

	pdf.SetFont("Arial", "", 8)
	pdf.SetTextColor(120, 120, 120)
	pdf.SetXY(15, pdf.GetY())
	pdf.CellFormat(80, 4, "(Nuginterior Team)", "", 0, "C", false, 0, "")
	pdf.SetXY(115, pdf.GetY()-4)
	pdf.CellFormat(80, 4, "(Owner / Client)", "", 1, "C", false, 0, "")

	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func drawKopSurat(pdf *gofpdf.Fpdf, cp CompanyProfile, uploadDir string) {
	pdf.SetY(10)
	logoFile := ""
	if cp.Logo != "" {
		logoFile = filepath.Join(uploadDir, filepath.Base(cp.Logo))
		if _, err := os.Stat(logoFile); os.IsNotExist(err) {
			logoFile = ""
		}
	}

	if logoFile != "" {
		pdf.Image(logoFile, 10, 10, 0, 12, false, "", 0, "")
		pdf.SetLeftMargin(28)
		pdf.SetX(28)
		pdf.SetY(10)
	} else {
		pdf.SetY(10)
	}

	pdf.SetFont("Arial", "B", 13)
	pdf.SetTextColor(0, 128, 128) // Teal
	pdf.CellFormat(0, 5, cp.Name, "", 1, "L", false, 0, "")

	pdf.SetFont("Arial", "", 8)
	pdf.SetTextColor(100, 100, 100)
	pdf.CellFormat(0, 4, "Premium Interior Design & Architecture Services", "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 4, fmt.Sprintf("Email: %s | Phone: %s", cp.Email, cp.Phone), "", 1, "L", false, 0, "")

	pdf.SetLeftMargin(10)
	pdf.SetX(10)
	pdf.SetY(25)

	pdf.SetDrawColor(0, 128, 128)
	pdf.SetLineWidth(0.6)
	pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
	pdf.Ln(4)
}

func drawMetadataHeader(pdf *gofpdf.Fpdf, am *dto.ApprovalMaterialResponse) {
	pdf.SetDrawColor(180, 180, 180)
	pdf.SetLineWidth(0.2)
	pdf.SetFillColor(250, 250, 250)

	yStart := pdf.GetY()
	boxW := 105.0
	boxH := 24.0

	// Draw outline box
	pdf.Rect(10, yStart, boxW, boxH, "DF")

	// Print metadata labels and values
	pdf.SetFont("Arial", "B", 8)
	pdf.SetTextColor(80, 80, 80)

	ownerVal := "-"
	projectVal := "-"
	lokasiVal := "-"
	if am.Order != nil {
		ownerVal = am.Order.NamaCustomer
		projectVal = strings.Title(strings.ToLower(am.Order.JenisInterior))
		lokasiVal = am.Order.Alamat
	}

	// Limit lokasi character size
	if len(lokasiVal) > 42 {
		lokasiVal = lokasiVal[:40] + "..."
	}

	dateVal := time.Now().Format("02 January 2006")

	// Labels
	labels := []string{"OWNER", "PROJECT", "LOKASI", "TANGGAL"}
	vals := []string{ownerVal, projectVal, lokasiVal, dateVal}

	for i := 0; i < 4; i++ {
		pdf.SetXY(12, yStart+2+float64(i)*5)
		pdf.SetFont("Arial", "B", 8)
		pdf.CellFormat(20, 5, labels[i], "", 0, "L", false, 0, "")
		pdf.SetFont("Arial", "", 8)
		pdf.CellFormat(70, 5, ": "+vals[i], "", 1, "L", false, 0, "")
	}

	pdf.SetY(yStart + boxH + 6)
}

func checkPageLimit(pdf *gofpdf.Fpdf, neededHeight float64) {
	y := pdf.GetY()
	// A4 is 297mm high, bottom margin 10mm -> limit 287mm
	if y+neededHeight > 280 {
		pdf.AddPage()
		pdf.SetY(15) // simple spacing on new page
	}
}

func renderCategoryTable(pdf *gofpdf.Fpdf, categoryTitle string, items []dto.ApprovalMaterialItemResponse, uploadDir string) error {
	if len(items) == 0 {
		return nil
	}

	checkPageLimit(pdf, 25)

	// Section Title
	pdf.SetFont("Arial", "B", 9)
	pdf.SetTextColor(0, 0, 0)
	pdf.CellFormat(0, 6, categoryTitle, "", 1, "L", false, 0, "")
	pdf.Ln(1)

	// Table Headers
	// Widths: NO(10), ITEM(35), FOTO(30), BRAND/SPEK(30), KODE MATERIAL(35), AREA(25), NOTE(25) = 190mm
	widths := []float64{10, 35, 30, 30, 35, 25, 25}
	headers := []string{"NO.", "ITEM", "FOTO", "BRAND/SPEK", "KODE MATERIAL", "AREA", "NOTE"}

	pdf.SetFillColor(240, 240, 240)
	pdf.SetDrawColor(180, 180, 180)
	pdf.SetTextColor(50, 50, 50)
	pdf.SetFont("Arial", "B", 8)

	for i, h := range headers {
		align := "C"
		if i == 1 {
			align = "L"
		}
		pdf.CellFormat(widths[i], 6, h, "1", 0, align, true, 0, "")
	}
	pdf.Ln(6)

	// Rows
	pdf.SetFont("Arial", "", 8)
	pdf.SetTextColor(60, 60, 60)

	for idx, item := range items {
		// Calculate dynamic row height
		brands := item.BrandSpek
		codes := item.KodeMaterial
		if len(brands) == 0 {
			brands = []string{"-"}
		}
		if len(codes) == 0 {
			codes = []string{"-"}
		}

		maxSubRows := len(brands)
		if len(codes) > maxSubRows {
			maxSubRows = len(codes)
		}

		// Row height is max(24mm, maxSubRows * 6mm)
		rowH := 24.0
		if float64(maxSubRows)*6.0 > rowH {
			rowH = float64(maxSubRows) * 6.0
		}

		checkPageLimit(pdf, rowH)

		yStart := pdf.GetY()
		xStart := 10.0

		// NO. Column
		pdf.Rect(xStart, yStart, widths[0], rowH, "D")
		pdf.SetXY(xStart, yStart)
		pdf.CellFormat(widths[0], rowH, fmt.Sprintf("%d", idx+1), "", 0, "C", false, 0, "")

		// ITEM Column
		xStart += widths[0]
		pdf.Rect(xStart, yStart, widths[1], rowH, "D")
		pdf.SetXY(xStart+2, yStart)
		// MultiCell if name is long
		pdf.MultiCell(widths[1]-4, 5, item.ItemName, "", "L", false)

		// FOTO Column
		xStart += widths[1]
		pdf.Rect(xStart, yStart, widths[2], rowH, "D")
		if item.Foto != "" {
			localFile := filepath.Join(uploadDir, filepath.Base(item.Foto))
			if _, err := os.Stat(localFile); err == nil {
				// Center photo inside column of width 30, height rowH
				imgW := 24.0
				imgH := 18.0
				if imgH > rowH-4 {
					imgH = rowH - 4
				}
				imgX := xStart + (widths[2]-imgW)/2
				imgY := yStart + (rowH-imgH)/2
				pdf.Image(localFile, imgX, imgY, imgW, imgH, false, "", 0, "")
			} else {
				pdf.SetXY(xStart, yStart)
				pdf.CellFormat(widths[2], rowH, "-", "", 0, "C", false, 0, "")
			}
		} else {
			pdf.SetXY(xStart, yStart)
			pdf.CellFormat(widths[2], rowH, "-", "", 0, "C", false, 0, "")
		}

		// BRAND/SPEK Column (Sub-rows)
		xStart += widths[2]
		pdf.Rect(xStart, yStart, widths[3], rowH, "D")
		brandH := rowH / float64(len(brands))
		for bIdx, brand := range brands {
			pdf.SetXY(xStart, yStart+float64(bIdx)*brandH)
			pdf.CellFormat(widths[3], brandH, brand, "1", 0, "C", false, 0, "")
		}

		// KODE MATERIAL Column (Sub-rows)
		xStart += widths[3]
		pdf.Rect(xStart, yStart, widths[4], rowH, "D")
		codeH := rowH / float64(len(codes))
		for cIdx, code := range codes {
			pdf.SetXY(xStart, yStart+float64(cIdx)*codeH)
			pdf.CellFormat(widths[4], codeH, code, "1", 0, "C", false, 0, "")
		}

		// AREA Column
		xStart += widths[4]
		pdf.Rect(xStart, yStart, widths[5], rowH, "D")
		pdf.SetXY(xStart+1, yStart+2)
		areaVal := item.Area
		if areaVal == "" {
			areaVal = "-"
		}
		pdf.MultiCell(widths[5]-2, 4, areaVal, "", "L", false)

		// NOTE Column
		xStart += widths[5]
		pdf.Rect(xStart, yStart, widths[6], rowH, "D")
		pdf.SetXY(xStart+1, yStart+2)
		noteVal := item.Notes
		if noteVal == "" {
			noteVal = "-"
		}
		pdf.MultiCell(widths[6]-2, 4, noteVal, "", "L", false)

		// Set cursor to next row
		pdf.SetXY(10, yStart+rowH)
	}

	pdf.Ln(4)
	return nil
}
