package database

import (
	"encoding/json"
	"log"
	"os"
	"path"
	"sun-panel/lib/cmn"
	"sun-panel/lib/password"
	"sun-panel/models"
	"time"

	"gorm.io/driver/mysql"
	_ "gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	_ "gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

const (
	MYSQL  = "mysql"
	SQLITE = "sqlite"
)

type DbClient interface {
	Connect() (db *gorm.DB, err error)
}

type MySQLConfig struct {
	Username    string
	Password    string
	Host        string
	Port        string
	Database    string
	WaitTimeout int
}

type SQLiteConfig struct {
	Filename string
}

func DbInit(dbClient DbClient) (db *gorm.DB, dbErr error) {
	db, dbErr = dbClient.Connect()
	if dbErr != nil {
		return
	}
	return
}

// Connect mysql连接
func (d *MySQLConfig) Connect() (db *gorm.DB, err error) {
	dsn := d.Username + ":" + d.Password + "@tcp(" + d.Host + ":" + d.Port + ")/" + d.Database + "?charset=utf8mb4&parseTime=True&loc=Local"
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: GetLogger(),
		NamingStrategy: schema.NamingStrategy{
			// TablePrefix:   "blog_",
			SingularTable: true,
		},
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	// 连接失败时 gorm.Open 可能返回 nil db 或无效的底层连接池，
	// 必须先检查错误再取底层 sql.DB，否则 nil 解引用 panic 掩盖真正的配置错误。
	if err != nil {
		return nil, err
	}
	sqlDb, dbErr := db.DB()
	if dbErr != nil {
		return nil, dbErr
	}
	sqlDb.SetMaxIdleConns(10)  // SetMaxIdleConns 设置空闲连接池中连接的最大数量
	sqlDb.SetMaxOpenConns(100) // SetMaxOpenConns 设置打开数据库连接的最大数量。
	wait_timeout := d.WaitTimeout
	sqlDb.SetConnMaxLifetime(time.Duration(wait_timeout * int(time.Second))) // SetConnMaxLifetime 设置了连接可复用的最大时间。
	return
}

// Connect sqllite3连接
func (d *SQLiteConfig) Connect() (db *gorm.DB, err error) {
	filePath := d.Filename
	exists := false
	if exists, err = cmn.PathExists(path.Dir(filePath)); err != nil {
		return
	} else {

		// 创建文件夹
		if !exists {
			if err = os.MkdirAll(path.Dir(filePath), 0700); err != nil {
				return
			}
		}

		db, err = gorm.Open(sqlite.Open(filePath), &gorm.Config{
			Logger: GetLogger(),
			NamingStrategy: schema.NamingStrategy{
				SingularTable: true,
			},
			DisableForeignKeyConstraintWhenMigrating: true,
		})
	}

	return
}

// 日志
func GetLogger() logger.Interface {
	return logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer（日志输出的目标，前缀和日志包含的内容——译者注）
		logger.Config{
			SlowThreshold:             time.Second, // 慢 SQL 阈值
			LogLevel:                  logger.Warn, // 日志级别
			IgnoreRecordNotFoundError: true,        // 忽略ErrRecordNotFound（记录未找到）错误
			Colorful:                  true,        // 彩色打印
		},
	)

}

// 创建数据库
func CreateDatabase(driver string, db *gorm.DB) error {

	// mysql特殊处理
	if driver == MYSQL {
		db = db.Set("gorm:table_options", "ENGINE=InnoDB")
	}

	// 创建数据表
	err := db.AutoMigrate(
		&models.User{},
		&models.SystemSetting{},
		&models.ItemIcon{},
		&models.UserConfig{},
		&models.File{},
		&models.ItemIconGroup{},
		&models.ModuleConfig{},
		&models.OpenApiToken{},
	)

	return err
}

// 初始化一个用户,一个用户都没有的时候创建一个
func NotFoundAndCreateUser(db *gorm.DB) error {
	fUser := models.User{}
	if err := db.First(&fUser).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return err
		}
		username := "admin@navio.local"
		fUser.Mail = username
		fUser.Username = username
		fUser.Name = username
		fUser.Status = 1
		fUser.Role = 1
		fUser.Password, _ = password.Hash("12345678")

		if errCreate := db.Create(&fUser).Error; errCreate != nil {
			return errCreate
		}
	}

	return nil
}

// legacySiteTitle 是上游历史版本写入数据库的默认站点标题。
const legacySiteTitle = "Sun-Panel"

// currentSiteTitle 是当前分支的默认站点标题。
const currentSiteTitle = "Navio"

// MigrateLegacySiteTitle 把数据库里仍为上游默认值的站点标题纠正为当前品牌。
// 只处理「恰好等于旧默认值」的情况，用户自行修改过的标题保持原样。
// 该操作幂等，可重复执行。
func MigrateLegacySiteTitle(db *gorm.DB) error {
	if db == nil {
		return nil
	}

	var row models.SystemSetting
	err := db.Model(&models.SystemSetting{}).
		Where("config_name = ?", "global_setting_site").
		First(&row).Error
	if err == gorm.ErrRecordNotFound {
		// 从未保存过站点设置，无需迁移（读取时会回退到代码默认值）
		return nil
	}
	if err != nil {
		return err
	}

	var site map[string]interface{}
	if err := json.Unmarshal([]byte(row.ConfigValue), &site); err != nil {
		// 配置不是合法 JSON 时不强行改写，避免破坏用户数据
		return nil
	}

	title, _ := site["title"].(string)
	if title != legacySiteTitle {
		return nil
	}

	site["title"] = currentSiteTitle
	updated, err := json.Marshal(site)
	if err != nil {
		return err
	}

	return db.Model(&models.SystemSetting{}).
		Where("id = ?", row.ID).
		Update("config_value", string(updated)).Error
}
