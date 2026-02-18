package utils

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/vfa-khuongdv/golang-cms/pkg/logger"
)

const (
	// MAX_CACHE_ENTRIES limits the size of sensitiveKeyCache to prevent memory leaks
	MAX_CACHE_ENTRIES = 100
)

// CensorSensitiveData recursively censors sensitive fields in complex data structures.
// It traverses maps, slices, structs, and pointers to mask values of fields whose names
// match any entry in maskFields (case-insensitive matching).
//
// Masking strategy:
//   - Strings: Shows first and last character with asterisks in between (e.g., "s****t")
//   - Short strings (≤2 chars): Fully masked with asterisks
//   - Non-string types: Converted to string or masked generically
//
// Note: Only exported struct fields can be censored due to reflection limitations.
// The function is thread-safe and does not modify the input data.
//
// Parameters:
//   - data: The data structure to censor (can be any type)
//   - maskFields: List of field/key names to censor (case-insensitive)
//
// Returns: A new data structure with sensitive fields censored.
func CensorSensitiveData(data any, maskFields []string) any {
	if data == nil {
		return nil
	}

	// Early return if no fields to mask
	if len(maskFields) == 0 {
		return data
	}

	val := reflect.ValueOf(data)

	switch val.Kind() {
	case reflect.Slice, reflect.Array:
		return censorSlice(data, maskFields)
	case reflect.Map:
		return censorMap(data, maskFields)
	case reflect.Struct:
		return censorStruct(data, maskFields)
	case reflect.Ptr:
		if val.IsNil() {
			return nil
		}
		return CensorSensitiveData(val.Elem().Interface(), maskFields)
	case reflect.String:
		return data
	default:
		return data
	}
}

// censorSlice recursively censors each element in a slice/array.
func censorSlice(data any, maskFields []string) any {
	val := reflect.ValueOf(data)

	// Handle arrays differently from slices
	var censoredSlice reflect.Value
	if val.Kind() == reflect.Array {
		censoredSlice = reflect.New(val.Type()).Elem()
	} else {
		censoredSlice = reflect.MakeSlice(val.Type(), val.Len(), val.Len())
	}

	for i := 0; i < val.Len(); i++ {
		item := val.Index(i).Interface()
		censoredItem := CensorSensitiveData(item, maskFields)
		censoredSlice.Index(i).Set(reflect.ValueOf(censoredItem))
	}

	return censoredSlice.Interface()
}

// censorMap recursively censors map entries based on keys.
func censorMap(data any, maskFields []string) any {
	val := reflect.ValueOf(data)
	censoredMap := reflect.MakeMap(val.Type())

	iter := val.MapRange()
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()

		keyStr := fmt.Sprintf("%v", key.Interface())

		var censoredValue reflect.Value
		if containsSensitiveKey(maskFields, keyStr) {
			// Mask the entire value if key is sensitive
			censoredValue = reflect.ValueOf(maskValue(value.Interface()))
		} else {
			censoredValue = reflect.ValueOf(CensorSensitiveData(value.Interface(), maskFields))
		}

		censoredMap.SetMapIndex(key, censoredValue)
	}

	return censoredMap.Interface()
}

// censorStruct recursively censors struct fields based on field names.
func censorStruct(data any, maskFields []string) any {
	val := reflect.ValueOf(data)
	typ := val.Type()
	censoredStruct := reflect.New(typ).Elem()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		if containsSensitiveKey(maskFields, fieldType.Name) {
			// Field needs to be masked
			if field.Kind() == reflect.Ptr {
				if field.IsNil() {
					censoredStruct.Field(i).Set(reflect.Zero(field.Type()))
				} else {
					maskedVal := maskValue(field.Elem().Interface())
					maskedValReflect := reflect.ValueOf(maskedVal)

					ptr := reflect.New(fieldType.Type.Elem())
					ptr.Elem().Set(matchedValOrZero(maskedValReflect, fieldType.Type.Elem()))
					censoredStruct.Field(i).Set(ptr)
				}
			} else {
				censoredStruct.Field(i).Set(matchedValOrZero(reflect.ValueOf(maskValue(field.Interface())), fieldType.Type))
			}
		} else {
			// Field does not need to be masked, process recursively
			censoredValue := CensorSensitiveData(field.Interface(), maskFields)
			if field.Kind() == reflect.Ptr {
				if field.IsNil() {
					censoredStruct.Field(i).Set(reflect.Zero(field.Type()))
				} else {
					ptr := reflect.New(fieldType.Type.Elem())
					ptr.Elem().Set(matchedValOrZero(reflect.ValueOf(censoredValue), fieldType.Type.Elem()))
					censoredStruct.Field(i).Set(ptr)
				}
			} else {
				censoredStruct.Field(i).Set(matchedValOrZero(reflect.ValueOf(censoredValue), fieldType.Type))
			}
		}
	}

	return censoredStruct.Interface()
}

// matchedValOrZero attempts to assign val to typ if compatible, otherwise returns zero value.
// This prevents panics when types are incompatible during reflection operations.
func matchedValOrZero(val reflect.Value, typ reflect.Type) reflect.Value {
	if val.Type().AssignableTo(typ) {
		return val
	}
	// Log warning about type mismatch (data loss)
	logger.Error(fmt.Sprintf("Type mismatch in censoring: cannot assign %v to %v, using zero value", val.Type(), typ))
	return reflect.Zero(typ)
}

// sensitiveKeyCache caches lowercase sensitive keys for O(1) lookup performance
// Protected by cacheMutex for thread-safe concurrent access
var (
	sensitiveKeyCache = make(map[string]map[string]bool)
	cacheMutex        sync.RWMutex
	onCacheWriteLock  = func() {}
)

// containsSensitiveKey checks if item matches any sensitive key (case-insensitive).
// Uses a cached map for O(1) lookups instead of O(n) slice iteration.
// Cache keys are sorted to avoid duplicates from different field orders.
func containsSensitiveKey(maskFields []string, item string) bool {
	if len(maskFields) == 0 {
		return false
	}

	// Sort maskFields to create consistent cache key (avoid ["a","b"] vs ["b","a"])
	sortedFields := make([]string, len(maskFields))
	copy(sortedFields, maskFields)
	sort.Strings(sortedFields)
	cacheKey := strings.Join(sortedFields, ",")

	// Try read lock first (most common case)
	cacheMutex.RLock()
	if cache, exists := sensitiveKeyCache[cacheKey]; exists {
		_, found := cache[strings.ToLower(item)]
		cacheMutex.RUnlock()
		return found
	}
	cacheMutex.RUnlock()

	// Cache miss - acquire write lock to build cache
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	onCacheWriteLock()

	// Double-check after acquiring write lock (another goroutine might have added it)
	if cache, exists := sensitiveKeyCache[cacheKey]; exists {
		_, found := cache[strings.ToLower(item)]
		return found
	}

	// Implement cache size limit to prevent memory leaks
	if len(sensitiveKeyCache) >= MAX_CACHE_ENTRIES {
		// Clear half of the cache (simple eviction strategy)
		count := 0
		for k := range sensitiveKeyCache {
			delete(sensitiveKeyCache, k)
			count++
			if count >= MAX_CACHE_ENTRIES/2 {
				break
			}
		}
	}

	// Build cache for this set of maskFields
	cache := make(map[string]bool, len(maskFields))
	for _, field := range maskFields {
		cache[strings.ToLower(field)] = true
	}
	sensitiveKeyCache[cacheKey] = cache

	_, found := cache[strings.ToLower(item)]
	return found
}

// maskValue masks sensitive values based on their type.
func maskValue(value any) any {
	switch v := value.(type) {
	case string:
		return maskString(v)
	case fmt.Stringer:
		return maskString(v.String())
	case []byte:
		return []byte(maskString(string(v)))
	case nil:
		return nil
	default:
		return maskReflectedValue(value)
	}
}

// maskString masks a string by replacing its middle characters with asterisks.
// For strings longer than 2 characters, it shows the first and last character.
// For shorter strings, it fully masks with asterisks.
func maskString(s string) string {
	if len(s) > 2 {
		maskLen := min(len(s)-2, 8) // Use built-in min() from Go 1.21+
		return string(s[0]) + strings.Repeat("*", maskLen) + string(s[len(s)-1])
	}
	return strings.Repeat("*", len(s))
}

// maskReflectedValue masks values of non-standard types using reflection.
// It handles slices, arrays, and structs while maintaining type safety.
func maskReflectedValue(value any) any {
	val := reflect.ValueOf(value)

	switch val.Kind() {
	case reflect.Slice, reflect.Array:
		// Create a new slice/array and mask each element based on its type
		var maskedSlice reflect.Value
		if val.Kind() == reflect.Array {
			maskedSlice = reflect.New(val.Type()).Elem()
		} else {
			maskedSlice = reflect.MakeSlice(val.Type(), val.Len(), val.Len())
		}

		for i := 0; i < val.Len(); i++ {
			elem := val.Index(i)
			maskedElem := maskElementByType(elem)
			if maskedElem.Type().AssignableTo(elem.Type()) {
				maskedSlice.Index(i).Set(maskedElem)
			} else {
				// If type mismatch, use zero value
				maskedSlice.Index(i).Set(reflect.Zero(elem.Type()))
			}
		}
		return maskedSlice.Interface()
	case reflect.Struct:
		maskedStruct := reflect.New(val.Type()).Elem()
		for i := 0; i < val.NumField(); i++ {
			field := maskedStruct.Field(i)
			if !field.CanSet() {
				continue // Skip unexported fields
			}
			switch field.Kind() {
			case reflect.String:
				field.SetString("*****")
			case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
				field.SetInt(0)
			case reflect.Bool:
				field.SetBool(false)
			default:
				field.Set(reflect.Zero(field.Type()))
			}
		}
		return maskedStruct.Interface()
	default:
		return "*****"
	}
}

// maskElementByType returns a masked value for a reflect.Value based on its type.
func maskElementByType(elem reflect.Value) reflect.Value {
	switch elem.Kind() {
	case reflect.String:
		return reflect.ValueOf("*****")
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		return reflect.ValueOf(0).Convert(elem.Type())
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		return reflect.ValueOf(uint(0)).Convert(elem.Type())
	case reflect.Float32, reflect.Float64:
		return reflect.ValueOf(0.0).Convert(elem.Type())
	case reflect.Bool:
		return reflect.ValueOf(false)
	default:
		return reflect.Zero(elem.Type())
	}
}
