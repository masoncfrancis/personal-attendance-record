package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const version = "0.1.1" 

func main() {
	// Load .env file if present
	_ = godotenv.Load()

	checkFlag := flag.Bool("check", false, "Run a single attendance check, update log and exit")
	checkFlagShort := flag.Bool("c", false, "Run a single attendance check, update log and exit (shorthand)")
	noSaveFlag := flag.Bool("no-save", false, "Run a single attendance check and print the result (no file saved)")
	noSaveFlagShort := flag.Bool("n", false, "Run a single attendance check and print the result (no file saved) (shorthand)")
	pathFlag := flag.Bool("path", false, "Display the path of the file where the attendance record is kept")
	pathFlagShort := flag.Bool("p", false, "Display the path of the file where the attendance record is kept (shorthand)")
	versionFlag := flag.Bool("version", false, "Display the software version")
	versionFlagShort := flag.Bool("v", false, "Display the software version (shorthand)")
	flag.Parse()

	if *versionFlag || *versionFlagShort {
		fmt.Println("personal-attendance-record (par) version", version)
		return
	}

	// Check required env vars
	requiredVars := []string{"PAR_FILE", "PAR_MODE"}
	for _, v := range requiredVars {
		if os.Getenv(v) == "" {
			fmt.Fprintf(os.Stderr, "%s environment variable not set\n", v)
			os.Exit(1)
		}
	}

	parFile := os.Getenv("PAR_FILE")

	if *pathFlag || *pathFlagShort {
		fmt.Println(parFile)
		return
	}

	if *noSaveFlag || *noSaveFlagShort {
		result, err := checkAttendance()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		if result {
			fmt.Println("In office")
		} else {
			fmt.Println("Not in office")
		}
		return
	}

	if *checkFlag || *checkFlagShort {
		err := runCheck(parFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		return
	}

	flag.Usage()
}

func runCheck(parFile string) error {
	mode := os.Getenv("PAR_MODE")
	if mode == "" {
		return errors.New("PAR_MODE environment variable not set")
	}

	var inOffice bool
	var err error

	switch strings.ToLower(mode) {
	case "corporateproxy":
		inOffice, err = checkCorporateProxy()
	case "checkurl":
		inOffice, err = checkURL()
	default:
		return fmt.Errorf("Unknown PAR_MODE: %s", mode)
	}
	if err != nil {
		return err
	}

	return logAttendance(parFile, inOffice)
}

func checkCorporateProxy() (bool, error) {
	url := os.Getenv("PAR_URL")
	if url == "" {
		url = "https://www.google.com"
	}
	proxyAddr := os.Getenv("PAR_PROXY_ADDRESS")
	if proxyAddr == "" {
		return false, errors.New("PAR_PROXY_ADDRESS environment variable not set")
	}
	// Remove URL scheme if present
	proxyAddr = stripScheme(proxyAddr)

	// Try to access the URL with no proxy
	transport := &http.Transport{Proxy: nil}
	client := &http.Client{Transport: transport, Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	urlErr := err
	if err == nil && resp.StatusCode == 200 {
		resp.Body.Close()
		// URL is reachable without proxy: user is NOT in office
		return false, nil
	}
	if resp != nil {
		resp.Body.Close()
	}

	// Try to ping the proxy (ICMP ping via shell)
	proxyHost := proxyAddr
	if i := strings.Index(proxyHost, ":"); i != -1 {
		proxyHost = proxyHost[:i]
	}
	pingErr := pingHost(proxyHost)
	if pingErr == nil {
		// URL is NOT reachable without proxy, but proxy IS reachable: user is IN office
		return true, nil
	}

	msg := ""
	if urlErr != nil {
		msg += fmt.Sprintf("Cannot reach URL (%s): %v. ", url, urlErr)
	} else {
		msg += fmt.Sprintf("Cannot reach URL (%s): status not 200. ", url)
	}
	msg += fmt.Sprintf("Cannot ping proxy host (%s): %v.", proxyHost, pingErr)
	return false, errors.New(strings.TrimSpace(msg))
}

func checkURL() (bool, error) {
	url := os.Getenv("PAR_URL")
	if url == "" {
		return false, errors.New("PAR_URL environment variable not set")
	}
	transport := &http.Transport{Proxy: nil}
	client := &http.Client{Transport: transport, Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err == nil && resp.StatusCode == 200 {
		resp.Body.Close()
		return true, nil // Can reach URL: in office
	}
	if resp != nil {
		resp.Body.Close()
	}
	return false, nil // Cannot reach URL: not in office
}

func checkAttendance() (bool, error) {
	mode := os.Getenv("PAR_MODE")
	if mode == "" {
		return false, errors.New("PAR_MODE environment variable not set")
	}

	switch strings.ToLower(mode) {
	case "corporateproxy":
		return checkCorporateProxy()
	case "checkurl":
		return checkURL()
	default:
		return false, fmt.Errorf("Unknown PAR_MODE: %s", mode)
	}
}

func logAttendance(parFile string, inOffice bool) error {
	increment := os.Getenv("PAR_RECORD_INCREMENTS")
	if increment == "" {
		increment = "daily"
	}
	now := time.Now().Local()
	var record []string
	humanTime := now.Format("2006-01-02 03:04:05 PM MST")
	if increment == "all" {
		record = []string{humanTime, boolToString(inOffice)}
		return appendCSV(parFile, record)
	}
	// daily: only one entry per day
	return upsertDailyCSV(parFile, now, inOffice)
}

func boolToString(b bool) string {
	if b {
		return "Y"
	}
	return "N"
}

func appendCSV(filename string, record []string) error {
	// Check if file exists and is empty (to write header)
	writeHeader := false
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		writeHeader = true
	} else {
		fi, err := os.Stat(filename)
		if err == nil && fi.Size() == 0 {
			writeHeader = true
		}
	}
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	if writeHeader {
		if err := w.Write([]string{"Date/Time", "In Office?"}); err != nil {
			return err
		}
	}
	if err := w.Write(record); err != nil {
		return err
	}
	w.Flush()
	return w.Error()
}

func upsertDailyCSV(filename string, now time.Time, inOffice bool) error {
	// Read all records
	var records [][]string
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	r := csv.NewReader(f)
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			break // ignore parse errors
		}
		records = append(records, rec)
	}
	// Check for header, add if missing
	header := []string{"Date/Time", "In Office?"}
	hasHeader := false
	if len(records) > 0 && len(records[0]) == 2 && records[0][0] == header[0] && records[0][1] == header[1] {
		hasHeader = true
	}
	if !hasHeader {
		records = append([][]string{header}, records...)
	}
	f.Truncate(0)
	f.Seek(0, 0)
	w := csv.NewWriter(f)
	now = now.Local()
	dateStr := now.Format("2006-01-02")
	humanTime := now.Format("2006-01-02 03:04:05 PM MST")
	updated := false
	// Start from 1 if header present
	startIdx := 1
	if !hasHeader {
		startIdx = 1
	}
	for i := startIdx; i < len(records); i++ {
		rec := records[i]
		if len(rec) > 0 && len(rec[0]) >= 10 && rec[0][:10] == dateStr {
			if rec[1] != "Y" && inOffice {
				records[i][1] = "Y"
			}
			updated = true
		}
	}
	if !updated {
		records = append(records, []string{humanTime, boolToString(inOffice)})
	}
	for _, rec := range records {
		w.Write(rec)
	}
	w.Flush()
	return w.Error()
}

type netDialer struct {
	timeout time.Duration
}

func (d *netDialer) Dial(network, address string) (io.Closer, error) {
	ch := make(chan struct {
		conn io.Closer
		err  error
	}, 1)
	go func() {
		conn, err := dial(network, address)
		ch <- struct {
			conn io.Closer
			err  error
		}{conn, err}
	}()
	select {
	case res := <-ch:
		return res.conn, res.err
	case <-time.After(d.timeout):
		return nil, errors.New("timeout")
	}
}

func dial(network, address string) (io.Closer, error) {
	return (&net.Dialer{}).Dial(network, address)
}

// Add this helper function:
func stripScheme(addr string) string {
	if strings.HasPrefix(addr, "http://") {
		return strings.TrimPrefix(addr, "http://")
	}
	if strings.HasPrefix(addr, "https://") {
		return strings.TrimPrefix(addr, "https://")
	}
	return addr
}

func pingHost(host string) error {
	// Use system ping command, 1 packet, 2s timeout
	cmd := exec.Command("ping", "-c", "1", "-W", "2", host)
	return cmd.Run()
}
