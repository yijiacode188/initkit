package utils

import (
	"fmt"
	"reflect"
	"regexp"
)

func getColumnNameFromGormTag(tag string) string {
	re := regexp.MustCompile(`(?i)column:([a-zA-Z0-9_]+)`)
	matches := re.FindStringSubmatch(tag)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}
func StructColumnToMap(v interface{}, ignore []string) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// 使用 reflect 获取传入结构体的反射值和类型
	val := reflect.ValueOf(v)
	typ := reflect.TypeOf(v)

	// 确保传入的是一个结构体
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected a struct but got %v", val.Kind())
	}

	// 将 ignore 列表转换为 map，方便快速查找
	ignoreMap := make(map[string]struct{})
	for _, field := range ignore {
		ignoreMap[field] = struct{}{}
	}

	// 遍历结构体的每个字段
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)

		// 获取字段的 GORM 标签
		gormTag := field.Tag.Get("gorm")
		columnName := getColumnNameFromGormTag(gormTag)
		if columnName == "" {
			continue
		}
		// 如果字段在 ignore 列表中，则跳过
		if _, ok := ignoreMap[columnName]; ok {
			continue
		}

		// 将字段名和对应的值加入 map 中
		result[columnName] = val.Field(i).Interface()
	}

	return result, nil
}

func StructAssignByTag(src, dst interface{}, ignore []string, tag string) error {
	srcVal := reflect.ValueOf(src).Elem()
	dstVal := reflect.ValueOf(dst).Elem()
	dstType := dstVal.Type()

	// 创建一个映射，将源结构体的json标签映射到目标结构体的字段名
	tagToFieldName := make(map[string]string)
	for i := 0; i < dstType.NumField(); i++ {
		field := dstType.Field(i)
		jsonTag := field.Tag.Get(tag)
		if jsonTag != "" {
			fieldName := field.Name
			tagToFieldName[jsonTag] = fieldName
		}
	}
	for i := 0; i < srcVal.NumField(); i++ {
		srcField := srcVal.Field(i)
		srcFieldType := srcField.Type()
		jsonTag := srcVal.Type().Field(i).Tag.Get(tag)
		if SliceStringContains(ignore, jsonTag) {
			continue
		}
		fieldName, exists := tagToFieldName[jsonTag]

		if !exists {
			// 如果目标结构体没有对应的字段，则忽略并继续下一个字段
			continue
		}
		dstField := dstVal.FieldByName(fieldName)
		if !dstField.IsValid() {
			// 如果找不到对应的字段，则返回错误
			return fmt.Errorf("cannot find field %q in destination struct", fieldName)
		}
		dstFieldType := dstField.Type()

		// 确保类型兼容
		if !srcFieldType.AssignableTo(dstFieldType) && !dstFieldType.AssignableTo(srcFieldType) {
			return fmt.Errorf("incompatible types between source and destination fields: %s and %s", srcFieldType, dstFieldType)
		}

		// 如果类型不同但兼容（例如，将int赋值给interface{}），则需要先转换类型
		if srcFieldType != dstFieldType {
			srcField = srcField.Convert(dstFieldType)
		}

		// 设置目标结构体的字段值
		dstField.Set(srcField)
	}

	return nil
}
