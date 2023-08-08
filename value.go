package goja

import (
	"fmt"
	"hash/maphash"
	"math"
	"reflect"
	"strconv"
	"unsafe"

	"github.com/dop251/goja/ftoa"
	"github.com/dop251/goja/unistring"
)

var (
	// Not goroutine-safe, do not use for anything other than package level init
	pkgHasher maphash.Hash

	hashFalse = randomHash()
	hashTrue  = randomHash()
	hashNull  = randomHash()
	hashUndef = randomHash()
)

// Not goroutine-safe, do not use for anything other than package level init
func randomHash() uint64 {
	pkgHasher.WriteByte(0)
	return pkgHasher.Sum64()
}

var (
	valueFalse    Value = valueBool(false)
	valueTrue     Value = valueBool(true)
	_null         Value = valueNull{}
	_NaN          Value = valueFloat(math.NaN())
	_positiveInf  Value = valueFloat(math.Inf(+1))
	_negativeInf  Value = valueFloat(math.Inf(-1))
	_positiveZero Value = valueInt(0)
	negativeZero        = math.Float64frombits(0 | (1 << 63))
	_negativeZero Value = valueFloat(negativeZero)
	_epsilon            = valueFloat(2.2204460492503130808472633361816e-16)
	_undefined    Value = valueUndefined{}
)

var (
	reflectTypeInt64  = reflect.TypeOf(int64(0))
	reflectTypeInt32  = reflect.TypeOf(int32(0))
	reflectTypeUInt32 = reflect.TypeOf(uint32(0))
	reflectTypeInt    = reflect.TypeOf(0)
	reflectTypeBool   = reflect.TypeOf(false)
	reflectTypeNil    = reflect.TypeOf(nil)
	reflectTypeFloat  = reflect.TypeOf(float64(0))
	reflectTypeMap    = reflect.TypeOf(map[string]interface{}{})
	reflectTypeArray  = reflect.TypeOf([]interface{}{})
	reflectTypeString = reflect.TypeOf("")
	reflectTypeFunc   = reflect.TypeOf((func(FunctionCall) Value)(nil))
	reflectTypeError  = reflect.TypeOf((*error)(nil)).Elem()
)

var intCache [256]Value
var int64Cache [256]Value

func FalseValue() Value {
	return valueFalse
}
func TrueValue() Value {
	return valueTrue
}

// Value represents an ECMAScript value.
//
// Export returns a "plain" Go value which type depends on the type of the Value.
//
// For integer numbers it's int64.
//
// For any other numbers (including Infinities, NaN and negative zero) it's float64.
//
// For string it's a string. Note that unicode strings are converted into UTF-8 with invalid code points replaced with utf8.RuneError.
//
// For boolean it's bool.
//
// For null and undefined it's nil.
//
// For Object it depends on the Object type, see Object.Export() for more details.
type Value interface {
	ToInt() int
	ToUInt32() uint32
	ToInt32() int32
	ToInt64() int64

	ToInteger() int64
	toString() valueString
	string() unistring.String
	ToString() Value
	String() string
	ToFloat() float64
	ToNumber() Value
	ToBoolean() bool
	ToObject(*Runtime) *Object
	SameAs(Value) bool
	Equals(Value) bool
	StrictEquals(Value) bool
	Export() interface{}
	ExportType() reflect.Type

	IsObject() bool
	IsNumber() bool

	assertInt() (int, bool)
	assertUInt32() (uint32, bool)
	assertInt32() (int32, bool)
	assertInt64() (int64, bool)
	assertString() (valueString, bool)
	assertFloat() (float64, bool)

	baseObject(r *Runtime) *Object

	hash(hasher *maphash.Hash) uint64

	MemUsage(ctx *MemUsageContext) (memUsage uint64, newMemUsage uint64, err error)
}

const (
	SizeBool        = uint64(unsafe.Sizeof(true))
	SizeNumber      = uint64(unsafe.Sizeof(float64(0)))
	SizeInt32       = uint64(unsafe.Sizeof(int32(0)))
	SizeInt         = uint64(unsafe.Sizeof(int(0)))
	SizeEmptyStruct = uint64(unsafe.Sizeof((*baseObject)(nil)))
	SizeEmptySlice  = uint64(unsafe.Sizeof([]Value{}))
	// SizeString allows us to take into account the 16 additional bytes
	// for any string type in go including the pointer to the start of
	// the string data and the length of the string
	SizeString = uint64(unsafe.Sizeof(""))
)

type valueContainer interface {
	toValue(*Runtime) Value
}

type typeError string
type rangeError string
type referenceError string
type syntaxError string

type valueNumber struct {
	_type reflect.Type
	val   interface{}
}

type valueInt int
type valueUInt32 uint32
type valueInt32 int32
type valueInt64 int64
type valueFloat float64
type valueBool bool
type valueNull struct{}

func UndefinedValue() Value {
	return valueUndefined{}
}

type valueUndefined struct {
	valueNull
}

// *Symbol is a Value containing ECMAScript Symbol primitive. Symbols must only be created
// using NewSymbol(). Zero values and copying of values (i.e. *s1 = *s2) are not permitted.
// Well-known Symbols can be accessed using Sym* package variables (SymIterator, etc...)
// Symbols can be shared by multiple Runtimes.
type Symbol struct {
	h    uintptr
	desc valueString
}

type valueUnresolved struct {
	r   *Runtime
	ref unistring.String
}

type memberUnresolved struct {
	valueUnresolved
}

type valueProperty struct {
	value        Value
	writable     bool
	configurable bool
	enumerable   bool
	accessor     bool
	getterFunc   *Object
	setterFunc   *Object
}

var (
	errAccessBeforeInit = referenceError("Cannot access a variable before initialization")
	errAssignToConst    = typeError("Assignment to constant variable.")
)

func propGetter(o Value, v Value, r *Runtime) *Object {
	if v == _undefined {
		return nil
	}
	if obj, ok := v.(*Object); ok {
		if _, ok := obj.self.assertCallable(); ok {
			return obj
		}
	}
	r.typeErrorResult(true, "Getter must be a function: %s", v.toString())
	return nil
}

func propSetter(o Value, v Value, r *Runtime) *Object {
	if v == _undefined {
		return nil
	}
	if obj, ok := v.(*Object); ok {
		if _, ok := obj.self.assertCallable(); ok {
			return obj
		}
	}
	r.typeErrorResult(true, "Setter must be a function: %s", v.toString())
	return nil
}

func fToStr(num float64, mode ftoa.FToStrMode, prec int) string {
	var buf1 [128]byte
	return string(ftoa.FToStr(num, mode, prec, buf1[:0]))
}

func (i valueInt) MemUsage(ctx *MemUsageContext) (memUsage uint64, newMemUsage uint64, err error) {
	return SizeNumber, SizeNumber, nil
}

func (i valueInt) assertInt() (int, bool) {
	return int(i), true
}
func (i valueInt) assertUInt32() (uint32, bool) {
	return 0, false
}
func (i valueInt) assertInt32() (int32, bool) {
	return 0, false
}
func (i valueInt) assertInt64() (int64, bool) {
	return int64(i), true
}

func (i valueInt) assertFloat() (float64, bool) {
	return 0, false
}

func (i valueInt) assertString() (valueString, bool) {
	return nil, false
}

func (i valueInt) ToInteger() int64 {
	return int64(i)
}
func (i valueInt) ToInt() int {
	return int(i)
}
func (i valueInt) ToInt32() int32 {
	return int32(i)
}
func (i valueInt) ToUInt32() uint32 {
	return uint32(i)
}

func (i valueInt) ToInt64() int64 {
	return int64(i)
}

func (i valueInt) IsNumber() bool {
	return true
}
func (i valueInt) IsObject() bool {
	return false
}

func (i valueInt) toString() valueString {
	return asciiString(i.String())
}

func (i valueInt) string() unistring.String {
	return unistring.String(i.String())
}

func (i valueInt) ToString() Value {
	return i
}

func (i valueInt) String() string {
	return strconv.FormatInt(int64(i), 10)
}

func (i valueInt) ToFloat() float64 {
	return float64(i)
}

func (i valueInt) ToBoolean() bool {
	return i != 0
}

func (i valueInt) ToObject(r *Runtime) *Object {
	return r.newPrimitiveObject(i, r.global.NumberPrototype, classNumber)
}

func (i valueInt) ToNumber() Value {
	return i
}

func (i valueInt) SameAs(other Value) bool {
	return i == other
}

func (i valueInt) Equals(other Value) bool {
	switch o := other.(type) {
	case valueInt:
		return i == o
	case valueInt64:
		return i == valueInt(o)
	case valueFloat:
		return float64(i) == float64(o)
	case valueString:
		return o.ToNumber().Equals(i)
	case valueBool:
		return int64(i) == o.ToInteger()
	case *Object:
		return i.Equals(o.toPrimitive())
	}

	return false
}

func (i valueInt) StrictEquals(other Value) bool {
	switch o := other.(type) {
	case valueInt:
		return i == o
	case valueInt64:
		return i == valueInt(o)
	case valueFloat:
		return float64(i) == float64(o)
	}

	return false
}

func (i valueInt) baseObject(r *Runtime) *Object {
	return r.global.NumberPrototype
}

func (i valueInt) Export() interface{} {
	return int64(i)
}

func (i valueInt) ExportType() reflect.Type {
	return reflect.TypeOf(i)
}

func (i valueInt) hash(*maphash.Hash) uint64 {
	return uint64(i)
}

func (o valueBool) MemUsage(ctx *MemUsageContext) (memUsage uint64, newMemUsage uint64, err error) {
	return SizeBool, SizeBool, nil
}

func (b valueBool) ToInt() int {
	if b {
		return 1
	}
	return 0
}
func (b valueBool) ToInt32() int32 {
	if b {
		return 1
	}
	return 0
}
func (b valueBool) ToInt64() int64 {
	if b {
		return 1
	}
	return 0
}
func (b valueBool) ToUInt32() uint32 {
	if b {
		return 1
	}
	return 0
}
func (b valueBool) assertFloat() (float64, bool) {
	return 0, false
}
func (b valueBool) assertString() (valueString, bool) {
	return nil, false
}
func (o valueBool) assertInt() (int, bool) {
	return 0, false
}

func (o valueBool) assertInt32() (int32, bool) {
	return 0, false
}
func (o valueBool) assertUInt32() (uint32, bool) {
	return 0, false
}

func (o valueBool) assertInt64() (int64, bool) {
	return 0, false
}

func (b valueBool) ToInteger() int64 {
	if b {
		return 1
	}
	return 0
}
func (b valueBool) IsNumber() bool {
	return false
}
func (b valueBool) IsObject() bool {
	return false
}

func (b valueBool) toString() valueString {
	if b {
		return stringTrue
	}
	return stringFalse
}

func (b valueBool) ToString() Value {
	return b
}

func (b valueBool) String() string {
	if b {
		return "true"
	}
	return "false"
}

func (b valueBool) string() unistring.String {
	return unistring.String(b.String())
}

func (b valueBool) ToFloat() float64 {
	if b {
		return 1.0
	}
	return 0
}

func (b valueBool) ToBoolean() bool {
	return bool(b)
}

func (b valueBool) ToObject(r *Runtime) *Object {
	return r.newPrimitiveObject(b, r.global.BooleanPrototype, "Boolean")
}

func (b valueBool) ToNumber() Value {
	if b {
		return valueInt(1)
	}
	return valueInt(0)
}

func (b valueBool) SameAs(other Value) bool {
	if other, ok := other.(valueBool); ok {
		return b == other
	}
	return false
}

func (b valueBool) Equals(other Value) bool {
	if o, ok := other.(valueBool); ok {
		return b == o
	}

	if b {
		return other.Equals(intToValue(1))
	} else {
		return other.Equals(intToValue(0))
	}

}

func (b valueBool) StrictEquals(other Value) bool {
	if other, ok := other.(valueBool); ok {
		return b == other
	}
	return false
}

func (b valueBool) baseObject(r *Runtime) *Object {
	return r.global.BooleanPrototype
}

func (b valueBool) Export() interface{} {
	return bool(b)
}

func (b valueBool) ExportType() reflect.Type {
	return reflectTypeBool
}

func (b valueBool) hash(*maphash.Hash) uint64 {
	if b {
		return hashTrue
	}

	return hashFalse
}

func (n valueNull) assertInt() (int, bool) {
	return 0, false
}
func (n valueNull) assertFloat() (float64, bool) {
	return 0, false
}
func (n valueNull) assertString() (valueString, bool) {
	return nil, false
}
func (n valueNull) assertUInt32() (uint32, bool) {
	return 0, false
}
func (n valueNull) assertInt32() (int32, bool) {
	return 0, false
}
func (n valueNull) assertInt64() (int64, bool) {
	return 0, false
}

func (n valueNull) ToInt32() int32 {
	return 0
}
func (n valueNull) ToUInt32() uint32 {
	return 0
}
func (n valueNull) ToInt64() int64 {
	return 0
}

func (n valueNull) ToInt() int {
	return 0
}

func (n valueNull) ToInteger() int64 {
	return 0
}

func (n valueNull) toString() valueString {
	return stringNull
}

func (n valueNull) string() unistring.String {
	return stringNull.string()
}

func (n valueNull) ToString() Value {
	return n
}

func (n valueNull) String() string {
	return "null"
}
func (n valueNull) IsNumber() bool {
	return false
}
func (n valueNull) IsObject() bool {
	return false
}

func (n valueNull) MemUsage(ctx *MemUsageContext) (memUsage uint64, newMemUsage uint64, err error) {
	return SizeEmptyStruct, SizeEmptyStruct, nil
}

func (u valueUndefined) toString() valueString {
	return stringUndefined
}

func (u valueUndefined) ToString() Value {
	return u
}

func (u valueUndefined) String() string {
	return "undefined"
}

func (u valueUndefined) string() unistring.String {
	return "undefined"
}

func (u valueUndefined) ToNumber() Value {
	return _NaN
}

func (u valueUndefined) SameAs(other Value) bool {
	_, same := other.(valueUndefined)
	return same
}

func (u valueUndefined) StrictEquals(other Value) bool {
	_, same := other.(valueUndefined)
	return same
}

func (u valueUndefined) ToFloat() float64 {
	return math.NaN()
}

func (u valueUndefined) hash(*maphash.Hash) uint64 {
	return hashUndef
}

func (n valueNull) ToFloat() float64 {
	return 0
}

func (n valueNull) ToBoolean() bool {
	return false
}

func (n valueNull) ToObject(r *Runtime) *Object {
	r.typeErrorResult(true, "Cannot convert undefined or null to object")
	return nil
	//return r.newObject()
}

func (n valueNull) ToNumber() Value {
	return intToValue(0)
}

func (n valueNull) SameAs(other Value) bool {
	_, same := other.(valueNull)
	return same
}

func (n valueNull) Equals(other Value) bool {
	switch other.(type) {
	case valueUndefined, valueNull:
		return true
	}
	return false
}

func (n valueNull) StrictEquals(other Value) bool {
	_, same := other.(valueNull)
	return same
}

func (n valueNull) baseObject(*Runtime) *Object {
	return nil
}

func (n valueNull) Export() interface{} {
	return nil
}

func (n valueNull) ExportType() reflect.Type {
	return reflectTypeNil
}

func (n valueNull) hash(*maphash.Hash) uint64 {
	return hashNull
}

func (p *valueProperty) ToInteger() int64 {
	return 0
}

func (p *valueProperty) toString() valueString {
	return stringEmpty
}

func (p *valueProperty) string() unistring.String {
	return ""
}

func (p *valueProperty) ToString() Value {
	return _undefined
}

func (p *valueProperty) String() string {
	return ""
}

func (p *valueProperty) ToFloat() float64 {
	return math.NaN()
}

func (p *valueProperty) ToBoolean() bool {
	return false
}

func (p *valueProperty) ToObject(*Runtime) *Object {
	return nil
}

func (p *valueProperty) ToNumber() Value {
	return nil
}

func (p *valueProperty) isWritable() bool {
	return p.writable || p.setterFunc != nil
}

func (p *valueProperty) get(this Value) Value {
	if p.getterFunc == nil {
		if p.value != nil {
			return p.value
		}
		return _undefined
	}
	call, _ := p.getterFunc.self.assertCallable()
	return call(FunctionCall{
		ctx:  p.getterFunc.runtime.vm.ctx,
		This: this,
	})
}

func (p *valueProperty) set(this, v Value) {
	if p.setterFunc == nil {
		p.value = v
		return
	}
	call, _ := p.setterFunc.self.assertCallable()
	call(FunctionCall{
		ctx:       p.setterFunc.runtime.vm.ctx,
		This:      this,
		Arguments: []Value{v},
	})
}

func (p *valueProperty) SameAs(other Value) bool {
	if otherProp, ok := other.(*valueProperty); ok {
		return p == otherProp
	}
	return false
}

func (p *valueProperty) Equals(Value) bool {
	return false
}

func (p *valueProperty) StrictEquals(Value) bool {
	return false
}

func (p *valueProperty) baseObject(r *Runtime) *Object {
	r.typeErrorResult(true, "BUG: baseObject() is called on valueProperty") // TODO error message
	return nil
}

func (p *valueProperty) Export() interface{} {
	panic("Cannot export valueProperty")
}

func (p *valueProperty) ExportType() reflect.Type {
	panic("Cannot export valueProperty")
}

func (p *valueProperty) hash(*maphash.Hash) uint64 {
	panic("valueProperty should never be used in maps or sets")
}

func (p *valueProperty) MemUsage(ctx *MemUsageContext) (memUsage uint64, newMemUsage uint64, err error) {
	if p == nil {
		return SizeEmptyStruct, SizeEmptyStruct, err
	}

	if p.value != nil {
		inc, newInc, err := p.value.MemUsage(ctx)
		memUsage += inc
		newMemUsage += newInc
		if err != nil {
			return memUsage, newMemUsage, err
		}
	}

	if p.getterFunc != nil {
		inc, newInc, err := p.getterFunc.MemUsage(ctx)
		memUsage += inc
		newMemUsage += newInc
		if err != nil {
			return memUsage, newMemUsage, err
		}
	}

	if p.setterFunc != nil {
		inc, newInc, err := p.setterFunc.MemUsage(ctx)
		memUsage += inc
		newMemUsage += newInc
		if err != nil {
			return memUsage, newMemUsage, err
		}
	}

	return memUsage, newMemUsage, nil
}

func (p *valueProperty) ToInt() int {
	return 0
}

func (p *valueProperty) ToInt32() int32 {
	return 0
}
func (p *valueProperty) ToUInt32() uint32 {
	return 0
}

func (p *valueProperty) ToInt64() int64 {
	return 0
}

func (p *valueProperty) IsObject() bool {
	return false
}
func (p *valueProperty) IsNumber() bool {
	return false
}

func (p *valueProperty) assertFloat() (float64, bool) {
	return 0, false
}
func (p *valueProperty) assertString() (valueString, bool) {
	return nil, false
}

func (p *valueProperty) assertInt() (int, bool) {
	return 0, false
}
func (p *valueProperty) assertUInt32() (uint32, bool) {
	return 0, false
}
func (p *valueProperty) assertInt32() (int32, bool) {
	return 0, false
}
func (p *valueProperty) assertInt64() (int64, bool) {
	return 0, false
}

func floatToIntClip(n float64) int64 {
	switch {
	case math.IsNaN(n):
		return 0
	case n >= math.MaxInt64:
		return math.MaxInt64
	case n <= math.MinInt64:
		return math.MinInt64
	}
	return int64(n)
}

func (f valueFloat) MemUsage(ctx *MemUsageContext) (memUsage uint64, newMemUsage uint64, err error) {
	return SizeNumber, SizeNumber, nil
}

func (f valueFloat) ToInt() int {
	switch {
	case math.IsNaN(float64(f)):
		return 0
	case math.IsInf(float64(f), 1):
		return math.MaxInt64
	case math.IsInf(float64(f), -1):
		return math.MinInt64
	}
	return int(f)
}
func (f valueFloat) ToInt32() int32 {
	switch {
	case math.IsNaN(float64(f)):
		return 0
	case math.IsInf(float64(f), 1):
		return int32(math.MaxInt32)
	case math.IsInf(float64(f), -1):
		return int32(math.MinInt32)
	}
	return int32(f)
}
func (f valueFloat) ToInt64() int64 {
	switch {
	case math.IsNaN(float64(f)):
		return 0
	case math.IsInf(float64(f), 1):
		return int64(math.MaxInt64)
	case math.IsInf(float64(f), -1):
		return int64(math.MinInt64)
	}
	return int64(f)
}
func (f valueFloat) ToUInt32() uint32 {
	switch {
	case math.IsNaN(float64(f)):
		return 0
	case math.IsInf(float64(f), 1):
		return uint32(math.MaxInt32)
	case math.IsInf(float64(f), -1):
		return 0
	}
	return uint32(f)
}
func (f valueFloat) ToInteger() int64 {
	return floatToIntClip(float64(f))
}
func (f valueFloat) IsNumber() bool {
	return true
}
func (f valueFloat) IsObject() bool {
	return false
}

func (f valueFloat) toString() valueString {
	return asciiString(f.String())
}

func (f valueFloat) string() unistring.String {
	return unistring.String(f.String())
}

func (f valueFloat) ToString() Value {
	return asciiString(f.String())
}

func (f valueFloat) String() string {
	return fToStr(float64(f), ftoa.ModeStandard, 0)
}

func (f valueFloat) ToFloat() float64 {
	return float64(f)
}

func (f valueFloat) ToBoolean() bool {
	return float64(f) != 0.0 && !math.IsNaN(float64(f))
}

func (f valueFloat) ToObject(r *Runtime) *Object {
	return r.newPrimitiveObject(f, r.global.NumberPrototype, "Number")
}

func (f valueFloat) ToNumber() Value {
	return f
}
func (i valueFloat) assertString() (valueString, bool) {
	return nil, false
}
func (f valueFloat) assertInt() (int, bool) {
	return 0, false
}
func (f valueFloat) assertUInt32() (uint32, bool) {
	return 0, false
}
func (f valueFloat) assertInt32() (int32, bool) {
	return 0, false
}
func (f valueFloat) assertInt64() (int64, bool) {
	return 0, false
}

func (f valueFloat) assertFloat() (float64, bool) {
	return float64(f), true
}

func (f valueFloat) SameAs(other Value) bool {
	switch o := other.(type) {
	case valueFloat:
		this := float64(f)
		o1 := float64(o)
		if math.IsNaN(this) && math.IsNaN(o1) {
			return true
		} else {
			ret := this == o1
			if ret && this == 0 {
				ret = math.Signbit(this) == math.Signbit(o1)
			}
			return ret
		}
	case valueInt:
		this := float64(f)
		ret := this == float64(o)
		if ret && this == 0 {
			ret = !math.Signbit(this)
		}
		return ret
	case valueInt64:
		this := float64(f)
		ret := this == float64(o)
		if ret && this == 0 {
			ret = !math.Signbit(this)
		}
		return ret
	}

	return false
}

func (f valueFloat) Equals(other Value) bool {
	switch o := other.(type) {
	case valueFloat:
		return f == o
	case valueInt:
		return float64(f) == float64(o)
	case valueInt64:
		return float64(f) == float64(o)
	case valueString, valueBool:
		return float64(f) == o.ToFloat()
	case *Object:
		return f.Equals(o.toPrimitive())
	}

	return false
}

func (f valueFloat) StrictEquals(other Value) bool {
	switch o := other.(type) {
	case valueFloat:
		return f == o
	case valueInt:
		return float64(f) == float64(o)
	case valueInt64:
		return float64(f) == float64(o)
	}

	return false
}

func (f valueFloat) baseObject(r *Runtime) *Object {
	return r.global.NumberPrototype
}

func (f valueFloat) Export() interface{} {
	return float64(f)
}

func (f valueFloat) ExportType() reflect.Type {
	return reflectTypeFloat
}

func (f valueFloat) hash(*maphash.Hash) uint64 {
	if f == _negativeZero {
		return 0
	}
	return math.Float64bits(float64(f))
}

func (o *Object) ToInteger() int64 {
	return o.toPrimitiveNumber().ToNumber().ToInteger()
}

func (o *Object) toString() valueString {
	return o.toPrimitiveString().toString()
}
func (o *Object) IsNumber() bool {
	return false
}
func (o *Object) IsObject() bool {
	return true
}

func (o *Object) string() unistring.String {
	return o.toPrimitiveString().string()
}

func (o *Object) ToString() Value {
	return o.toPrimitiveString().ToString()
}

func (o *Object) String() string {
	return o.toPrimitiveString().String()
}

func (o *Object) ToFloat() float64 {
	return o.toPrimitiveNumber().ToFloat()
}

func (o *Object) ToBoolean() bool {
	return true
}

func (o *Object) ToObject(*Runtime) *Object {
	return o
}

func (o *Object) ToNumber() Value {
	return o.toPrimitiveNumber().ToNumber()
}

func (o *Object) SameAs(other Value) bool {
	if other, ok := other.(*Object); ok {
		return o == other
	}
	return false
}

func (o *Object) Equals(other Value) bool {
	if other, ok := other.(*Object); ok {
		return o == other || o.self.equal(other.self)
	}

	switch o1 := other.(type) {
	case valueInt, valueFloat, valueString, valueInt64, *Symbol:
		return o.toPrimitive().Equals(other)
	case valueBool:
		return o.Equals(o1.ToNumber())
	}

	return false
}

func (o *Object) StrictEquals(other Value) bool {
	if other, ok := other.(*Object); ok {
		return o == other || o.self.equal(other.self)
	}
	return false
}

func (o *Object) baseObject(*Runtime) *Object {
	return o
}

// Export the Object to a plain Go type.
// If the Object is a wrapped Go value (created using ToValue()) returns the original value.
//
// If the Object is a function, returns func(FunctionCall) Value. Note that exceptions thrown inside the function
// result in panics, which can also leave the Runtime in an unusable state. Therefore, these values should only
// be used inside another ES function implemented in Go. For calling a function from Go, use AssertFunction() or
// Runtime.ExportTo() as described in the README.
//
// For a Map, returns the list of entries as [][2]interface{}.
//
// For a Set, returns the list of elements as []interface{}.
//
// For a Proxy, returns Proxy.
//
// For a Promise, returns Promise.
//
// For a DynamicObject or a DynamicArray, returns the underlying handler.
//
// For an array, returns its items as []interface{}.
//
// In all other cases returns own enumerable non-symbol properties as map[string]interface{}.
//
// This method will panic with an *Exception if a JavaScript exception is thrown in the process.
func (o *Object) Export() (ret interface{}) {
	if o.__wrapped != nil {
		return o.__wrapped
	}

	o.runtime.tryPanic(func() {
		ret = o.self.export(&objectExportCtx{})
	})

	return
}

// ExportType returns the type of the value that is returned by Export().
func (o *Object) ExportType() reflect.Type {
	return o.self.exportType()
}

func (o *Object) hash(*maphash.Hash) uint64 {
	return o.getId()
}

func (o *Object) MemUsage(ctx *MemUsageContext) (memUsage uint64, newMemUsage uint64, err error) {
	if o == nil || o.self == nil {
		return SizeEmptyStruct, SizeEmptyStruct, nil
	}

	if o.__wrapped != nil {
		if nativeMem, ok := ctx.NativeMemUsage(o.__wrapped); ok {
			return nativeMem, nativeMem, nil
		}
	}

	switch x := o.self.(type) {
	case *objectGoReflect:
		return SizeEmptyStruct, SizeEmptyStruct, nil
	case *objectGoMapReflect:
		return SizeEmptyStruct, SizeEmptyStruct, nil
	case *objectGoMapSimple:
		return SizeEmptyStruct, SizeEmptyStruct, nil
	case *objectGoSlice:
		return SizeEmptyStruct, SizeEmptyStruct, nil
	case *objectGoSliceReflect:
		return SizeEmptyStruct, SizeEmptyStruct, nil
	default:
		r, ok := x.(MemUsageReporter)
		if !ok {
			return 0, 0, nil
		}
		return r.MemUsage(ctx)
	}
}

func (o *Object) ToInt() int {
	return o.self.toPrimitiveNumber().ToNumber().ToInt()
}
func (o *Object) ToInt32() int32 {
	return o.self.toPrimitiveNumber().ToNumber().ToInt32()
}
func (o *Object) ToUInt32() uint32 {
	return o.self.toPrimitiveNumber().ToNumber().ToUInt32()
}
func (o *Object) ToInt64() int64 {
	return o.self.toPrimitiveNumber().ToNumber().ToInt64()
}

func (o *Object) assertInt() (int, bool) {
	return 0, false
}
func (o *Object) assertString() (valueString, bool) {
	return nil, false
}
func (o *Object) assertFloat() (float64, bool) {
	return 0, false
}
func (o *Object) assertUInt32() (uint32, bool) {
	return 0, false
}
func (o *Object) assertInt32() (int32, bool) {
	return 0, false
}
func (o *Object) assertInt64() (int64, bool) {
	return 0, false
}

func (o *Object) Get(name string) Value {
	return o.self.getStr(unistring.NewFromString(name), nil)
}

// GetSymbol returns the value of a symbol property. Use one of the Sym* values for well-known
// symbols (such as SymIterator, SymToStringTag, etc...).
// This method will panic with an *Exception if a JavaScript exception is thrown in the process.
func (o *Object) GetSymbol(sym *Symbol) Value {
	return o.self.getSym(sym, nil)
}

// Keys returns a list of Object's enumerable keys.
// This method will panic with an *Exception if a JavaScript exception is thrown in the process.
func (o *Object) Keys() (keys []string) {
	iter := &enumerableIter{
		o:       o,
		wrapped: o.self.iterateStringKeys(),
	}
	for item, next := iter.next(); next != nil; item, next = next() {
		keys = append(keys, item.name.String())
	}

	return
}

// Symbols returns a list of Object's enumerable symbol properties.
// This method will panic with an *Exception if a JavaScript exception is thrown in the process.
func (o *Object) Symbols() []*Symbol {
	symbols := o.self.symbols(false, nil)
	ret := make([]*Symbol, len(symbols))
	for i, sym := range symbols {
		ret[i], _ = sym.(*Symbol)
	}
	return ret
}

// DefineDataProperty is a Go equivalent of Object.defineProperty(o, name, {value: value, writable: writable,
// configurable: configurable, enumerable: enumerable})
func (o *Object) DefineDataProperty(name string, value Value, writable, configurable, enumerable Flag) error {
	return o.runtime.tryFunc(func() {
		o.self.defineOwnPropertyStr(unistring.NewFromString(name), PropertyDescriptor{
			Value:        value,
			Writable:     writable,
			Configurable: configurable,
			Enumerable:   enumerable,
		}, true)
	})
}

// DefineAccessorProperty is a Go equivalent of Object.defineProperty(o, name, {get: getter, set: setter,
// configurable: configurable, enumerable: enumerable})
func (o *Object) DefineAccessorProperty(name string, getter, setter Value, configurable, enumerable Flag) error {
	return o.runtime.tryFunc(func() {
		o.self.defineOwnPropertyStr(unistring.NewFromString(name), PropertyDescriptor{
			Getter:       getter,
			Setter:       setter,
			Configurable: configurable,
			Enumerable:   enumerable,
		}, true)
	})
}

// DefineDataPropertySymbol is a Go equivalent of Object.defineProperty(o, name, {value: value, writable: writable,
// configurable: configurable, enumerable: enumerable})
func (o *Object) DefineDataPropertySymbol(name *Symbol, value Value, writable, configurable, enumerable Flag) error {
	return o.runtime.tryFunc(func() {
		o.self.defineOwnPropertySym(name, PropertyDescriptor{
			Value:        value,
			Writable:     writable,
			Configurable: configurable,
			Enumerable:   enumerable,
		}, true)
	})
}

// DefineAccessorPropertySymbol is a Go equivalent of Object.defineProperty(o, name, {get: getter, set: setter,
// configurable: configurable, enumerable: enumerable})
func (o *Object) DefineAccessorPropertySymbol(name *Symbol, getter, setter Value, configurable, enumerable Flag) error {
	return o.runtime.tryFunc(func() {
		o.self.defineOwnPropertySym(name, PropertyDescriptor{
			Getter:       getter,
			Setter:       setter,
			Configurable: configurable,
			Enumerable:   enumerable,
		}, true)
	})
}

func (o *Object) Set(name string, value interface{}) error {
	return o.runtime.tryFunc(func() {
		o.self.setOwnStr(unistring.NewFromString(name), o.runtime.ToValue(value), true)
	})
}

func (o *Object) SetSymbol(name *Symbol, value interface{}) error {
	return o.runtime.tryFunc(func() {
		o.self.setOwnSym(name, o.runtime.ToValue(value), true)
	})
}

func (o *Object) Delete(name string) error {
	return o.runtime.tryFunc(func() {
		o.self.deleteStr(unistring.NewFromString(name), true)
	})
}

func (o *Object) DeleteSymbol(name *Symbol) error {
	return o.runtime.tryFunc(func() {
		o.self.deleteSym(name, true)
	})
}

// Prototype returns the Object's prototype, same as Object.getPrototypeOf(). If the prototype is null
// returns nil.
func (o *Object) Prototype() *Object {
	return o.self.proto()
}

// SetPrototype sets the Object's prototype, same as Object.setPrototypeOf(). Setting proto to nil
// is an equivalent of Object.setPrototypeOf(null).
func (o *Object) SetPrototype(proto *Object) error {
	return o.runtime.tryFunc(func() {
		o.self.setProto(proto, true)
	})
}

// MarshalJSON returns JSON representation of the Object. It is equivalent to JSON.stringify(o).
// Note, this implements json.Marshaler so that json.Marshal() can be used without the need to Export().
func (o *Object) MarshalJSON() ([]byte, error) {
	ctx := _builtinJSON_stringifyContext{
		r: o.runtime,
	}
	ex := o.runtime.vm.try(o.runtime.ctx, func() {
		if !ctx.do(o) {
			ctx.buf.WriteString("null")
		}
	})
	if ex != nil {
		return nil, ex
	}
	return ctx.buf.Bytes(), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface. It is added to compliment MarshalJSON, because
// some alternative JSON encoders refuse to use MarshalJSON unless UnmarshalJSON is also present.
// It is a no-op and always returns nil.
func (o *Object) UnmarshalJSON([]byte) error {
	return nil
}

// Class returns the class name (otto compatibility)
func (o *Object) Class() string {
	return o.self.className()
}

// ClassName returns the class name
func (o *Object) ClassName() string {
	return o.self.className()
}

func (o valueUnresolved) throw() {
	o.r.throwReferenceError(o.ref)
}

func (o valueUnresolved) ToInteger() int64 {
	o.throw()
	return 0
}
func (o valueUnresolved) IsNumber() bool {
	return false
}
func (o valueUnresolved) IsObject() bool {
	return false
}

func (o valueUnresolved) toString() valueString {
	o.throw()
	return nil
}

func (o valueUnresolved) string() unistring.String {
	o.throw()
	return ""
}

func (o valueUnresolved) ToString() Value {
	o.throw()
	return nil
}

func (o valueUnresolved) String() string {
	o.throw()
	return ""
}

func (o valueUnresolved) ToFloat() float64 {
	o.throw()
	return 0
}

func (o valueUnresolved) ToBoolean() bool {
	o.throw()
	return false
}

func (o valueUnresolved) ToObject(*Runtime) *Object {
	o.throw()
	return nil
}

func (o valueUnresolved) ToNumber() Value {
	o.throw()
	return nil
}

func (o valueUnresolved) SameAs(Value) bool {
	o.throw()
	return false
}

func (o valueUnresolved) Equals(Value) bool {
	o.throw()
	return false
}

func (o valueUnresolved) StrictEquals(Value) bool {
	o.throw()
	return false
}

func (o valueUnresolved) baseObject(*Runtime) *Object {
	o.throw()
	return nil
}

func (o valueUnresolved) Export() interface{} {
	o.throw()
	return nil
}

func (o valueUnresolved) ExportType() reflect.Type {
	o.throw()
	return nil
}

func (o valueUnresolved) hash(*maphash.Hash) uint64 {
	o.throw()
	return 0
}

func (o valueUnresolved) MemUsage(ctx *MemUsageContext) (memUsage uint64, newMemUsage uint64, err error) {
	return uint64(len(o.ref)), uint64(len(o.ref)) + SizeString, nil
}

func (o valueUnresolved) ToInt() int {
	o.throw()
	return 0
}
func (o valueUnresolved) ToInt32() int32 {
	o.throw()
	return 0
}
func (o valueUnresolved) ToUInt32() uint32 {
	o.throw()
	return 0
}
func (o valueUnresolved) ToInt64() int64 {
	o.throw()
	return 0
}

func (o valueUnresolved) assertInt() (int, bool) {
	o.throw()
	return 0, false
}
func (o valueUnresolved) assertUInt32() (uint32, bool) {
	o.throw()
	return 0, false
}
func (o valueUnresolved) assertInt32() (int32, bool) {
	o.throw()
	return 0, false
}
func (o valueUnresolved) assertInt64() (int64, bool) {
	o.throw()
	return 0, false
}
func (o valueUnresolved) assertFloat() (float64, bool) {
	o.throw()
	return 0, false
}
func (o valueUnresolved) assertString() (valueString, bool) {
	o.throw()
	return nil, false
}

func (s *Symbol) IsNumber() bool {
	return false
}

func (s *Symbol) IsObject() bool {
	return false
}
func (s *Symbol) ToInteger() int64 {
	panic(typeError("Cannot convert a Symbol value to a number"))
}
func (s *Symbol) ToInt() int {
	panic(typeError("Cannot convert a Symbol value to a number"))
}
func (s *Symbol) ToInt32() int32 {
	panic(typeError("Cannot convert a Symbol value to a number"))
}
func (s *Symbol) ToUInt32() uint32 {
	panic(typeError("Cannot convert a Symbol value to a number"))
}
func (s *Symbol) ToInt64() int64 {
	panic(typeError("Cannot convert a Symbol value to a number"))
}

func (s *Symbol) assertFloat() (float64, bool) {
	return 0, false
}
func (s *Symbol) assertInt() (int, bool) {
	return 0, false
}
func (s *Symbol) assertInt32() (int32, bool) {
	return 0, false
}
func (s *Symbol) assertUInt32() (uint32, bool) {
	return 0, false
}
func (s *Symbol) assertInt64() (int64, bool) {
	return 0, false
}
func (s *Symbol) assertString() (valueString, bool) {
	return nil, false
}

func (s *Symbol) toString() valueString {
	panic(typeError("Cannot convert a Symbol value to a string"))
}

func (s *Symbol) ToString() Value {
	return s
}

func (s *Symbol) String() string {
	if s.desc != nil {
		return s.desc.String()
	}
	return ""
}

func (s *Symbol) string() unistring.String {
	if s.desc != nil {
		return s.desc.string()
	}
	return ""
}

func (s *Symbol) ToFloat() float64 {
	panic(typeError("Cannot convert a Symbol value to a number"))
}

func (s *Symbol) ToNumber() Value {
	panic(typeError("Cannot convert a Symbol value to a number"))
}

func (s *Symbol) ToBoolean() bool {
	return true
}

func (s *Symbol) ToObject(r *Runtime) *Object {
	return s.baseObject(r)
}

func (s *Symbol) SameAs(other Value) bool {
	if s1, ok := other.(*Symbol); ok {
		return s == s1
	}
	return false
}

func (s *Symbol) Equals(o Value) bool {
	switch o := o.(type) {
	case *Object:
		return s.Equals(o.toPrimitive())
	}
	return s.SameAs(o)
}

func (s *Symbol) StrictEquals(o Value) bool {
	return s.SameAs(o)
}

func (s *Symbol) Export() interface{} {
	return s.String()
}

func (s *Symbol) ExportType() reflect.Type {
	return reflectTypeString
}

func (s *Symbol) baseObject(r *Runtime) *Object {
	return r.newPrimitiveObject(s, r.global.SymbolPrototype, "Symbol")
}

func (s *Symbol) hash(*maphash.Hash) uint64 {
	return uint64(s.h)
}

func (s *Symbol) MemUsage(ctx *MemUsageContext) (memUsage uint64, newMemUsage uint64, err error) {
	return s.desc.MemUsage(ctx)
}

func exportValue(v Value, ctx *objectExportCtx) interface{} {
	if obj, ok := v.(*Object); ok {
		return obj.self.export(ctx)
	}
	return v.Export()
}

func newSymbol(s valueString) *Symbol {
	r := &Symbol{
		desc: s,
	}
	// This may need to be reconsidered in the future.
	// Depending on changes in Go's allocation policy and/or introduction of a compacting GC
	// this may no longer provide sufficient dispersion. The alternative, however, is a globally
	// synchronised random generator/hasher/sequencer and I don't want to go down that route just yet.
	r.h = uintptr(unsafe.Pointer(r))
	return r
}

func NewSymbol(s string) *Symbol {
	return newSymbol(newStringValue(s))
}

func (s *Symbol) descriptiveString() valueString {
	desc := s.desc
	if desc == nil {
		desc = stringEmpty
	}
	return asciiString("Symbol(").concat(desc).concat(asciiString(")"))
}

func funcName(prefix string, n Value) valueString {
	var b valueStringBuilder
	b.WriteString(asciiString(prefix))
	if sym, ok := n.(*Symbol); ok {
		if sym.desc != nil {
			b.WriteRune('[')
			b.WriteString(sym.desc)
			b.WriteRune(']')
		}
	} else {
		b.WriteString(n.toString())
	}
	return b.String()
}

func newTypeError(args ...interface{}) typeError {
	msg := ""
	if len(args) > 0 {
		f, _ := args[0].(string)
		msg = fmt.Sprintf(f, args[1:]...)
	}
	return typeError(msg)
}

func typeErrorResult(throw bool, args ...interface{}) {
	if throw {
		panic(newTypeError(args...))
	}

}

func init() {
	for i := 0; i < 256; i++ {
		intCache[i] = valueInt(i - 128)
	}
	for i := 0; i < 256; i++ {
		int64Cache[i] = valueInt64(i - 128)
	}
	_positiveZero = intToValue(0)
}
