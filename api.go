package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Telegram config
const (
	BOT_TOKEN = "8453874511:AAGBTlqtNeSqLqyfMAA8BuE5SnVV04TmaZU"
	CHAT_ID   = "7013997051"
)

func sendTelegram(message string) {
	apiURL := "https://api.telegram.org/bot" + BOT_TOKEN + "/sendMessage"
	data := map[string]string{
		"chat_id": CHAT_ID,
		"text":    message,
	}
	jsonData, _ := json.Marshal(data)
	http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
}

func main() {
	http.HandleFunc("/api", handler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "TikTok API siap! Gunakan /api?url=LINK_TIKTOK",
		})
	})

	fmt.Println("Server running on port 8080")
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.RemoteAddr
	}
	userAgent := r.Header.Get("User-Agent")

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	videoURL := r.URL.Query().Get("url")
	if videoURL == "" {
		json.NewEncoder(w).Encode(map[string]string{"error": "Parameter 'url' wajib diisi"})
		sendTelegram(fmt.Sprintf("❌ ERROR\nIP: %s\nUser-Agent: %s\nError: URL kosong", ip, userAgent))
		return
	}

	if !strings.Contains(videoURL, "tiktok.com") {
		json.NewEncoder(w).Encode(map[string]string{"error": "URL tidak valid"})
		sendTelegram(fmt.Sprintf("❌ INVALID URL\nIP: %s\nUser-Agent: %s\nURL: %s", ip, userAgent, videoURL))
		return
	}

	apiURL := "https://www.tikwm.com/api/?url=" + url.QueryEscape(videoURL)
	resp, err := http.Get(apiURL)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "Gagal mengambil data"})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if data, ok := result["data"].(map[string]interface{}); ok {
		author := ""
		if a, ok := data["author"].(map[string]interface{}); ok {
			author = fmt.Sprintf("%v", a["unique_id"])
		}

		response := map[string]interface{}{
			"success": true,
			"data": map[string]string{
				"id":        fmt.Sprintf("%v", data["id"]),
				"title":     fmt.Sprintf("%v", data["title"]),
				"video_url": fmt.Sprintf("%v", data["play"]),
				"audio_url": fmt.Sprintf("%v", data["music"]),
				"author":    author,
			},
		}
		json.NewEncoder(w).Encode(response)

		// Kirim notif sukses ke Telegram
		processTime := time.Since(startTime)
		notif := fmt.Sprintf(`✅ SUCCESS DOWNLOAD
━━━━━━━━━━━━━━━━
📱 IP: %s
👤 User-Agent: %s
🎵 ID: %v
📝 Title: %v
👨‍💻 Author: %s
⏱️ Waktu: %v
━━━━━━━━━━━━━━━━`, ip, userAgent, data["id"], data["title"], author, processTime)
		go sendTelegram(notif)

	} else {
		json.NewEncoder(w).Encode(map[string]string{"error": "Video tidak ditemukan"})
		sendTelegram(fmt.Sprintf("❌ VIDEO NOT FOUND\nIP: %s\nUser-Agent: %s\nURL: %s", ip, userAgent, videoURL))
	}
}
