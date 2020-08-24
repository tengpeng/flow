package main

import (
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	log "github.com/sirupsen/logrus"
)

var db *gorm.DB

func main() {
	log.Info("Bayesnote flow started")

	os.Remove("flow.db")

	initDB()
	setUpTestDB()

	testFlow()
	//init internal flow data

	// db, err := gorm.Open("sqlite3", "flow.db")
	// if err != nil {
	// 	panic("failed to connect database")
	// }
	// defer db.Close()

	// // Migrate the schema
	// db.AutoMigrate(&Product{})

	// // Create
	// db.Create(&Product{Code: "L1212", Price: 1000})

	// // Read
	// var product Product
	// db.First(&product, 1)                   // find product with id 1
	// db.First(&product, "code = ?", "L1212") // find product with code l1212

	// // Update - update product's price to 2000
	// db.Model(&product).Update("Price", 2000)

	// // Delete - delete product
	// db.Delete(&product)
}

func testFlow() {
	var f flow
	db.First(&f, 1)
	f.generateDep()

	f.run()
}
