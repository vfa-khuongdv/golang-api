package utils

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type stringerValue struct {
	val string
}

func (s stringerValue) String() string {
	return s.val
}

func resetSensitiveKeyCache() {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	sensitiveKeyCache = make(map[string]map[string]bool)
}

func TestContainsSensitiveKey_InternalBranches(t *testing.T) {
	t.Run("EmptyMaskFields", func(t *testing.T) {
		resetSensitiveKeyCache()
		assert.False(t, containsSensitiveKey(nil, "password"))
		assert.False(t, containsSensitiveKey([]string{}, "password"))
	})

	t.Run("EvictionBranchWhenCacheIsFull", func(t *testing.T) {
		resetSensitiveKeyCache()

		cacheMutex.Lock()
		for i := range MAX_CACHE_ENTRIES {
			sensitiveKeyCache[fmt.Sprintf("cache-key-%d", i)] = map[string]bool{"password": true}
		}
		cacheMutex.Unlock()

		found := containsSensitiveKey([]string{"password", "token"}, "password")
		assert.True(t, found)

		cacheMutex.RLock()
		cacheLen := len(sensitiveKeyCache)
		cacheMutex.RUnlock()
		assert.LessOrEqual(t, cacheLen, (MAX_CACHE_ENTRIES/2)+1)
	})

	t.Run("DoubleCheckBranchAfterWriteLock", func(t *testing.T) {
		resetSensitiveKeyCache()
		originalHook := onCacheWriteLock
		t.Cleanup(func() {
			onCacheWriteLock = originalHook
		})

		maskFields := []string{"password", "token"}
		onCacheWriteLock = func() {
			cacheKey := "password,token"
			sensitiveKeyCache[cacheKey] = map[string]bool{"password": true, "token": true}
			onCacheWriteLock = func() {}
		}

		found := containsSensitiveKey(maskFields, "password")
		assert.True(t, found)
		assert.True(t, containsSensitiveKey(maskFields, "token"))
		resetSensitiveKeyCache()
	})
}

func TestMaskValue_InternalBranches(t *testing.T) {
	t.Run("StringerAndNil", func(t *testing.T) {
		assert.Equal(t, "s****t", maskValue(stringerValue{val: "secret"}))
		assert.Nil(t, maskValue(nil))
	})
}

func TestMaskReflectedValue_InternalBranches(t *testing.T) {
	t.Run("StructBranch", func(t *testing.T) {
		type sample struct {
			Name   string
			Age    int
			Active bool
			Meta   map[string]string
			secret string
		}

		in := sample{
			Name:   "john",
			Age:    42,
			Active: true,
			Meta:   map[string]string{"k": "v"},
			secret: "private",
		}

		out := maskReflectedValue(in).(sample)
		assert.Equal(t, "*****", out.Name)
		assert.Equal(t, 0, out.Age)
		assert.False(t, out.Active)
		assert.Nil(t, out.Meta)
	})

	t.Run("SliceTypeMismatchBranch", func(t *testing.T) {
		type customBool bool
		in := []customBool{true, false}
		out := maskReflectedValue(in).([]customBool)
		assert.Equal(t, []customBool{false, false}, out)
	})

	t.Run("MaskElementByTypeBranches", func(t *testing.T) {
		uintResult := maskElementByType(reflect.ValueOf(uint16(9)))
		assert.Equal(t, uint16(0), uintResult.Interface())

		structResult := maskElementByType(reflect.ValueOf(struct{ Value string }{Value: "x"}))
		assert.Equal(t, struct{ Value string }{}, structResult.Interface())
	})
}

func TestCensorInternalBranches(t *testing.T) {
	t.Run("ArrayBranchInCensorSlice", func(t *testing.T) {
		in := [2]string{"ab", "cd"}
		out := censorSlice(in, []string{"password"}).([2]string)
		assert.Equal(t, in, out)
	})

	t.Run("NilPointerInNonSensitiveStructField", func(t *testing.T) {
		type sample struct {
			Name *string
		}

		in := sample{Name: nil}
		out := censorStruct(in, []string{"password"}).(sample)
		assert.Nil(t, out.Name)
	})
}
