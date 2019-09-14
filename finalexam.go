package main

import (
	"database/sql"
	"fmt"
	"os"
	"log"
	"strconv"
	"net/http"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type Customer struct {
	ID     int `json:"id"`
	Name  string `json:"name"`
	Email  string `json:"email"`
	Status string `json:"status"`
}


type MyApp struct{
	MyDB *sql.DB
}

func authMiddleware(c *gin.Context) {
	token := c.GetHeader("Authorization")
	fmt.Println("token :", token)
	if token != "token2019" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		c.Abort()
		return
	}
	c.Next()
}

func (app MyApp) createCustHandler(c *gin.Context) {
	var cs Customer

 	err := c.ShouldBindJSON(&cs)
 	if err != nil {
  		c.JSON(http.StatusBadRequest, err.Error())
  		return
 	}
	
	var customer []Customer

 	customer = append(customer, cs)
	row := app.MyDB.QueryRow("INSERT INTO customer (name, email, status) VALUES ($1, $2, $3) RETURNING id, name, email, status", cs.Name , cs.Email , cs.Status)
	
	ooo := Customer{}
	err = row.Scan(&ooo.ID, &ooo.Name, &ooo.Email, &ooo.Status)
	if err != nil {
	 fmt.Println("can't scan id", err)
	 return
	}
   
	c.JSON(http.StatusCreated, ooo)
}

func (app MyApp) getCustHandler(c *gin.Context) {
	idx := c.Param("id")
	ii, _ := strconv.Atoi(idx)
	row, err := app.MyDB.Query("SELECT id, name, email, status FROM customer WHERE id=$1", ii)
	if err != nil {
		log.Fatal("can't prepare statment SELECT", err)
	}

	row.Next() 
 	ooo := Customer{}
	err = row.Scan(&ooo.ID, &ooo.Name, &ooo.Email, &ooo.Status)
	if err != nil {
		log.Fatal("Can not scan row into variables", err)
	}
	c.JSON(http.StatusOK, ooo)

}

func (app MyApp) getAllCustHandler(c *gin.Context) {
	stmt, err := app.MyDB.Prepare("SELECT id, name, email, status FROM customer")
	if err != nil {
		log.Fatal("Can not prepare query all customer statement", err)
	}
	rows, err := stmt.Query()
	if err != nil {
		log.Fatal("Can not query all customer", err)
	}

	ooo := []Customer{}
	for rows.Next() {
		var id int
		var name, email, status string
		err := rows.Scan(&id, &name, &email, &status)
		if err != nil {
			log.Fatal("Can not scan row into variable", err)
		}
		ooo = append(ooo,Customer{id, name, email, status})
	}
	c.JSON(http.StatusOK, ooo)
}

func (app MyApp) updateCustHandler(c *gin.Context) {
	idx := c.Param("id")
	ii, _ := strconv.Atoi(idx)

	var cs Customer
	err := c.ShouldBindJSON(&cs)
	if err != nil {
		 c.JSON(http.StatusBadRequest, err.Error())
		 return
	}

	_, err = app.MyDB.Query("UPDATE customer SET name=$2 , email=$3 , status=$4 WHERE id=$1", ii, cs.Name, cs.Email, cs.Status)
	//ooo := Customer{}
	//err = row.Scan(&ooo.ID, &ooo.Name, &ooo.Email, &ooo.Status)
	if err != nil {
	 	fmt.Println("can't update id", err)
	 	return
	}
	c.JSON(http.StatusOK, cs)
}

func (app MyApp) deleteCustHandler(c *gin.Context) {
	idx:= c.Param("id")
	ii, _ := strconv.Atoi(idx)

	row, err := app.MyDB.Query("DELETE FROM customer WHERE id=$1;", ii)

	if err != nil {
		log.Fatal("Can not prepare statement delete", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return	
	}

	row.Next() 

	if err != nil {
		log.Fatal("Error excute delete ", err)
	}
	c.JSON(http.StatusOK, gin.H{"message":"customer deleted"})
}


func main() {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Can not connect database, error", err)
	}
	defer db.Close()
	app := MyApp{db}

	createTb := `
	CREATE TABLE IF NOT EXISTS customer (
		id SERIAL PRiMARY KEY,
		name TEXT,
		email TEXT,
		status TEXT
	);
	`
	_, err = db.Exec(createTb)

	if err != nil {
		log.Fatal("Can not create table, error", err)		
	}

	r := gin.Default()

	r.Use(authMiddleware)
	r.POST("/customers", app.createCustHandler)
	r.GET("/customers/:id", app.getCustHandler)
	r.GET("/customers", app.getAllCustHandler)
	r.PUT("/customers/:id", app.updateCustHandler)
	r.DELETE("/customers/:id", app.deleteCustHandler)
	r.Run(":2019") 

}

