package tables

import (
	"database/sql"

	"github.com/pkg/errors"
)

func init() {
	MigrationClient.AddMigration(Up_20220818101352, Down_20220818101352)
}

func Up_20220818101352(tx *sql.Tx) error {
	logger.Info.Println("Increasing width of software.vendor...")

	//-----------------
	// Add temp column.
	//-----------------
	if _, err := tx.Exec(
		`ALTER TABLE software ADD COLUMN vendor_wide varchar(114) NULL, ALGORITHM=INPLACE, LOCK=NONE`); err != nil {
		return errors.Wrapf(err, "creating temp column for vendor")
	}

	//---------------------
	// Add uniq constraint
	//---------------------
	if _, err := tx.Exec(
		"ALTER TABLE software ADD constraint unq_name UNIQUE (name, version, source, `release`, vendor_wide, arch)"); err != nil {
		return errors.Wrapf(err, "adding new uniquess constraint")
	}

	//------------------
	// Update in batches
	//------------------
	const updateStmt = `UPDATE software SET vendor_wide = vendor WHERE vendor_wide IS NULL LIMIT 500`
	for {
		res, err := tx.Exec(updateStmt)
		if err != nil {
			return errors.Wrapf(err, "updating temp vendor column")
		}
		affected, err := res.RowsAffected()
		if err != nil {
			return errors.Wrapf(err, "updating temp vendor column")
		}
		if affected == 0 {
			break
		}
	}

	//----------------
	// Drop old index
	//----------------
	if _, err := tx.Exec(`ALTER TABLE software DROP KEY name`); err != nil {
		return errors.Wrapf(err, "dropping old index")
	}

	//------------------
	// Rename old column
	//------------------
	if _, err := tx.Exec(`ALTER TABLE software CHANGE vendor vendor_old varchar(32) DEFAULT '' NOT NULL, ALGORITHM=INPLACE, LOCK=NONE`); err != nil {
		return errors.Wrapf(err, "dropping old column")
	}

	// ---------------
	// Rename column
	// ---------------
	if _, err := tx.Exec(
		`ALTER TABLE software CHANGE vendor_wide vendor varchar(114) DEFAULT '' NOT NULL, ALGORITHM=INPLACE, LOCK=NONE`); err != nil {
		return errors.Wrapf(err, "dropping old column")
	}

	logger.Info.Println("Done increasing width of software.vendor...")
	return nil
}

func Down_20220818101352(tx *sql.Tx) error {
	return nil
}