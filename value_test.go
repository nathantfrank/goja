package goja

import (
	"testing"

	"github.com/dop251/goja/unistring"
)

func TestIntSameAsInt(t *testing.T) {
	if !valueInt(5).SameAs(valueInt(5)) {
		t.Fatal("values are not equal")
	}
}

func TestIntStrictEqualsInt64(t *testing.T) {
	if !valueInt(5).StrictEquals(valueInt64(5)) {
		t.Fatal("values are not equal")
	}
}

func TestIntStrictEqualsFloat(t *testing.T) {
	if !valueInt(5).StrictEquals(valueFloat(5.0)) {
		t.Fatal("values are not equal")
	}
}

func TestIntZeroStrictEqualsFloatZero(t *testing.T) {
	if !valueInt(0).StrictEquals(valueFloat(0.0)) {
		t.Fatal("values are not equal")
	}
}

func TestFloatArrayIncludes(t *testing.T) {
	vm := New()
	res, err := vm.RunString(`[0.0].includes(0)`)
	if err != nil {
		t.Fatalf("unexpected error %s", err.Error())
	}
	if !res.SameAs(valueBool(true)) {
		t.Fatal("value not found in array")
	}

	res, err = vm.RunString(`[0.0].includes(-0)`)
	if err != nil {
		t.Fatalf("unexpected error %s", err.Error())
	}
	if !res.SameAs(valueBool(true)) {
		t.Fatal("value not found in array")
	}

	res, err = vm.RunString(`[0].includes(0.0)`)
	if err != nil {
		t.Fatalf("unexpected error %s", err.Error())
	}
	if !res.SameAs(valueBool(true)) {
		t.Fatal("value not found in array")
	}
}

func TestValueMemUsage(t *testing.T) {
	tests := []struct {
		name           string
		val            MemUsageReporter
		expectedMem    uint64
		expectedNewMem uint64
		errExpected    error
	}{
		{
			name:           "should have a value of SizeNumber given a valueInt",
			val:            valueInt(99),
			expectedMem:    SizeEmptyStruct,
			expectedNewMem: SizeEmptyStruct,
			errExpected:    nil,
		},
		{
			name:           "should have a value of SizeBool given a valueBool",
			val:            valueBool(true),
			expectedMem:    SizeBool,
			expectedNewMem: SizeBool,
			errExpected:    nil,
		},
		{
			name:           "should have a value of SizeEmptyStruct given a valueNull",
			val:            valueNull{},
			expectedMem:    SizeEmptyStruct,
			expectedNewMem: SizeEmptyStruct,
			errExpected:    nil,
		},
		{
			name:           "should have a value of SizeNumber given a valueFloat",
			val:            valueFloat(3.3),
			expectedMem:    SizeNumber,
			expectedNewMem: SizeNumber,
			errExpected:    nil,
		},
		{
			name:           "should account for ref given a valueUnresolved",
			val:            valueUnresolved{ref: "test"},
			expectedMem:    4,
			expectedNewMem: 4 + SizeString,
			errExpected:    nil,
		},
		{
			name:           "should account for desc given a Symbol",
			val:            &Symbol{desc: newStringValue("test")},
			expectedMem:    4,
			expectedNewMem: 4 + SizeString,
			errExpected:    nil,
		},
		{
			name:           "should return 0 given a Symbol with nil desc",
			val:            &Symbol{},
			expectedMem:    0,
			expectedNewMem: 0,
			errExpected:    nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			total, newTotal, err := tc.val.MemUsage(NewMemUsageContext(New(), 100, 100, 100, 100, nil))
			if err != tc.errExpected {
				t.Fatalf("Unexpected error. Actual: %v Expected: %v", err, tc.errExpected)
			}
			if err != nil && tc.errExpected != nil && err.Error() != tc.errExpected.Error() {
				t.Fatalf("Errors do not match. Actual: %v Expected: %v", err, tc.errExpected)
			}
			if total != tc.expectedMem {
				t.Fatalf("Unexpected memory return. Actual: %v Expected: %v", total, tc.expectedMem)
			}
			if newTotal != tc.expectedNewMem {
				t.Fatalf("Unexpected new memory return. Actual: %v Expected: %v", newTotal, tc.expectedNewMem)
			}
		})
	}
}

func TestValuePropertyMemUsage(t *testing.T) {
	tests := []struct {
		name           string
		val            *valueProperty
		expectedMem    uint64
		expectedNewMem uint64
		errExpected    error
	}{
		{
			name:           "should have a value of SizeEmptyStruct given nil valueProperty",
			val:            nil,
			expectedMem:    SizeEmptyStruct,
			expectedNewMem: SizeEmptyStruct,
			errExpected:    nil,
		},
		{
			name:           "should account for value given a valueProperty with non-empty value",
			val:            &valueProperty{value: valueInt(99)},
			expectedMem:    SizeInt,
			expectedNewMem: SizeInt,
			errExpected:    nil,
		},
		{
			name: "should account for getterFunc given a valueProperty with non-empty getterFunc",
			val: &valueProperty{
				getterFunc: &Object{
					self: &baseObject{propNames: []unistring.String{"test"}, values: map[unistring.String]Value{"test": valueInt(99)}},
				},
			},
			expectedMem:    SizeEmptyStruct + SizeEmptyStruct + 4,
			expectedNewMem: SizeEmptyStruct + SizeEmptyStruct + (4 + SizeString),
			errExpected:    nil,
		},
		{
			name: "should account for setterFunc given a valueProperty with non-empty setterFunc",
			val: &valueProperty{
				setterFunc: &Object{
					self: &baseObject{propNames: []unistring.String{"test"}, values: map[unistring.String]Value{"test": valueInt(99)}},
				},
			},
			expectedMem:    SizeEmptyStruct + SizeEmptyStruct + 4,
			expectedNewMem: SizeEmptyStruct + SizeEmptyStruct + (4 + SizeString),
			errExpected:    nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			total, newTotal, err := tc.val.MemUsage(NewMemUsageContext(New(), 100, 100, 100, 100, nil))
			if err != tc.errExpected {
				t.Fatalf("Unexpected error. Actual: %v Expected: %v", err, tc.errExpected)
			}
			if err != nil && tc.errExpected != nil && err.Error() != tc.errExpected.Error() {
				t.Fatalf("Errors do not match. Actual: %v Expected: %v", err, tc.errExpected)
			}
			if total != tc.expectedMem {
				t.Fatalf("Unexpected memory return. Actual: %v Expected: %v", total, tc.expectedMem)
			}
			if newTotal != tc.expectedNewMem {
				t.Fatalf("Unexpected new memory return. Actual: %v Expected: %v", newTotal, tc.expectedNewMem)
			}
		})
	}
}

func TestObjectMemUsage(t *testing.T) {
	tests := []struct {
		name           string
		val            *Object
		expectedMem    uint64
		expectedNewMem uint64
		errExpected    error
	}{
		{
			name:           "should have a value of SizeEmptyStruct given nil Object",
			val:            nil,
			expectedMem:    SizeEmptyStruct,
			expectedNewMem: SizeEmptyStruct,
			errExpected:    nil,
		},
		{
			name:           "should have a value of SizeEmptyStruct given an empty Object with nil self",
			val:            &Object{},
			expectedMem:    SizeEmptyStruct,
			expectedNewMem: SizeEmptyStruct,
			errExpected:    nil,
		},
		{
			name:           "should account for __wrapped given an Object with non-empty __wrapped",
			val:            &Object{__wrapped: 99},
			expectedMem:    SizeInt,
			expectedNewMem: SizeInt,
			errExpected:    nil,
		},
		{
			name:           "should have a value of SizeEmptyStruct given an Object with self of type objectGoReflect",
			val:            &Object{self: &objectGoReflect{}},
			expectedMem:    SizeEmptyStruct,
			expectedNewMem: SizeEmptyStruct,
			errExpected:    nil,
		},
		{
			name:           "should have a value of SizeEmptyStruct given an Object with self of type objectGoMapReflect",
			val:            &Object{self: &objectGoMapReflect{}},
			expectedMem:    SizeEmptyStruct,
			expectedNewMem: SizeEmptyStruct,
			errExpected:    nil,
		},
		{
			name:           "should have a value of SizeEmptyStruct given an Object with self of type objectGoMapSimple",
			val:            &Object{self: &objectGoMapSimple{}},
			expectedMem:    SizeEmptyStruct,
			expectedNewMem: SizeEmptyStruct,
			errExpected:    nil,
		},
		{
			name:           "should have a value of SizeEmptyStruct given an Object with self of type objectGoSlice",
			val:            &Object{self: &objectGoSlice{}},
			expectedMem:    SizeEmptyStruct,
			expectedNewMem: SizeEmptyStruct,
			errExpected:    nil,
		},
		{
			name:           "should have a value of SizeEmptyStruct given an Object with self of type objectGoSliceReflect",
			val:            &Object{self: &objectGoSliceReflect{}},
			expectedMem:    SizeEmptyStruct,
			expectedNewMem: SizeEmptyStruct,
			errExpected:    nil,
		},
		{
			name: "should have a value of SizeEmptyStruct given an Object with self of type objectGoSliceReflect",
			val: &Object{
				self: &baseObject{propNames: []unistring.String{"test"}, values: map[unistring.String]Value{"test": valueInt(99)}},
			},
			// baseObject overhead + value
			expectedMem: SizeEmptyStruct + (4 + SizeInt),
			// baseObject overhead + value with string overhead
			expectedNewMem: SizeEmptyStruct + (4 + SizeString + SizeInt),
			errExpected:    nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			total, newTotal, err := tc.val.MemUsage(NewMemUsageContext(New(), 100, 100, 100, 100, nil))
			if err != tc.errExpected {
				t.Fatalf("Unexpected error. Actual: %v Expected: %v", err, tc.errExpected)
			}
			if err != nil && tc.errExpected != nil && err.Error() != tc.errExpected.Error() {
				t.Fatalf("Errors do not match. Actual: %v Expected: %v", err, tc.errExpected)
			}
			if total != tc.expectedMem {
				t.Fatalf("Unexpected memory return. Actual: %v Expected: %v", total, tc.expectedMem)
			}
			if newTotal != tc.expectedNewMem {
				t.Fatalf("Unexpected new memory return. Actual: %v Expected: %v", newTotal, tc.expectedNewMem)
			}
		})
	}
}
