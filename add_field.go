package gotool

import "reflect"

type IAddField interface {
	IsNeedAddStructField(field reflect.Type) bool
	AddStructFields(reflect.Type) []reflect.StructField
	GetStructFieldValue(val reflect.Value) ([]reflect.Value, error)

	GetMapKeyValueToAdd(val reflect.Value) (keys, values []reflect.Value, err error)
}

type DefaultAddFiledAPI struct {
}

func (d DefaultAddFiledAPI) IsNeedAddStructField(field reflect.Type) bool {
	return false
}

func (d DefaultAddFiledAPI) AddStructFields(r reflect.Type) []reflect.StructField {
	return nil
}

func (d DefaultAddFiledAPI) GetStructFieldValue(val reflect.Value) ([]reflect.Value, error) {
	return nil, nil
}

func (d DefaultAddFiledAPI) GetMapKeyValueToAdd(val reflect.Value) (keys, values []reflect.Value, err error) {
	return nil, nil, nil
}

func AddField(data any, api IAddField) (any, error) {
	val := reflect.ValueOf(data)
	ff := newFillField(api)
	newVal, err := ff.fill(val)
	if err != nil {
		return nil, err
	}
	return newVal.Interface(), nil
}

type fillField struct {
	api IAddField
}

func newFillField(api IAddField) *fillField {
	return &fillField{api: api}
}

func (f *fillField) fill(val reflect.Value) (reflect.Value, error) {
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() == reflect.Interface {
		val = val.Elem()
	}
	switch val.Kind() {
	case reflect.Slice:
		newZeroVal := f.addFieldZeroValue(val.Type())
		newSlice := reflect.MakeSlice(reflect.SliceOf(newZeroVal.Elem()), val.Len(), val.Len())
		for i := 0; i < val.Len(); i++ {
			newVal, err := f.fill(val.Index(i))
			if err != nil {
				return reflect.Value{}, err
			}
			newSlice.Index(i).Set(newVal)
		}
		return newSlice, nil
	case reflect.Map:
		if val.IsNil() || val.Len() == 0 {
			return val, nil
		}
		return f.addFieldInMap(val)

	case reflect.Struct:
		sf, sv, err := f.getStructFieldAndVal(val)
		if err != nil {
			return reflect.Value{}, err
		}
		newElem := reflect.New(reflect.StructOf(sf)).Elem()
		for i := 0; i < newElem.NumField(); i++ {
			if newElem.Field(i).CanSet() {
				newElem.Field(i).Set(sv[i])
			}
		}
		return newElem, nil
	}
	return val, nil
}

func (f *fillField) addFieldZeroValue(val reflect.Type) reflect.Type {
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	switch val.Kind() {
	case reflect.Slice:
		return reflect.SliceOf(f.addFieldZeroValue(val.Elem()))
	case reflect.Struct:
		sf := f.getStructFieldType(val)
		return reflect.StructOf(sf)
	}
	return val
}

func (f *fillField) getStructFieldType(typ reflect.Type) []reflect.StructField {
	var sf []reflect.StructField

	for i := 0; i < typ.NumField(); i++ {
		v := typ.Field(i).Type
		switch v.Kind() {
		case reflect.Struct, reflect.Slice:
			newTyp := f.addFieldZeroValue(typ.Field(i).Type)
			sf = append(sf, reflect.StructField{
				Name:      typ.Field(i).Name,
				Type:      newTyp,
				Tag:       typ.Field(i).Tag,
				Anonymous: typ.Field(i).Anonymous,
			})
		default:
			sf = append(sf, typ.Field(i))
		}
	}

	if f.api.IsNeedAddStructField(typ) {
		sf = append(sf, f.api.AddStructFields(typ)...)
	}
	return sf
}

func (f *fillField) getStructFieldAndVal(val reflect.Value) ([]reflect.StructField, []reflect.Value, error) {
	var sf []reflect.StructField
	var sv []reflect.Value
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		item := val.Field(i)
		if item.Kind() == reflect.Interface {
			item = item.Elem()
		}
		switch item.Kind() {
		case reflect.Struct, reflect.Slice:
			newVal, err := f.fill(item)
			if err != nil {
				return nil, nil, err
			}

			sf = append(sf, reflect.StructField{
				Name:      typ.Field(i).Name,
				Type:      newVal.Type(),
				Tag:       typ.Field(i).Tag,
				Anonymous: typ.Field(i).Anonymous,
			})
			sv = append(sv, newVal)

		default:
			sf = append(sf, typ.Field(i))
			sv = append(sv, item)
		}
	}

	if f.api.IsNeedAddStructField(val.Type()) {
		sf = append(sf, f.api.AddStructFields(val.Type())...)
		vals, err := f.api.GetStructFieldValue(val)
		if err != nil {
			return nil, nil, err
		}
		sv = append(sv, vals...)
	}
	return sf, sv, nil
}

func (f *fillField) addFieldInMap(val reflect.Value) (reflect.Value, error) {
	addKeys, addValues, err := f.api.GetMapKeyValueToAdd(val)
	if err != nil {
		return reflect.Value{}, err
	}
	if len(addKeys) != len(addValues) {
		panic("keys must match values")
	}
	if len(addValues) == 0 {
		return val, nil
	}
	newMap := reflect.MakeMapWithSize(val.Type(), val.Len()+len(addValues))
	iter := val.MapRange()
	for iter.Next() {
		newMap.SetMapIndex(iter.Key(), iter.Value())
	}
	for i, v := range addKeys {
		newMap.SetMapIndex(v, addValues[i])
	}
	return newMap, nil
}
