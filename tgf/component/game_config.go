package component

import (
	"github.com/cornelk/hashmap"
	"github.com/thkhxm/tgf/db"
	"github.com/thkhxm/tgf/log"
	"github.com/thkhxm/tgf/util"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
)

//***************************************************
//@Link  https://github.com/thkhxm/tgf
//@Link  https://gitee.com/timgame/tgf
//@QQ群 7400585
//author tim.huang<thkhxm@gmail.com>
//@Description
//2023/4/10
//***************************************************

var confPath = "./conf/json"

var contextDataManager db.IAutoCacheService[string, []byte]
var cacheDataManager db.IAutoCacheService[string, *hashmap.Map[string, interface{}]]

var newLock = &sync.Mutex{}

func GetGameConf[Val any](id string) (res Val) {
	t := util.ReflectType[Val]()
	key := t.Name()
	data, _ := cacheDataManager.Get(key)

	tmp, _ := data.Get(id)
	return tmp.(Val)
}

// GetAllGameConf [Val any]
// @Description: 不建议使用这个函数除非特殊需求，建议使用RangeGameConf
func GetAllGameConf[Val any]() (res []Val) {
	t := util.ReflectType[Val]()
	key := t.Name()
	data, _ := cacheDataManager.Get(key)
	res = make([]Val, 0, data.Len())
	data.Range(func(s string, i interface{}) bool {
		res = append(res, i)
		return true
	})
	return res
}

// RangeGameConf [Val any]
// @Description:
// @param f
func RangeGameConf[Val any](f func(s string, i Val) bool) {
	t := util.ReflectType[Val]()
	key := t.Name()
	data, _ := cacheDataManager.Get(key)
	ff := func(a string, b interface{}) bool {
		return f(a, b)
	}
	data.Range(ff)
}

func getCacheGameConfData(key string) {
	//data := cacheDataManager.Get(key)
	//if data == nil {
	//
	//}
}

// LoadGameConf [Val any]
//
//	 泛型传入自动生成的配置即可
//		@Description: 预加载
func LoadGameConf[Val any]() {
	t := util.ReflectType[Val]()
	key := t.Name()
	context, _ := contextDataManager.Get(key)
	data, _ := util.StrToAny[[]Val](util.ConvertStringByByteSlice(context))
	cc := hashmap.New[string, interface{}]()
	for _, d := range data {
		rd := reflect.ValueOf(d).Elem()
		id := rd.FieldByName("Id")
		cc.Set(id.Interface().(string), d)
	}
	cacheDataManager.Set(cc, key)
	log.DebugTag("GameConf", "load game conf , name=%v", t.Name())
}

func WithConfPath(path string) {
	confPath, _ = filepath.Abs(path)
	log.InfoTag("GameConf", "set game json file path=%v", confPath)
}

func InitGameConfToMem() {
	builder := db.NewAutoCacheBuilder[string, []byte]()
	builder.WithMemCache(0)
	contextDataManager = builder.New()
	//
	cacheBuilder := db.NewAutoCacheBuilder[string, *hashmap.Map[string, interface{}]]()
	cacheBuilder.WithMemCache(0)
	cacheDataManager = cacheBuilder.New()
	//
	files := util.GetFileList(confPath, ".json")
	for _, filePath := range files {
		file, err := os.Open(filePath)
		if err != nil {
			log.WarnTag("GameConf", "game json file [%v] open error %v", filePath, err)
			continue
		}
		context, err := io.ReadAll(file)
		if err != nil {
			log.WarnTag("GameConf", "game json file [%v] read error %v", filePath, err)
			continue
		}
		_, fileName := filepath.Split(filePath)
		contextDataManager.Set(context, strings.Split(fileName, `.`)[0]+"Conf")
		file.Close()
	}
}
