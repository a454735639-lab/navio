package initialize

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sun-panel/biz"
	"sun-panel/global"
	"sun-panel/initialize/cUserToken"
	"sun-panel/initialize/config"
	"sun-panel/initialize/database"
	"sun-panel/initialize/lang"
	"sun-panel/initialize/other"
	"sun-panel/initialize/probe"
	"sun-panel/initialize/redis"
	"sun-panel/initialize/runlog"
	"sun-panel/initialize/systemSettingCache"
	"sun-panel/initialize/userToken"
	"sun-panel/lib/cmn"
	"sun-panel/lib/password"
	"sun-panel/models"
	"sun-panel/structs"
	"time"

	"log"

	"github.com/gin-gonic/gin"
)

type CMDParams struct {
	CfgPath  string // 配置文件目录
	CfgInit  bool   // 配置初始化
	PwdReset bool   // 重置密码
}

var DB_DRIVER = database.SQLITE
var cmdParams CMDParams

func getDefaultConfig() map[string]map[string]string {
	return map[string]map[string]string{
		"base": {
			"http_address":     "0.0.0.0",
			"http_port":        "3002",
			"https_port":       "3003",
			"source_path":      "./uploads",      // 存放文件的路径
			"source_temp_path": "./runtime/temp", // 存放文件的缓存路径
			"web_path":         "./web",
		},

		"probe": {
			"enabled":   "false",
			"port":      "8443",
			"cert_path": "",
			"key_path":  "",
		},

		// 此为隐藏配置
		"pro_auth": {
			"host": "https://auth.sun-panel.top",
		},
	}
}

func InitApp() error {
	Logo()
	gin.SetMode(global.RUNCODE) // GIN 运行模式

	// 日志
	if logger, err := runlog.InitRunlog(global.RUNCODE, "running.log"); err != nil {
		log.Panicln("Log initialization error", err)
		panic(err)
	} else {
		global.Logger = logger
	}

	cmdParams = CommandInit()

	// 如果为命令可能会终止运行
	CommandRun(cmdParams)

	// 配置初始化
	{
		if config, err := config.ConfigInit(cmdParams.CfgPath, false); err != nil {
			global.Logger.Errorln("Configuration initialization error", err)
			return err
		} else {
			config.Default = getDefaultConfig()
			global.Config = config
		}
	}

	// 多语言初始化
	lang.LangInit("zh-cn") // en-us

	DatabaseConnect()

	// Redis 连接
	{
		// 判断是否有使用redis的驱动，没有将不连接
		cacheDrive := global.Config.GetValueString("base", "cache_drive")
		queueDrive := global.Config.GetValueString("base", "queue_drive")
		if cacheDrive == "redis" || queueDrive == "redis" {
			redisConfig := structs.IniConfigRedis{}
			global.Config.GetSection("redis", &redisConfig)
			rdb, err := redis.InitRedis(redis.Options{
				Addr:     redisConfig.Address,
				Password: redisConfig.Password,
				DB:       redisConfig.Db,
			})

			if err != nil {
				log.Panicln("Redis initialization error", err)
				panic(err)
				// return err
			}
			global.RedisDb = rdb
		}
	}

	// 初始化用户token
	global.UserToken = userToken.InitUserToken()
	tokenExpires := global.Config.GetValueIntOrDefault("base", "token_expires", 168) // 默认7天
	global.CUserToken = cUserToken.InitCUserToken(time.Duration(tokenExpires) * time.Hour)

	// 其他的初始化
	global.VerifyCodeCachePool = other.InitVerifyCodeCachePool()
	global.SystemSetting = systemSettingCache.InItSystemSettingCache()
	global.SystemMonitor = global.NewCache[interface{}](5*time.Hour, -1, "systemMonitorCache")

	// 注册与统计程序
	biz.ProAuth = *biz.NewProAuth()
	biz.ProAuth.Start()

	// 可选的网络位置探针（TLS，默认关闭）
	probe.Start()
	return nil
}

func DatabaseConnect() {
	// 数据库连接 - 开始
	var dbClientInfo database.DbClient
	databaseDrive := global.Config.GetValueStringOrDefault("base", "database_drive")
	if databaseDrive == database.MYSQL {
		dbClientInfo = &database.MySQLConfig{
			Username:    global.Config.GetValueStringOrDefault("mysql", "username"),
			Password:    global.Config.GetValueStringOrDefault("mysql", "password"),
			Host:        global.Config.GetValueStringOrDefault("mysql", "host"),
			Port:        global.Config.GetValueStringOrDefault("mysql", "port"),
			Database:    global.Config.GetValueStringOrDefault("mysql", "db_name"),
			WaitTimeout: global.Config.GetValueInt("mysql", "wait_timeout"),
		}
	} else {
		dbClientInfo = &database.SQLiteConfig{
			Filename: global.Config.GetValueStringOrDefault("sqlite", "file_path"),
		}
	}

	if db, err := database.DbInit(dbClientInfo); err != nil {
		log.Panicln("Database initialization error:", err)
		panic(err)
	} else {
		global.Db = db
		models.Db = global.Db
	}

	if err := database.CreateDatabase(databaseDrive, global.Db); err != nil {
		log.Println("Database initialization error:", err)
		log.Panicln("Database initialization error", err)
	}

	database.NotFoundAndCreateUser(global.Db)

	// 修正历史版本写入数据库的旧默认站点标题（幂等，只改恰好等于旧默认值的记录）
	if err := database.MigrateLegacySiteTitle(global.Db); err != nil {
		// 迁移失败不应阻断启动，仅记录日志
		log.Println("Site title migration skipped:", err)
	}
}

// 命令行参数初始化
func CommandInit() CMDParams {
	flag.StringVar(&cmdParams.CfgPath, "config-path", "./conf", "Configuration file directory")
	flag.BoolVar(&cmdParams.CfgInit, "config-reset", false, "Create or reset the configuration file")
	flag.BoolVar(&cmdParams.PwdReset, "password-reset", false, "Reset the password of the first admin account")
	flag.Parse()
	return cmdParams
}

// 命令行运行
func CommandRun(arg CMDParams) {
	if cmdParams.CfgInit {
		// 生成配置文件
		fmt.Println("Generating configuration file")
		_, initError := config.ConfigInit(cmdParams.CfgPath, true)
		if initError != nil {
			fmt.Println("ERROR: config init:", initError.Error())
			os.Exit(1) // 务必退出
		}

		fmt.Println("The configuration file has been created ", filepath.Join(cmdParams.CfgPath, "conf.ini"), "Please modify according to your own needs")
		os.Exit(0) // 务必退出
	} else if cmdParams.PwdReset {
		// 重置密码

		// 配置初始化
		config, initError := config.ConfigInit(cmdParams.CfgPath, false)
		if initError != nil {
			fmt.Println("ERROR: config init:", initError.Error())
			os.Exit(1) // 务必退出
		}
		global.Config = config

		DatabaseConnect()
		userInfo := models.User{}
		if err := global.Db.Where("role=?", 1).Order("id").First(&userInfo).Error; err != nil {
			fmt.Println("ERROR", err.Error())
			os.Exit(1) // 务必退出
		}

		newPassword := "12345678"

		updateInfo := models.User{
			Password: func() string { h, _ := password.Hash(newPassword); return h }(),
			Token:    "",
		}
		// 重置第一个管理员的密码
		if err := global.Db.Select("Password", "Token").Where("id=?", userInfo.ID).Updates(&updateInfo).Error; err != nil {
			fmt.Println("ERROR", err.Error())
			os.Exit(1) // 务必退出
		}

		fmt.Println("The password has been successfully reset. Here is the account information")
		fmt.Println("Username ", userInfo.Username)
		fmt.Println("Password ", newPassword)
		os.Exit(0) // 务必退出
	} else {
		// 正常运行
		return
	}
	// 无匹配的命令
	os.Exit(0) // 务必退出
}

func Logo() {
	fmt.Println("    _   __      _          ")
	fmt.Println("   / | / /___ _(_)___  ____ ")
	fmt.Println("  /  |/ / __ `/ / __ \\/ __ \\")
	fmt.Println(" / /|  / /_/ / / /_/ / /_/ /")
	fmt.Println("/_/ |_/\\__,_/_/\\____/\\____/ ")
	fmt.Println("")

	versionInfo := cmn.GetSysVersionInfo()
	fmt.Println("Version:", versionInfo.Version)
	fmt.Println("Welcome to Navio.")
	fmt.Println("Project address:", "https://github.com/a454735639-lab/navio")

}
