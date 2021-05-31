package wharfyml

type InputType int

const (
	String = InputType(iota + 1)
	Choice
	Number
	Password
)

var strToInputType = map[string]InputType{
	"string":   String,
	"choice":   Choice,
	"number":   Number,
	"password": Password,
}

var inputTypeToString = map[InputType]string{}

func init() {
	for str, st := range strToInputType {
		inputTypeToString[st] = str
	}
}

func (t InputType) String() string {
	return inputTypeToString[t]
}

func ParseInputType(name string) (InputType, bool) {
	inputType, ok := strToInputType[name]
	return inputType, ok
}