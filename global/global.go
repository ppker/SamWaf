package global

import (
	"SamWaf/cache"
	"SamWaf/common/gwebsocket"
	"SamWaf/common/queue"
	"SamWaf/model"
	"SamWaf/model/spec"
	"SamWaf/wafnotify"
	"SamWaf/wafowasp"
	"SamWaf/wafsnowflake"
	"github.com/bytedance/godlp/dlpheader"
	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
	"github.com/oschwald/geoip2-golang"
	"gorm.io/gorm"
	"strconv"
	"time"
)

const (
	GWAF_NAME   = "SamWaf"
	Version_num = 1
)

var (

	/*本机信息**/
	GWAF_RUNTIME_IP          string = "127.0.0.1" //本机当前外网IP
	GWAF_RUNTIME_AREA        string = ""          //本机当前所在区域
	GWAF_RUNTIME_SERVER_TYPE bool   = false       //当前是是否以服务形式启动

	GWAF_RUNTIME_NEW_VERSION      string = ""      //最新版本号
	GWAF_RUNTIME_NEW_VERSION_DESC string = ""      //最新版本描述
	GWAF_RUNTIME_WIN7_VERSION     string = "false" //是否是win7内部版本
	GWAF_RUNTIME_QPS              uint64 = 0       //当前qps
	GWAF_RUNTIME_LOG_PROCESS      uint64 = 0       //log 处理速度

	GWAF_RUNTIME_DNS_SERVER  string = "119.29.29.29" //反向查询DNS的IP
	GWAF_RUNTIME_DNS_TIMEOUT int64  = 500            // DNS 查询超时时间 单位毫秒

	GWAF_RUNTIME_RECORD_LOG_TYPE string = "all" // 记录日志形式： 全部(all),非正常(abnormal)
	GWAF_RUNTIME_IS_UPDATETING   bool   = false //是否正在升级中

	GWAF_RUNTIME_CURRENT_EXEPATH                 string = "" //当前程序运行路径
	GWAF_RUNTIME_CURRENT_WEBPORT                 string = "" //当前程序所占用端口
	GWAF_RUNTIME_CURRENT_TUNNELPORT              string = "" //当前程序所占用隧道端口
	GWAF_RUNTIME_CURRENT_EXPORT_DB_LOG_FILE_PATH string = "" //生成的日志导出文件路径

	GWAF_RUNTIME_SSL_EXPIRE_CHECK bool = false //SSL过期检测是否正在运行
	GWAF_RUNTIME_SSL_SYNC_HOST    bool = false //主机同步信息到过期检测是否正在运行
	/**
	遥测数据
	*/
	GWAF_MEASURE_PROCESS_DEQUEENGINE cache.WafOnlyLockWriteData //遥测数据-队列处理时间

	GWAF_GLOBAL_HOST_NAME   string   = "全局网站:0"    //全局网站
	GWAF_GLOBAL_HOST_CODE   string   = "0"         //全局网站代码
	GWAF_LOCAL_DB           *gorm.DB               //通用本地数据库，尊重用户隐私
	GWAF_LOCAL_LOG_DB       *gorm.DB               //通用本地数据库存日志数据，尊重用户隐私
	GWAF_LOCAL_STATS_DB     *gorm.DB               //通用本地数据库存放统计数据，尊重用户隐私
	GWAF_REMOTE_DB          *gorm.DB               //仅当用户使用云数据库
	GWAF_LOCAL_SERVER_PORT  int      = 26666       // 本地local端口
	GWAF_LOCAL_INDEX_HTML   string                 //本地首页HTML信息
	GWAF_USER_CODE          string                 // 当前识别号
	GWAF_CUSTOM_SERVER_NAME string                 // 当前服务器自定义名称
	GWAF_TENANT_ID          string   = "SamWafCom" // 当前租户ID

	//管理端访问控制
	GWAF_IP_WHITELIST string = "0.0.0.0/0,::/0" //IP白名单 后台默认放行所有

	//zlog 日志相关信息
	GWAF_LOG_OUTPUT_FORMAT       string              = "console"  //zlog输出格式 控制台格式console,json格式
	GWAF_LOG_DEBUG_ENABLE        bool                = false      //是否开启debug日志，默认关闭
	GWAF_LOG_DEBUG_DB_ENABLE     bool                = false      //是否开启数据库日志，默认关闭
	GWAF_RELEASE                 string              = "false"    // 当前是否为发行版
	GWAF_RELEASE_VERSION_NAME    string              = "20241028" // 发行版的版本号名称
	GWAF_RELEASE_VERSION         string              = "v1.0.0"   // 发行版的版本号
	GWAF_LAST_UPDATE_TIME        time.Time                        // 上次时间
	GWAF_LAST_TIME_UNIX          int64               = 0          // 上次时间戳
	GWAF_NOTICE_ENABLE           bool                = false      // 是否开启通知
	GWAF_CAN_EXPORT_DOWNLOAD_LOG bool                = false      //是否可以导出下载日志
	GWAF_DLP                     dlpheader.EngineAPI              // 脱敏引擎
	GWAF_DLP_CONFIG              string                           // 脱敏引擎配置数据

	GWAF_OWASP *wafowasp.WafOWASP //owasp引擎
	/**链聚合**/
	GWAF_CHAN_HOST                                  = make(chan model.Hosts, 10)         //主机链
	GWAF_CHAN_ENGINE                                = make(chan int, 10)                 //引擎链
	GWAF_CHAN_MSG                                   = make(chan spec.ChanCommonHost, 10) //全局通讯包
	GWAF_CHAN_COMMON_MSG                            = make(chan spec.ChanCommon, 10)     //全局共用通讯包
	GWAF_CHAN_UPDATE                                = make(chan int, 10)                 //升级后处理链
	GWAF_CHAN_SENSITIVE                             = make(chan int, 10)                 //敏感词处理链
	GWAF_CHAN_SSL                                   = make(chan string, 10)              //证书处理链
	GWAF_CHAN_SSLOrder                              = make(chan spec.ChanSslOrder, 10)   //SSL证书申请
	GWAF_CHAN_SSL_EXPIRE_CHECK                      = make(chan int, 10)                 //SSL证书到期检测
	GWAF_CHAN_SYNC_HOST_TO_SSL_EXPIRE               = make(chan int, 10)                 //同步已存在主机到SSL证书检测任务里
	GWAF_CHAN_TASK                                  = make(chan string, 10)              //手工执行任务
	GWAF_CHAN_CLEAR_CC_WINDOWS                      = make(chan int, 10)                 //清除cc缓存信息
	GWAF_CHAN_CLEAR_CC_IP                           = make(chan string, 10)              //清除cc缓存信息IP
	GWAF_QUEUE_SHUTDOWN_SIGNAL        chan struct{} = make(chan struct{})                // 队列关闭信号
	GWAF_CHAN_CREATE_LOG_INDEX                      = make(chan string, 10)              // 创建日志索引

	GWAF_SHUTDOWN_SIGNAL bool = false // 系统关闭信号
	/*****CACHE相关*********/
	GCACHE_WAFCACHE      *cache.WafCache      //cache
	GCACHE_WECHAT_ACCESS string          = "" //微信访问密钥

	/*********HTTP相关**************/
	GWAF_HTTP_SENSITIVE_REPLACE_STRING = "**" //HTTP 敏感内容替换成

	/*********IP相关**************/
	GCACHE_IP_CBUFF            []byte         // IP相关缓存
	GCACHE_IP_V6_COUNTRY_CBUFF []byte         // IPv6国家相关缓存
	GCACHE_IPV4_SEARCHER       *xdb.Searcher  //IPV4得查询器
	GCACHE_IPV6_SEARCHER       *geoip2.Reader // IPV6得查询器

	GDATA_DELETE_INTERVAL int64 = 180 // 删除180天前的数据

	/****队列相关*****/
	GQEQUE_DB              *queue.Queue //正常DB队列
	GQEQUE_UPDATE_DB       *queue.Queue //正常DB 更新队列
	GQEQUE_LOG_DB          *queue.Queue //日志DB队列
	GQEQUE_STATS_DB        *queue.Queue //统计DB队列
	GQEQUE_STATS_UPDATE_DB *queue.Queue //统计更新DB队列
	GQEQUE_MESSAGE_DB      *queue.Queue //发送消息队列

	/*******通知相关*************/
	GNOTIFY_KAKFA_SERVICE           *wafnotify.WafNotifyService                                  //通知服务
	GNOTIFY_SEND_MAX_LIMIT_MINTUTES                             = time.Duration(5) * time.Minute // 规则相关信息最大发送抑止 默认5分钟

	/*******日志记录相关*************/
	GWEBLOG_VERSION = 20250112 //weblog 日志版本号
	/*********SSL相关*************/
	GSSL_HTTP_CHANGLE_PATH string = "/.well-known/acme-challenge/" // http01证书验证路径

	/******数据库处理参数*****/
	GDATA_BATCH_INSERT       int64               = 100         //最大批量插入数量
	GDATA_SHARE_DB_SIZE      int64               = 100 * 10000 //100w 进行分库 100*10000
	GDATA_SHARE_DB_FILE_SIZE int64               = 1024        //1024M 进行分库
	GDATA_CURRENT_CHANGE     bool                = false       //当前是否正在切换
	GDATA_CURRENT_LOG_DB_MAP map[string]*gorm.DB               //当前自定义的数据库连接 TODO 如果用户打开了多个 会不会影响内存
	/******WebSocket*********/
	GWebSocket *gwebsocket.WebSocketOnline

	//升级相关
	GUPDATE_VERSION_URL        string = "https://update.samwaf.com/"                                       // 官方下载
	GUPDATE_GITHUB_VERSION_URL string = "https://update.samwaf.com/samwaf_beta_update/latest_release.json" //从gitHub自动下载镜像
	//GUPDATE_GITHUB_VERSION_URL string = "http://10.0.2.2:8111/beta_update_linux/latest_release.json"
	// http://127.0.0.1:8111/beta_update/latest_release.json
	// http://10.0.2.2:8111/beta_update_linux/latest_release.json
	GWAF_SNOWFLAKE_GEN *wafsnowflake.Snowflake //雪花算法

	//任务开关信息
	GWAF_SWITCH_TASK_COUNTER bool

	//数据密码
	GWAF_PWD_COREDB = "3Y)(27EtO^tK8Bj~CORE" //加密
	GWAF_PWD_STATDB = "3Y)(27EtO^tK8Bj~STAT" //加密
	GWAF_PWD_LOGDB  = "3Y)(27EtO^tK8Bj~LOG"  //加密

	//默认创建的账户和密码
	GWAF_DEFAULT_ACCOUNT      string = "admin"         //默认创建的账户
	GWAF_DEFAULT_ACCOUNT_PWD  string = "admin868"      //默认创建的密码
	GWAF_DEFAULT_ACCOUNT_SALT string = "%@311*abDop#*" //盐值

	//通讯加密
	GWAF_COMMUNICATION_KEY = []byte("7E@u*has$d*@s5YX") //通讯加密密钥

	//资源下载
	GWAF_BOT_IP_URL_MAIN string = "https://raw.githubusercontent.com/samwafgo/SamWafBotIPDatabase/main/allowlist/index.json"

	/**
	中心管控部分
	*/
	GWAF_CENTER_ENABLE        string                 = "false"                    //中心管控激活状态
	GWAF_CENTER_URL           string                 = "http://127.0.0.1:26666"   //中心管控默认URL
	GWAF_REG_INFO             model.RegistrationInfo                              //当前注册信息
	GWAF_REG_VERSION                                 = "v1"                       //注册信息版本
	GWAF_REG_KEY                                     = []byte("5F!vion$k@a7QZ&)") //注册信息加密密钥
	GWAF_REG_PUBLIC_KEY       string                 = ""                         //注册信息加密公钥
	GWAF_REG_TMP_REG          []byte                                              //用户传来的信息
	GWAF_REG_FREE_COUNT       int64                  = 99999                      //免费版授权用户数
	GWAF_REG_CUR_CLIENT_COUNT int64                  = 3                          //当前客户端用户数
)

func GetCurrentVersionInt() int {
	version, _ := strconv.Atoi(GWAF_RELEASE_VERSION)
	return version
}
