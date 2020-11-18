package golug_util

import "testing"

func TestJsonDiff(t *testing.T) {
	JsonDiff(`
{
"a":{
"b":{
"c":1
}
}

}

`, "{}", 2)
}
