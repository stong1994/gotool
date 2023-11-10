package gotool_test

import (
	"fmt"
	"github.com/stong1994/gotool"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestAddField(t *testing.T) {
	t.Run("addUserName", func(t *testing.T) {
		u := User{
			UserID:  "abc",
			UserAge: 12,
		}
		u2, _ := gotool.AddField(u, addUserName{})
		assert.Equal(t, reflect.Struct, reflect.ValueOf(u2).Kind())
		assert.Equal(t, getUserName(u.UserID), reflect.ValueOf(u2).FieldByName("UserName").Interface())
	})
	t.Run("noAddUserName", func(t *testing.T) {
		u := UserNoID{
			UserAge: 12,
		}
		u2, _ := gotool.AddField(u, addUserName{})
		assert.Equal(t, reflect.Struct, reflect.ValueOf(u2).Kind())
		assert.Equal(t, reflect.Value{}, reflect.ValueOf(u2).FieldByName("UserName"))
	})
	t.Run("addUserNameInList", func(t *testing.T) {
		u := []User{
			{
				UserID:  "abc",
				UserAge: 12,
			},
			{
				UserID:  "def",
				UserAge: 20,
			},
		}
		u2, _ := gotool.AddField(u, addUserName{})
		assert.Equal(t, reflect.Slice, reflect.ValueOf(u2).Kind())
		for i := 0; i < reflect.ValueOf(u2).Len(); i++ {
			assert.Equal(t, getUserName(u[i].UserID), reflect.ValueOf(u2).Index(i).FieldByName("UserName").Interface())
		}
	})
	t.Run("addUserNameInEmptyList", func(t *testing.T) {
		var u []User
		u2, _ := gotool.AddField(u, addUserName{})
		assert.Equal(t, reflect.Slice, reflect.ValueOf(u2).Kind())
		assert.Equal(t, 0, reflect.ValueOf(u2).Len())
	})

	// map
	t.Run("addUserNameInMap", func(t *testing.T) {
		user := map[string]any{
			"user_id": "a",
		}
		u2, _ := gotool.AddField(user, addUserName{})
		assert.Equal(t, reflect.Map, reflect.ValueOf(u2).Kind())
		assert.Equal(t, getUserName(user["user_id"].(string)), u2.(map[string]any)["user_name"].(string))
	})

	t.Run("noAddUserNameInMap", func(t *testing.T) {
		user := map[string]any{
			"user_age": 12,
		}
		u2, _ := gotool.AddField(user, addUserName{})
		assert.Equal(t, reflect.Map, reflect.ValueOf(u2).Kind())
		assert.Nil(t, u2.(map[string]any)["user_name"])
	})
	t.Run("addUserNameInMapInSlice", func(t *testing.T) {
		user := []map[string]any{
			{
				"user_id": "a",
			},
			{
				"user_id": "b",
			},
			{
				"user_age": 20,
			},
		}
		u2, _ := gotool.AddField(user, addUserName{})
		assert.Equal(t, reflect.Slice, reflect.ValueOf(u2).Kind())
		for i, v := range u2.([]map[string]any) {
			if user[i]["user_id"] != nil {
				assert.Equal(t, getUserName(user[i]["user_id"].(string)), v["user_name"].(string))
			} else {
				assert.Equal(t, nil, v["user_name"])
			}
		}
	})

}

type User struct {
	UserID  string
	UserAge int
}

type UserNoID struct {
	UserAge int
}

type addUserName struct{}

func (a addUserName) GetMapKeyValueToAdd(val reflect.Value) (keys, values []reflect.Value, err error) {
	iter := val.MapRange()
	for iter.Next() {
		if iter.Key().Interface() == "user_id" {
			keys = append(keys, reflect.ValueOf("user_name"))
			values = append(values, reflect.ValueOf(getUserName(iter.Value().Interface().(string))))
		}
	}
	return
}

func (a addUserName) IsNeedAddStructField(field reflect.Type) bool {
	for i := 0; i < field.NumField(); i++ {
		if field.Field(i).Name == "UserID" {
			return true
		}
	}
	return false
}

func (a addUserName) AddStructFields(reflect.Type) []reflect.StructField {
	return []reflect.StructField{
		{
			Name: "UserName",
			Type: reflect.TypeOf(""),
			Tag:  `json:"user_name"`,
		},
	}
}

func (a addUserName) GetStructFieldValue(val reflect.Value) ([]reflect.Value, error) {
	var fields []reflect.Value
	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		if typ.Field(i).Name == "UserID" {
			name := getUserName(val.Field(i).Interface().(string))
			fields = append(fields, reflect.ValueOf(name))
		}
	}
	return fields, nil
}

func getUserName(userID string) string {
	return fmt.Sprintf("userName_%s", userID)
}
