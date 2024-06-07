package models

import (
	"errors"
	"video_stream/db"
	"video_stream/utils"
)

type User struct {
	ID       int64  `json:"id"`
	Email    string `binding:"required" json:"email"`
	Password string `binding:"required" json:"password"`
}

func (u *User) Save() error {
	query := "INSERT INTO users (email, password) VALUES ($1, $2)"
	stmt, err := db.DB.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	hashedPassword, err := utils.HashPassword(u.Password)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(u.Email, hashedPassword)
	if err != nil {
		return err
	}

	//var userId int64
	//err = stmt.QueryRow(u.Email, u.Password).Scan(&userId)
	//if err != nil {
	//	return err
	//}
	//u.ID = userId
	return nil
}

func (u *User) ValidateCredentials() error {
	query := "SELECT id, password FROM users WHERE email = $1"
	row := db.DB.QueryRow(query, u.Email)

	var retrievedPassword string
	err := row.Scan(&u.ID, &retrievedPassword)
	if err != nil {
		return errors.New("Invalid email or password")
	}

	passwordIsValid := utils.CheckPasswordHash(u.Password, retrievedPassword)
	if !passwordIsValid {
		return errors.New("Invalid email or password")
	}

	return nil
}
