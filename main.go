package main

import (
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

// Config 定义了与 config.yml 文件结构相对应的结构体
type Config struct {
	Server struct {
		Port string `yaml:"port"` // 服务器监听端口
	} `yaml:"server"`
	Storage struct {
		TmpPath string `yaml:"tmp_path"` // 文件存储临时路径
	} `yaml:"storage"`
	Random struct {
		StringLength int `yaml:"string_length"` // 随机字符串长度
	} `yaml:"random"`
}

func main() {
	gin.SetMode(gin.ReleaseMode) // 设置 Gin 框架为发布模式
	r := gin.Default()           // 创建默认的 Gin 路由器

	// 从 config.yml 文件加载配置
	config, err := loadConfig("config.yml")
	if err != nil {
		panic(err) // 如果配置文件加载失败，则终止程序
	}

	// 设置静态资源目录和 HTML 模板
	r.Static("/static", "./static")
	r.LoadHTMLFiles("index.html")

	rand.Seed(time.Now().UnixNano()) // 初始化随机数生成器

	setupRoutes(r, config) // 配置路由

	r.Run(config.Server.Port) // 启动服务器并监听配置文件指定的端口
}

// setupRoutes 配置路由和处理函数
func setupRoutes(r *gin.Engine, config Config) {
	r.GET("/", func(c *gin.Context) {
		randomString := generateRandomString(config.Random.StringLength) // 生成指定长度的随机字符串
		c.Redirect(http.StatusFound, "/"+randomString)                   // 重定向到随机字符串对应的 URL
	})

	r.GET("/:path", func(c *gin.Context) {
		path := c.Param("path")
		filePath := filepath.Join(config.Storage.TmpPath, path) // 构造文件完整路径
		if err := ensureFileExists(filePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		fileContent, err := os.ReadFile(filePath) // 读取文件内容
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.HTML(http.StatusOK, "index.html", gin.H{"title": path, "body": string(fileContent)}) // 使用 HTML 模板渲染并返回内容
	})

	r.POST("/:path", func(c *gin.Context) {
		body, err := c.GetRawData() // 获取请求体数据
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading request body"})
			return
		}
		path := c.Param("path")
		filePath := filepath.Join(config.Storage.TmpPath, path)
		if err := ensureFileExists(filePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if err := os.WriteFile(filePath, body, 0644); err != nil { // 将数据写入文件
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error writing to file"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "Success"}) // 返回成功状态
	})
}

// loadConfig 从指定路径加载配置文件并解析
func loadConfig(path string) (Config, error) {
	var config Config
	data, err := ioutil.ReadFile(path) // 读取文件内容
	if err != nil {
		return config, err
	}
	err = yaml.Unmarshal(data, &config) // 解析 YAML 数据到结构体
	return config, err
}

// generateRandomString 生成指定长度的随机字符串
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))] // 从字符集中随机选取字符
	}
	return string(b)
}

// ensureFileExists 确保指定的文件存在，如果不存在则创建
func ensureFileExists(filePath string) error {
	dir := filepath.Dir(filePath) // 获取文件所在的目录
	if err := os.MkdirAll(dir, 0755); err != nil { // 创建目录，如果不存在
		return err
	}
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		_, err = os.Create(filePath) // 创建文件，如果不存在
	}
	return err
}
