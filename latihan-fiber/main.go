package main

import "fmt"

func main() {
	var nama string = "Reyhan Ammar"
	var umur int = 21
	var ipk float64 = 3.7
	var aktif bool = true
	hobi := []string{"bermain game", "olahraga", "coding"}

	fmt.Println("=== Variabel ===")
	fmt.Println("Nama :", nama)
	fmt.Println("Umur :", umur)
	fmt.Println("IPK  :", ipk)
	fmt.Println("Aktif:", aktif)
	fmt.Println("Hobi :", hobi)

	nilaiMahasiswa := make(map[string]float64)

	nilaiMahasiswa["Alfin"] = 85.5
	nilaiMahasiswa["Ody"] = 90.0
	nilaiMahasiswa["Ilyas"] = 78.0

	fmt.Println("\n=== Map setelah diisi ===")
	fmt.Println(nilaiMahasiswa)

	fmt.Println("\n=== Cek keberadaan Nilai ===")
	if nilai, ada := nilaiMahasiswa["Alfin"]; ada {
		fmt.Println("Nilai Alfin ditemukan:", nilai)
	} else {
		fmt.Println("Alfin tidak ada di map")
	}

	if nilai, ada := nilaiMahasiswa["Ody"]; ada {
		fmt.Println("Nilai Ody ditemukan:", nilai)
	} else {
		fmt.Println("Ody belum punya nilai")
	}

	delete(nilaiMahasiswa, "Ilyas")
	fmt.Println("\n=== Map setelah Ilyas dihapus ===")
	fmt.Println(nilaiMahasiswa)
	
	fmt.Println("\n=== Menelusuri semua data ===")
	for nama, nilai := range nilaiMahasiswa {
		fmt.Printf("%s: %.1f\n", nama, nilai)
	}
}