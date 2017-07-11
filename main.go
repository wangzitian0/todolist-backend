package main

import (
    "gopkg.in/gin-gonic/gin.v1"
    "github.com/jinzhu/gorm"
    _ "github.com/jinzhu/gorm/dialects/sqlite"
    "net/http"
    "strconv"
    "fmt"
)

func Database() *gorm.DB {
    //open a db connection
    db, err := gorm.Open("sqlite3", "gorm.db")
    if err != nil {
        panic("failed to connect database")
    }
    return db
}

type Todo struct {
    gorm.Model
    Title     string `form:"title" json:"title" binding:"exists,required"`
    Priority  int `form:"priority" json:"priority" binding:"exists"`
    Completed bool `form:"completed" json:"completed" binding:"exists"`
}

type TodoResponse struct {
    ID        uint `json:"id" xml:"id"`
    Title     string `json:"title" xml:"title"`
    Priority  int `json:"priority" xml:"priority"`
    Completed bool `json:"completed" xml:"completed"`
}

func (todo Todo) Response() (interface{}){
    return TodoResponse{
        ID: todo.ID,
        Title: todo.Title,
        Priority: todo.Priority,
        Completed: todo.Completed,
    }
}

func CreateTodo(c *gin.Context) {
    var todo Todo
    //fmt.Println(todo)
    if err := c.Bind(&todo); err != nil {
        //fmt.Println(err)
        c.JSON(http.StatusBadRequest, gin.H{"Data binding error" : err})
        return
    }
    db := Database()
    defer db.Close()

    if err := db.Save(&todo).Error; err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"Database error" : err})
        return
    }
    c.JSON(http.StatusCreated, todo.Response())
    //注释掉上面一句，取消下面一句可以返回 xml
    //c.XML(http.StatusCreated, todo.Response())
}

func FetchAllTodo(c *gin.Context) {
    var todos []Todo
    var _todos []interface{}

    db := Database()

    page_str := c.Query("page");
    page, err := strconv.Atoi(page_str);
    if  err!=nil {
        page = 0
    }
    per_page_str := c.Query("per_page")
    per_page, err := strconv.Atoi(per_page_str);
    if  err!=nil {
        per_page = 10
    }
    db = db.Limit(per_page).Offset(per_page * page)

    min_priority_str := c.Query("min_priority")
    min_priority, err := strconv.Atoi(min_priority_str);
    if  err==nil {
        db = db.Where("priority >= ?", min_priority)
    }
    db.Find(&todos)

    if (len(todos) <= 0) {
        c.JSON(http.StatusNotFound, gin.H{"error" : "No todo found!"})
        return
    }

    //transforms the todos for building a good response
    for _, todo := range todos {
        _todos = append(_todos, todo.Response())
    }
    c.JSON(http.StatusOK, _todos)
    //c.XML(http.StatusOK, _todos)
}

func FetchSingleTodo(c *gin.Context) {
    var todo Todo
    todoId := c.Param("id")

    db := Database()
    db.First(&todo, todoId)

    if (todo.ID == 0) {
        c.JSON(http.StatusNotFound, gin.H{"error" : "No todo found!"})
        return
    }

    c.JSON(http.StatusOK, todo.Response())
}

func UpdateTodo(c *gin.Context) {
    var todo, todoTmp Todo
    todoId := c.Param("id")
    db := Database()
    db.First(&todoTmp, todoId)

    if (todoTmp.ID == 0) {
        c.JSON(http.StatusNotFound, gin.H{"message" : "No todo found!"})
    }else if err := c.Bind(&todo); err == nil {
        fmt.Println(todoTmp)
        todo.ID = todoTmp.ID
        db.Save(&todo)
        c.JSON(http.StatusOK, todo.Response())
    }else {
        c.JSON(http.StatusBadRequest, gin.H{"Bind Error" : err })
    }
}

func PartialUpdateTodo(c *gin.Context) {
    var todo Todo
    todoId := c.Param("id")
    db := Database()
    db.First(&todo, todoId)

    if (todo.ID == 0) {
        c.JSON(http.StatusNotFound, gin.H{"message" : "No todo found!"})
    }else if err := c.Bind(&todo); err == nil {
        //--- 可以手工一个个域更新，也可以一次性更新 todo ---
        //db.Model(&todo).Update("title", c.PostForm("title"))
        //db.Model(&todo).Update("completed", c.PostForm("completed"))
        db.Model(&todo).Update(todo)
        c.JSON(http.StatusOK, todo.Response())
    }else {
        c.JSON(http.StatusBadRequest, gin.H{"Bind Error" : err })
    }
}

func DeleteTodo(c *gin.Context) {
    var todo Todo
    todoId := c.Param("id")
    db := Database()
    db.First(&todo, todoId)

    if (todo.ID == 0) {
        c.JSON(http.StatusNotFound, gin.H{ "error" : "No todo found!"})
        return
    }

    db.Delete(&todo)
    c.JSON(http.StatusOK, gin.H{"message" : "Todo deleted successfully!"})
}
func main() {
    db := Database()
    db.AutoMigrate(&Todo{})
    defer db.Close()

    //db.Create(&Todo{Title: "L1212", Completed: 1})
    router := gin.Default()
    v1 := router.Group("/api/v1/todos")
    {
        v1.POST("/", CreateTodo)
        v1.GET("/", FetchAllTodo)
        v1.GET("/:id", FetchSingleTodo)
        v1.PUT("/:id", UpdateTodo)
        v1.PATCH("/:id", PartialUpdateTodo)
        v1.DELETE("/:id", DeleteTodo)
    }
    router.Run()
}
