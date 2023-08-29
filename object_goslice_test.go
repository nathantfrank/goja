package goja

import (
	"testing"
)

func TestGoSliceBasic(t *testing.T) {
	const SCRIPT = `
	var sum = 0;
	for (var i = 0; i < a.length; i++) {
		sum += a[i];
	}
	sum;
	`
	r := New()
	r.Set("a", []interface{}{1, 2, 3, 4})
	v, err := r.RunString(SCRIPT)
	if err != nil {
		t.Fatal(err)
	}
	if i := v.ToInteger(); i != 10 {
		t.Fatalf("Expected 10, got: %d", i)
	}
}

func TestGoSliceIn(t *testing.T) {
	const SCRIPT = `
	var idx = "";
	for (var i in a) {
		idx += i;
	}
	idx;
	`
	r := New()
	r.Set("a", []interface{}{1, 2, 3, 4})
	v, err := r.RunString(SCRIPT)
	if err != nil {
		t.Fatal(err)
	}
	if i := v.String(); i != "0123" {
		t.Fatalf("Expected '0123', got: '%s'", i)
	}
}

func TestGoSliceExpand(t *testing.T) {
	const SCRIPT = `
	var l = a.length;
	for (var i = 0; i < l; i++) {
		a[l + i] = a[i] * 2;
	}

	var sum = 0;
	for (var i = 0; i < a.length; i++) {
		sum += a[i];
	}
	sum;
	`
	r := New()
	a := []interface{}{int64(1), int64(2), int64(3), int64(4)}
	r.Set("a", &a)
	v, err := r.RunString(SCRIPT)
	if err != nil {
		t.Fatal(err)
	}
	sum := int64(0)
	for _, v := range a {
		sum += v.(int64)
	}
	if i := v.ToInteger(); i != sum {
		t.Fatalf("Expected %d, got: %d", sum, i)
	}
}

func TestGoSliceProtoMethod(t *testing.T) {
	const SCRIPT = `
	a.join(",")
	`

	r := New()
	a := []interface{}{1, 2, 3, 4}
	r.Set("a", a)
	ret, err := r.RunString(SCRIPT)
	if err != nil {
		t.Fatal(err)
	}
	if s := ret.String(); s != "1,2,3,4" {
		t.Fatalf("Unexpected result: '%s'", s)
	}
}

func TestGoSliceSetLength(t *testing.T) {
	r := New()
	a := []interface{}{1, 2, 3, 4}
	r.Set("a", &a)
	_, err := r.RunString(`
	'use strict';
	a.length = 3;
	if (a.length !== 3) {
		throw new Error("length="+a.length);
	}
	if (a[3] !== undefined) {
		throw new Error("a[3](1)="+a[3]);
	}
	a.length = 5;
	if (a.length !== 5) {
		throw new Error("length="+a.length);
	}
	if (a[3] !== null) {
		throw new Error("a[3](2)="+a[3]);
	}
	if (a[4] !== null) {
		throw new Error("a[4]="+a[4]);
	}
	`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGoSliceProto(t *testing.T) {
	r := New()
	a := []interface{}{1, nil, 3}
	r.Set("a", &a)
	r.testScriptWithTestLib(`
	var proto = [,2,,4];
	Object.setPrototypeOf(a, proto);
	assert.sameValue(a[1], null, "a[1]");
	assert.sameValue(a[3], 4, "a[3]");
	var desc = Object.getOwnPropertyDescriptor(a, "1");
	assert.sameValue(desc.value, null, "desc.value");
	assert(desc.writable, "writable");
	assert(desc.enumerable, "enumerable");
	assert(!desc.configurable, "configurable");
	var v5;
	Object.defineProperty(proto, "5", {
		set: function(v) {
			v5 = v;
		}
	});
	a[5] = "test";
	assert.sameValue(v5, "test", "v5");
	`, _undefined, t)
}

func TestGoSliceProtoProto(t *testing.T) {
	r := New()
	a := []interface{}{1, nil, 3}
	proto := []interface{}{1, 2, 3, 4}
	r.Set("a", &a)
	r.Set("proto", proto)
	_, err := r.RunString(`
	"use strict";
	var protoproto = Object.create(null);
	Object.defineProperty(protoproto, "3", {
		value: 42
	});
	Object.setPrototypeOf(proto, protoproto);
	Object.setPrototypeOf(a, proto);
	a[3] = 11;
	if (a[3] !== 11) {
		throw new Error("a[3]=" + a[3]);
	}
	`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGoSliceDelete(t *testing.T) {
	r := New()
	a := []interface{}{1, nil, 3}
	r.Set("a", a)
	v, err := r.RunString(`
	delete a[0] && delete a[1] && delete a[3];
	`)
	if err != nil {
		t.Fatal(err)
	}
	if v != valueTrue {
		t.Fatalf("not true: %v", v)
	}
}

func TestGoSlicePop(t *testing.T) {
	r := New()
	a := []interface{}{1, nil, 3}
	r.Set("a", &a)
	v, err := r.RunString(`
	a.pop()
	`)
	if err != nil {
		t.Fatal(err)
	}
	if !v.SameAs(intToValue(3)) {
		t.Fatal(v)
	}
}

func TestGoSlicePopNoPtr(t *testing.T) {
	r := New()
	a := []interface{}{1, nil, 3}
	r.Set("a", a)
	v, err := r.RunString(`
	a.pop()
	`)
	if err != nil {
		t.Fatal(err)
	}
	if !v.SameAs(intToValue(3)) {
		t.Fatal(v)
	}
}

func TestGoSliceShift(t *testing.T) {
	r := New()
	a := []interface{}{1, nil, 3}
	r.Set("a", &a)
	v, err := r.RunString(`
	a.shift()
	`)
	if err != nil {
		t.Fatal(err)
	}
	if !v.SameAs(intToValue(1)) {
		t.Fatal(v)
	}
}

func TestGoSliceLengthProperty(t *testing.T) {
	vm := New()
	vm.Set("s", []interface{}{2, 3, 4})
	_, err := vm.RunString(`
	if (!s.hasOwnProperty("length")) {
		throw new Error("hasOwnProperty() returned false");
	}
	let desc = Object.getOwnPropertyDescriptor(s, "length");
	if (desc.value !== 3 || !desc.writable || desc.enumerable || desc.configurable) {
		throw new Error("incorrect property descriptor: " + JSON.stringify(desc));
	}
	`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGoSliceSort(t *testing.T) {
	vm := New()
	s := []interface{}{4, 2, 3}
	vm.Set("s", &s)
	_, err := vm.RunString(`s.sort()`)
	if err != nil {
		t.Fatal(err)
	}
	if len(s) != 3 {
		t.Fatalf("len: %d", len(s))
	}
	if s[0] != 2 || s[1] != 3 || s[2] != 4 {
		t.Fatalf("val: %v", s)
	}
}

func TestGoSliceToString(t *testing.T) {
	vm := New()
	s := []interface{}{4, 2, 3}
	vm.Set("s", &s)
	res, err := vm.RunString("`${s}`")
	if err != nil {
		t.Fatal(err)
	}
	if exp := res.Export(); exp != "4,2,3" {
		t.Fatal(exp)
	}
}

func TestGoSliceMemUsage(t *testing.T) {
	vm := New()
	tests := []struct {
		name           string
		val            *objectGoSlice
		expectedMem    uint64
		expectedNewMem uint64
		errExpected    error
	}{
		{
			name: "should account for each value given a non-empty slice",
			val: &objectGoSlice{
				baseObject: baseObject{
					val: &Object{runtime: vm},
				},
				data: &[]interface{}{
					valueInt(99),
					valueInt(99),
				},
			},
			// overhead + values
			expectedMem: SizeEmptyStruct + SizeInt*2,
			// overhead + values
			expectedNewMem: SizeEmptyStruct + SizeInt*2,
			errExpected:    nil,
		},
		{
			name: "should account for each value given a slice with go ints",
			val: &objectGoSlice{
				baseObject: baseObject{
					val: &Object{runtime: vm},
				},
				data: &[]interface{}{
					99,
					99,
				},
			},
			// overhead + values
			expectedMem: SizeEmptyStruct + SizeInt*2,
			// overhead + values
			expectedNewMem: SizeEmptyStruct + SizeInt*2,
			errExpected:    nil,
		},
		{
			name: "should account for each value given slice with a nil value",
			val: &objectGoSlice{
				baseObject: baseObject{
					val: &Object{runtime: vm},
				},
				data: &[]interface{}{nil},
			},
			// overhead + null
			expectedMem: SizeEmptyStruct + SizeEmptyStruct,
			// overhead + null
			expectedNewMem: SizeEmptyStruct + SizeEmptyStruct,
			errExpected:    nil,
		},
		{
			name: "should account for nested slices",
			val: &objectGoSlice{
				baseObject: baseObject{
					val: &Object{runtime: vm},
				},
				data: &[]interface{}{
					[]interface{}{
						valueInt(99),
						valueInt(99),
					},
				},
			},
			// overhead + (value + len("length") + "length".value + prototype + ints)
			expectedMem: SizeEmptyStruct + (SizeEmptyStruct + 6 + SizeEmptyStruct + (SizeEmptyStruct + SizeEmptyStruct) + SizeNumber*2),
			// overhead + (value + len("length") with string overhead + "length".value + prototype + ints)
			expectedNewMem: SizeEmptyStruct + (SizeEmptyStruct + (6 + SizeString) + SizeEmptyStruct + (SizeEmptyStruct + SizeEmptyStruct) + SizeNumber*2),
			errExpected:    nil,
		},
		{
			name: "should account for nested pointer slice",
			val: &objectGoSlice{
				baseObject: baseObject{
					val: &Object{runtime: vm},
				},
				data: &[]interface{}{
					&[]interface{}{
						valueInt(99),
						valueInt(99),
					},
				},
			},
			// overhead + (value + len("length") + "length".value + prototype + ints)
			expectedMem: SizeEmptyStruct + (SizeEmptyStruct + 6 + SizeEmptyStruct + (SizeEmptyStruct + SizeEmptyStruct) + SizeNumber*2),
			// overhead + (value + len("length") with string overhead + "length".value + prototype + ints)
			expectedNewMem: SizeEmptyStruct + (SizeEmptyStruct + (6 + SizeString) + SizeEmptyStruct + (SizeEmptyStruct + SizeEmptyStruct) + SizeNumber*2),
			errExpected:    nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			total, newTotal, err := tc.val.MemUsage(NewMemUsageContext(vm, 100, 100, 100, 100, nil))
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
