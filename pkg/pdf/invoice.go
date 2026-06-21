package pdf

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/claudio-nehemia/interior_backend/internal/entity"
	"github.com/jung-kurt/gofpdf/v2"
)

// CompanyProfile represents the dynamic company data for PDF rendering.
type CompanyProfile struct {
	Name        string
	Director    string
	Logo        string
	Address     string
	BankName    string
	BankAccount string
	BankHolder  string
	Email       string
	Phone       string
}

// Helper function to format currency with Indonesian thousand separator (dots)
func formatRupiah(val float64) string {
	s := fmt.Sprintf("%.0f", val)
	n := len(s)
	if n <= 3 {
		return s
	}
	var buf bytes.Buffer
	for i, c := range s {
		buf.WriteRune(c)
		if (n-i-1)%3 == 0 && i != n-1 {
			buf.WriteByte('.')
		}
	}
	return buf.String()
}

// GenerateCommitmentFeeInvoice creates a professional PDF receipt for Commitment Fee payments.
func GenerateCommitmentFeeInvoice(order *entity.Order, m *entity.Moodboard, fee *entity.CommitmentFee, cp CompanyProfile, uploadDir string) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	
	// =========================================================================
	// PAGE 1: SURAT PERNYATAAN KOMITMEN (COMMITMENT FEE)
	// =========================================================================
	pdf.AddPage()
	pdf.SetMargins(20, 20, 20)
	pdf.SetTextColor(50, 50, 50)

	// Document Title
	pdf.SetFont("Arial", "B", 13)
	pdf.CellFormat(0, 8, "SURAT PERNYATAAN KOMITMEN (COMMITMENT FEE)", "", 1, "C", false, 0, "")
	
	pdf.SetFont("Arial", "", 9)
	pdf.CellFormat(0, 5, fmt.Sprintf("Nomor Dokumen: DOC/CF/%s/%04d", order.NomorOrder, fee.ID), "", 1, "C", false, 0, "")
	pdf.Ln(8)

	// Format Indonesian date
	const (
		Jan = "Januari"
		Feb = "Februari"
		Mar = "Maret"
		Apr = "April"
		May = "Mei"
		Jun = "Juni"
		Jul = "Juli"
		Aug = "Agustus"
		Sep = "September"
		Oct = "Oktober"
		Nov = "November"
		Dec = "Desember"
	)
	months := []string{"", Jan, Feb, Mar, Apr, May, Jun, Jul, Aug, Sep, Oct, Nov, Dec}
	t := time.Now()
	todayStr := fmt.Sprintf("%d %s %d", t.Day(), months[t.Month()], t.Year())

	// Body 1
	pdf.SetFont("Arial", "", 10)
	body1 := fmt.Sprintf("Pada hari ini, tanggal %s, kami yang bertanda tangan di bawah ini menyatakan kesepakatan komitmen proyek desain interior:", todayStr)
	pdf.MultiCell(0, 5.5, body1, "", "J", false)
	pdf.Ln(4)

	// Details grid
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(40, 6, "Nama Klien", "", 0, "L", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(0, 6, fmt.Sprintf(": %s", order.NamaCustomer), "", 1, "L", false, 0, "")

	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(40, 6, "Nama Project", "", 0, "L", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(0, 6, fmt.Sprintf(": %s", order.NamaProject), "", 1, "L", false, 0, "")

	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(40, 6, "Nomor Order", "", 0, "L", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(0, 6, fmt.Sprintf(": %s", order.NomorOrder), "", 1, "L", false, 0, "")

	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(40, 6, "Alamat Proyek", "", 0, "L", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(0, 6, fmt.Sprintf(": %s", order.Alamat), "", 1, "L", false, 0, "")
	pdf.Ln(4)

	feeVal := 0.0
	if fee.TotalFee != nil {
		feeVal = *fee.TotalFee
	}
	totalFeeStr := fmt.Sprintf("Rp %s", formatRupiah(feeVal))

	// Body 2 & 3 & 4
	body2 := fmt.Sprintf("Bahwa pihak klien telah menyatakan kesepakatan dan komitmen atas pengerjaan konsep desain visual proyek dengan membayar Commitment Fee sebesar %s (Rupiah) sebagai prasyarat dimulainya pengerjaan master blueprint desain final.", totalFeeStr)
	pdf.MultiCell(0, 5.5, body2, "", "J", false)
	pdf.Ln(4)

	body3 := "Commitment fee ini bersifat mengikat dan akan diperhitungkan sebagai pengurang dari total biaya jasa desain interior pada saat pelunasan termin berikutnya."
	pdf.MultiCell(0, 5.5, body3, "", "J", false)
	pdf.Ln(4)

	body4 := "Demikian surat pernyataan komitmen ini dibuat dengan kesadaran penuh dari kedua belah pihak untuk dapat dipergunakan sebagaimana mestinya."
	pdf.MultiCell(0, 5.5, body4, "", "J", false)
	pdf.Ln(18)

	// Signatures Page 1
	pdf.CellFormat(85, 5, "Hormat Kami,", "", 0, "C", false, 0, "")
	pdf.CellFormat(85, 5, "Menyetujui,", "", 1, "C", false, 0, "")
	pdf.Ln(18)
	
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(85, 5, cp.Name, "", 0, "C", false, 0, "")
	pdf.CellFormat(85, 5, order.NamaCustomer, "", 1, "C", false, 0, "")

	pdf.SetFont("Arial", "", 8)
	pdf.SetTextColor(120, 120, 120)
	pdf.CellFormat(85, 4, "(Finance / Customer Service)", "", 0, "C", false, 0, "")
	pdf.CellFormat(85, 4, "(Pelanggan / Klien)", "", 1, "C", false, 0, "")

	// =========================================================================
	// PAGE 2: INVOICE / KWITANSI COMMITMENT FEE
	// =========================================================================
	pdf.AddPage()
	pdf.SetMargins(15, 15, 15)
	pdf.SetTextColor(80, 80, 80)

	// Kop Surat / Header
	logoFile := ""
	if cp.Logo != "" {
		logoFile = filepath.Join(uploadDir, filepath.Base(cp.Logo))
		if _, err := os.Stat(logoFile); os.IsNotExist(err) {
			logoFile = ""
		}
	}

	if logoFile != "" {
		pdf.Image(logoFile, 15, 12, 0, 15, false, "", 0, "")
		pdf.SetLeftMargin(33)
		pdf.SetX(33)
		pdf.SetY(12)
	} else {
		pdf.SetY(12)
	}

	pdf.SetFont("Arial", "B", 16)
	pdf.SetTextColor(0, 128, 128) // Teal color
	pdf.CellFormat(0, 7, cp.Name, "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(100, 100, 100)
	pdf.CellFormat(0, 4, "Premium Interior Design & Architecture Services", "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 4, fmt.Sprintf("Email: %s | Phone: %s", cp.Email, cp.Phone), "", 1, "L", false, 0, "")

	// Reset margins
	pdf.SetLeftMargin(15)
	pdf.SetX(15)
	pdf.SetY(30)
	
	pdf.SetDrawColor(0, 128, 128)
	pdf.SetLineWidth(0.8)
	pdf.Line(15, pdf.GetY(), 195, pdf.GetY())
	pdf.Ln(6)

	// Title
	pdf.SetFont("Arial", "B", 14)
	pdf.SetTextColor(50, 50, 50)
	pdf.CellFormat(0, 8, "INVOICE / KWITANSI COMMITMENT FEE", "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(0, 5, fmt.Sprintf("No. Invoice: INV/CF/%s/%04d", order.NomorOrder, fee.ID), "", 1, "C", false, 0, "")
	pdf.Ln(6)

	// Invoice Information Grid
	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(0, 128, 128)
	pdf.CellFormat(90, 6, "INFORMASI KELAYAKAN PROYEK", "", 0, "L", false, 0, "")
	pdf.CellFormat(90, 6, "Rincian Tagihan", "", 1, "L", false, 0, "")
	pdf.SetLineWidth(0.2)
	pdf.SetDrawColor(200, 200, 200)
	pdf.Line(15, pdf.GetY(), 195, pdf.GetY())
	pdf.Ln(2)

	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(80, 80, 80)
	
	// Row 1
	pdf.CellFormat(30, 5, "Nomor Order:", "", 0, "L", false, 0, "")
	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(60, 5, order.NomorOrder, "", 0, "L", false, 0, "")
	pdf.SetFont("Arial", "", 9)
	pdf.CellFormat(30, 5, "Tanggal Tagihan:", "", 0, "L", false, 0, "")
	pdf.CellFormat(60, 5, time.Now().Format("02 Jan 2006"), "", 1, "L", false, 0, "")

	// Row 2
	pdf.CellFormat(30, 5, "Nama Project:", "", 0, "L", false, 0, "")
	pdf.CellFormat(60, 5, order.NamaProject, "", 0, "L", false, 0, "")
	pdf.CellFormat(30, 5, "Status Bayar:", "", 0, "L", false, 0, "")
	pdf.SetFont("Arial", "B", 9)
	if fee.PaymentStatus == "completed" {
		pdf.SetTextColor(16, 124, 65)
		pdf.CellFormat(60, 5, "LUNAS (Completed)", "", 1, "L", false, 0, "")
	} else {
		pdf.SetTextColor(200, 120, 0)
		pdf.CellFormat(60, 5, "MENUNGGU PEMBAYARAN", "", 1, "L", false, 0, "")
	}
	pdf.SetTextColor(80, 80, 80)
	pdf.SetFont("Arial", "", 9)

	// Row 3
	pdf.CellFormat(30, 5, "Pelanggan:", "", 0, "L", false, 0, "")
	pdf.CellFormat(60, 5, order.NamaCustomer, "", 0, "L", false, 0, "")
	pdf.CellFormat(30, 5, "Metode Bayar:", "", 0, "L", false, 0, "")
	pdf.CellFormat(60, 5, "Bank Transfer", "", 1, "L", false, 0, "")

	// Row 4
	pdf.CellFormat(30, 5, "Alamat Proyek:", "", 0, "L", false, 0, "")
	pdf.CellFormat(150, 5, order.Alamat, "", 1, "L", false, 0, "")
	pdf.Ln(6)

	// Table Header
	pdf.SetFillColor(0, 128, 128)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(15, 8, "No", "1", 0, "C", true, 0, "")
	pdf.CellFormat(115, 8, "Deskripsi Layanan", "1", 0, "L", true, 0, "")
	pdf.CellFormat(50, 8, "Total Biaya", "1", 1, "R", true, 0, "")

	// Table Body
	pdf.SetTextColor(80, 80, 80)
	pdf.SetFont("Arial", "", 9)
	pdf.CellFormat(15, 8, "1", "1", 0, "C", false, 0, "")
	pdf.CellFormat(115, 8, "Pembayaran Awal (Commitment Fee) untuk Pembuatan Desain Interior Proyek: "+order.NamaProject, "1", 0, "L", false, 0, "")
	pdf.CellFormat(50, 8, fmt.Sprintf("Rp %s", formatRupiah(feeVal)), "1", 1, "R", false, 0, "")
	
	// Total Row
	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(50, 50, 50)
	pdf.CellFormat(130, 8, "Total Tagihan", "1", 0, "R", false, 0, "")
	pdf.SetTextColor(0, 128, 128)
	pdf.CellFormat(50, 8, fmt.Sprintf("Rp %s", formatRupiah(feeVal)), "1", 1, "R", false, 0, "")
	pdf.Ln(8)

	// Note / Bank Account
	pdf.SetFont("Arial", "B", 9)
	pdf.SetTextColor(0, 128, 128)
	pdf.CellFormat(0, 5, "INFORMASI PEMBAYARAN TRANSFER BANK:", "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(80, 80, 80)
	pdf.CellFormat(0, 5, fmt.Sprintf("Bank: %s", cp.BankName), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 5, fmt.Sprintf("Nomor Rekening: %s", cp.BankAccount), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 5, fmt.Sprintf("Atas Nama: %s", cp.BankHolder), "", 1, "L", false, 0, "")
	pdf.Ln(10)

	// Signatures Page 2
	pdf.SetFont("Arial", "", 9)
	pdf.CellFormat(90, 5, "Hormat Kami,", "", 0, "C", false, 0, "")
	pdf.CellFormat(90, 5, "Pelanggan,", "", 1, "C", false, 0, "")
	pdf.Ln(15)
	
	pdf.SetFont("Arial", "BU", 9)
	pdf.CellFormat(90, 5, cp.Name, "", 0, "C", false, 0, "")
	pdf.CellFormat(90, 5, order.NamaCustomer, "", 1, "C", false, 0, "")

	pdf.SetFont("Arial", "", 8)
	pdf.SetTextColor(120, 120, 120)
	pdf.CellFormat(90, 4, "(Finance / Customer Service)", "", 0, "C", false, 0, "")
	pdf.CellFormat(90, 4, "(Tanda Tangan & Nama Terang)", "", 1, "C", false, 0, "")

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
