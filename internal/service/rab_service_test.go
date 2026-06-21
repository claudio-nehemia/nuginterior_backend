package service

import (
	"testing"

	"github.com/claudio-nehemia/interior_backend/internal/entity"
)

func TestMapToRABResponse(t *testing.T) {
	// Create mock entities
	rab := entity.RAB{
		ID:            1,
		InputItemID:   2,
		OrderID:       3,
		MarkupGeneral: 10,
		Status:        "draft",
		Rooms: []entity.RABRoom{
			{
				ID:          1,
				NamaRuangan: "Living Room",
				Qty:         2,
				Panjang:     2.0,
				Lebar:       2.0,
				Tinggi:      2.0,
				Markup:      10, // 10% room markup
				BahanBakus: []entity.RABRoomBahanBaku{
					{
						BahanBakuID: 101,
						HargaDasar:  100000,
						HargaJasa:   50000,
						Markup:      10,
					},
				},
				FinishingDalams: []entity.RABRoomFinishing{
					{
						ItemID: 201,
						Type:   "dalam",
						Harga:  50000,
						Markup: 10,
					},
				},
				FinishingLuars: []entity.RABRoomFinishing{
					{
						ItemID: 202,
						Type:   "luar",
						Harga:  50000,
						Markup: 10,
					},
				},
				Aksesoris: []entity.RABRoomAksesoris{
					{
						ItemID: 301,
						Qty:    3,
						Harga:  10000,
						Markup: 20, // 20% accessory markup
					},
				},
			},
		},
	}

	// Instantiate a dummy service (since mapping only accesses memory structs, DB is nil)
	s := &rabService{}

	// 1. Test Internal Mode
	{
		res := s.mapToRABResponse(rab, "internal")
		room := res.Rooms[0]

		// Total Bahan Baku = 100000
		// Total Finishing = 50000 (Dalam) + 50000 (Luar) = 100000
		// Room Markup = 10%
		// Harga Dasar = (100000 + 100000) * (1 + 0.10) = 220000
		expectedHargaDasar := 220000.0
		if mathAbs(room.HargaDasar-expectedHargaDasar) > 0.001 {
			t.Errorf("Internal Mode: Expected HargaDasar %.2f, got %.2f", expectedHargaDasar, room.HargaDasar)
		}

		// Harga Satuan = Harga Dasar * Volume (2 * 2 * 2) * Qty (2) = 220000 * 8 * 2 = 3520000
		expectedHargaSatuan := 3520000.0
		if mathAbs(room.HargaSatuan-expectedHargaSatuan) > 0.001 {
			t.Errorf("Internal Mode: Expected HargaSatuan %.2f, got %.2f", expectedHargaSatuan, room.HargaSatuan)
		}

		// Aksesoris: 10000 * (1 + 0.20) * 3 = 12000 * 3 = 36000
		// Room Harga Total = Harga Satuan + Aksesoris = 3520000 + 36000 = 3556000
		expectedHargaTotal := 3556000.0
		if mathAbs(room.HargaTotal-expectedHargaTotal) > 0.001 {
			t.Errorf("Internal Mode: Expected HargaTotal %.2f, got %.2f", expectedHargaTotal, room.HargaTotal)
		}
	}

	// 2. Test Vendor Mode
	{
		res := s.mapToRABResponse(rab, "vendor")
		room := res.Rooms[0]

		// Markup = 0%
		// Harga Dasar = 100000 (Bahan Baku) + 100000 (Finishing) = 200000
		expectedHargaDasar := 200000.0
		if mathAbs(room.HargaDasar-expectedHargaDasar) > 0.001 {
			t.Errorf("Vendor Mode: Expected HargaDasar %.2f, got %.2f", expectedHargaDasar, room.HargaDasar)
		}

		// Harga Satuan = 200000 * 8 * 2 = 3200000
		expectedHargaSatuan := 3200000.0
		if mathAbs(room.HargaSatuan-expectedHargaSatuan) > 0.001 {
			t.Errorf("Vendor Mode: Expected HargaSatuan %.2f, got %.2f", expectedHargaSatuan, room.HargaSatuan)
		}

		// Aksesoris: 10000 * 3 = 30000
		// Room Harga Total = 3200000 + 30000 = 3230000
		expectedHargaTotal := 3230000.0
		if mathAbs(room.HargaTotal-expectedHargaTotal) > 0.001 {
			t.Errorf("Vendor Mode: Expected HargaTotal %.2f, got %.2f", expectedHargaTotal, room.HargaTotal)
		}
	}

	// 3. Test Jasa Mode
	{
		res := s.mapToRABResponse(rab, "jasa")
		room := res.Rooms[0]

		// Uses HargaJasa for materials: 50000.
		// Finishing = 100000.
		// Markup = 0%
		// Harga Dasar = 50000 (Jasa BB) + 100000 (Finishing) = 150000
		expectedHargaDasar := 150000.0
		if mathAbs(room.HargaDasar-expectedHargaDasar) > 0.001 {
			t.Errorf("Jasa Mode: Expected HargaDasar %.2f, got %.2f", expectedHargaDasar, room.HargaDasar)
		}

		// Harga Satuan = 150000 * 8 * 2 = 2400000
		expectedHargaSatuan := 2400000.0
		if mathAbs(room.HargaSatuan-expectedHargaSatuan) > 0.001 {
			t.Errorf("Jasa Mode: Expected HargaSatuan %.2f, got %.2f", expectedHargaSatuan, room.HargaSatuan)
		}

		// Aksesoris is removed (0)
		// Room Harga Total = 2400000
		expectedHargaTotal := 2400000.0
		if mathAbs(room.HargaTotal-expectedHargaTotal) > 0.001 {
			t.Errorf("Jasa Mode: Expected HargaTotal %.2f, got %.2f", expectedHargaTotal, room.HargaTotal)
		}
	}
}

func mathAbs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
