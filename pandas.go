/*
	对 gota 简单扩展
*/

package pandas

import (
	"fmt"
	"gitee.com/jn-qq/go-tools/data"
	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
	"github.com/shakinm/xlsReader/xls"
	"github.com/xuri/excelize/v2"
	"path/filepath"
	"strings"
)

// DataFrame is https://github.com/go-gota/gota/dataframe simple extend
type DataFrame struct {
	dataframe.DataFrame
	filePath string
	sheet    []string
}

const (
	XLSX = "xlsx"
	XLS  = "xls"
)

// Read 读取excel文档，返回 DataFrame
//
//	path: 文档路径 sheet: 读取的表格，默认Sheet1
func Read(path string, sheet ...string) DataFrame {

	// 设置默认读取工作部名称
	if sheet == nil {
		sheet = append(sheet, "Sheet1")
	}

	// 构造
	df := DataFrame{filePath: path, sheet: sheet}

	// 读取数据
	switch strings.Split(filepath.Base(path), ".")[1] {

	case XLS:
		df.readFormXLS()

	case XLSX:
		df.readFromXLSX()

	default:
		df.Err = fmt.Errorf("不支持的文件类型")
	}

	return df
}

// ReadFromXLSX 从 XLSX 格式文档里读取数据
func (d *DataFrame) readFromXLSX() {
	// 读取文档
	f, err := excelize.OpenFile(d.filePath)
	if err != nil {
		d.Err = err
		return
	}
	defer func(f *excelize.File) {
		err := f.Close()
		if err != nil {
			return
		}
	}(f)

	var header []string
	var xlsxData [][]string

	// 遍历sheet
	for _, sn := range f.GetSheetList() {

		// 跳过
		if !data.Contains(d.sheet, sn) {
			continue
		}

		// 获取表头与行列
		var colNum []int
		var hd []string

		// 行迭代器
		rows, err := f.Rows(sn)
		if err != nil {
			d.Err = err
			return
		}
		for rows.Next() {
			row, err := rows.Columns()
			if err != nil {
				d.Err = err
				return
			}

			if data.IsEmpty(row) {
				continue
			}

			// 获取表头与行列
			if hd == nil {
				for i, r := range row {
					if len(colNum) == 0 {
						if r != "" {
							colNum = append(colNum, i)
						}
					} else if len(colNum) == 1 {
						if r == "" {
							colNum = append(colNum, i)
							break
						}
						if i == len(row)-1 {
							colNum = append(colNum, i+1)
						}
					}
				}

				hd = append(hd, row[colNum[0]:colNum[1]]...)
				if header == nil {
					header = append(header, hd...)
				} else {
					if !data.Equal(header, hd) {
						d.Err = fmt.Errorf("读取多工作簿发现表头不一致,请确认")
						return
					}
				}
				continue
			}

			// 获取数据体
			xlsxData = append(xlsxData, row[colNum[0]:colNum[1]])

		}
		_ = rows.Close()
	}

	d.DataFrame = dataframe.LoadRecords(append([][]string{header}, xlsxData...))
}

// WriteXLSX 将 DataFrame 写入 XLSX 文档中
func (d *DataFrame) WriteXLSX(path string) error {
	// 使用 NewFile 新建 Excel 工作薄，新创建的工作簿中会默认包含一个名为 Sheet1 的工作表
	f := excelize.NewFile()
	defer func(f *excelize.File) {
		_ = f.Close()
	}(f)

	// 整理数据
	for i, colVals := range d.MapCols() {
		if err := f.SetSheetCol("Sheet1", data.Int2AAA(i+1)+"1", &colVals); err != nil {
			return err
		}
	}

	// 根据指定路径保存文件
	if err := f.SaveAs(path); err != nil {
		return err
	}

	return nil

}

// 从 XLS 格式文档里读取数据
func (d *DataFrame) readFormXLS() {

	workbook, err := xls.OpenFile(d.filePath)
	if err != nil {
		d.Err = err
		return
	}

	var header []string
	var xlsxData [][]string

	for _, sn := range workbook.GetSheets() {
		// 跳过
		if !data.Contains(d.sheet, sn.GetName()) {
			continue
		}

		// 获取表头与行列
		var colNum []int
		var hd []string

		for _, rw := range sn.GetRows() {
			var row []string
			for _, cellData := range rw.GetCols() {
				row = append(row, cellData.GetString())
			}

			// 空行跳过
			if data.IsEmpty(row) {
				continue
			}

			// 获取表头与行列
			if hd == nil {
				for i, r := range row {
					if len(colNum) == 0 {
						if r != "" {
							colNum = append(colNum, i)
						}
					} else if len(colNum) == 1 {
						if r == "" {
							colNum = append(colNum, i)
							break
						}
						if i == len(row)-1 {
							colNum = append(colNum, i+1)
						}
					}
				}

				hd = append(hd, row[colNum[0]:colNum[1]]...)
				if header == nil {
					header = append(header, hd...)
				} else {
					if !data.Equal(header, hd) {
						d.Err = fmt.Errorf("读取多工作簿发现表头不一致,请确认")
						return
					}
				}
				continue
			}

			// 获取数据体
			xlsxData = append(xlsxData, row[colNum[0]:colNum[1]])

		}

	}

	d.DataFrame = dataframe.LoadRecords(append([][]string{header}, xlsxData...))
}

// MapCols 将 DataFrame 转化为 列切片
func (d *DataFrame) MapCols(cols ...string) [][]any {
	if cols == nil {
		cols = d.Names()
	}

	var pd [][]any
	for _, colName := range cols {

		var col []any
		// 写入表头
		col = append(col, colName)

		// 遍历列的值
		seriesCol := d.Col(colName)
		for ii := 0; ii < seriesCol.Len(); ii++ {
			col = append(col, seriesCol.Elem(ii).Val())
		}
		pd = append(pd, col)
	}

	return pd

}

func (d *DataFrame) FormatCols(f func(elem any) any, colName ...string) {

	// 获取指定列数据集
	ds := d.Select(colName)

	// 整理格式
	ds = ds.Capply(func(elems series.Series) series.Series {
		var newElems []any
		for i := 0; i < elems.Len(); i++ {
			// 处理数据
			elem := f(elems.Elem(i).Val())
			// 返回新数据针
			newElems = append(newElems, elem)
		}
		return series.New(newElems, elems.Type(), elems.Name)
	})

	// 更新数据集
	for _, name := range ds.Names() {
		d.DataFrame = d.Mutate(ds.Col(name))
	}
}
