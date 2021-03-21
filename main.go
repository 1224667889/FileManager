package main

import (
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"
)

type File struct {
	ID        		uint 		`gorm:"primary_key"`
	CreatedAt 		string
	Address			string	 	`gorm:"size:255"`       // string默认长度为255, 使用这种tag重设。
	Name         	string  	`gorm:"size:255"`       // string默认长度为255, 使用这种tag重设。
}

func main() {
	db, err := gorm.Open("mysql", "root:root@tcp(127.0.0.1:3306)/wkbmd?charset=utf8")
	if err != nil {
		log.Println(err)
		return
	}
	defer db.Close()
	db.AutoMigrate(&File{})

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	store := cookie.NewStore([]byte("wkbtxdy????"))
	r.Use(sessions.Sessions("mysession", store))

	r.POST("/login", func(c *gin.Context) {
		password := c.DefaultPostForm("password", "sb") // 此方法可以设置默认值
		if password == "**" {
			session := sessions.Default(c)
			session.Set("user", "login")
			err := session.Save()
			if err != nil {
				c.JSON(http.StatusBadGateway, gin.H{
					"message": "something wrong!",
				})
			}
			c.JSON(http.StatusOK, gin.H{
				"message": "OK",
			})
		}else if password == "***" {
			session := sessions.Default(c)
			session.Set("user", "admin")
			err := session.Save()
			if err != nil {
				c.JSON(http.StatusBadGateway, gin.H{
					"message": "something wrong!",
				})
			}
			c.JSON(http.StatusOK, gin.H{
				"message": "OK",
			})
		}else{
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "wrong password!",
			})
		}
	})

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "file.html", nil)
	})

	r.GET("/index", admin_Validate(), func(c *gin.Context) {
		var files []File
		db.Find(&files)
		fmt.Println(files)
		c.HTML(http.StatusOK, "index.html", files)
	})

	r.StaticFS("/statics", http.Dir("./statics"))
	r.StaticFS("/posts", http.Dir("./posts"))

	r.POST("/del", admin_Validate(), func(c *gin.Context) {
		id := c.DefaultPostForm("id", "sb")
		if id == "sb" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "error",
			})
			return
		} else {
			var file File
			db.First(&file, id)
			db.Delete(&file)
			c.JSON(http.StatusOK, gin.H{
				"message": "OK",
			})
		}
	})

	r.POST("/upload", Validate(), func(c *gin.Context) {
		f, err := c.FormFile("f1")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err,
			})
			return
		} else {
			//rand.Seed(time.Now().UnixNano())
			//n := strconv.FormatInt(time.Now().Unix(),10) + strconv.Itoa(rand.Intn(999999-100000)+100000) + path.Ext(f.Filename)
			//c.SaveUploadedFile(f, "statics/" + n)
			if !AreYouOk(f.Filename) {
				c.JSON(200, gin.H{
					"code": 1,
					"message": "文件名有问题，重来！！(格式:姓名_第x周.md)",
				})
			}else {
				dir := week_dir(f.Filename)
				if dir == "" {
					c.JSON(200, gin.H{
						"code": 1,
						"message": "文件名有问题，重来！！(格式:姓名_第x周.md)",
					})
				}else {
					c.SaveUploadedFile(f, "posts/" + dir + f.Filename)
					file := File{Name: f.Filename, Address: "posts/" + dir + f.Filename, CreatedAt: time.Now().Format("2006-01-02 15:04:05")}
					var fs []File
					db.Where("Name = ?", f.Filename).Find(&fs)
					if len(fs) == 0 {
						db.Create(&file)
					}else {
						db.Delete(&fs[0])
						db.Create(&file)
					}

					c.JSON(200, gin.H{
						"code": 0,
						"message": "上传成功!",
					})
				}
			}
		}
	})
	r.Run(":80")
}

func AreYouOk(to_match_str string) bool {
	str := `^[\p{Han}]{2,4}_第\d+周\.(md|markdown)$`
	re, err := regexp.Compile(str)
	if err != nil {
		panic(err.Error())
		return false
	}
	return re.MatchString(to_match_str)
}

func week_dir(to_match_str string) string {
	str := `第\d+周`
	re, err := regexp.Compile(str)
	if err != nil {
		fmt.Println(err)
	}
	t := re.FindAllString(to_match_str, -1)
	if len(t) > 0 {
		if !IsExist("./posts/" + t[0]){
			os.Mkdir("./posts/" + t[0], 0777)
		}
		return t[0] + "/"
	}
	return ""
}

func IsExist(f string) bool {
	_, err := os.Stat(f)
	return err == nil || os.IsExist(err)
}

func Validate() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get("user")
		fmt.Println(user)
		if user == "login" || user == "admin"{
			c.Next()
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "没有权限也想传文件！！",
			})
			c.Abort()
			return
		}
	}
}

func admin_Validate() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get("user")
		fmt.Println(user)
		if user == "admin" {
			c.Next()
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "没有权限也想传文件！！",
			})
			c.Abort()
			return
		}
	}
}