package handler

import (
	"net/mail"
	"time"

	"github.com/adhtanjung/go_rest_api/database"
	"github.com/adhtanjung/go_rest_api/helper"
	"github.com/adhtanjung/go_rest_api/middleware"
	"github.com/adhtanjung/go_rest_api/model"
	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func valid(email string) bool {
	_, err := mail.ParseAddress(email)

	return err == nil

}

// Get All Users from db
func GetAllUsers(c *fiber.Ctx) error {
	db := database.DB.Db
	var users []model.User

	// find all users in the database
	db.Find(&users)

	// If no user found, return an error
	if len(users) == 0 {
		return c.Status(404).JSON(fiber.Map{"status": "error", "message": "Users not found", "data": nil})
	}

	// return users
	return c.Status(200).JSON(fiber.Map{"status": "sucess", "message": "Users Found", "data": users})
}

// GetSingleUser from db
func GetSingleUser(c *fiber.Ctx) error {
	db := database.DB.Db

	// get id params
	id := c.Params("id")

	var user model.User

	// find single user in the database by id
	db.Find(&user, "id = ?", id)

	if user.ID == uuid.Nil {
		return c.Status(404).JSON(fiber.Map{"status": "error", "message": "User not found", "data": nil})
	}

	return c.Status(200).JSON(fiber.Map{"status": "success", "message": "User Found", "data": user})
}

func CreateUser(c *fiber.Ctx) error {
	db := database.DB.Db
	user := new(model.User)

	// Store the body in the user and return error if encountered
	if err := c.BodyParser(user); err != nil {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Something's wrong with your input", "data": err})
	}

	// Check if username is already taken, if true return error
	if err := db.Where("username= ?", &user.Username).First(&user).Error; err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Username already taken", "data": err})
	}

	// Check if email is already taken, if true return error
	if err := db.Where("email= ?", &user.Email).First(&user).Error; err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Email already taken", "data": err})
	}

	hashed, err := helper.HashPassword(user.Password)

	// replace with a hashed password
	user.Password = hashed

	err = db.Create(&user).Error
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Could not create user", "data": err})
	}

	// Return the created user
	return c.Status(201).JSON(fiber.Map{"status": "success", "message": "User has created", "data": user})
}

func UpdateUser(c *fiber.Ctx) error {
	type updateUser struct {
		Username string `json:"username"`
	}

	db := database.DB.Db

	var user model.User

	// get id params
	id := c.Params("id")

	// find single user in the database by id
	db.Find(&user, "id = ?", id)

	if user.ID == uuid.Nil {
		return c.Status(404).JSON(fiber.Map{"status": "error", "message": "User not found", "data": nil})
	}

	var updateUserData updateUser
	err := c.BodyParser(&updateUserData)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Something's wrong with your input", "data": err})
	}

	user.Username = updateUserData.Username

	// Save the Changes
	db.Save(&user)

	// Return the updated user
	return c.Status(200).JSON(fiber.Map{"status": "success", "message": "users Found", "data": user})

}

func DeleteUserByID(c *fiber.Ctx) error {
	db := database.DB.Db
	var user model.User

	// get id params
	id := c.Params("id")

	// find single user in the database by id
	db.Find(&user, "id = ?", id)

	if user.ID == uuid.Nil {
		return c.Status(404).JSON(fiber.Map{"status": "error", "message": "User not found", "data": nil})

	}

	err := db.Delete(&user, "id = ?", id).Error

	if err != nil {
		return c.Status(404).JSON(fiber.Map{"status": "error", "message": "Failed to delete user", "data": nil})
	}

	return c.Status(200).JSON(fiber.Map{"status": "success", "message": "User deleted"})
}

// Login user

func Login(c *fiber.Ctx) error {
	db := database.DB.Db
	var user model.User

	var input middleware.LoginInput

	// binding user input to a struct
	if err := c.BodyParser(&input); err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	// set a variable depending on the condition
	var query string
	if valid(input.Identity) {
		query = "email= ?"
	} else {
		query = "username= ?"
	}

	if err := db.Where(query, input.Identity).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status": "error", "message": "User does not exists",
		})

	}

	identity := input.Identity
	pass := input.Password

	if !helper.ValidatePassword(pass, user.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status": "error", "messvalidage": "Password incorrect",
		})
	}

	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["identity"] = identity
	claims["admin"] = true
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	t, err := token.SignedString([]byte("secret"))
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{"status": "success", "message": "Success login", "token": t})

}
