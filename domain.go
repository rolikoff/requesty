package main

import (
	"database/sql"
)

const DomainTableCreationQuery = `CREATE TABLE IF NOT EXISTS domains
(
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
	timestamp INTEGER NOT NULL,
    name TEXT NOT NULL,
    counter INTEGER NOT NULL
)`

type domain struct {
	Id        int    `json:"-"`
	Timestamp int    `json:"-"`
	Name      string `json:"name"`
	Requests  int    `json:"requests"`
}

// Checks if the domain is already in the DB, if so, then the requests counter field will be updated.
// If there's no such domain record found, the func will insted a new domain record into the DB.
func (d *domain) upsertDomain(db *sql.DB) error {
	foundDomain, _ := getDomain(db, &d.Timestamp, &d.Name)

	if foundDomain.Id == 0 {
		// well, looks like there's no domain with the same name and timestamp at the DB. Inserting a new record.
		d.createDomain(db)
	} else {
		// looks like we have the domain under the same timestamp already, lets update the requests counter.
		foundDomain.Requests = foundDomain.Requests + d.Requests
		foundDomain.updateDomainCounter(db)
	}

	return nil
}

// Creates a new domain record in the DB.
func (d *domain) createDomain(db *sql.DB) error {
	err := db.QueryRow(
		"INSERT INTO domains(timestamp, name, counter) VALUES(?, ?, ?) RETURNING id",
		d.Timestamp, d.Name, d.Requests).Scan(&d.Id)

	if err != nil {
		return err
	}

	return nil
}

// Updates the domain's counter.
func (d *domain) updateDomainCounter(db *sql.DB) error {
	_, err := db.Exec(
		"UPDATE domains SET counter =$1 WHERE id=$2",
		d.Requests, d.Id)

	if err != nil {
		return err
	}

	return nil
}

// Returns a domain object with the data that currently persist in DB for provided timestamp and domain name.
func getDomain(db *sql.DB, timestamp *int, name *string) (domain, error) {
	foundDomain := domain{}

	err := db.QueryRow(
		"SELECT id, name, counter, timestamp FROM domains WHERE name=? AND timestamp=?",
		name, timestamp).Scan(&foundDomain.Id, &foundDomain.Name, &foundDomain.Requests, &foundDomain.Timestamp)

	if err != nil {
		if err != sql.ErrNoRows {
			return foundDomain, err
		}
		return foundDomain, nil
	}

	return foundDomain, nil
}

func getTopTenDomains(db *sql.DB, fromTimestamp, toTimeStamp int64) ([]domain, error) {
	var domains []domain
	// Top 10 domains last round minute (with several requests). If now 54min and 30 sec. We need stats for 53 minutes.
	rows, err := db.Query(
		`SELECT name, SUM(counter) 
		   FROM domains WHERE timestamp >= ? AND timestamp < ? 
		   GROUP BY name 
		   ORDER BY counter DESC 
		   LIMIT 10`,
		fromTimestamp, toTimeStamp)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		d := domain{}
		if err := rows.Scan(&d.Name, &d.Requests); err != nil {
			return domains, err
		}
		domains = append(domains, d)
	}
	if err = rows.Err(); err != nil {
		return domains, err
	}

	return domains, nil
}
