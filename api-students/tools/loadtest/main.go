package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"api-students/app/model"
	"api-students/config"
	"api-students/database"
	"api-students/helper"
)

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        200,
		MaxIdleConnsPerHost: 200,
	},
}

// fire menjalankan n pekerjaan secara serentak dan menghitung hasilnya.
//
// Setiap goroutine menunggu di depan "gerbang" start yang sama. Gerbang
// baru dibuka setelah SELURUH goroutine siap, sehingga request benar-benar
// berangkat hampir bersamaan — bukan bergiliran sesuai urutan dibuatnya.
func fire(n int, job func(i int) string) map[string]int {
	var wg sync.WaitGroup
	start := make(chan struct{})
	results := make(chan string, n)

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int, out chan<- string) {
			defer wg.Done()
			<-start
			out <- job(i)
		}(i, results)
	}

	close(start) // buka gerbang
	wg.Wait()    // tunggu SEMUA goroutine selesai
	close(results)

	tally := map[string]int{}
	for r := range results {
		tally[r]++
	}
	return tally
}

func printTally(tally map[string]int) {
	fmt.Println("hasil request:")
	keys := make([]string, 0, len(tally))
	for k := range tally {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("  %-25s %d\n", k, tally[k])
	}
}

func parseResponse(resp *http.Response) string {
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var res struct {
		Success bool   `json:"success"`
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	_ = json.Unmarshal(body, &res)

	if resp.StatusCode == http.StatusOK {
		return "200 OK"
	}
	if resp.StatusCode == http.StatusCreated {
		return "201 OK"
	}
	if res.Code != "" {
		return fmt.Sprintf("%d %s", resp.StatusCode, res.Code)
	}
	return fmt.Sprintf("%d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
}

func loadStudents(ctx context.Context, pool *pgxpool.Pool) ([]model.Student, error) {
	rows, err := pool.Query(ctx, `SELECT id, nim, role FROM students WHERE nim LIKE 'mhs%' ORDER BY nim`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []model.Student
	for rows.Next() {
		var s model.Student
		if err := rows.Scan(&s.ID, &s.NIM, &s.Role); err != nil {
			return nil, err
		}
		students = append(students, s)
	}
	return students, rows.Err()
}

func generateTokens(students []model.Student, jwtManager *helper.JWTManager) ([]string, error) {
	tokens := make([]string, len(students))
	for i, s := range students {
		token, err := jwtManager.GenerateAccess(s)
		if err != nil {
			return nil, fmt.Errorf("gagal membuat token untuk %s: %w", s.NIM, err)
		}
		tokens[i] = token
	}
	return tokens, nil
}

func main() {
	mode := flag.String("mode", "crowd", "Mode pengujian: crowd | credits | swap")
	baseURL := flag.String("url", "http://localhost:3000", "Base URL server target")
	flag.Parse()

	config.LoadEnv()

	jwtSecret := config.GetEnv("JWT_SECRET", "")
	if len(jwtSecret) < 32 {
		fmt.Println("Error: JWT_SECRET tidak valid atau kurang dari 32 karakter")
		return
	}

	ctx := context.Background()
	pool, err := database.NewPool(ctx)
	if err != nil {
		fmt.Printf("Error menghubungkan ke database: %v\n", err)
		return
	}
	defer pool.Close()

	students, err := loadStudents(ctx, pool)
	if err != nil {
		fmt.Printf("Error memuat akun mahasiswa: %v\n", err)
		return
	}

	if len(students) < 60 {
		fmt.Printf("Peringatan: ditemukan %d akun mahasiswa berawalan 'mhs' (dibutuhkan minimal 60).\n", len(students))
		fmt.Println("Pastikan migration seed sudah dijalankan.")
	}

	jwtManager := helper.NewJWTManager(jwtSecret, config.GetEnv("JWT_ISSUER", "students"), 24*time.Hour)
	tokens, err := generateTokens(students, jwtManager)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	switch *mode {
	case "crowd":
		runCrowd(ctx, pool, *baseURL, students, tokens)
	case "credits":
		runCredits(ctx, pool, *baseURL, students, tokens)
	case "swap":
		runSwap(ctx, pool, *baseURL, students, tokens)
	default:
		fmt.Printf("Mode tidak dikenal: %s. Pilihan: crowd, credits, swap\n", *mode)
	}
}

// Mode crowd: 60 mahasiswa berebut kelas IF399 (kuota 5)
func runCrowd(ctx context.Context, pool *pgxpool.Pool, baseURL string, students []model.Student, tokens []string) {
	var courseID, quota int
	err := pool.QueryRow(ctx, "SELECT id, quota FROM courses WHERE code = 'IF399'").Scan(&courseID, &quota)
	if err != nil {
		fmt.Printf("Error mengambil data kelas IF399: %v\n", err)
		return
	}

	n := 60
	if len(students) < n {
		n = len(students)
	}

	fmt.Printf("%d mahasiswa berebut kelas IF399 secara serentak...\n", n)

	start := time.Now()
	tally := fire(n, func(i int) string {
		url := fmt.Sprintf("%s/api/v1/courses/%d/enroll", baseURL, courseID)
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewBufferString("{}"))
		if err != nil {
			return fmt.Sprintf("ERR %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tokens[i])

		resp, err := httpClient.Do(req)
		if err != nil {
			return fmt.Sprintf("ERR %v", err)
		}
		return parseResponse(resp)
	})
	elapsed := time.Since(start)

	printTally(tally)
	fmt.Printf("waktu total : %v\n", elapsed)

	// Verifikasi database
	var enrolledCount, rowCount int
	_ = pool.QueryRow(ctx, "SELECT enrolled_count FROM courses WHERE id = $1", courseID).Scan(&enrolledCount)
	_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM enrollments WHERE course_id = $1", courseID).Scan(&rowCount)

	status := "KONSISTEN"
	if enrolledCount != rowCount || enrolledCount > quota {
		status = "TIDAK KONSISTEN"
	}
	fmt.Printf("IF399 kuota=%d enrolled_count=%d baris enrollments=%d -> %s\n", quota, enrolledCount, rowCount, status)
}

// Mode credits: 10 mahasiswa, masing-masing mendaftar ke 12 kelas sekaligus (batas 24 SKS)
func runCredits(ctx context.Context, pool *pgxpool.Pool, baseURL string, students []model.Student, tokens []string) {
	rows, err := pool.Query(ctx, "SELECT id, code, credits FROM courses WHERE code != 'IF399' ORDER BY code")
	if err != nil {
		fmt.Printf("Error mengambil daftar kelas: %v\n", err)
		return
	}
	defer rows.Close()

	type courseInfo struct {
		id      int
		code    string
		credits int
	}
	var courses []courseInfo
	for rows.Next() {
		var c courseInfo
		if err := rows.Scan(&c.id, &c.code, &c.credits); err != nil {
			fmt.Printf("Error membaca row kelas: %v\n", err)
			return
		}
		courses = append(courses, c)
	}

	numStudents := 10
	if len(students) < numStudents {
		numStudents = len(students)
	}
	numCourses := len(courses)
	totalRequests := numStudents * numCourses

	fmt.Printf("%d mahasiswa, masing-masing mendaftar ke %d kelas 3 SKS secara serentak (batas 24 SKS)...\n",
		numStudents, numCourses)

	start := time.Now()
	tally := fire(totalRequests, func(idx int) string {
		studentIdx := idx / numCourses
		courseIdx := idx % numCourses

		url := fmt.Sprintf("%s/api/v1/courses/%d/enroll", baseURL, courses[courseIdx].id)
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewBufferString("{}"))
		if err != nil {
			return fmt.Sprintf("ERR %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tokens[studentIdx])

		resp, err := httpClient.Do(req)
		if err != nil {
			return fmt.Sprintf("ERR %v", err)
		}
		return parseResponse(resp)
	})
	elapsed := time.Since(start)

	printTally(tally)
	fmt.Printf("waktu total : %v\n", elapsed)

	// Verifikasi SKS setiap mahasiswa
	allConsistent := true
	for i := 0; i < numStudents; i++ {
		var totalSKS int
		_ = pool.QueryRow(ctx, `
			SELECT COALESCE(SUM(c.credits), 0)
			FROM enrollments e
			JOIN courses c ON c.id = e.course_id
			WHERE e.student_id = $1
		`, students[i].ID).Scan(&totalSKS)

		if totalSKS > 24 {
			allConsistent = false
			fmt.Printf("Mahasiswa %s melebihi batas SKS: total %d SKS\n", students[i].NIM, totalSKS)
		}
	}

	if allConsistent {
		fmt.Println("seluruh 10 mahasiswa <= 24 SKS -> KONSISTEN")
	} else {
		fmt.Println("ada mahasiswa > 24 SKS -> TIDAK KONSISTEN")
	}
}

// Mode swap: 20 pasang mahasiswa pindah kelas berlawanan arah (IF101 <-> IF102)
func runSwap(ctx context.Context, pool *pgxpool.Pool, baseURL string, students []model.Student, tokens []string) {
	var if101ID, if102ID, if101Quota, if102Quota int
	err := pool.QueryRow(ctx, "SELECT id, quota FROM courses WHERE code = 'IF101'").Scan(&if101ID, &if101Quota)
	if err != nil {
		fmt.Printf("Error mengambil data kelas IF101: %v\n", err)
		return
	}
	err = pool.QueryRow(ctx, "SELECT id, quota FROM courses WHERE code = 'IF102'").Scan(&if102ID, &if102Quota)
	if err != nil {
		fmt.Printf("Error mengambil data kelas IF102: %v\n", err)
		return
	}

	if len(students) < 40 {
		fmt.Printf("Jumlah mahasiswa tidak mencukupi untuk uji swap (dibutuhkan 40, tersedia %d)\n", len(students))
		return
	}

	// Persiapkan enrollment awal untuk 20 pasang mahasiswa (40 mahasiswa):
	// mhs001..mhs020 di IF101, mhs021..mhs040 di IF102
	tx, err := pool.Begin(ctx)
	if err != nil {
		fmt.Printf("Error memulai transaksi persiapan swap: %v\n", err)
		return
	}

	var studentIDs []int
	for i := 0; i < 40; i++ {
		studentIDs = append(studentIDs, students[i].ID)
	}

	_, _ = tx.Exec(ctx, "DELETE FROM enrollments WHERE course_id IN ($1, $2) AND student_id = ANY($3)", if101ID, if102ID, studentIDs)
	for i := 0; i < 20; i++ {
		_, _ = tx.Exec(ctx, "INSERT INTO enrollments (course_id, student_id) VALUES ($1, $2)", if101ID, students[i].ID)
	}
	for i := 20; i < 40; i++ {
		_, _ = tx.Exec(ctx, "INSERT INTO enrollments (course_id, student_id) VALUES ($1, $2)", if102ID, students[i].ID)
	}
	_, _ = tx.Exec(ctx, "UPDATE courses c SET enrolled_count = (SELECT COUNT(*) FROM enrollments e WHERE e.course_id = c.id) WHERE id IN ($1, $2)", if101ID, if102ID)
	_ = tx.Commit(ctx)

	var deadlocksBefore int
	_ = pool.QueryRow(ctx, "SELECT COALESCE(sum(deadlocks), 0) FROM pg_stat_database WHERE datname = current_database()").Scan(&deadlocksBefore)

	fmt.Println("20 pasang mahasiswa pindah kelas berlawanan arah secara serentak...")

	start := time.Now()
	tally := fire(40, func(i int) string {
		fromID := if101ID
		toID := if102ID
		if i >= 20 {
			fromID = if102ID
			toID = if101ID
		}

		reqBody, _ := json.Marshal(map[string]int{
			"from_course_id": fromID,
			"to_course_id":   toID,
		})

		url := fmt.Sprintf("%s/api/v1/enrollments/swap", baseURL)
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBody))
		if err != nil {
			return fmt.Sprintf("ERR %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tokens[i])

		resp, err := httpClient.Do(req)
		if err != nil {
			return fmt.Sprintf("ERR %v", err)
		}
		return parseResponse(resp)
	})
	elapsed := time.Since(start)

	var deadlocksAfter int
	_ = pool.QueryRow(ctx, "SELECT COALESCE(sum(deadlocks), 0) FROM pg_stat_database WHERE datname = current_database()").Scan(&deadlocksAfter)

	printTally(tally)
	fmt.Printf("waktu total : %v\n", elapsed)
	fmt.Printf("deadlock terdeteksi PostgreSQL selama uji: %d\n", deadlocksAfter-deadlocksBefore)

	var if101Count, if101Rows, if102Count, if102Rows int
	_ = pool.QueryRow(ctx, "SELECT enrolled_count FROM courses WHERE id = $1", if101ID).Scan(&if101Count)
	_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM enrollments WHERE course_id = $1", if101ID).Scan(&if101Rows)
	_ = pool.QueryRow(ctx, "SELECT enrolled_count FROM courses WHERE id = $1", if102ID).Scan(&if102Count)
	_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM enrollments WHERE course_id = $1", if102ID).Scan(&if102Rows)

	s101 := "KONSISTEN"
	if if101Count != if101Rows || if101Count > if101Quota {
		s101 = "TIDAK KONSISTEN"
	}
	s102 := "KONSISTEN"
	if if102Count != if102Rows || if102Count > if102Quota {
		s102 = "TIDAK KONSISTEN"
	}

	fmt.Printf("IF101 kuota=%d enrolled_count=%d baris enrollments=%d -> %s\n", if101Quota, if101Count, if101Rows, s101)
	fmt.Printf("IF102 kuota=%d enrolled_count=%d baris enrollments=%d -> %s\n", if102Quota, if102Count, if102Rows, s102)
}
