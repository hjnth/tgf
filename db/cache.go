package db

import (
	"encoding/json"
	"github.com/thkhxm/tgf"
	"github.com/thkhxm/tgf/log"
	"github.com/thkhxm/tgf/util"
	"time"
)

//***************************************************
//@Link  https://github.com/thkhxm/tgf
//@Link  https://gitee.com/timgame/tgf
//@QQ群 7400585
//author tim.huang<thkhxm@gmail.com>
//@Description
//2023/2/24
//***************************************************

var cache iCacheService

var cacheModule = tgf.CacheModuleRedis

type iCacheService interface {
	Get(key string) (res string)
	Set(key string, val any, timeout time.Duration)
	GetMap(key string) map[string]string
	PutMap(key, filed, val string, timeout time.Duration)
	Del(key string)
	DelNow(key string)
	GetList(key string, start, end int64) (res []string, err error)
	SetList(key string, l []interface{}, timeout time.Duration)
	AddListItem(key string, val string)
}

type IAutoCacheService[Key cacheKey, Val any] interface {
	Get(key ...Key) (val Val, err error)
	Set(val Val, key ...Key) (success bool)
	Push(key ...Key)
	Remove(key ...Key) (success bool)
	Reset() IAutoCacheService[Key, Val]
}

// Get [Res any]
// @Description: 通过二级缓存获取数据
// @param key
// @return res
func Get[Res any](key string) (res Res, success bool) {
	val := cache.Get(key)
	if val != "" {
		res, _ = util.StrToAny[Res](val)
		success = true
	}
	return
}

func Set(key string, val any, timeout time.Duration) {
	switch val.(type) {
	case interface{}:
		data, _ := json.Marshal(val)
		cache.Set(key, data, timeout)
	default:
		cache.Set(key, val, timeout)
	}
}

func GetMap[Key cacheKey, Val any](key string) (res map[Key]Val, success bool) {
	data := cache.GetMap(key)
	if data != nil && len(data) > 0 {
		res = make(map[Key]Val, len(data))
		for k, v := range data {
			kk, _ := util.StrToAny[Key](k)
			vv, _ := util.StrToAny[Val](v)
			res[kk] = vv
		}
		success = true
	}
	return
}

func PutMap[Key cacheKey, Val any](key string, field Key, val Val, timeout time.Duration) {
	f, _ := util.AnyToStr(field)
	v, _ := util.AnyToStr(val)
	cache.PutMap(key, f, v, timeout)
}

func GetList[Res any](key string) []Res {
	if res, err := cache.GetList(key, 0, -1); err == nil {
		data := make([]Res, len(res))
		for i, r := range res {
			data[i], _ = util.StrToAny[Res](r)
		}
		return data
	}
	return nil
}

func GetListLimit[Res any](key string, start, end int64) []Res {
	if res, err := cache.GetList(key, start, end); err == nil {
		data := make([]Res, len(res))
		for i, r := range res {
			data[i], _ = util.StrToAny[Res](r)
		}
		return data
	}
	return nil
}

func AddListItem[Val any](key string, timeout time.Duration, val ...Val) (err error) {
	data := make([]interface{}, len(val))
	for i, v := range val {
		a, e := util.AnyToStr(v)
		if e != nil {
			err = e
			return
		}
		data[i] = a
	}
	cache.SetList(key, data, timeout)
	return
}

func Del(key string) {
	cache.Del(key)
}

func DelNow(key string) {
	cache.DelNow(key)
}

// AutoCacheBuilder [Key comparable,Val any]
// @Description: 自动化缓存Builder
type AutoCacheBuilder[Key cacheKey, Val any] struct {

	//数据是否在本地存储
	mem bool

	//

	//数据是否缓存
	cache bool
	//获取唯一key的拼接函数
	keyFun string

	//

	//数据是否持久化
	longevity         bool
	longevityInterval time.Duration
	//
	//是否自动清除过期数据
	autoClear        bool
	cacheTimeOut     time.Duration
	memTimeOutSecond int64
}

func (this *AutoCacheBuilder[Key, Val]) New() IAutoCacheService[Key, Val] {
	var ()
	manager := &autoCacheManager[Key, Val]{}
	manager.builder = this
	manager.InitStruct()
	return manager
}

func (this *AutoCacheBuilder[Key, Val]) WithAutoCache(cacheKey string, cacheTimeOut time.Duration) *AutoCacheBuilder[Key, Val] {
	var ()
	this.cache = true
	this.keyFun = cacheKey

	if cacheTimeOut > 0 {
		this.autoClear = true
		this.cacheTimeOut = cacheTimeOut
	}

	return this
}

func (this *AutoCacheBuilder[Key, Val]) WithMemCache(memTimeOutSecond uint32) *AutoCacheBuilder[Key, Val] {
	var ()
	this.mem = true
	if memTimeOutSecond>>31 == 1 {
		memTimeOutSecond = 0
	}
	if memTimeOutSecond != 0 {
		this.autoClear = true
	}
	this.memTimeOutSecond = int64(memTimeOutSecond)

	return this
}

func (this *AutoCacheBuilder[Key, Val]) WithLongevityCache(updateInterval time.Duration) *AutoCacheBuilder[Key, Val] {
	this.longevity = true
	if updateInterval < time.Second {
		log.WarnTag("orm", "updateInterval minimum is 1 second")
		updateInterval = time.Second
	}
	this.longevityInterval = updateInterval
	return this
}

// NewDefaultAutoCacheManager [Key comparable, Val any]
//
//	@Description: 创建一个默认的自动化数据管理，默认不包含持久化数据落地(mysql)，包含本地缓存，cache缓存(redis)
//	@param cacheKey cache缓存使用的组合key，例如user:1001 那么这里应该传入user即可，拼装方式为cacheKey:key
//	@return IAutoCacheService [Key comparable, Val any] 返回一个全新的自动化数据缓存管理对象
func NewDefaultAutoCacheManager[Key cacheKey, Val any](cacheKey string) IAutoCacheService[Key, Val] {
	builder := &AutoCacheBuilder[Key, Val]{}
	builder.keyFun = cacheKey
	builder.mem = true
	builder.autoClear = true
	builder.cache = true
	builder.cacheTimeOut = time.Hour * 24 * 3
	builder.memTimeOutSecond = 60 * 60 * 3
	builder.longevity = false
	return builder.New()
}

// NewLongevityAutoCacheManager [Key comparable, Val any]
//
//	@Description: 创建一个持久化的自动化数据管理，包含持久化数据落地(mysql)，包含本地缓存，cache缓存(redis)
//	@param cacheKey
//	@param tableName
//	@return IAutoCacheService [Key comparable, Val any]
func NewLongevityAutoCacheManager[Key cacheKey, Val IModel](cacheKey string) IAutoCacheService[Key, Val] {
	builder := &AutoCacheBuilder[Key, Val]{}
	builder.keyFun = cacheKey
	builder.mem = true
	builder.autoClear = true
	builder.cache = true
	builder.cacheTimeOut = time.Hour * 24 * 3
	builder.memTimeOutSecond = 60 * 60 * 3
	builder.longevity = true
	return builder.New()
}

// NewAutoCacheManager [Key comparable, Val any]
// @Description: 创建一个持久化的自动化数据管理，包含本地缓存，不包含持久化数据落地(mysql)，cache缓存(redis)
func NewAutoCacheManager[Key cacheKey, Val any]() IAutoCacheService[Key, Val] {
	builder := &AutoCacheBuilder[Key, Val]{}
	builder.keyFun = ""
	builder.mem = true
	builder.cache = false
	builder.longevity = false
	return builder.New()
}

func NewAutoCacheBuilder[Key cacheKey, Val any]() *AutoCacheBuilder[Key, Val] {
	builder := &AutoCacheBuilder[Key, Val]{}
	builder.mem = true
	builder.memTimeOutSecond = 60 * 60 * 3
	return builder
}

func WithCacheModule(module tgf.CacheModule) {
	cacheModule = module
}

func run() {
	switch cacheModule {
	case tgf.CacheModuleRedis:
		cache = newRedisService()
	case tgf.CacheModuleClose:
		return
	}
	//初始化mysql
	initMySql()
}