package common

// Region represents ECS region
type Region string

// Constants of region definition
const (
	Hangzhou     = Region("cn-hangzhou")
	Qingdao      = Region("cn-qingdao")
	Beijing      = Region("cn-beijing")
	Hongkong     = Region("cn-hongkong")
	Shenzhen     = Region("cn-shenzhen")
	USWest1      = Region("us-west-1")
	USEast1      = Region("us-east-1")
	APSouthEast1 = Region("ap-southeast-1")
	Shanghai     = Region("cn-shanghai")
)

var ValidRegions = []Region{Hangzhou, Qingdao, Beijing, Shenzhen, Hongkong, Shanghai, USWest1, USEast1, APSouthEast1}
