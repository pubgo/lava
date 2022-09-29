package version

import "github.com/coreos/go-semver/semver"

func CommitID() string {
	return commitID
}

func Version() string {
	return version
}

func Semver() *semver.Version {
	return semver.New(version)
}

func BuildTime() string {
	return buildTime
}

func Data() string {
	return data
}

func Domain() string {
	return domain
}

func Tag() string {
	return tag
}

func Project() string {
	return project
}

// DeviceID 设备ID
func DeviceID() string {
	return deviceID
}

// InstanceID service instance id
func InstanceID() string {
	return instanceID
}
