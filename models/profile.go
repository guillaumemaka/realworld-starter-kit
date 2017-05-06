package models

import "fmt"

// Profile is the profile structure
type Profile struct {
	ID        uint   `db:"id" json:"-"`
	Username  string `db:"username" json:"username"`
	Bio       string `db:"bio" json:"bio"`
	Image     string `db:"image" json:"image"`
	Following bool   `db:"following" json:"following"`
}

// ProfileResponse is the required response json
type ProfileResponse struct {
	Profile *Profile `json:"profile"`
}

const (
	qGetProfileByUsername = `SELECT 
	id,username, bio,image, 
	CASE WHEN uf.usr_following_id IS null THEN 0 ELSE 1 END AS following 
	FROM users u
	LEFT OUTER JOIN usr_following uf 
		ON u.id = uf.usr_following_id
		and uf.usr_id = ?
	WHERE u.username = ?`
	qFollowUser   = `INSERT INTO usr_following (usr_id, usr_following_id) VALUES (?,?)`
	qUnfollowUser = `DELETE FROM usr_following WHERE usr_id=? AND usr_following_id=?`
)

// GetProfileByUsername returns a profile for the supplied username
func (adb *AppDB) GetProfileByUsername(username string, whosasking uint) (*Profile, error) {
	var p Profile
	if err := adb.DB.QueryRow(qGetProfileByUsername, whosasking, username).Scan(&p.ID, &p.Username, &p.Bio, &p.Image, &p.Following); err != nil {
		return &p, err
	}
	return &p, nil
}

// FollowUser adds a record to allow users to follow each other
func (adb *AppDB) FollowUser(currentUserID, followUserID uint) error {
	stmt, err := adb.DB.Prepare(qFollowUser)
	if err != nil {
		return err
	}
	res, err := stmt.Exec(currentUserID, followUserID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return fmt.Errorf("Did not insert row for %d to follow %d", currentUserID, followUserID)
	}
	return nil
}

// UnfollowUser adds a record to allow users to follow each other
func (adb *AppDB) UnfollowUser(currentUserID, followUserID uint) error {
	stmt, err := adb.DB.Prepare(qUnfollowUser)
	if err != nil {
		return err
	}
	res, err := stmt.Exec(currentUserID, followUserID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return fmt.Errorf("Did not delete row for %d to unfollow %d", currentUserID, followUserID)
	}
	return nil
}
