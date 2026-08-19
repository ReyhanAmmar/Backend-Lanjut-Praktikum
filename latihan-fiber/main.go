package main

import "fmt"

func swap(a, b *int) {
	*a, *b = *b, *a
}

func updateSlice(s *[]string, newItem string) {
	*s = append(*s, newItem)
}

func PassbyValue(x int) {
	x = x + 100
}

func PassbyPointer(x *int) {
	*x = *x + 100
}

func main() {
	fmt.Println("=== swap dengan pointer ===")
	a, b := 10, 20
	fmt.Println("Sebelum swap:", a, b)
	swap(&a, &b)
	fmt.Println("Sesudah swap:", a, b)

	fmt.Println("\n=== updateSlice dengan pointer ===")
	daftarBuah := []string{"apel", "jeruk"}
	fmt.Println("Sebelum update:", daftarBuah)
	updateSlice(&daftarBuah, "mangga")
	fmt.Println("Sesudah update:", daftarBuah)

	fmt.Println("\n=== Perbandingan pass by value vs pass by pointer ===")
	angka := 5
	fmt.Println("Nilai awal:", angka)

	PassbyValue(angka)
	fmt.Println("Setelah PassbyValue:", angka)

	PassbyPointer(&angka)
	fmt.Println("Setelah PassbyPointer:", angka)
}