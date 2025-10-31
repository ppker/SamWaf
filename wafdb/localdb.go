package wafdb

import (
	"SamWaf/common/uuid"
	"SamWaf/common/zlog"
	"SamWaf/customtype"
	"SamWaf/global"
	"SamWaf/innerbean"
	"SamWaf/model"
	"SamWaf/model/baseorm"
	"SamWaf/utils"
	"context"
	"database/sql"
	"fmt"
	"gorm.io/gorm/logger"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	//"github.com/kangarooxin/gorm-webplugin-crypto"
	//"github.com/kangarooxin/gorm-webplugin-crypto/strategy"
	gowxsqlite3 "github.com/samwafgo/go-wxsqlite3"
	"github.com/samwafgo/sqlitedriver"
	"gorm.io/gorm"
)

func InitCoreDb(currentDir string) (bool, error) {
	if currentDir == "" {
		currentDir = utils.GetCurrentDir()
	}
	// 判断备份目录是否存在，不存在则创建
	if _, err := os.Stat(currentDir + "/data/"); os.IsNotExist(err) {
		if err := os.MkdirAll(currentDir+"/data/", os.ModePerm); err != nil {
			zlog.Error("创建data目录失败:", err)
			return false, err
		}
	}
	if global.GWAF_LOCAL_DB == nil {
		path := currentDir + "/data/local.db"
		// 检查数据库文件是否存在
		isNewDb := false
		if _, err := os.Stat(path); os.IsNotExist(err) {
			isNewDb = true
			zlog.Debug("本地主数据库文件不存在，将创建新数据库")
		}
		// 检查文件是否存在
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			// 文件存在的逻辑，使用工具函数进行备份
			backupDir := currentDir + "/data/backups"
			_, err := utils.BackupFile(path, backupDir, "local_backup", 10)
			if err != nil {
				zlog.Error("备份数据库文件失败:", err)
			}
		}

		key := url.QueryEscape(global.GWAF_PWD_COREDB)
		dns := fmt.Sprintf("%s?_db_key=%s", path, key)
		db, err := gorm.Open(sqlite.Open(dns), &gorm.Config{})
		if err != nil {
			panic("failed to connect database")
		}
		// 启用 WAL 模式
		_ = db.Exec("PRAGMA journal_mode=WAL;")

		// 创建自定义日志记录器
		gormLogger := NewGormZLogger()
		if global.GWAF_LOG_DEBUG_DB_ENABLE == true {
			gormLogger = gormLogger.LogMode(logger.Info).(*GormZLogger)
			// 启用调试模式
			db = db.Session(&gorm.Session{
				Logger: gormLogger,
			})
		}
		global.GWAF_LOCAL_DB = db
		s, err := db.DB()
		s.Ping()
		//db.Use(crypto.NewCryptoPlugin())
		// 注册默认的AES加解密策略
		//crypto.RegisterCryptoStrategy(strategy.NewAesCryptoStrategy("3Y)(27EtO^tK8Bj~"))
		// Migrate the schema
		db.AutoMigrate(&model.Hosts{})
		db.AutoMigrate(&model.Rules{})

		//隐私处理
		db.AutoMigrate(&model.LDPUrl{})

		//白名单处理
		db.AutoMigrate(&model.IPAllowList{})
		db.AutoMigrate(&model.URLAllowList{})

		//限制处理
		db.AutoMigrate(&model.IPBlockList{})
		db.AutoMigrate(&model.URLBlockList{})

		//抵抗CC
		db.AutoMigrate(&model.AntiCC{})

		//waf自身账号
		db.AutoMigrate(&model.TokenInfo{})
		db.AutoMigrate(&model.Account{})

		//系统参数
		db.AutoMigrate(&model.SystemConfig{})

		//延迟信息
		db.AutoMigrate(&model.DelayMsg{})

		//分库信息表
		db.AutoMigrate(&model.ShareDb{})

		//中心管控数据
		db.AutoMigrate(&model.Center{})

		//敏感词管理
		db.AutoMigrate(&model.Sensitive{})

		//负载均衡
		db.AutoMigrate(&model.LoadBalance{})

		//SSL证书
		db.AutoMigrate(&model.SslConfig{})

		//IPTag
		db.AutoMigrate(&model.IPTag{})

		//自动任务
		db.AutoMigrate(&model.BatchTask{})

		//SSL证书申请订单
		db.AutoMigrate(&model.SslOrder{})

		//SSL到期检测
		db.AutoMigrate(&model.SslExpire{})

		//HTTP AUTH
		db.AutoMigrate(&model.HttpAuthBase{})

		//任务
		db.AutoMigrate(&model.Task{})

		//自定义拦截界面
		db.AutoMigrate(&model.BlockingPage{})

		//OTP
		db.AutoMigrate(&model.Otp{})

		//密钥信息
		db.AutoMigrate(&model.PrivateInfo{})

		//密钥分组信息
		db.AutoMigrate(&model.PrivateGroup{})

		//缓存规则
		db.AutoMigrate(&model.CacheRule{})

		//隧道
		db.AutoMigrate(&model.Tunnel{})

		//CA服务器信息
		db.AutoMigrate(&model.CaServerInfo{})

		global.GWAF_LOCAL_DB.Callback().Query().Before("gorm:query").Register("tenant_plugin:before_query", before_query)
		global.GWAF_LOCAL_DB.Callback().Query().Before("gorm:update").Register("tenant_plugin:before_update", before_update)

		//重启需要删除无效规则
		db.Where("user_code = ? and rule_status = 999", global.GWAF_USER_CODE).Delete(model.Rules{})

		pathCoreSql(db)
		return isNewDb, nil
	} else {
		return false, nil
	}
}

func InitLogDb(currentDir string) (bool, error) {
	if currentDir == "" {
		currentDir = utils.GetCurrentDir()
	}
	if global.GWAF_LOCAL_LOG_DB == nil {
		path := currentDir + "/data/local_log.db"

		// 检查数据库文件是否存在
		isNewDb := false
		if _, err := os.Stat(path); os.IsNotExist(err) {
			isNewDb = true
			zlog.Debug("本地日志数据库文件不存在，将创建新数据库")
		}

		key := url.QueryEscape(global.GWAF_PWD_LOGDB)
		dns := fmt.Sprintf("%s?_db_key=%s", path, key)
		db, err := gorm.Open(sqlite.Open(dns), &gorm.Config{})
		if err != nil {
			panic("failed to connect database")
		}
		// 启用 WAL 模式
		_ = db.Exec("PRAGMA journal_mode=WAL;")
		// 创建自定义日志记录器
		gormLogger := NewGormZLogger()
		if global.GWAF_LOG_DEBUG_DB_ENABLE == true {
			gormLogger = gormLogger.LogMode(logger.Info).(*GormZLogger)
			// 启用调试模式
			db = db.Session(&gorm.Session{
				Logger: logger.Default.LogMode(logger.Info), // 设置为Info表示启用调试模式
			})
		}
		global.GWAF_LOCAL_LOG_DB = db
		//logDB.Use(crypto.NewCryptoPlugin())
		// 注册默认的AES加解密策略
		//crypto.RegisterCryptoStrategy(strategy.NewAesCryptoStrategy("3Y)(27EtO^tK8Bj~"))
		// Migrate the schema
		//统计处理
		db.AutoMigrate(&innerbean.WebLog{})
		db.AutoMigrate(&model.AccountLog{})
		db.AutoMigrate(&model.WafSysLog{})
		db.AutoMigrate(&model.OneKeyMod{})
		global.GWAF_LOCAL_LOG_DB.Callback().Query().Before("gorm:query").Register("tenant_plugin:before_query", before_query)
		global.GWAF_LOCAL_LOG_DB.Callback().Query().Before("gorm:update").Register("tenant_plugin:before_update", before_update)

		pathLogSql(db)
		var total int64 = 0
		global.GWAF_LOCAL_DB.Model(&model.ShareDb{}).Count(&total)
		if total == 0 {

			var logtotal int64 = 0
			global.GWAF_LOCAL_LOG_DB.Model(&innerbean.WebLog{}).Count(&logtotal)

			sharDbBean := model.ShareDb{
				BaseOrm: baseorm.BaseOrm{
					Id:          uuid.GenUUID(),
					USER_CODE:   global.GWAF_USER_CODE,
					Tenant_ID:   global.GWAF_TENANT_ID,
					CREATE_TIME: customtype.JsonTime(time.Now()),
					UPDATE_TIME: customtype.JsonTime(time.Now()),
				},
				DbLogicType: "log",
				StartTime:   customtype.JsonTime(time.Now()),
				EndTime:     customtype.JsonTime(time.Now()),
				FileName:    "local_log.db",
				Cnt:         logtotal,
			}
			global.GWAF_LOCAL_DB.Create(sharDbBean)
		}

		return isNewDb, nil
	} else {
		return false, nil
	}
}

// 手工切换日志数据源
func InitManaulLogDb(currentDir string, custFileName string) {
	if currentDir == "" {
		currentDir = utils.GetCurrentDir()
	}
	if global.GDATA_CURRENT_LOG_DB_MAP[custFileName] == nil {
		zlog.Debug("初始化自定义的库", custFileName)
		path := currentDir + "/data/" + custFileName
		key := url.QueryEscape(global.GWAF_PWD_LOGDB)
		dns := fmt.Sprintf("%s?_db_key=%s", path, key)
		db, err := gorm.Open(sqlite.Open(dns), &gorm.Config{})
		if err != nil {
			panic("failed to connect database")
		}
		// 启用 WAL 模式
		_ = db.Exec("PRAGMA journal_mode=WAL;")
		// 创建自定义日志记录器
		gormLogger := NewGormZLogger()
		if global.GWAF_LOG_DEBUG_DB_ENABLE == true {
			gormLogger = gormLogger.LogMode(logger.Info).(*GormZLogger)
			// 启用调试模式
			db = db.Session(&gorm.Session{
				Logger: logger.Default.LogMode(logger.Info), // 设置为Info表示启用调试模式
			})
		}
		global.GDATA_CURRENT_LOG_DB_MAP[custFileName] = db
		//logDB.Use(crypto.NewCryptoPlugin())
		// 注册默认的AES加解密策略
		//crypto.RegisterCryptoStrategy(strategy.NewAesCryptoStrategy("3Y)(27EtO^tK8Bj~"))
		// Migrate the schema
		//统计处理
		db.AutoMigrate(&innerbean.WebLog{})
		db.AutoMigrate(&model.AccountLog{})
		db.AutoMigrate(&model.WafSysLog{})
		db.AutoMigrate(&model.OneKeyMod{})

		global.GDATA_CURRENT_LOG_DB_MAP[custFileName].Callback().Query().Before("gorm:query").Register("tenant_plugin:before_query", before_query)
		global.GDATA_CURRENT_LOG_DB_MAP[custFileName].Callback().Query().Before("gorm:update").Register("tenant_plugin:before_update", before_update)

	} else {
		zlog.Debug("自定义的库已存在", custFileName)
	}
}

func InitStatsDb(currentDir string) (bool, error) {
	if currentDir == "" {
		currentDir = utils.GetCurrentDir()
	}
	if global.GWAF_LOCAL_STATS_DB == nil {
		path := currentDir + "/data/local_stats.db"
		// 检查数据库文件是否存在
		isNewDb := false
		if _, err := os.Stat(path); os.IsNotExist(err) {
			isNewDb = true
			zlog.Debug("本地统计数据库文件不存在，将创建新数据库")
		}
		key := url.QueryEscape(global.GWAF_PWD_STATDB)
		dns := fmt.Sprintf("%s?_db_key=%s", path, key)
		db, err := gorm.Open(sqlite.Open(dns), &gorm.Config{})
		if err != nil {
			panic("failed to connect database")
		}
		// 启用 WAL 模式
		_ = db.Exec("PRAGMA journal_mode=WAL;")
		// 创建自定义日志记录器
		gormLogger := NewGormZLogger()
		if global.GWAF_LOG_DEBUG_DB_ENABLE == true {
			gormLogger = gormLogger.LogMode(logger.Info).(*GormZLogger)
			// 启用调试模式
			db = db.Session(&gorm.Session{
				Logger: logger.Default.LogMode(logger.Info), // 设置为Info表示启用调试模式
			})
		}
		global.GWAF_LOCAL_STATS_DB = db
		//db.Use(crypto.NewCryptoPlugin())
		// 注册默认的AES加解密策略
		//crypto.RegisterCryptoStrategy(strategy.NewAesCryptoStrategy("3Y)(27EtO^tK8Bj~"))
		// Migrate the schema
		//统计处理
		db.AutoMigrate(&model.StatsTotal{})
		db.AutoMigrate(&model.StatsDay{})
		db.AutoMigrate(&model.StatsIPDay{})
		db.AutoMigrate(&model.StatsIPCityDay{})
		//IPTag
		db.AutoMigrate(&model.IPTag{})
		global.GWAF_LOCAL_STATS_DB.Callback().Query().Before("gorm:query").Register("tenant_plugin:before_query", before_query)
		global.GWAF_LOCAL_STATS_DB.Callback().Query().Before("gorm:update").Register("tenant_plugin:before_update", before_update)

		pathStatsSql(db)

		return isNewDb, nil
	} else {
		return false, nil
	}
}

func before_query(db *gorm.DB) {
	db.Where("tenant_id = ? and user_code=? ", global.GWAF_TENANT_ID, global.GWAF_USER_CODE)
}
func before_update(db *gorm.DB) {
}

// 在线备份
func BackupDatabase(db *gorm.DB, backupFile string) error {
	// 获取底层的 sql.DB 对象
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	// 获取源数据库的连接
	srcConn, err := sqlDB.Conn(context.Background())
	if err != nil {
		return err
	}
	defer srcConn.Close()

	// 获取底层的 SQLiteConn 对象
	var srcSQLiteConn *gowxsqlite3.SQLiteConn
	err = srcConn.Raw(func(driverConn interface{}) error {
		// 将 driverConn 转换为 *wxsqlite3.SQLiteConn
		sqliteConn, ok := driverConn.(*gowxsqlite3.SQLiteConn)
		if !ok {
			return fmt.Errorf("not a SQLite connection")
		}
		srcSQLiteConn = sqliteConn
		return nil
	})
	if err != nil {
		return err
	}

	// 打开目标数据库连接
	destConn, err := sql.Open("sqlite3", backupFile)
	if err != nil {
		return err
	}
	defer destConn.Close()

	// 获取目标数据库的连接
	destSqlConn, err := destConn.Conn(context.Background())
	if err != nil {
		return err
	}
	defer destSqlConn.Close()

	// 获取目标数据库的 SQLiteConn 对象
	var destSQLiteConn *gowxsqlite3.SQLiteConn
	err = destSqlConn.Raw(func(driverConn interface{}) error {
		// 将 driverConn 转换为 *wxsqlite3.SQLiteConn
		sqliteConn, ok := driverConn.(*gowxsqlite3.SQLiteConn)
		if !ok {
			return fmt.Errorf("not a SQLite connection")
		}
		destSQLiteConn = sqliteConn
		return nil
	})
	if err != nil {
		return err
	}

	// 执行备份
	backup, err := destSQLiteConn.Backup("main", srcSQLiteConn, "main")
	if err != nil {
		return err
	}
	defer backup.Finish()

	// 执行备份步骤 (-1 代表全部备份)
	for {
		b, stepErr := backup.Step(-1) // 备份指定多个页面 -1 是所有
		if b == false {
			zlog.Debug("backup fail", stepErr)
			if stepErr != nil {
				return stepErr
			}
		} else {
			break
		}

	}

	fmt.Println("Backup completed successfully")
	return nil
}

// cleanupOldBackups 清理旧的备份文件，只保留最新的n个
func cleanupOldBackups(backupDir string, keepCount int) {
	// 获取备份目录中的所有文件
	files, err := os.ReadDir(backupDir)
	if err != nil {
		zlog.Error("读取备份目录失败:", err)
		return
	}

	// 筛选出数据库备份文件
	var backupFiles []os.DirEntry
	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), "local_backup_") && filepath.Ext(file.Name()) == ".db" {
			backupFiles = append(backupFiles, file)
		}
	}

	// 如果备份文件数量不超过保留数量，则不需要删除
	if len(backupFiles) <= keepCount {
		return
	}

	// 按文件修改时间排序（从旧到新）
	sort.Slice(backupFiles, func(i, j int) bool {
		infoI, err := backupFiles[i].Info()
		if err != nil {
			return false
		}
		infoJ, err := backupFiles[j].Info()
		if err != nil {
			return false
		}
		return infoI.ModTime().Before(infoJ.ModTime())
	})

	// 删除多余的旧文件
	for i := 0; i < len(backupFiles)-keepCount; i++ {
		filePath := filepath.Join(backupDir, backupFiles[i].Name())
		err := os.Remove(filePath)
		if err != nil {
			zlog.Error("删除旧备份文件失败:", err, filePath)
		} else {
			zlog.Info("已删除旧备份文件:", filePath)
		}
	}
}
