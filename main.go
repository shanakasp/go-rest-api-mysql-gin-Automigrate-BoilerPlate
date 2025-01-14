package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var db *sql.DB
var err error

func main() {
	db, err = sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/restgoq")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	// Auto migrate the database schema
	autoMigrate()

	router := gin.Default()

	router.POST("/users", createUser) // POST /users
	router.GET("/user/:id", getUser)   // GET /users/:id
	router.GET("/users", getUsers)     // GET /users
	router.PUT("/user/:id", updateUser) // PUT /users/:id
	router.DELETE("/users/:id", deleteUser) // DELETE /users/:id
	
	router.Run(":8080")
}

func autoMigrate() {
	// Create users table if it doesn't exist
	query := `
	CREATE TABLE IF NOT EXISTS userss (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		deleted_at TIMESTAMP NULL DEFAULT NULL
	);`
	_, err = db.Exec(query)
	if err != nil {
		panic(err.Error())
	}
}

func createUser(c *gin.Context) {

	var body struct {
	    Name string
	}
	c.Bind(&body)

	//prepare SQL statment
	stmt, err := db.Prepare("INSERT INTO userss(name) VALUES(?)")
	if err != nil {
	    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer stmt.Close()

	// Get last inserted ID
	id, err := stmt.Exec(body.Name)
	if err != nil {
	    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	//return the created user as JSON
	user := User {
		ID : fmt.Sprintf("%d", id),
		Name : body.Name,
	}

	c.JSON(http.StatusCreated, user)
}

func getUsers(c *gin.Context) {
    c.Header("Content-Type", "application/json")

	rows, err := db.Query("SELECT id, name FROM userss")
	if err != nil {
	    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
	    var user User
	    if err := rows.Scan(&user.ID, &user.Name); err != nil {
	        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	        return
	    }
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
	    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	    return
	}

	c.JSON(http.StatusOK, users)
}

func getUser(c *gin.Context){
	c.Header("Content-Type", "application/json")

	id := c.Param("id")

	row := db.QueryRow("SELECT id, name FROM userss WHERE id = ?", id)

	var user User
	if err := row.Scan(&user.ID, &user.Name); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		
	    return
	}

	c.JSON(http.StatusOK, user)
}

func updateUser(c *gin.Context) {
	id := c.Param("id")

	var body struct {
		Name string `json:"name"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stmt, err := db.Prepare("UPDATE userss SET name = ? WHERE id = ? AND deleted_at IS NULL")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(body.Name, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("User with ID = %s was updated", id)})
}

func deleteUser(c*gin.Context) {
id:= c.Param("id")

stmt, err := db.Prepare("DELETE FROM userss WHERE id = ?")
if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
}
defer stmt.Close()

_, err = stmt.Exec(id)
if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
}
c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("User with ID = %s was deleted", id)})
}
