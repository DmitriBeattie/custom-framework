package slice

import (
	"errors"
	"reflect"
)

func CreateMapFromSliceOfPtrToStruct(sliceOfPtrToStruct interface{}, keyFieldName string) (interface{}, error) {
	if reflect.Slice != reflect.TypeOf(sliceOfPtrToStruct).Kind() {
		return nil, errors.New("Param is not a slice")
	}

	val := reflect.ValueOf(sliceOfPtrToStruct)

	res := make(map[interface{}]interface{}, val.Len())

	for i := 0; i < val.Len(); i++ {
		if val.Index(i).Kind() != reflect.Ptr {
			return nil, errors.New("Slice contains type differs from ptr")
		}

		if val.Index(i).Elem().Kind() != reflect.Struct {
			return nil, errors.New("Ptr isn't points to struct")
		}

		keyFieldVal := val.Index(i).Elem().FieldByName(keyFieldName)

		if !keyFieldVal.IsValid() {
			return nil, errors.New("Not found specified field in struct")
		}

		res[keyFieldVal.Interface()] = val.Index(i).Interface()
	}

	return res, nil
}

func IndexSliceOfStruct(
	sliceOfStruct interface{},
	isSliceOfPtrToStruct bool,
	elemCreator func(val ...interface{}) (string, error),
	ignoreElemPattern string,
	fieldsName ...string,
) (map[string]interface{}, error) {

	if reflect.Slice != reflect.TypeOf(sliceOfStruct).Kind() {
		return nil, errors.New("Param is not a slice")
	}

	val := reflect.ValueOf(sliceOfStruct)

	res := make(map[string]interface{}, val.Len())

	for i := 0; i < val.Len(); i++ {
		var v reflect.Value

		switch isSliceOfPtrToStruct {
		case false:
			v = val.Index(i)
		case true:
			v = val.Index(i).Elem()
		}

		elem, err := createElementFromStruct(v, elemCreator, fieldsName...)
		if err != nil {
			return nil, err
		}

		if elem == ignoreElemPattern {
			continue
		}

		res[elem] = val.Index(i).Interface()
	}

	return res, nil
}

func createElementFromStruct(
	val reflect.Value,
	elemCreator func(val ...interface{}) (string, error),
	fieldsName ...string,
) (string, error) {
	if val.Kind() != reflect.Struct {
		return "", errors.New("Ptr isn't points to struct")
	}

	fieldsValues := make([]interface{}, 0, len(fieldsName))
	for j := range fieldsName {
		keyFieldVal := val.FieldByName(fieldsName[j])
		if !keyFieldVal.IsValid() {
			return "", errors.New("Not found specified field in struct")
		}

		fieldsValues = append(fieldsValues, keyFieldVal.Interface())
	}

	return elemCreator(fieldsValues...)
}

func ExtractIndexedSliceFromSliceOfStruct(
	sliceOfStruct interface{},
	isSliceOfPtrToStruct bool,
	elemCreator func(val ...interface{}) (string, error),
	ignoreElementPattern string,
	fieldsName ...string) ([]string, map[int]string, error) {

	if reflect.Slice != reflect.TypeOf(sliceOfStruct).Kind() {
		return nil, nil, errors.New("Param is not a slice")
	}

	val := reflect.ValueOf(sliceOfStruct)

	resSlice := make([]string, 0, val.Len())
	resMap := make(map[int]string, val.Len())
	avoidDuplicate := make(map[string]struct{}, val.Len())

	for i := 0; i < val.Len(); i++ {
		var v reflect.Value

		switch isSliceOfPtrToStruct {
		case false:
			v = val.Index(i)
		case true:
			v = val.Index(i).Elem()
		}

		elem, err := createElementFromStruct(v, elemCreator, fieldsName...)
		if err != nil {
			return nil, nil, err
		}

		if elem == ignoreElementPattern {
			continue
		}

		if _, ok := avoidDuplicate[elem]; !ok {
			resSlice = append(resSlice, elem)
			avoidDuplicate[elem] = struct{}{}
		}
		resMap[i] = elem
	}

	return resSlice, resMap, nil
}