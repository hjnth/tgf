package util

import (
	"fmt"
	"github.com/xuri/excelize/v2"
	"os"
	"path/filepath"
	"text/template"
	"time"
)

//***************************************************
//@Link  https://github.com/thkhxm/tgf
//@Link  https://gitee.com/timgame/tgf
//@QQ群 7400585
//author tim.huang<thkhxm@gmail.com>
//@Description
//2023/4/10
//***************************************************

var (
	excelToJsonPath = ""
	excelToGoPath   = ""
	excelPath       = ""
	goPackage       = "conf"
	fileExt         = ".xlsx"
)

// ExcelExport
// @Description: Excel导出json文件
func ExcelExport() {
	fmt.Println("---------------start export------------------")
	fmt.Println("")
	fmt.Println("")
	files := GetFileList(excelPath, fileExt)
	structs := make([]*configStruct, 0)
	for _, file := range files {
		structs = append(structs, parseFile(file)...)
	}
	//
	if excelToGoPath != "" {
		toGolang(structs)
	}
	fmt.Println("")
	fmt.Println("")
	fmt.Println("---------------end export-------------------")

}

// SetExcelToJsonPath
// @Description: 设置Excel导出Json地址
// @param path
func SetExcelToJsonPath(path string) {
	excelToJsonPath, _ = filepath.Abs(path)
	fmt.Println("set excel to json path", excelToJsonPath)
}

// SetExcelToGoPath
// @Description: 设置Excel导出Go地址
// @param path
func SetExcelToGoPath(path string) {
	excelToGoPath, _ = filepath.Abs(path)
	fmt.Println("set excel to go path", excelToGoPath)
}

// SetExcelPath
// @Description: 设置Excel文件所在路径
func SetExcelPath(path string) {
	excelPath, _ = filepath.Abs(path)
	fmt.Println("set excel file path", excelPath)
}

// to golang
func toGolang(metalist []*configStruct) {
	tpl := fmt.Sprintf(`
//Auto generated by tgf util
//created at %v

package %v
		{{range .}}
type {{.StructName}}Conf struct {
		{{range .Fields}}
		//{{.Des}}
		{{.Key}}	{{.Typ}}
        {{end}}
}
{{end}}`, time.Now().String(), goPackage)
	t := template.New("ConfigStruct")
	tp, _ := t.Parse(tpl)
	file, err := os.Create(excelToGoPath + string(filepath.Separator) + "conf_struct.go")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	tp.Execute(file, metalist)
}

//

type configStruct struct {
	StructName string
	Fields     []*meta
	Version    string
}

type meta struct {
	Key string
	Idx int
	Typ string
	Des string
}

type rowdata []interface{}

func parseFile(file string) []*configStruct {
	fmt.Println("excel file [", file, "]")
	xlsx, err := excelize.OpenFile(file)
	if err != nil {
		panic(err.Error())
	}
	sheets := xlsx.GetSheetList()

	rs := make([]*configStruct, len(sheets))

	for i, s := range sheets {
		rows, err := xlsx.GetRows(s)
		if err != nil {
			return nil
		}
		if len(rows) < 5 {
			return nil
		}

		colNum := len(rows[1])
		metaList := make([]*meta, 0, colNum)
		dataList := make([]rowdata, 0, len(rows)-4)
		version := ""
		for line, row := range rows {
			switch line {
			case 0: // sheet 名
				version = row[0]
			case 1: // col name
				for idx, colname := range row {
					metaList = append(metaList, &meta{Key: colname, Idx: idx})
				}
			case 2: // data type
				for idx, typ := range row {
					metaList[idx].Typ = typ
				}
			case 3: // desc
				for idx, des := range row {
					metaList[idx].Des = des
				}
			default: //>= 4 row data
				data := make(rowdata, colNum)
				for k := 0; k < colNum; k++ {
					if k < len(row) {
						data[k] = row[k]
					}
				}
				dataList = append(dataList, data)
			}
		}
		jsonFile := fmt.Sprintf("%s.json", s)
		if excelToJsonPath != "" {
			err = output(jsonFile, toJson(dataList, metaList))
		}
		if err != nil {
			fmt.Println(err)
		}
		result := &configStruct{}
		result.Fields = metaList
		result.StructName = s
		result.Version = version
		rs[i] = result
		fmt.Println("excel export : json file", jsonFile, "golang struct :", s+"Conf", "[", version, "]")
	}
	return rs
}

const (
	fileType_string     = "string"
	fileType_time       = "time"
	fileType_arrayInt32 = "[]int32"
)

func toJson(datarows []rowdata, metalist []*meta) string {
	ret := "["
	for _, row := range datarows {
		ret += "\n\t{"
		for idx, meta := range metalist {
			ret += fmt.Sprintf("\n\t\t\"%s\":", meta.Key)
			switch meta.Typ {
			case fileType_string:
				if row[idx] == nil {
					ret += "\"\""
				} else {
					ret += fmt.Sprintf("\"%s\"", row[idx])
				}
			case fileType_time:
			case fileType_arrayInt32:
				if row[idx] == nil || row[idx] == "" {
					ret += "[]"
				} else {
					ret += fmt.Sprintf("%s", row[idx])
				}
			default:
				if row[idx] == nil || row[idx] == "" {
					ret += "0"
				} else {
					ret += fmt.Sprintf("%s", row[idx])
				}
			}
			ret += ","
		}
		ret = ret[:len(ret)-1]
		ret += "\n\t},"
	}
	ret = ret[:len(ret)-1]
	ret += "\n]"
	return ret
}

func output(filename string, str string) error {

	f, err := os.OpenFile(excelToJsonPath+string(filepath.Separator)+filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(str)
	if err != nil {
		return err
	}

	return nil
}

type TemplateKeyValueData struct {
	FieldName interface{}
	Values    interface{}
}

func JsonToKeyValueGoFile(packageName, fileName, outPath, fieldType string, data []TemplateKeyValueData) {

	tpl := fmt.Sprintf(`
//Auto generated by tgf util
//created at %v

package %v
const(
	{{range .}}
	{{.FieldName}} = "{{.Values}}"
    {{end}}
)
`, time.Now().String(), packageName)
	if "string" != fieldType {
		tpl = fmt.Sprintf(`
//Auto generated by tgf util
//created at %v

package %v
const(
	{{range .}}
	{{.FieldName}} %v = {{.Values}}
    {{end}}
)
`, time.Now().String(), packageName, fieldType)
	}
	t := template.New("KeyValueStruct")
	tp, _ := t.Parse(tpl)
	s, _ := filepath.Abs(outPath)
	file, err := os.Create(s + string(filepath.Separator) + fileName + ".go")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	tp.Execute(file, data)
}
