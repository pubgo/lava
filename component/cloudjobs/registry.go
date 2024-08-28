package cloudjobs

import "google.golang.org/protobuf/proto"

var subjects = make(map[string]proto.Message)

func RegisterSubject(subject string, subType proto.Message) any {
	subjects[subject] = subType
	return nil
}
