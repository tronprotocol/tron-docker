package utils

import (
	"archive/tar"
	"compress/gzip"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/tronprotocol/tron-docker/config"
	"golang.org/x/net/html"
)

// Function to fetch and parse HTML, extracting absolute links
func fetchAndExtractLinks(webURL string) ([]string, error) {
	resp, err := http.Get(webURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %s", resp.Status)
	}

	parsedBaseURL, err := url.Parse(webURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %v", err)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error parsing HTML: %v", err)
	}

	var links []string
	var extractLinks func(*html.Node)
	extractLinks = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					// Resolve relative URLs to absolute URLs
					resolvedURL := parsedBaseURL.ResolveReference(&url.URL{Path: attr.Val}).String()
					links = append(links, resolvedURL)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractLinks(c)
		}
	}

	extractLinks(doc)
	return links, nil
}

func generateDateList() []string {
	var dateList []string
	curTime := time.Now()

	// Loop for 180 days (from 1 to 179, inclusive)
	for i := 1; i < 180; i++ {
		// Calculate the date by subtracting i days from the current time
		startDate := curTime.Add(-time.Duration(i) * 24 * time.Hour)
		// Format the date as "yyyyMMdd"
		dateStr := startDate.Format("20060102")

		// If it's the first 30 days, add all dates
		if i < 30 {
			dateList = append(dateList, dateStr)
		} else {
			// Otherwise, check if the day is 10, 20, or 30 and add it
			if startDate.Day() == 10 || startDate.Day() == 20 || startDate.Day() == 30 {
				dateList = append(dateList, dateStr)
			}
		}
	}

	return dateList
}

// extractBackupPart extracts the last part of the path from a given URL.
func extractBackupPart(rawURL string) (string, error) {
	// Parse the input URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %v", err)
	}

	// Extract the last part of the path
	backupPart := path.Base(parsedURL.Path)

	// Remove trailing slash if present
	backupPart = strings.TrimSuffix(backupPart, "/")

	return backupPart, nil
}

func ShowSnapshotListForNile() error {
	lists := generateDateList()

	fmt.Println("Available backup:")
	for _, v := range lists {
		fmt.Println("  backup" + v)
	}
	return nil
}

func getSnapshotList(domain string, https bool) ([]string, error) {
	webURL := "http://" + domain
	if https {
		webURL = "https://" + domain
	}

	links, err := fetchAndExtractLinks(webURL)
	if err != nil {
		return nil, err
	}

	var snapshots []string
	for _, link := range links {
		basePath, err := extractBackupPart(link)
		if err != nil {
			return nil, err
		}
		if len(basePath) == 0 {
			continue
		}

		snapshots = append(snapshots, basePath)
	}

	return snapshots, nil
}

func ShowSnapshotList(domain string, https bool) error {
	snapshots, err := getSnapshotList(domain, https)
	if err != nil {
		return err
	}

	fmt.Println("Available backup:")
	for _, snapshot := range snapshots {
		fmt.Println("  " + snapshot)
	}

	return nil
}

func GetLatestSnapshot(domain string, https bool) (string, error) {
	snapshots, err := getSnapshotList(domain, https)
	if err != nil {
		return "", err
	}

	// Remove "backup" prefix and store cleaned dates
	var dates []string
	for _, backup := range snapshots {
		dates = append(dates, strings.TrimPrefix(backup, "backup"))
	}

	// Sort dates in ascending order
	sort.Strings(dates)

	// Find the largest date
	largestDate := dates[len(dates)-1]

	return "backup" + largestDate, nil
}

func GetLatestNileSnapshot(domain string, https bool) (string, error) {
	snapshots := generateDateList()

	// Remove "backup" prefix and store cleaned dates
	var dates []string
	for _, backup := range snapshots {
		dates = append(dates, strings.TrimPrefix(backup, "backup"))
	}

	// Sort dates in ascending order
	sort.Strings(dates)

	// Find the largest date
	largestDate := dates[len(dates)-1]

	return "backup" + largestDate, nil
}

// getFileNameFromURL extracts the file name from the URL path
func getFileNameFromURL(fileURL string) (string, error) {
	parsedURL, err := url.Parse(fileURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %v", err)
	}
	return path.Base(parsedURL.Path), nil
}

// calculateMD5 computes the MD5 checksum of a file
func calculateMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to calculate MD5: %v", err)
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// readMD5FromFile parses the expected MD5 checksum from the file formatted as "md5sum  filename"
func readMD5FromFile(md5FilePath string) (string, string, error) {
	file, err := os.Open(md5FilePath)
	if err != nil {
		return "", "", fmt.Errorf("failed to open checksum file: %v", err)
	}
	defer file.Close()

	var md5Hash, filename string
	_, err = fmt.Fscanf(file, "%s %s", &md5Hash, &filename)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse MD5 from file: %v", err)
	}

	md5Hash = strings.TrimSpace(md5Hash)
	filename = strings.TrimSpace(filename)
	return md5Hash, filename, nil
}

// formatSize converts bytes to a human-readable string with appropriate units.
func formatSize(size int64) string {
	const (
		_          = iota
		KB float64 = 1 << (10 * iota)
		MB
		GB
	)

	switch {
	case size >= int64(GB):
		return fmt.Sprintf("%.2f GB", float64(size)/GB)
	case size >= int64(MB):
		return fmt.Sprintf("%.2f MB", float64(size)/MB)
	case size >= int64(KB):
		return fmt.Sprintf("%.2f KB", float64(size)/KB)
	default:
		return fmt.Sprintf("%d Bytes", size)
	}
}

// downloadFileWithProgress downloads a file from the given URL and shows progress with estimated time
func DownloadFileWithProgress(fileURL, md5FilePath string) (string, error) {
	expectedMD5 := ""
	expectedFilename := ""

	// If md5FilePath is provided, read the expected MD5 and filename
	if md5FilePath != "" {
		var err error
		expectedMD5, expectedFilename, err = readMD5FromFile(md5FilePath)
		if err != nil {
			return "", fmt.Errorf("error reading MD5 file: %v", err)
		}
	}

	// Extract the original file name from the URL
	filename, err := getFileNameFromURL(fileURL)
	if err != nil {
		return "", fmt.Errorf("failed to get file name: %v", err)
	}

	// If an MD5 file was provided, check if filenames match
	if md5FilePath != "" && filename != expectedFilename {
		return "", fmt.Errorf("filename mismatch: expected %s, got %s", expectedFilename, filename)
	}

	fmt.Println("Downloading file:", fileURL)

	// Create the file to save the content
	outFile, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer outFile.Close()

	// Send HTTP GET request
	resp, err := http.Get(fileURL)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %v", err)
	}
	defer resp.Body.Close()

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	// Get the content length for progress tracking
	contentLength := resp.ContentLength
	if contentLength <= 0 {
		return "", fmt.Errorf("unable to determine file size")
	}

	// Initialize progress bar with estimated time
	bar := progressbar.NewOptions64(
		contentLength,
		progressbar.OptionSetDescription(fmt.Sprintf("Downloading... (Total: %s)", formatSize(contentLength))),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionThrottle(10*time.Millisecond), // More frequent updates
		progressbar.OptionClearOnFinish(),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionSetElapsedTime(true),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionShowIts(),
		progressbar.OptionOnCompletion(func() {
			fmt.Println("Download complete:", filename)
		}),
	)

	// Read the response body in smaller chunks for frequent updates
	buffer := make([]byte, 32*1024) // 32KB buffer for finer updates
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			_, writeErr := outFile.Write(buffer[:n])
			if writeErr != nil {
				return "", fmt.Errorf("failed to write to file: %v", writeErr)
			}

			_ = bar.Add(n) // Update progress bar with exact number of bytes written
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("error reading data: %v", err)
		}
	}

	/// If an MD5 file was provided, perform checksum verification
	if md5FilePath != "" {
		// Calculate MD5 checksum of the downloaded file
		fmt.Println("Verifying MD5 checksum...")
		calculatedMD5, err := calculateMD5(filename)
		if err != nil {
			return "", fmt.Errorf("error calculating MD5: %v", err)
		}

		fmt.Println("Expected MD5:", expectedMD5)
		fmt.Println("Calculated MD5:", calculatedMD5)

		if calculatedMD5 == expectedMD5 {
			fmt.Println("MD5 verification successful!")
		} else {
			fmt.Println("MD5 verification failed!")
			return "", fmt.Errorf("checksum mismatch")
		}
	}

	return filename, nil
}

func GenerateSnapshotDownloadURL(domain, backup, nType string) string {
	if nType == "full" {
		if IsNile(domain) {
			return config.SnapshotDataSource[config.STNileLevel][domain].DownloadURL + "/" + backup + "/FullNode_output-directory.tgz"
		}
		return "http://" + domain + "/" + backup + "/FullNode_output-directory.tgz"
	} else if nType == "lite" {
		if IsNile(domain) {
			return config.SnapshotDataSource[config.STNileLevel][domain].DownloadURL + "/" + backup + "/LiteFullNode_output-directory.tgz"
		}
		return "http://" + domain + "/" + backup + "/LiteFullNode_output-directory.tgz"
	}
	return ""
}

func GenerateSnapshotMD5DownloadURL(domain, backup, nType string) string {
	if nType == "full" {
		if IsNile(domain) {
			return config.SnapshotDataSource[config.STNileLevel][domain].DownloadURL + "/" + backup + "/FullNode_output-directory.tgz.md5sum"
		}
		return "http://" + domain + "/" + backup + "/FullNode_output-directory.tgz.md5sum"
	} else if nType == "lite" {
		if IsNile(domain) {
			return config.SnapshotDataSource[config.STNileLevel][domain].DownloadURL + "/" + backup + "/LiteFullNode_output-directory.tgz.md5sum"
		}
		return "http://" + domain + "/" + backup + "/LiteFullNode_output-directory.tgz.md5sum"
	}
	return ""
}

func GetDownloadedSnapshotName(domain, backup, nType string) string {
	if nType == "full" {
		return "FullNode_output-directory.tgz"
	} else if nType == "lite" {
		return "LiteFullNode_output-directory.tgz"
	}
	return ""
}

// ExtractTgzWithStatus extracts a `.tgz` file into a directory with status updates
func ExtractTgzWithStatus(tgzFile, destDir string) error {
	// Open `.tgz` file
	file, err := os.Open(tgzFile)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Create Gzip reader
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %v", err)
	}
	defer gzReader.Close()

	// Create Tar reader
	tr := tar.NewReader(gzReader)

	// Track total bytes extracted
	var totalExtractedSize int64
	var totalFilesExtracted int64

	// Start extracting and show status updates
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break // Done
		}
		if err != nil {
			return fmt.Errorf("error reading tar: %v", err)
		}

		sanitizedName := filepath.Clean(header.Name)
		if strings.Contains(sanitizedName, "..") || strings.HasPrefix(sanitizedName, "/") || strings.HasPrefix(sanitizedName, "../") {
			return fmt.Errorf("invalid file path in archive: %s", header.Name)
		}

		// Target file path
		target := filepath.Join(destDir, sanitizedName)
		// Ensure the target path is still within the destination directory
		if !strings.HasPrefix(target, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("attempted directory traversal: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to create directory: %v", err)
			}
		case tar.TypeReg:
			// Ensure parent directory exists
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory: %v", err)
			}

			// Create file
			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed to create file: %v", err)
			}
			defer outFile.Close()

			// Read data in chunks and write to the output file
			buf := make([]byte, 32*1024) // 32KB buffer for reading
			for {
				n, err := tr.Read(buf)
				if n > 0 {
					// Write the data to the file
					if _, writeErr := outFile.Write(buf[:n]); writeErr != nil {
						return fmt.Errorf("failed to write file: %v", writeErr)
					}

					// Update the total extracted size and file count
					totalExtractedSize += int64(n)
				}
				if err == io.EOF {
					break
				}
				if err != nil {
					return fmt.Errorf("error reading file: %v", err)
				}
			}

			// Increment the file count
			totalFilesExtracted++
		case tar.TypeSymlink:
			// Create symlink
			if err := os.Symlink(header.Linkname, target); err != nil {
				return fmt.Errorf("failed to create symlink: %v", err)
			}
		case tar.TypeLink:
			// Create hard link
			linkTarget := filepath.Join(destDir, header.Linkname)
			// Sanitize the link target to prevent directory traversal
			sanitizedLinkTarget := filepath.Clean(linkTarget)
			if !strings.HasPrefix(sanitizedLinkTarget, filepath.Clean(destDir)+string(os.PathSeparator)) {
				return fmt.Errorf("attempted directory traversal in hard link: %s -> %s", header.Name, header.Linkname)
			}

			if err := os.Link(linkTarget, target); err != nil {
				return fmt.Errorf("failed to create hard link: %v", err)
			}
		default:
			fmt.Printf("Skipping unknown file type: %v\n", header.Typeflag)
		}

		// Display the current status
		fmt.Printf("\rExtracted %d files, %s bytes total", totalFilesExtracted, formatSize(totalExtractedSize))
	}

	// Print completion message
	fmt.Printf("\nExtraction complete: %d files extracted, %s bytes total\n", totalFilesExtracted, formatSize(totalExtractedSize))
	return nil
}
