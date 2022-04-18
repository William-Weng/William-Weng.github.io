package main

import (
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"william/utility"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const sqliteDatabasePath = "./material/sqliteDB.sqlite3" // 把資料庫存放在material資料夾下

type DatabaseType int8

const (
	Sqlite3 DatabaseType = 0 // 支援Sqlite3
	MySql   DatabaseType = 1 // 支援MySql
)

// 資料表的長相
type Product struct {
	gorm.Model
	Code  string `json:"code" gorm:"index:idx_name,unique"`
	Price uint   `json:"price"`
}

// 使用Gmail寄信的資訊
type gmailInformation struct {
	fromMail string
	toMail   string
	password string
	host     string
	port     uint
	title    string
	message  string
}

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

func main() {
	go cronSchedule()
	error := initSetting()
	utility.Println(error)
}

// 初始設定
func initSetting() error {

	database, error := Sqlite3.createDatabase(sqliteDatabasePath)

	if error != nil {
		return error
	}

	error = createTables(database)

	if error != nil {
		return error
	}

	initRouter(database)

	return nil
}

// 初始化Router相關設定
func initRouter(database *gorm.DB) {

	router := gin.Default()
	router.MaxMultipartMemory = 8 << 20

	registerWebAPI(router, database)

	router.Run(":8080")
}

// 建立資料庫
func (dbType DatabaseType) createDatabase(path string) (*gorm.DB, error) {

	var database *gorm.DB
	var error error

	switch dbType {
	case Sqlite3:
		database, error = gorm.Open(sqlite.Open(path), &gorm.Config{})
	case MySql:
		break
	}

	return database, error
}

// 建立表格
func createTables(database *gorm.DB) error {
	return database.AutoMigrate(&Product{})
}

// 註冊API
func registerWebAPI(router *gin.Engine, database *gorm.DB) {
	insertProduct(router, database)
	selectProduct(router, database)
	deleteProduct(router, database)
	partialUpdateProduct(router, database)
	selectAllProduct(router, database)
	mailServer(router, database)
}

// MARK: - WebAPI
// <POST>新增單一商品 => http://localhost:8080/product + {"code":"iPhone","price":20000}
func insertProduct(router *gin.Engine, database *gorm.DB) {

	var emptyProduct Product

	router.POST("/product", func(context *gin.Context) {

		dictionary := utility.RequestBodyToMap(context)
		result, error := emptyProduct._Insert(database, dictionary)

		utility.ContextJSON(context, http.StatusOK, result, error)
	})
}

// <GET>搜尋單一商品 => http://localhost:8080/product/87
func selectProduct(router *gin.Engine, database *gorm.DB) {

	var emptyProduct Product

	router.GET("/product/:id", func(context *gin.Context) {

		id, error := utility.StringToInt(context.Param("id"))
		result := emptyProduct._Select(database, uint(id))

		if result.ID == 0 {
			utility.ContextJSON(context, http.StatusOK, nil, error)
			return
		}

		utility.ContextJSON(context, http.StatusOK, result, error)
	})
}

// <DELETE>刪除單一商品 => http://localhost:8080/product/87
func deleteProduct(router *gin.Engine, database *gorm.DB) {

	var emptyProduct Product

	router.DELETE("/product/:id", func(context *gin.Context) {

		id, error := utility.StringToInt(context.Param("id"))
		isSuccess := emptyProduct._Delete(database, uint(id))
		result := map[string]interface{}{"isSuccess": isSuccess}

		utility.ContextJSON(context, http.StatusOK, result, error)
	})
}

// [<PATCH>更新單一商品部分內容](https://ihower.tw/blog/archives/6483) => http://localhost:8080/product/87
func partialUpdateProduct(router *gin.Engine, database *gorm.DB) {

	var emptyProduct Product

	router.PATCH("/product/:id", func(context *gin.Context) {

		id, error := utility.StringToInt(context.Param("id"))
		dictionary := utility.RequestBodyToMap(context)

		result := emptyProduct._PartialUpdate(database, uint(id), dictionary)
		utility.ContextJSON(context, http.StatusOK, result, error)
	})
}

// [<GET>搜尋部分商品](https://stackoverflow.com/questions/51534285/how-to-access-gorm-model-id/71865857) => http://localhost:8080/product?price=1487&code=iPhone
func selectAllProduct(router *gin.Engine, database *gorm.DB) {

	var emptyProduct Product

	router.GET("/product", func(context *gin.Context) {

		code := context.Query("code")
		price := context.Query("price")

		dictionary := map[string]interface{}{
			"code":  code,
			"price": price,
		}

		utility.Println(dictionary)

		result := emptyProduct._SelectAll(database, dictionary)
		utility.ContextJSON(context, http.StatusOK, result, nil)
	})
}

// [<POST>寄垃圾信](https://gist.github.com/jpillora/cb46d183eca0710d909a) => http://localhost:8080/mail + {"to":"linuxice0609@gmail.com","title":"垃垃信測試","message":"2022/4/14 Golang smtp test!!!!"}
func mailServer(router *gin.Engine, database *gorm.DB) {

	router.POST("/mail", func(context *gin.Context) {

		dictionary := utility.RequestBodyToMap(context)

		info := gmailInformation{
			fromMail: "linuxice0609@gmail.com",
			password: "<password>",
			host:     "smtp.gmail.com",
			port:     587,
			toMail:   dictionary["to"].(string),
			title:    dictionary["title"].(string),
			message:  dictionary["message"].(string),
		}

		result, error := gmailServer(info)
		utility.ContextJSON(context, http.StatusOK, result, error)
	})
}

// MARK: - [Product小工具](https://gorm.io/zh_CN/docs/query.html)
// 新增單一商品 => ["code":"iPhone","price":20000]
func (_product *Product) _Insert(database *gorm.DB, dictionary map[string]interface{}) (map[string]interface{}, error) {

	code := dictionary["code"].(string)
	price := uint(dictionary["price"].(float64))

	isSuccess := false
	error := database.Create(&Product{Code: code, Price: price}).Error

	if error == nil {
		isSuccess = true
	}

	result := map[string]interface{}{"isSuccess": isSuccess}
	return result, error
}

// 搜尋單一商品 => id = 87
func (_product *Product) _Select(database *gorm.DB, id uint) Product {

	var product Product
	database.Take(&product, "id=?", id)

	return product
}

// 刪除單一商品 => http://localhost:8080/product/87
func (_product *Product) _Delete(database *gorm.DB, id uint) bool {

	product := _product._Select(database, id)
	database.Delete(&product, id)

	return product.ID != 0
}

// 更新單一商品部分內容 => http://localhost:8080/product/87
func (_product *Product) _PartialUpdate(database *gorm.DB, id uint, dictionary map[string]interface{}) map[string]interface{} {

	isSuccess := false
	product := _product._Select(database, id)
	error := database.Model(&product).Updates(dictionary).Error

	if error == nil {
		isSuccess = true
	}

	result := map[string]interface{}{}
	result["isSuccess"] = isSuccess

	return result
}

// 搜尋部分商品 (過濾條件) => http://localhost:8080/product
func (_product *Product) _SelectAll(database *gorm.DB, dictionary map[string]interface{}) []Product {

	var products []Product
	var _database = database

	code := dictionary["code"].(string)
	price, _ := utility.StringToInt(dictionary["price"].(string))

	if len(code) != 0 {
		_database = _database.Where("code LIKE ?", "%"+code+"%")
	}

	if price > 0 {
		_database = _database.Where("price >= ?", price)
	}

	_database.Find(&products)

	return products
}

// [使用Gmail寄信](https://support.google.com/mail/?p=InvalidSecondFactor)
func gmailServer(info gmailInformation) (map[string]interface{}, error) {

	var isSuccess bool = false
	var error error = nil

	smtpServer := fmt.Sprintf("%s:%d", info.host, info.port)
	message := fmt.Sprintf("From: %s\nTo: %s\nSubject: %s\n\n %s", info.fromMail, info.toMail, info.title, info.message)
	authentication := smtp.PlainAuth("", info.fromMail, info.password, info.host)

	error = smtp.SendMail(smtpServer, authentication, info.fromMail, []string{info.toMail}, []byte(message))

	if error == nil {
		isSuccess = true
	}

	result := map[string]interface{}{"isSuccess": isSuccess}
	return result, error
}

// [定時排程任務](https://www.readfog.com/a/1637371620314157056)
func cronSchedule() {

	schedule := cron.New(cron.WithSeconds())
	index := 1
	schedule.AddFunc("* * * * * *", func() {
		utility.Println("每一秒執行一次", index)
		index++
	})

	schedule.Start()
	// time.Sleep(time.Minute * 1)
	select {}
}
