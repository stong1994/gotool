# gotool

## dynamic add field for struct and map
Implement `gotool.IAddField` should be the first thing:
```go
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

```
Now, we can add field for struct:
```go
t.Run("addUserName", func(t *testing.T) {
    u := User{
        UserID:  "abc",
        UserAge: 12,
    }
    u2, _ := gotool.AddField(u, addUserName{})
    assert.Equal(t, reflect.Struct, reflect.ValueOf(u2).Kind())
    assert.Equal(t, getUserName(u.UserID), reflect.ValueOf(u2).FieldByName("UserName").Interface())
})
```

Or, we can add field for map:
```go
t.Run("addUserNameInMap", func(t *testing.T) {
    user := map[string]any{
        "user_id": "a",
    }
    u2, _ := gotool.AddField(user, addUserName{})
    assert.Equal(t, reflect.Map, reflect.ValueOf(u2).Kind())
    assert.Equal(t, getUserName(user["user_id"].(string)), u2.(map[string]any)["user_name"].(string))
})
```