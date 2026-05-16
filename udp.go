package main

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Cara penggunaan: go run udp_max.go <IP> <PORT> <TIME_IN_SECONDS>")
		fmt.Println("Contoh         : go run udp_max.go 127.0.0.1 8080 10")
		os.Exit(1)
	}

	targetIP := os.Args[1]
	targetPort := os.Args[2]
	durationStr := os.Args[3]

	durationSec, err := strconv.Atoi(durationStr)
	if err != nil || durationSec <= 0 {
		fmt.Println("Error: Durasi waktu harus berupa angka positif.")
		os.Exit(1)
	}

	targetAddress := fmt.Sprintf("%s:%s", targetIP, targetPort)

	// Menggunakan semua Core CPU yang tersedia di komputermu
	numCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPU)

	fmt.Printf("Membuka 🚀 MAX POWER dengan %d Core CPU...\n", numCPU)
	fmt.Printf("Mengirim paket UDP tanpa henti ke %s selama %d detik...\n", targetAddress, durationSec)

	// Data/Payload berukuran ~1KB (standar untuk test bandwidth)
	payload := make([]byte, 1024)
	for i := range payload {
		payload[i] = 'X'
	}

	// Counter untuk menghitung total paket menggunakan atomic (aman untuk multi-thread)
	var totalPackets uint64

	// Channel untuk memberikan sinyal berhenti ke semua goroutine
	done := make(chan struct{})

	// Menjalankan fungsi pengirim di setiap Core CPU secara bersamaan
	for i := 0; i < numCPU; i++ {
		go func() {
			// Membuat koneksi per thread agar tidak rebutan socket
			conn, err := net.Dial("udp", targetAddress)
			if err != nil {
				return
			}
			defer conn.Close()

			for {
				select {
				case <-done:
					return
				default:
					// Kirim tanpa sleep/jeda sama sekali
					_, err := conn.Write(payload)
					if err == nil {
						atomic.AddUint64(&totalPackets, 1)
					}
				}
			}
		}()
	}

	// Timer durasi aplikasi berjalan
	time.Sleep(time.Duration(durationSec) * time.Second)

	// Sinyalkan semua goroutine untuk berhenti
	close(done)

	fmt.Println("\n[Selesai] Pengujian Max Power Berhenti.")
	fmt.Printf("Total Paket Terkirim: %d paket\n", atomic.LoadUint64(&totalPackets))
}
