package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
        "crypto/sha256"
)


var (
	AppName = "godown"
	Author  = "Iksan rumasoreng"
)

const expectedAuthorHash = "5a2994a034bbe76f7b70ed69b249d06bedaaa23c0bf47d60b4d4a5698319d704"

func verifyIntegrity() bool {
	hash := sha256.Sum256([]byte(Author))
	return fmt.Sprintf("%x", hash) == expectedAuthorHash
}


// List keyword pengabaian (preview/thumb/ads)
var ignoreKeywords = []string{"preview", "thumb", "sample", "trailer", "poster", "advert"}

func isIgnored(link string) bool {
	lower := strings.ToLower(link)
	for _, kw := range ignoreKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

func fetchHTML(targetURL, referer string) (string, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	if referer != "" {
		req.Header.Set("Referer", referer)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// Universal Scraper Engine
func extractUniversalMediaURL(pageURL string) (string, error) {
	html, err := fetchHTML(pageURL, pageURL)
	if err != nil {
		return "", fmt.Errorf("gagal mengambil halaman: %w", err)
	}

	// STAGE 1: Cek Tag <video> atau <source>
	fmt.Println("  [Stage 1] Memindai Tag <video> / <source>...")
	reVideo := regexp.MustCompile(`(?i)<(?:video|source)[^>]+src=["']([^"']+)["']`)
	matchesVideo := reVideo.FindAllStringSubmatch(html, -1)
	for _, m := range matchesVideo {
		if len(m) > 1 && !isIgnored(m[1]) {
			return fixURL(m[1], pageURL), nil
		}
	}

	// STAGE 2: Cek Variabel JavaScript Player (JWPlayer, Video.js, Clappr, HLS.js)
	fmt.Println("  [Stage 2] Memindai Variabel JS Player (file:, source:, src:)...")
	reJS := regexp.MustCompile(`(?i)(?:file|source|src|url|playlist)\s*:\s*["']([^"']+\.(?:mp4|m3u8)[^"']*)["']`)
	matchesJS := reJS.FindAllStringSubmatch(html, -1)
	for _, m := range matchesJS {
		if len(m) > 1 {
			clean := strings.ReplaceAll(m[1], `\/`, `/`)
			if !isIgnored(clean) {
				return fixURL(clean, pageURL), nil
			}
		}
	}

	// STAGE 3: Cek <iframe> embed player & Scrape Iframe tersebut
	fmt.Println("  [Stage 3] Memindai Tag <iframe> Embed...")
	reIframe := regexp.MustCompile(`(?i)<iframe[^>]+src=["']([^"']+)["']`)
	matchesIframe := reIframe.FindAllStringSubmatch(html, -1)
	for _, m := range matchesIframe {
		if len(m) > 1 {
			iframeURL := fixURL(m[1], pageURL)
			if !isIgnored(iframeURL) && strings.HasPrefix(iframeURL, "http") {
				fmt.Printf("   -> Ditemukan iframe: %s (Mencoba scraping iframe...)\n", iframeURL)
				iframeHTML, err := fetchHTML(iframeURL, pageURL)
				if err == nil {
					// Cari stream di dalam iframe
					reIframeMedia := regexp.MustCompile(`https?://[^"'\s\\]+\.(?:mp4|m3u8)[^"'\s\\]*`)
					if media := reIframeMedia.FindString(iframeHTML); media != "" && !isIgnored(media) {
						return media, nil
					}
				}
			}
		}
	}

	// STAGE 4: Fallback Regex Pola .mp4 atau .m3u8 umum
	fmt.Println("  [Stage 4] Fallback Scanning Link .mp4 / .m3u8...")
	reGeneral := regexp.MustCompile(`https?://[^"'\s\\]+\.(?:mp4|m3u8)[^"'\s\\]*`)
	matchesGeneral := reGeneral.FindAllString(html, -1)
	for _, m := range matchesGeneral {
		clean := strings.ReplaceAll(m, `\/`, `/`)
		if !isIgnored(clean) {
			return clean, nil
		}
	}

	return "", fmt.Errorf("tidak dapat menemukan link stream video di situs ini")
}

// Normalisasi URL relatif menjadi URL absolut
func fixURL(target, base string) string {
	target = strings.TrimSpace(target)
	if strings.HasPrefix(target, "//") {
		return "https:" + target
	}
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		return target
	}
	baseURL, err := url.Parse(base)
	if err != nil {
		return target
	}
	relURL, err := url.Parse(target)
	if err != nil {
		return target
	}
	return baseURL.ResolveReference(relURL).String()
}

func promptInput(reader *bufio.Reader, promptText string) string {
	fmt.Print(promptText)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func main() {
        // 🔒 1. PROTEKSI HASH & BRANDING (Tambahkan di paling atas)
        if !verifyIntegrity() {
                fmt.Println("❌ [ERROR] Modifikasi ilegal terdeteksi! Nama pembuat asli telah diubah.")
                os.Exit(1)
        }
        fmt.Printf("⚡ Powered by %s (%s)\n\n", AppName, Author)

        // ----------------------------------------------------
        // 🚀 2. LOGIKA UTAMA (Kode Asli Kamu)
        // ----------------------------------------------------
        if len(os.Args) < 2 {
                fmt.Println("use: godown <URL_WEBSITE>")
                return
        }

        pageURL := os.Args[1]
        reader := bufio.NewReader(os.Stdin)

        fmt.Printf("🔍 Universal Scraping in: %s ...\n", pageURL)
        videoURL, err := extractUniversalMediaURL(pageURL)
        if err != nil {
                fmt.Printf("❌ [ERROR SCRAPE] %v\n", err)
                fmt.Println("💡 note: If the site uses protection, consider using yt-dlp backend.")
                return
        }

        fmt.Printf("\n🔗 direct link caught : %s\n\n", videoURL)

        defaultFileName := "video_download.mp4"
        if strings.Contains(videoURL, ".m3u8") {
                defaultFileName = "video_download.mp4"
        }

        inputName := promptInput(reader, fmt.Sprintf("✏️  enter file name [Default: %s]: ", defaultFileName))
        fileName := defaultFileName
        if inputName != "" {
                if !strings.HasSuffix(inputName, ".mp4") && !strings.HasSuffix(inputName, ".mkv") {
                        inputName += ".mp4"
                }
                fileName = inputName
        }

        defaultDir := "/sdcard/Download"
        if _, err := os.Stat(defaultDir); os.IsNotExist(err) {
                defaultDir, _ = os.Getwd()
        }

        inputDir := promptInput(reader, fmt.Sprintf("📁 enter directory [Default: %s]: ", defaultDir))
        targetDir := defaultDir
        if inputDir != "" {
                targetDir = inputDir
        }

        _ = os.MkdirAll(targetDir, 0755)
        fullOutputPath := filepath.Join(targetDir, fileName)

        fmt.Printf("\n🚀 downloading ...\n")
        fmt.Printf("📍 target directory  : %s\n\n", fullOutputPath)

        cmd := exec.Command("aria2c",
                "-x16",
                "-s16",
                "-k1M",
                "--user-agent=Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
                fmt.Sprintf("--referer=%s", pageURL),
                "-d", targetDir,
                "-o", fileName,
                videoURL,
        )

        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        cmd.Stdin = os.Stdin

        err = cmd.Run()
        if err != nil {
                fmt.Printf("\n❌ [ERROR] download failed: %v\n", err)
                return
        }

        fmt.Printf("\n✅ [DONE] save in: %s\n", fullOutputPath)
}

