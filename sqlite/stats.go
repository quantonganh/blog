package sqlite

import (
	"archive/zip"
	"database/sql"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"

	"github.com/quantonganh/blog"
	"github.com/rs/zerolog"
)

const (
	ip2LocationFileName    = "IP2LOCATION-LITE-DB1.CSV"
	ip2LocationZipFileName = ip2LocationFileName + ".zip"
)

type statService struct {
	logger zerolog.Logger
	db     *DB
}

func NewStatService(logger zerolog.Logger, db *DB) blog.StatService {
	return &statService{
		logger: logger,
		db:     db,
	}
}

// Insert inserts new activity into SQLite
func (s *statService) Insert(e *blog.Event) error {
	tx, err := s.db.sqlDB.Begin()
	if err != nil {
		return fmt.Errorf("failed to start a transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if err := tx.Rollback(); err != nil {
				s.logger.Error().Err(err).Msg("error rollbacking")
			}
			return
		}
		if err := tx.Commit(); err != nil {
			s.logger.Error().Err(err).Msg("error committing")
		}
	}()

	_, err = tx.Exec("INSERT INTO activities (user_id, ip_address, country, browser, os, referer, url, time) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		e.UserID, e.IP, e.Country, e.Browser, e.OS, e.Referer, e.URL, e.Time)
	if err != nil {
		return fmt.Errorf("failed to insert into activities table: %w", err)
	}

	return nil
}

func (s *statService) ImportIP2LocationDB(token string) error {
	imported, err := checkIP2LocationData(s.db.sqlDB)
	if err != nil {
		s.logger.Error().Err(err).Msg("error checking IP2Location data")
	}
	if !imported {
		if err := downloadIP2LocationDB(token); err != nil {
			return err
		}

		cmd := exec.Command("sqlite3", "db/stats.db", "-cmd", fmt.Sprintf(".import --csv --skip 1 %s ip2location", ip2LocationFileName))
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("error importing CSV data into ip2location table: %s: %w", string(output), err)
		}
		defer os.Remove(ip2LocationFileName)

		_, err = s.db.sqlDB.Exec("INSERT INTO migrations (name) VALUES ('ip2location')")
		if err != nil {
			return fmt.Errorf("error marking migration as applied: %w", err)
		}
	}

	return nil
}

func checkIP2LocationData(db *sql.DB) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM ip2location LIMIT 1);`

	var exists int
	err := db.QueryRow(query).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return exists == 1, nil
}

func downloadIP2LocationDB(token string) error {
	resp, err := http.Get(fmt.Sprintf("https://www.ip2location.com/download/?token=%s&file=DB1LITE", token))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	file, err := os.Create(ip2LocationZipFileName)
	if err != nil {
		return fmt.Errorf("error creating ip2Location file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	r, err := zip.OpenReader(ip2LocationZipFileName)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, file := range r.File {
		if file.Name != ip2LocationFileName {
			continue
		}

		outFile, err := os.Create(ip2LocationFileName)
		if err != nil {
			return err
		}
		defer outFile.Close()

		rc, err := file.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		_, err = io.Copy(outFile, rc)
		if err != nil {
			return err
		}
	}

	if err := os.Remove(ip2LocationZipFileName); err != nil {
		return err
	}

	return nil
}

func (s *statService) GetCountryFromIP(ip string) (string, error) {
	ipInteger, err := ipToInteger(ip)
	if err != nil {
		return "", err
	}

	var country string
	if err := s.db.sqlDB.QueryRow(`
		SELECT country FROM ip2location WHERE ? BETWEEN start_ip AND end_ip ORDER BY end_ip LIMIT 1
	`, ipInteger).Scan(&country); err != nil {
		return "", err
	}
	if country == "-" {
		country = "Unknown"
	}

	return country, nil
}

func (s *statService) Top10VisitedPages() ([]blog.PageStats, error) {
	rows, err := s.db.sqlDB.Query("SELECT url, count(id) AS visits FROM activities GROUP BY url ORDER BY count(id) DESC LIMIT 10;")
	if err != nil {
		return nil, err
	}

	var pages []blog.PageStats
	for rows.Next() {
		var p blog.PageStats
		if err := rows.Scan(&p.URL, &p.Visits); err != nil {
			return nil, err
		}
		pages = append(pages, p)
	}

	return pages, nil
}

func (s *statService) Top10Referers(domain string) ([]blog.RefererStats, error) {
	rows, err := s.db.sqlDB.Query(`
SELECT
	CAST((COUNT(id) * 100.0 / (SELECT COUNT(id) FROM activities)) AS int) AS Share,
	referer,
	count(id) AS visits
FROM activities
WHERE referer NOT LIKE '%' || ? || '%'
GROUP BY referer
ORDER BY count(id) DESC
LIMIT 10;`, domain)
	if err != nil {
		return nil, err
	}

	var referers []blog.RefererStats
	for rows.Next() {
		var r blog.RefererStats
		if err := rows.Scan(&r.Share, &r.Referer, &r.Visits); err != nil {
			return nil, err
		}
		referers = append(referers, r)
	}

	return referers, nil
}

func (s *statService) Top10Countries() ([]blog.CountryStats, error) {
	rows, err := s.db.sqlDB.Query(`
SELECT
	CAST((COUNT(id) * 100.0 / (SELECT COUNT(id) FROM activities)) AS int) AS Share,
	country,
	COUNT(id) AS visits
FROM activities
GROUP BY country
ORDER BY visits DESC
LIMIT 10;`)
	if err != nil {
		return nil, err
	}

	var countries []blog.CountryStats
	for rows.Next() {
		var c blog.CountryStats
		if err := rows.Scan(&c.Share, &c.Country, &c.Visits); err != nil {
			return nil, err
		}
		countries = append(countries, c)
	}

	return countries, nil
}

func (s *statService) Top10Browsers() ([]blog.BrowserStats, error) {
	rows, err := s.db.sqlDB.Query(`
SELECT
	CAST((COUNT(id) * 100.0 / (SELECT COUNT(id) FROM activities)) AS int) AS Share,
	browser,
	COUNT(id) AS visits
FROM activities
GROUP BY browser
ORDER BY visits DESC
LIMIT 10;`)
	if err != nil {
		return nil, err
	}

	var browsers []blog.BrowserStats
	for rows.Next() {
		var b blog.BrowserStats
		if err := rows.Scan(&b.Share, &b.Browser, &b.Visits); err != nil {
			return nil, err
		}
		browsers = append(browsers, b)
	}

	return browsers, nil
}

func (s *statService) Top10OperatingSystems() ([]blog.OSStats, error) {
	rows, err := s.db.sqlDB.Query(`
SELECT
	ROUND((COUNT(id) * 100.0 / (SELECT COUNT(id) FROM activities)), 0) AS Share,
	os,
	COUNT(id) AS visits
FROM activities
GROUP BY os
ORDER BY visits DESC
LIMIT 10;`)
	if err != nil {
		return nil, err
	}

	var operatingSystems []blog.OSStats
	for rows.Next() {
		var os blog.OSStats
		if err := rows.Scan(&os.Share, &os.OS, &os.Visits); err != nil {
			return nil, err
		}
		operatingSystems = append(operatingSystems, os)
	}

	return operatingSystems, nil
}

func ipToInteger(ipAddr string) (uint32, error) {
	parsedIP := net.ParseIP(ipAddr)
	if parsedIP == nil {
		return 0, fmt.Errorf("invalid IP address: %s", ipAddr)
	}

	ipBytes := parsedIP.To4()
	if ipBytes == nil {
		return 0, fmt.Errorf("not an IPv4 address: %s", ipAddr)
	}

	ipInteger := uint32(ipBytes[0])<<24 | uint32(ipBytes[1])<<16 | uint32(ipBytes[2])<<8 | uint32(ipBytes[3])

	return ipInteger, nil
}
