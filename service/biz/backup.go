package biz

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"sun-panel/global"
	"sun-panel/initialize/database"
	"sun-panel/lib/cmn"
	"sun-panel/lib/file"
	"sun-panel/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// ---------- 备份配置（与镜像 v1.8.1 字段一致） ----------

type BackupConfigPath struct {
	Uploads     string `json:"uploads"`
	CustomStyle string `json:"customStyle"`
	Sqlite      string `json:"sqlite"`
}

type BackupConfig struct {
	Name                  string           `json:"name"`
	ConfigVersion         string           `json:"configVersion"`
	SunPanelVersion       string           `json:"sunPanelVersion"`
	SunPanelLowestVersion string           `json:"sunPanelLowestVersion"`
	Time                  string           `json:"time"`
	Paths                 BackupConfigPath `json:"paths"`
}

const (
	BackupConfigVersion          = "1.0"
	BackupDirName                = "runtime/backup" // 备份 zip 落盘目录（相对工作目录）
	BackupTempDir                = "runtime/temp/backup"
	BackupConfigFileName         = "backup.json"
	ErrBackupIncompatibleVersion = "backup version incompatible"
)

// ---------- 版本比较 ----------

// parseVersion 取 "1.8.1-beta25" / "20|1.8.1" 中的数字版本段 [major, minor, patch]
func parseVersion(v string) []int {
	v = strings.TrimSpace(v)
	if i := strings.Index(v, "|"); i >= 0 {
		v = v[i+1:]
	}
	// 截掉首段数字之后的非数字尾巴（如 -beta25）
	if i := strings.IndexAny(v, "-+"); i >= 0 {
		v = v[:i]
	}
	parts := strings.Split(v, ".")
	nums := make([]int, 3)
	for j := 0; j < 3 && j < len(parts); j++ {
		nums[j], _ = strconv.Atoi(strings.TrimSpace(parts[j]))
	}
	return nums
}

func versionLess(a, b []int) bool {
	for i := 0; i < 3; i++ {
		if a[i] != b[i] {
			return a[i] < b[i]
		}
	}
	return false
}

// VersionCompatibilityCheck 判断备份能否恢复：
//
//	备份版本 <= 当前版本：直接可恢复；
//	备份版本 >  当前版本：若 lowestVersion <= 当前版本，返回 compatible=true 且 newer=true（由调用方决定是否强制）；
//	否则不可恢复。
func VersionCompatibilityCheck(cfg BackupConfig) (compatible, newer bool, err error) {
	cur := parseVersion(cmn.GetSysVersionInfo().Version)
	backup := parseVersion(cfg.SunPanelVersion)
	lowest := parseVersion(cfg.SunPanelLowestVersion)

	switch {
	case !versionLess(cur, backup): // cur >= backup
		return true, false, nil
	case cfg.SunPanelLowestVersion == "" || !versionLess(cur, lowest): // cur >= lowest
		return true, true, nil
	default:
		return false, false, fmt.Errorf("%s: current %s < lowest %s",
			ErrBackupIncompatibleVersion, cmn.GetSysVersionInfo().Version, cfg.SunPanelLowestVersion)
	}
}

// ---------- 备份 ----------

func buildBackupConfig() BackupConfig {
	return BackupConfig{
		Name:                  "sun-panel",
		ConfigVersion:         BackupConfigVersion,
		SunPanelVersion:       cmn.GetSysVersionInfo().Version,
		SunPanelLowestVersion: cmn.GetSysVersionInfo().Version,
		Time:                  cmn.GetTime(),
		Paths: BackupConfigPath{
			Uploads:     "uploads",
			CustomStyle: "custom",
			Sqlite:      "sqlite/database.db",
		},
	}
}

func saveBackupConfig(dir string, cfg BackupConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, BackupConfigFileName), data, 0644)
}

func loadBackupConfig(dir string) (cfg BackupConfig, err error) {
	data, err := os.ReadFile(filepath.Join(dir, BackupConfigFileName))
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &cfg)
	return
}

// backupToTemp 把 uploads / custom / sqlite 汇总到临时目录
func backupToTemp(tmpDir string, cfg BackupConfig) error {
	sourcePath := global.Config.GetValueString("base", "source_path")
	customPath := orDefault(global.Config.GetValueString("base", "custom_style_path"), "./conf/custom")
	sqlitePath := orDefault(global.Config.GetValueString("sqlite", "file_path"), "./conf/database/database.db")

	_ = cfg
	if err := os.MkdirAll(tmpDir, os.ModePerm); err != nil {
		return err
	}
	// 目录可能尚未创建（从未上传过文件/未设置自定义样式），缺失时跳过
	if exists, _ := cmn.PathExists(sourcePath); exists {
		if err := file.CopyDir(sourcePath, filepath.Join(tmpDir, "uploads")); err != nil {
			return fmt.Errorf("copy uploads: %w", err)
		}
	}
	if exists, _ := cmn.PathExists(customPath); exists {
		if err := file.CopyDir(customPath, filepath.Join(tmpDir, "custom")); err != nil {
			return fmt.Errorf("copy custom: %w", err)
		}
	}
	if exists, _ := cmn.PathExists(sqlitePath); exists {
		if err := file.CopyFile(sqlitePath, filepath.Join(tmpDir, "sqlite", "database.db")); err != nil {
			return fmt.Errorf("copy sqlite: %w", err)
		}
	} else {
		return fmt.Errorf("sqlite db not found: %s", sqlitePath)
	}
	return nil
}

// backupPathToZip 把临时目录压成 zip（缺失的子目录自动跳过）
func backupPathToZip(tmpDir, zipPath string) error {
	z := file.ZipFile{}
	items := map[string]string{
		"sqlite":             filepath.Join(tmpDir, "sqlite"),
		BackupConfigFileName: filepath.Join(tmpDir, BackupConfigFileName),
	}
	// 全新实例可能从未上传过文件，uploads 目录不存在时跳过
	if exists, _ := cmn.PathExists(filepath.Join(tmpDir, "uploads")); exists {
		items["uploads"] = filepath.Join(tmpDir, "uploads")
	}
	if exists, _ := cmn.PathExists(filepath.Join(tmpDir, "custom")); exists {
		items["custom"] = filepath.Join(tmpDir, "custom")
	}
	return z.Create(items, zipPath)
}

// BackupToZip 入口：备份并打包，返回文件名（不含路径）
func BackupToZip() (fileName string, err error) {
	cfg := buildBackupConfig()
	ts := time.Now().Format("20060102150405")
	fileName = fmt.Sprintf("sun-panel-backup-%s.zip", ts)

	tmpDir := filepath.Join(BackupTempDir, ts)
	defer os.RemoveAll(tmpDir)

	if err = backupToTemp(tmpDir, cfg); err != nil {
		return
	}
	if err = saveBackupConfig(tmpDir, cfg); err != nil {
		return
	}
	zipPath := filepath.Join(BackupDirName, fileName)
	if err = os.MkdirAll(BackupDirName, os.ModePerm); err != nil {
		return
	}
	if err = backupPathToZip(tmpDir, zipPath); err != nil {
		return
	}
	return
}

// ---------- 恢复 ----------

// ExtractBackZip 解压上传的备份包
func ExtractBackZip(zipPath, destDir string) error {
	z := file.ZipFile{}
	return z.Extract(zipPath, destDir)
}

// ExtractBackupZipAndLoadConfig 解压并读取 backup.json
func ExtractBackupZipAndLoadConfig(zipPath string) (cfg BackupConfig, dir string, err error) {
	dir = filepath.Join(BackupTempDir, "restore-"+time.Now().Format("20060102150405"))
	if err = ExtractBackZip(zipPath, dir); err != nil {
		return
	}
	cfg, err = loadBackupConfig(dir)
	return
}

// deleteSqliteDatabaseImportantData 恢复场景保留：清空敏感表。
// 当前恢复实现采用整库文件替换，仅在 mysql 等无法整库替换的场景下由调用方使用。
func deleteSqliteDatabaseImportantData(db *gorm.DB) error {
	for _, table := range []string{"open_api_token", "user_config", "file"} {
		if err := db.Exec(fmt.Sprintf("DELETE FROM `%s`", table)).Error; err != nil {
			return err
		}
	}
	return nil
}

// reopenSqlite 关闭当前连接并用原配置重新打开 sqlite
func reopenSqlite(filePath string) error {
	if sqlDB, err := global.Db.DB(); err == nil {
		_ = sqlDB.Close()
	}
	db, err := gorm.Open(sqlite.Open(filePath), &gorm.Config{
		Logger:                                   database.GetLogger(),
		NamingStrategy:                           schema.NamingStrategy{SingularTable: true},
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return err
	}
	// 同时更新两处全局句柄，避免残留已关闭的旧连接
	global.Db = db
	models.Db = db
	return nil
}

// Recovery 恢复备份：sqlite 整库替换 + 还原 uploads / custom。
// force=false 且备份版本比当前新（但兼容）时返回 lowestSunPanelVersion 供前端确认。
func Recovery(zipPath string, force bool) (needConfirm bool, lowestSunPanelVersion string, err error) {
	cfg, dir, err := ExtractBackupZipAndLoadConfig(zipPath)
	if err != nil {
		return
	}
	defer os.RemoveAll(dir)

	compatible, newer, cerr := VersionCompatibilityCheck(cfg)
	if cerr != nil {
		return false, cfg.SunPanelLowestVersion, cerr
	}
	if !compatible {
		return false, cfg.SunPanelLowestVersion, fmt.Errorf("%s", ErrBackupIncompatibleVersion)
	}
	if newer && !force {
		return true, cfg.SunPanelLowestVersion, nil
	}

	// 校验 backup.json 内路径：必须先确保解析后仍位于解压目录内，才能参与复制。
	// 这层校验阻断 ../ 穿越与绝对路径，避免把宿主机任意文件复制进静态可访问目录。
	sqliteSrc, pathErr := safeBackupPath(dir, cfg.Paths.Sqlite)
	if pathErr != nil {
		return false, "", pathErr
	}
	uploadsSrc, pathErr := safeBackupPath(dir, cfg.Paths.Uploads)
	if pathErr != nil {
		return false, "", pathErr
	}
	customSrc, pathErr := safeBackupPath(dir, cfg.Paths.CustomStyle)
	if pathErr != nil {
		return false, "", pathErr
	}

	// 拒绝符号链接：即便路径字符串合规，若解压目录内是 symlink 仍会指向目录外
	for _, p := range []string{sqliteSrc, uploadsSrc, customSrc} {
		if info, lerr := os.Lstat(p); lerr == nil && info.Mode()&os.ModeSymlink != 0 {
			return false, "", fmt.Errorf("symbolic link not allowed in backup: %s", p)
		}
	}

	// 预检：所有待复制的源路径都必须真实存在且可读，避免在替换数据库之后
	// 才发现 uploads/custom 源缺失，留下「新库 + 旧文件」的半恢复状态。
	if fi, serr := os.Stat(sqliteSrc); serr != nil || fi.IsDir() {
		return false, "", fmt.Errorf("invalid sqlite backup: %s", sqliteSrc)
	}
	for _, p := range []string{uploadsSrc, customSrc} {
		if exists, _ := cmn.PathExists(p); exists {
			if _, serr := os.Stat(p); serr != nil {
				return false, "", fmt.Errorf("unreadable backup path %s: %v", p, serr)
			}
		}
	}

	// 事务性回滚：把即将被覆盖的 sqlite 先存一份副本，任意步骤失败时回滚。
	sqlitePath := orDefault(global.Config.GetValueString("sqlite", "file_path"), "./conf/database/database.db")
	isSqlite := strings.EqualFold(orDefault(global.Config.GetValueString("base", "database_drive"), "sqlite"), "sqlite")
	var sqliteBackup string
	if isSqlite {
		if exists, _ := cmn.PathExists(sqlitePath); exists {
			sqliteBackup = sqlitePath + ".recovery-bak"
			if err = file.CopyFile(sqlitePath, sqliteBackup); err != nil {
				return false, "", fmt.Errorf("backup current db before recovery: %w", err)
			}
			defer os.Remove(sqliteBackup)
		}
	}

	rollback := func() {
		if sqliteBackup != "" {
			_ = reopenSqlite(sqlitePath) // 关闭可能指向半状态的连接
			_ = file.CopyFile(sqliteBackup, sqlitePath)
			_ = reopenSqlite(sqlitePath)
		}
	}

	// 1) sqlite 整库替换
	if isSqlite {
		if err = reopenSqlite(sqlitePath); err != nil {
			return
		}
		if err = file.CopyFile(sqliteSrc, sqlitePath); err != nil {
			rollback()
			return
		}
		if err = reopenSqlite(sqlitePath); err != nil {
			rollback()
			return
		}
	}

	// 2) 还原 uploads（备份包里可能没有该目录）
	sourcePath := global.Config.GetValueString("base", "source_path")
	if exists, _ := cmn.PathExists(uploadsSrc); exists {
		if err = file.CopyDir(uploadsSrc, sourcePath); err != nil {
			rollback()
			return
		}
	}

	// 3) 还原 custom（同上，可能缺失）
	customPath := orDefault(global.Config.GetValueString("base", "custom_style_path"), "./conf/custom")
	if exists, _ := cmn.PathExists(customSrc); exists {
		if err = file.CopyDir(customSrc, customPath); err != nil {
			rollback()
			return
		}
	}
	return
}

// safeBackupPath 校验备份包内 backup.json 提供的路径，确保解析后仍位于解压目录 base 内。
// backup.json 完全来自用户上传的 zip，属于不可信输入；若不加校验，../../ 穿越可把
// 宿主机任意文件（conf.ini、probe/.secret 等）复制进 /uploads 静态目录造成泄露。
func safeBackupPath(base, rel string) (string, error) {
	if rel == "" {
		return "", fmt.Errorf("empty backup path")
	}
	if filepath.IsAbs(rel) {
		return "", fmt.Errorf("absolute path not allowed in backup: %s", rel)
	}
	// 统一分隔符后再清理，避免 Windows 下反斜杠被当作普通字符绕过判断
	normalized := strings.ReplaceAll(rel, "\\", "/")
	joined := filepath.Join(base, filepath.FromSlash(normalized))

	absBase, err := filepath.Abs(base)
	if err != nil {
		return "", err
	}
	absJoined, err := filepath.Abs(joined)
	if err != nil {
		return "", err
	}
	relPath, err := filepath.Rel(absBase, absJoined)
	if err != nil || relPath == ".." || strings.HasPrefix(relPath, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("path escapes backup directory: %s", rel)
	}
	return joined, nil
}

func orDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}
