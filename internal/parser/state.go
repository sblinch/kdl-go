package parser

import (
	"strconv"
)

type parserState int

const (
	stateDocument parserState = iota
	stateNode
	stateNodeParams
	stateNodeEnd
	stateArgProp
	stateProperty
	statePropertyValue
	stateChildren
	stateTypeAnnot
	stateTypeDone
)

func (p parserState) String() string {
	switch p {
	case stateDocument:
		return "stateDocument"
	case stateNode:
		return "stateNode"
	case stateNodeParams:
		return "stateNodeParams"
	case stateNodeEnd:
		return "stateNodeEnd"
	case stateArgProp:
		return "stateArgProp"
	case stateProperty:
		return "stateProperty"
	case statePropertyValue:
		return "statePropertyValue"
	case stateChildren:
		return "stateChildren"
	case stateTypeAnnot:
		return "stateTypeAnnot"
	case stateTypeDone:
		return "stateTypeDone"
	default:
		return strconv.FormatInt(int64(p), 10)
	}
}
