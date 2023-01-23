package parsergenerator

type rawType string

type anyType interface{}

type baseType string

type array struct {
	of interface{}
}
type mapType struct {
	from baseType
	to   interface{}
}
