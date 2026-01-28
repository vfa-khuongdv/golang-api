package utils_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vfa-khuongdv/golang-cms/internal/utils"
)

type secretStringer struct {
	secret string
}

func (s secretStringer) String() string {
	return s.secret
}

func TestCensorSensitiveData(t *testing.T) {
	maskFields := []string{"password", "apiKey"}

	t.Run("Nil value input", func(t *testing.T) {
		type TestInput struct {
			Name     string
			Password *string
			Deeps    struct {
				Password *string
			}
			DeepsPtr *struct {
				Password *string
			}
		}

		passwordPtr := "12345"
		passwordPtrMasked := "1***5"

		input := TestInput{
			Name: "test", Password: nil,
			Deeps: struct {
				Password *string
			}{Password: nil},
			DeepsPtr: &struct {
				Password *string
			}{Password: &passwordPtr}}

		expected := TestInput{Name: "test", Password: nil, Deeps: struct {
			Password *string
		}{Password: nil}, DeepsPtr: &struct {
			Password *string
		}{Password: &passwordPtrMasked}}

		result := utils.CensorSensitiveData(input, maskFields)

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}

		// nil value input without maskFields
		result = utils.CensorSensitiveData(nil, maskFields)
		assert.Nil(t, result, "Expected nil result for nil input without maskFields")

	})

	t.Run("Nil input", func(t *testing.T) {
		var input interface{} = nil
		result := utils.CensorSensitiveData(input, maskFields)
		if result != nil {
			t.Errorf("Expected nil, got %v", result)
		}
	})
	t.Run("Nil value input without maskFields", func(t *testing.T) {
		type TestInput struct {
			Name     string
			DeepsPtr *struct {
				Name *string
			}
		}
		// input
		input := TestInput{
			Name:     "test",
			DeepsPtr: &struct{ Name *string }{Name: nil},
		}
		// expected
		expected := TestInput{
			Name:     "test",
			DeepsPtr: &struct{ Name *string }{Name: nil},
		}
		// call function
		result := utils.CensorSensitiveData(input, nil)
		// check result
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}

	})

	t.Run("Map with sensitive keys", func(t *testing.T) {
		input := map[string]string{"password": "secret", "username": "user"}
		expected := map[string]string{"password": "s****t", "username": "user"}
		result := utils.CensorSensitiveData(input, maskFields)
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("Struct with sensitive fields", func(t *testing.T) {
		type User struct {
			Password string
			Username string
		}
		input := User{Password: "secret", Username: "user"}
		expected := User{Password: "s****t", Username: "user"}
		result := utils.CensorSensitiveData(input, maskFields).(User)
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("Pointer to struct", func(t *testing.T) {
		type User struct {
			Password string
			Username string
		}
		input := &User{Password: "secret", Username: "user"}
		expected := User{Password: "s****t", Username: "user"}
		result := utils.CensorSensitiveData(input, maskFields)
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("Nested structures", func(t *testing.T) {
		type Profile struct {
			APIKey  string
			Details map[string]string
		}
		type User struct {
			Password string
			Profile  Profile
		}
		input := User{
			Password: "mypassword",
			Profile: Profile{
				APIKey: "12345",
				Details: map[string]string{
					"username": "user",
					"email":    "user@example.com",
				},
			},
		}
		expected := User{
			Password: "m********d",
			Profile: Profile{
				APIKey: "1***5",
				Details: map[string]string{
					"username": "user",
					"email":    "user@example.com",
				},
			},
		}
		result := utils.CensorSensitiveData(input, maskFields).(User)
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("Slice of structs", func(t *testing.T) {
		type User struct {
			Password string
			Username string
		}
		input := []User{
			{Password: "secret", Username: "user"},
		}
		expected := []User{
			{Password: "s****t", Username: "user"},
		}
		result := utils.CensorSensitiveData(input, maskFields).([]User)
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("Ptr value is nil", func(t *testing.T) {
		type testStruct struct {
			Field1 string
			Field2 string
		}

		var ptr *testStruct = nil

		result := utils.CensorSensitiveData(ptr, []string{"Field1"})
		if result != nil {
			t.Errorf("Expected nil result for nil pointer input, got: %#v", result)
		}
	})

	t.Run("Mask string fields in struct", func(t *testing.T) {
		type User struct {
			Password string
			Username string
		}

		maskFields := []string{"Password"}

		input := User{Password: "1", Username: "user"}
		expected := User{
			Password: "*", // Assuming the mask function replaces the password with a single asterisk
			Username: "user",
		}

		result := utils.CensorSensitiveData(input, maskFields).(User)

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %+v, got %+v", expected, result)
		}
	})

	t.Run("Default case with unsupported type", func(t *testing.T) {
		maskFields := []string{"Password"}

		tests := []struct {
			name  string
			input any
			want  any
		}{
			{
				name:  "int input returns same int",
				input: 42,
				want:  42,
			},
			{
				name:  "float input returns same float",
				input: 3.14,
				want:  3.14,
			},
			{
				name:  "bool input returns same bool",
				input: true,
				want:  true,
			},
			{
				name:  "complex input returns same complex",
				input: complex(1, 2),
				want:  complex(1, 2),
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := utils.CensorSensitiveData(tt.input, maskFields)
				if got != tt.want {
					t.Errorf("CensorSensitiveData() = %v, want %v", got, tt.want)
				}
			})
		}
	})

	t.Run("Test with value is type stringer", func(t *testing.T) {

		type testStruct struct {
			Secret string
			Other  string
		}

		maskFields := []string{"Secret"}

		input := testStruct{
			Secret: secretStringer{secret: "verysecret"}.String(), // truyền chuỗi từ Stringer
			Other:  "public",
		}

		got := utils.CensorSensitiveData(input, maskFields)

		result, ok := got.(testStruct)
		if !ok {
			t.Fatalf("expected testStruct, got %T", got)
		}

		expectedMasked := "v********t"
		if result.Secret != expectedMasked {
			t.Errorf("expected masked Secret %q, got %q", expectedMasked, result.Secret)
		}

		if result.Other != input.Other {
			t.Errorf("expected Other to be unchanged: got %q", result.Other)
		}

	})

	t.Run("Test with []byte input", func(t *testing.T) {
		type byteStruct struct {
			Secret []byte
			Other  string
		}

		maskFields := []string{"Secret"}

		input := byteStruct{
			Secret: []byte("verysecret"),
			Other:  "public",
		}
		expected := byteStruct{
			Secret: []byte("v********t"),
			Other:  "public",
		}

		result := utils.CensorSensitiveData(input, maskFields).(byteStruct)

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("Test early return with empty maskFields", func(t *testing.T) {
		type User struct {
			Password string
			Email    string
		}

		input := User{Password: "secret123", Email: "user@example.com"}

		// Call with empty maskFields - should return unchanged
		result := utils.CensorSensitiveData(input, []string{})
		assert.Equal(t, input, result, "Should return unchanged data when maskFields is empty")

		// Call with nil maskFields - should return unchanged
		result2 := utils.CensorSensitiveData(input, nil)
		assert.Equal(t, input, result2, "Should return unchanged data when maskFields is nil")
	})

	t.Run("Test containsSensitiveKey with case insensitivity", func(t *testing.T) {
		maskFields := []string{"Password", "APIKey", "Token"}

		tests := []struct {
			name     string
			input    map[string]string
			expected map[string]string
		}{
			{
				name:     "lowercase keys should match",
				input:    map[string]string{"password": "secret", "username": "user"},
				expected: map[string]string{"password": "s****t", "username": "user"},
			},
			{
				name:     "uppercase keys should match",
				input:    map[string]string{"PASSWORD": "secret", "username": "user"},
				expected: map[string]string{"PASSWORD": "s****t", "username": "user"},
			},
			{
				name:     "mixed case keys should match",
				input:    map[string]string{"PaSsWoRd": "secret", "username": "user"},
				expected: map[string]string{"PaSsWoRd": "s****t", "username": "user"},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := utils.CensorSensitiveData(tt.input, maskFields)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("Test non-string slice masking", func(t *testing.T) {
		type TestData struct {
			Numbers  []int
			Flags    []bool
			Floats   []float64
			Username string
		}

		maskFields := []string{"Numbers", "Flags", "Floats"}

		input := TestData{
			Numbers:  []int{1, 2, 3},
			Flags:    []bool{true, false},
			Floats:   []float64{1.1, 2.2},
			Username: "user",
		}

		result := utils.CensorSensitiveData(input, maskFields).(TestData)

		// Verify slices are masked to zero values
		assert.Equal(t, []int{0, 0, 0}, result.Numbers, "Int slice should be masked to zeros")
		assert.Equal(t, []bool{false, false}, result.Flags, "Bool slice should be masked to false")
		assert.Equal(t, []float64{0.0, 0.0}, result.Floats, "Float slice should be masked to zeros")
		assert.Equal(t, "user", result.Username, "Non-masked field should remain unchanged")
	})

	t.Run("Test maskElementByType for various types", func(t *testing.T) {
		type ComplexStruct struct {
			IntField    int
			UintField   uint
			FloatField  float64
			BoolField   bool
			StringField string
		}

		maskFields := []string{"IntField", "UintField", "FloatField", "BoolField", "StringField"}

		input := ComplexStruct{
			IntField:    42,
			UintField:   100,
			FloatField:  3.14,
			BoolField:   true,
			StringField: "secret",
		}

		expected := ComplexStruct{
			IntField:    0,
			UintField:   0,
			FloatField:  0.0,
			BoolField:   false,
			StringField: "s****t",
		}

		result := utils.CensorSensitiveData(input, maskFields).(ComplexStruct)
		assert.Equal(t, expected, result)
	})

	t.Run("Test deeply nested structure with mixed types", func(t *testing.T) {
		type Address struct {
			Street string
			City   string
		}

		type Contact struct {
			Email   string
			Phone   string
			Address Address
		}

		type User struct {
			Username string
			Password string
			Contact  Contact
			APIKey   string
		}

		maskFields := []string{"Password", "Email", "Phone", "APIKey"}

		input := User{
			Username: "john",
			Password: "secret123",
			Contact: Contact{
				Email: "john@example.com",
				Phone: "1234567890",
				Address: Address{
					Street: "123 Main St",
					City:   "NYC",
				},
			},
			APIKey: "key123456",
		}

		result := utils.CensorSensitiveData(input, maskFields).(User)

		assert.Equal(t, "john", result.Username)
		assert.Contains(t, result.Password, "*")
		assert.NotEqual(t, "secret123", result.Password)
		assert.Contains(t, result.Contact.Email, "*")
		assert.Contains(t, result.Contact.Phone, "*")
		assert.Equal(t, "123 Main St", result.Contact.Address.Street)
		assert.Equal(t, "NYC", result.Contact.Address.City)
		assert.Contains(t, result.APIKey, "*")
	})

	t.Run("Test slice of maps", func(t *testing.T) {
		maskFields := []string{"password", "token"}

		input := []map[string]string{
			{"username": "user1", "password": "pass1"},
			{"username": "user2", "token": "token123"},
		}

		result := utils.CensorSensitiveData(input, maskFields).([]map[string]string)

		assert.Equal(t, "user1", result[0]["username"])
		assert.Contains(t, result[0]["password"], "*")
		assert.Equal(t, "user2", result[1]["username"])
		assert.Contains(t, result[1]["token"], "*")
	})

	t.Run("Test map with nested slices", func(t *testing.T) {
		type User struct {
			Username string
			Password string
		}

		maskFields := []string{"Password"}

		input := map[string][]User{
			"users": {
				{Username: "user1", Password: "pass1"},
				{Username: "user2", Password: "pass2"},
			},
		}

		result := utils.CensorSensitiveData(input, maskFields).(map[string][]User)

		assert.Equal(t, "user1", result["users"][0].Username)
		assert.Contains(t, result["users"][0].Password, "*")
		assert.Equal(t, "user2", result["users"][1].Username)
		assert.Contains(t, result["users"][1].Password, "*")
	})

	t.Run("Test string masking edge cases", func(t *testing.T) {
		type TestCase struct {
			Name     string
			Input    string
			Expected string
		}

		tests := []TestCase{
			{Name: "Empty string", Input: "", Expected: ""},
			{Name: "Single character", Input: "a", Expected: "*"},
			{Name: "Two characters", Input: "ab", Expected: "**"},
			{Name: "Three characters", Input: "abc", Expected: "a*c"},
			{Name: "Long string", Input: "verylongpassword", Expected: "v********d"},
		}

		for _, tc := range tests {
			t.Run(tc.Name, func(t *testing.T) {
				type Data struct {
					Secret string
				}
				input := Data{Secret: tc.Input}
				result := utils.CensorSensitiveData(input, []string{"Secret"}).(Data)
				assert.Equal(t, tc.Expected, result.Secret)
			})
		}
	})

	t.Run("Test array vs slice behavior", func(t *testing.T) {
		type TestStruct struct {
			SliceData [3]string
			Username  string
		}

		maskFields := []string{"SliceData"}

		input := TestStruct{
			SliceData: [3]string{"secret1", "secret2", "secret3"},
			Username:  "user",
		}

		result := utils.CensorSensitiveData(input, maskFields).(TestStruct)

		// Check that array elements are censored
		assert.Contains(t, result.SliceData[0], "*")
		assert.Contains(t, result.SliceData[1], "*")
		assert.Contains(t, result.SliceData[2], "*")
		assert.Equal(t, "user", result.Username)
	})

	t.Run("Test caching mechanism performance", func(t *testing.T) {
		maskFields := []string{"password", "token", "apikey", "secret"}

		type User struct {
			Username string
			Password string
		}

		// First call - builds cache
		input1 := map[string]string{"username": "user1", "password": "pass1"}
		result1 := utils.CensorSensitiveData(input1, maskFields)
		assert.NotNil(t, result1)

		// Second call with same maskFields - should use cache
		input2 := map[string]string{"username": "user2", "token": "token123"}
		result2 := utils.CensorSensitiveData(input2, maskFields)
		assert.NotNil(t, result2)

		// Verify both results are correctly censored
		r1 := result1.(map[string]string)
		r2 := result2.(map[string]string)

		assert.Contains(t, r1["password"], "*")
		assert.Equal(t, "user1", r1["username"])
		assert.Contains(t, r2["token"], "*")
		assert.Equal(t, "user2", r2["username"])
	})

	t.Run("Test with interface{} values in map", func(t *testing.T) {
		maskFields := []string{"password", "count"}

		input := map[string]interface{}{
			"username": "user",
			"password": "secret",
			"count":    42,
			"active":   true,
		}

		result := utils.CensorSensitiveData(input, maskFields).(map[string]interface{})

		assert.Equal(t, "user", result["username"])
		assert.Contains(t, result["password"], "*")
		assert.Equal(t, "*****", result["count"]) // Non-string gets generic mask
		assert.Equal(t, true, result["active"])
	})

}
