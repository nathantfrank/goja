package goja

import (
	"testing"
)

func TestGomapProp(t *testing.T) {
	const SCRIPT = `
	o.a + o.b;
	`
	r := New()
	r.Set("o", map[string]interface{}{
		"a": 40,
		"b": 2,
	})
	v, err := r.RunString(SCRIPT)
	if err != nil {
		t.Fatal(err)
	}
	if i := v.ToInteger(); i != 42 {
		t.Fatalf("Expected 42, got: %d", i)
	}
}

func TestGomapEnumerate(t *testing.T) {
	const SCRIPT = `
	var hasX = false;
	var hasY = false;
	for (var key in o) {
		switch (key) {
		case "x":
			if (hasX) {
				throw "Already have x";
			}
			hasX = true;
			break;
		case "y":
			if (hasY) {
				throw "Already have y";
			}
			hasY = true;
			break;
		default:
			throw "Unexpected property: " + key;
		}
	}
	hasX && hasY;
	`
	r := New()
	r.Set("o", map[string]interface{}{
		"x": 40,
		"y": 2,
	})
	v, err := r.RunString(SCRIPT)
	if err != nil {
		t.Fatal(err)
	}

	if !v.StrictEquals(valueTrue) {
		t.Fatalf("Expected true, got %v", v)
	}
}

func TestGomapDeleteWhileEnumerate(t *testing.T) {
	const SCRIPT = `
	var hasX = false;
	var hasY = false;
	for (var key in o) {
		switch (key) {
		case "x":
			if (hasX) {
				throw "Already have x";
			}
			hasX = true;
			delete o.y;
			break;
		case "y":
			if (hasY) {
				throw "Already have y";
			}
			hasY = true;
			delete o.x;
			break;
		default:
			throw "Unexpected property: " + key;
		}
	}
	hasX && !hasY || hasY && !hasX;
	`
	r := New()
	r.Set("o", map[string]interface{}{
		"x": 40,
		"y": 2,
	})
	v, err := r.RunString(SCRIPT)
	if err != nil {
		t.Fatal(err)
	}

	if !v.StrictEquals(valueTrue) {
		t.Fatalf("Expected true, got %v", v)
	}
}

func TestGomapInstanceOf(t *testing.T) {
	const SCRIPT = `
	(o instanceof Object) && !(o instanceof Error);
	`
	r := New()
	r.Set("o", map[string]interface{}{})
	v, err := r.RunString(SCRIPT)
	if err != nil {
		t.Fatal(err)
	}

	if !v.StrictEquals(valueTrue) {
		t.Fatalf("Expected true, got %v", v)
	}
}

func TestGomapTypeOf(t *testing.T) {
	const SCRIPT = `
	typeof o;
	`
	r := New()
	r.Set("o", map[string]interface{}{})
	v, err := r.RunString(SCRIPT)
	if err != nil {
		t.Fatal(err)
	}

	if !v.StrictEquals(asciiString("object")) {
		t.Fatalf("Expected object, got %v", v)
	}
}

func TestGomapProto(t *testing.T) {
	const SCRIPT = `
	o.hasOwnProperty("test");
	`
	r := New()
	r.Set("o", map[string]interface{}{
		"test": 42,
	})
	v, err := r.RunString(SCRIPT)
	if err != nil {
		t.Fatal(err)
	}

	if !v.StrictEquals(valueTrue) {
		t.Fatalf("Expected true, got %v", v)
	}
}

func TestGoMapExtensibility(t *testing.T) {
	const SCRIPT = `
	"use strict";
	o.test = 42;
	Object.preventExtensions(o);
	o.test = 43;
	try {
		o.test1 = 42;
	} catch (e) {
		if (!(e instanceof TypeError)) {
			throw e;
		}
	}
	o.test === 43 && o.test1 === undefined;
	`

	r := New()
	r.Set("o", map[string]interface{}{})
	v, err := r.RunString(SCRIPT)
	if err != nil {
		if ex, ok := err.(*Exception); ok {
			t.Fatal(ex.String())
		} else {
			t.Fatal(err)
		}
	}

	if !v.StrictEquals(valueTrue) {
		t.Fatalf("Expected true, got %v", v)
	}

}

func TestGoMapWithProto(t *testing.T) {
	vm := New()
	m := map[string]interface{}{
		"t": "42",
	}
	vm.Set("m", m)
	vm.testScriptWithTestLib(`
	(function() {
	'use strict';
	var proto = {};
	var getterAllowed = false;
	var setterAllowed = false;
	var tHolder = "proto t";
	Object.defineProperty(proto, "t", {
		get: function() {
			if (!getterAllowed) throw new Error("getter is called");
			return tHolder;
		},
		set: function(v) {
			if (!setterAllowed) throw new Error("setter is called");
			tHolder = v;
		}
	});
	var t1Holder;
	Object.defineProperty(proto, "t1", {
		get: function() {
			return t1Holder;
		},
		set: function(v) {
			t1Holder = v;
		}
	});
	Object.setPrototypeOf(m, proto);
	assert.sameValue(m.t, "42");
	m.t = 43;
	assert.sameValue(m.t, 43);
	t1Holder = "test";
	assert.sameValue(m.t1, "test");
	m.t1 = "test1";
	assert.sameValue(m.t1, "test1");
	delete m.t;
	getterAllowed = true;
	assert.sameValue(m.t, "proto t", "after delete");
	setterAllowed = true;
	m.t = true;
	assert.sameValue(m.t, true);
	assert.sameValue(tHolder, true);
	Object.preventExtensions(m);
	assert.throws(TypeError, function() {
		m.t2 = 1;
	});
	m.t1 = "test2";
	assert.sameValue(m.t1, "test2");
	})();
	`, _undefined, t)
}

func TestGoMapProtoProp(t *testing.T) {
	const SCRIPT = `
	(function() {
	"use strict";
	var proto = {};
	Object.defineProperty(proto, "ro", {value: 42});
	Object.setPrototypeOf(m, proto);
	assert.throws(TypeError, function() {
		m.ro = 43;
	});
	Object.defineProperty(m, "ro", {value: 43});
	assert.sameValue(m.ro, 43);
	})();
	`

	r := New()
	r.Set("m", map[string]interface{}{})
	r.testScriptWithTestLib(SCRIPT, _undefined, t)
}

func TestGoMapProtoPropChain(t *testing.T) {
	const SCRIPT = `
	(function() {
	"use strict";
	var p1 = Object.create(null);
	m.__proto__ = p1;
	
	Object.defineProperty(p1, "test", {
		value: 42
	});
	
	Object.defineProperty(m, "test", {
		value: 43,
		writable: true,
	});
	var o = Object.create(m);
	o.test = 44;
	assert.sameValue(o.test, 44);

	var sym = Symbol(true);
	Object.defineProperty(p1, sym, {
		value: 42
	});
	
	Object.defineProperty(m, sym, {
		value: 43,
		writable: true,
	});
	o[sym] = 44;
	assert.sameValue(o[sym], 44);
	})();
	`

	r := New()
	r.Set("m", map[string]interface{}{})
	r.testScriptWithTestLib(SCRIPT, _undefined, t)
}

func TestGoMapUnicode(t *testing.T) {
	const SCRIPT = `
	Object.setPrototypeOf(m, s);
	if (m.Тест !== "passed") {
		throw new Error("m.Тест: " + m.Тест);
	}
	m["é"];
	`
	type S struct {
		Тест string
	}
	vm := New()
	m := map[string]interface{}{
		"é": 42,
	}
	s := S{
		Тест: "passed",
	}
	vm.Set("m", m)
	vm.Set("s", &s)
	res, err := vm.RunString(SCRIPT)
	if err != nil {
		t.Fatal(err)
	}
	if res == nil || !res.StrictEquals(valueInt(42)) {
		t.Fatalf("Unexpected value: %v", res)
	}
}

func TestGoMapMemUsage(t *testing.T) {
	vm := New()
	vmCtx := NewMemUsageContext(vm, 100, 100, 100, 100, nil)

	nestedMap := map[string]interface{}{
		"subTest1": valueInt(99),
		"subTest2": valueInt(99),
	}

	// The baseObject is quite large when ToValue is called due to the initialization of
	// objectGoMapSimple. The init sets the object prototype with all its associated
	// functions and fields. Calculating ahead of time for test case
	nestedMapAsObject := vm.ToValue(nestedMap)
	nestedMapMemUsage, nestedMapNewMemUsage, err := nestedMapAsObject.MemUsage(vmCtx)
	if err != nil {
		t.Fatalf("Unexpected error. Actual: %v Expected: %v", err, nil)
	}

	tests := []struct {
		name           string
		val            *objectGoMapSimple
		expectedMem    uint64
		expectedNewMem uint64
		errExpected    error
	}{
		{
			name: "should account for each key value pair given a non-empty object",
			val: &objectGoMapSimple{
				baseObject: baseObject{
					val: &Object{runtime: vm},
				},
				data: map[string]interface{}{
					"test0": valueInt(99),
					"test1": valueInt(99),
				},
			},
			// baseObject overhead + len("testN") + value
			expectedMem: SizeEmptyStruct + (5+SizeInt)*2,
			// baseObject overhead + len("testN") with string overhead + value
			expectedNewMem: SizeEmptyStruct + ((5+SizeString)+SizeInt)*2,
			errExpected:    nil,
		},
		{
			name: "should account for each key value pair given a map with native ints",
			val: &objectGoMapSimple{
				baseObject: baseObject{
					val: &Object{runtime: vm},
				},
				data: map[string]interface{}{
					"test0": 99,
					"test1": 99,
				},
			},
			// baseObject overhead + len("testN") + value
			expectedMem: SizeEmptyStruct + (5+SizeInt)*2,
			// baseObject overhead + len("testN") with string overhead + value
			expectedNewMem: SizeEmptyStruct + ((5+SizeString)+SizeInt)*2,
			errExpected:    nil,
		},
		{
			name: "should account for each key value pair given map with a nil value",
			val: &objectGoMapSimple{
				baseObject: baseObject{
					val: &Object{runtime: vm},
				},
				data: map[string]interface{}{
					"test": nil,
				},
			},
			// overhead + len("test") + null
			expectedMem: SizeEmptyStruct + 4 + SizeEmptyStruct,
			// overhead + len("test") with string overhead + null
			expectedNewMem: SizeEmptyStruct + (4 + SizeString) + SizeEmptyStruct,
			errExpected:    nil,
		},
		{
			name: "should account for nested key value pairs",
			val: &objectGoMapSimple{
				baseObject: baseObject{
					val: &Object{runtime: vm},
				},
				data: map[string]interface{}{
					"test": nestedMap,
				},
			},
			// overhead + len("test") + (Object prototype + values)
			expectedMem: SizeEmptyStruct + 4 + nestedMapMemUsage,
			// overhead + len("testN") with string overhead + (Object prototype with overhead + values with string overhead)
			expectedNewMem: SizeEmptyStruct + (4 + SizeString) + nestedMapNewMemUsage,
			errExpected:    nil,
		},
		{
			name: "should account for nested pointer of key value pairs",
			val: &objectGoMapSimple{
				baseObject: baseObject{
					val: &Object{runtime: vm},
				},
				data: map[string]interface{}{
					"test": &nestedMap,
				},
			},
			// overhead + len("test") + nested overhead
			expectedMem: SizeEmptyStruct + 4 + SizeEmptyStruct,
			// overhead + len("testN") with string overhead + nested overhead
			expectedNewMem: SizeEmptyStruct + (4 + SizeString) + SizeEmptyStruct,
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
